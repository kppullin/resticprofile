package main

import (
	"log"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/remote"
)

func setupRemoteLogger(flags commandLineFlags) {
	client := remote.NewClient(flags.parentPort)
	logger := clog.NewLogger(client)
	client.SetLogPrefix("elevated user: ")
	clog.SetDefaultLogger(logger)
}

func setupFileLogger(flags commandLineFlags) (*clog.FileHandler, error) {
	fileHandler, err := clog.NewFileHandler(flags.logFile, "", log.LstdFlags)
	if err != nil {
		return nil, err
	}
	logger := newFilteredLogger(flags, fileHandler)
	// default logger added with level filtering
	clog.SetDefaultLogger(logger)
	// but return fileHandler (so we can close it at the end)
	return fileHandler, nil
}

func setupConsoleLogger(flags commandLineFlags) {
	consoleHandler := clog.NewConsoleHandler("", log.LstdFlags)
	if flags.theme != "" {
		consoleHandler.SetTheme(flags.theme)
	}
	if flags.noAnsi {
		consoleHandler.Colouring(false)
	}
	logger := newFilteredLogger(flags, consoleHandler)
	clog.SetDefaultLogger(logger)
}

func newFilteredLogger(flags commandLineFlags, handler clog.Handler) *clog.Logger {
	if flags.quiet && flags.verbose {
		coin := ""
		if randomBool() {
			coin = "verbose"
			flags.quiet = false
		} else {
			coin = "quiet"
			flags.verbose = false
		}
		// the logger hasn't been created yet, so we call the handler directly
		handler.LogEntry(clog.LogEntry{
			Level:  clog.LevelWarning,
			Format: "you specified -quiet (-q) and -verbose (-v) at the same time. So let's flip a coin! and selection is ... %s.",
			Values: []interface{}{coin},
		})
	}
	minLevel := clog.LevelInfo
	if flags.quiet {
		minLevel = clog.LevelWarning
	} else if flags.verbose {
		minLevel = clog.LevelDebug
	}
	// now create and return the logger
	return clog.NewLogger(clog.NewLevelFilter(minLevel, handler))
}
