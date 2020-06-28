package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"text/tabwriter"
	"time"

	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/filesearch"
	"github.com/creativeprojects/resticprofile/priority"
)

// These fields are populated by the goreleaser build
var (
	version = "0.7.2-dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func init() {
	rand.Seed(time.Now().UnixNano() - time.Now().Unix())
}

func main() {
	var err error
	defer showPanicData()

	flagset, flags := loadFlags()
	if flags.help {
		flagset.Usage()
		return
	}
	setLoggerFlags(flags)
	banner()

	// Deprecated in version 0.7.0
	// Keep for compatibility with version 0.6.1
	if flags.selfUpdate {
		err = confirmAndSelfUpdate(flags.verbose)
		if err != nil {
			clog.Error(err)
			os.Exit(1)
		}
		return
	}

	// resticprofile own commands (configuration file NOT loaded)
	if len(flags.resticArgs) > 0 {
		if isOwnCommand(flags.resticArgs[0], false) {
			err = runOwnCommand(nil, flags.resticArgs[0], flags, flags.resticArgs[1:])
			if err != nil {
				clog.Error(err)
				os.Exit(1)
			}
			return
		}
	}

	configFile, err := filesearch.FindConfigurationFile(flags.config)
	if err != nil {
		clog.Error(err)
		os.Exit(1)
	}

	c, err := config.LoadFile(configFile)
	if err != nil {
		clog.Error("cannot load configuration file:", err)
		os.Exit(1)
	}

	global, err := c.GetGlobalSection()
	if err != nil {
		clog.Error("cannot load global configuration:", err)
		os.Exit(1)
	}

	err = setPriority(global.Nice, global.Priority)
	if err != nil {
		clog.Warning(err)
	}

	if global.IONice {
		err = priority.SetIONice(global.IONiceClass, global.IONiceLevel)
		if err != nil {
			clog.Warning(err)
		}
	}

	resticBinary, err := filesearch.FindResticBinary(global.ResticBinary)
	if err != nil {
		clog.Error("cannot find restic:", err)
		clog.Warning("you can specify the path of the restic binary in the global section of the configuration file (restic-binary)")
		os.Exit(1)
	}

	// The remaining arguments are going to be sent to the restic command line
	resticArguments := flags.resticArgs
	resticCommand := global.DefaultCommand
	if len(resticArguments) > 0 {
		resticCommand = resticArguments[0]
		resticArguments = resticArguments[1:]
	}

	// resticprofile own commands (with configuration file)
	if isOwnCommand(resticCommand, true) {
		err = runOwnCommand(c, resticCommand, flags, resticArguments)
		if err != nil {
			clog.Error(err)
			os.Exit(1)
		}
		return
	}

	if c.HasProfile(flags.name) {
		// Single profile run
		runProfile(c, global, flags, flags.name, resticBinary, resticArguments, resticCommand)

	} else if c.HasGroup(flags.name) {
		// Group run
		group, err := c.LoadGroup(flags.name)
		if err != nil {
			clog.Errorf("cannot load group '%s': %v", flags.name, err)
		}
		if group != nil && len(group) > 0 {
			for i, profileName := range group {
				clog.Debugf("[%d/%d] Starting profile '%s' from group '%s'", i+1, len(group), profileName, flags.name)
				runProfile(c, global, flags, profileName, resticBinary, resticArguments, resticCommand)
			}
		}

	} else {
		clog.Errorf("profile or group not found '%s'", flags.name)
		displayProfiles(c)
		displayGroups(c)
		os.Exit(1)
	}
}

func setLoggerFlags(flags commandLineFlags) {
	if flags.theme != "" {
		clog.SetTheme(flags.theme)
	}
	if flags.noAnsi {
		clog.Colorize(false)
	}

	if flags.quiet && flags.verbose {
		coin := ""
		if randomBool() {
			coin = "verbose"
			flags.quiet = false
		} else {
			coin = "quiet"
			flags.verbose = false
		}
		clog.Warningf("you specified -quiet (-q) and -verbose (-v) at the same time. So let's flip a coin! and selection is ... %s.", coin)
	}
	if flags.quiet {
		clog.Quiet()
	}
	if flags.verbose {
		clog.Verbose()
	}
}

func banner() {
	clog.Infof("resticprofile %s compiled with %s", version, runtime.Version())
}

func setPriority(nice int, class string) error {
	var err error

	if class != "" {
		if classID, ok := constants.PriorityValues[class]; ok {
			err = priority.SetClass(classID)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("incorrect value '%s' for priority in global section", class)
		}
		return nil
	}
	if nice != 0 {
		err = priority.SetNice(nice)
		if err != nil {
			return err
		}
	}
	return nil
}

func runProfile(c *config.Config, global *config.Global, flags commandLineFlags, profileName string, resticBinary string, resticArguments []string, resticCommand string) {
	var err error

	profile, err := c.LoadProfile(profileName)
	if err != nil {
		clog.Warning(err)
	}
	if profile == nil {
		clog.Errorf("cannot load profile '%s'", profileName)
		os.Exit(1)
	}

	// Send the quiet/verbose down to restic as well (override profile configuration)
	if flags.quiet {
		profile.Quiet = true
		profile.Verbose = false
	}
	if flags.verbose {
		profile.Verbose = true
		profile.Quiet = false
	}

	// All files in the configuration are relative to the configuration file, NOT the folder where resticprofile is started
	// So we need to fix all relative files
	rootPath := filepath.Dir(c.GetConfigFile())
	if rootPath != "." {
		clog.Debugf("files in configuration are relative to '%s'", rootPath)
	}
	profile.SetRootPath(rootPath)

	// Specific case for the "host" flag where an empty value should be replaced by the hostname
	hostname := "none"
	currentHost, err := os.Hostname()
	if err == nil {
		hostname = currentHost
	}
	profile.SetHost(hostname)

	// Catch CTR-C keypress
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	wrapper := newResticWrapper(
		resticBinary,
		global.Initialize || profile.Initialize,
		profile,
		resticCommand,
		resticArguments,
		sigChan,
	)
	err = wrapper.runProfile()
	if err != nil {
		clog.Error(err)
		os.Exit(1)
	}
}

// randomBool returns true for Heads and false for Tails
func randomBool() bool {
	return rand.Int31n(10000) < 5000
}

func showPanicData() {
	if r := recover(); r != nil {
		message := `
===============================================================
uh-oh! resticprofile crashed miserably :-(
Please can you open an issue on github including these details:
===============================================================
`
		fmt.Fprintf(os.Stderr, message)
		w := tabwriter.NewWriter(os.Stderr, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "os", runtime.GOOS)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "arch", runtime.GOARCH)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "version", version)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "commit", commit)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "compiled", date)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "by", builtBy)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "error", r)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "stack", string(debug.Stack()))
		w.Flush()
		fmt.Fprint(os.Stderr, "===============================================================\n")
	}
}
