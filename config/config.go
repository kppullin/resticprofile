package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/creativeprojects/resticprofile/array"
	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config wraps up a viper configuration object
type Config struct {
	keyDelim   string
	format     string
	configFile string
	viper      *viper.Viper
	groups     map[string][]string
}

// This is where things are getting hairy:
//
// Most configuration file formats allow only one declaration per section
// This is not the case for HCL where you can declare a bloc multiple times:
//
// "global" {
//   key1 = "value"
// }
//
// "global" {
//   key2 = "value"
// }
//
// For that matter, viper creates an slice of maps instead of a map for the other configuration file formats
// This configOptionHCL deals with the slice to merge it into a single map
var (
	configOption = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))
	configOptionHCL = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		sliceOfMapsToMapHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))
)

// newConfig instantiate a new Config object
func newConfig(format string) *Config {
	return &Config{
		keyDelim: ".",
		format:   format,
		viper:    viper.New(),
	}
}

// LoadFile loads configuration from file
func LoadFile(configFile string) (*Config, error) {
	format := filepath.Ext(configFile)
	if strings.HasPrefix(format, ".") {
		format = format[1:]
	}
	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("cannot open configuration file for reading: %w", err)
	}
	c := newConfig(format)
	c.configFile = configFile
	err = c.load(file)
	if err != nil {
		return c, err
	}
	return c, nil
}

// Load configuration from reader
func Load(input io.Reader, format string) (*Config, error) {
	c := newConfig(format)
	err := c.load(input)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (c *Config) load(input io.Reader) error {
	// For compatibility with the previous versions, a .conf file is TOML format
	if c.format == "conf" {
		c.format = "toml"
	}
	c.viper.SetConfigType(c.format)
	err := c.viper.ReadConfig(input)
	if err != nil {
		return fmt.Errorf("cannot parse %s configuration: %w", c.format, err)
	}
	return nil
}

// AllKeys returns all keys holding a value, regardless of where they are set.
// Nested keys are separated by a "."
func (c *Config) AllKeys() []string {
	return c.viper.AllKeys()
}

// IsSet checks if the key contains a value
func (c *Config) IsSet(key string) bool {
	if strings.Contains(key, ".") {
		clog.Warningf("it should not search for a subkey: %s", key)
	}
	return c.viper.IsSet(key)
}

// GetConfigFile returns the config file used
func (c *Config) GetConfigFile() string {
	return c.configFile
}

// Get the value from the key
func (c *Config) Get(key string) interface{} {
	return c.viper.Get(key)
}

// HasProfile returns true if the profile exists in the configuration
func (c *Config) HasProfile(profileKey string) bool {
	return c.IsSet(profileKey)
}

// ProfileGroups returns all groups from the configuration
func (c *Config) ProfileGroups() map[string][]string {
	groups := make(map[string][]string, 0)
	if !c.IsSet(constants.SectionConfigurationGroups) {
		return nil
	}
	err := c.unmarshalKey(constants.SectionConfigurationGroups, &groups)
	if err != nil {
		return nil
	}
	return groups
}

// ProfileSections returns a list of profiles with all the sections defined inside each
func (c *Config) ProfileSections() map[string][]string {
	allKeys := c.AllKeys()
	if allKeys == nil || len(allKeys) == 0 {
		return nil
	}
	profileSections := make(map[string][]string, 0)
	for _, keys := range allKeys {
		keyPath := strings.SplitN(keys, ".", 3)
		if len(keyPath) > 0 {
			if keyPath[0] == constants.SectionConfigurationGlobal || keyPath[0] == constants.SectionConfigurationGroups {
				continue
			}
			var commands []string
			var found bool
			if commands, found = profileSections[keyPath[0]]; !found {
				commands = make([]string, 0)
			} else {
				commands = profileSections[keyPath[0]]
			}
			// If there are more than two keys, it means the second key is a group of keys, so it's a "command" definition
			if len(keyPath) > 2 {
				if _, found = array.FindString(commands, keyPath[1]); !found {
					commands = append(commands, keyPath[1])
				}
			}
			profileSections[keyPath[0]] = commands
		}
	}
	return profileSections
}

// GetGlobalSection returns the global configuration
func (c *Config) GetGlobalSection() (*Global, error) {
	global := newGlobal()
	err := c.unmarshalKey(constants.SectionConfigurationGlobal, global)
	if err != nil {
		return nil, err
	}
	return global, nil
}

// HasGroup returns true if the group of profiles exists in the configuration
func (c *Config) HasGroup(groupKey string) bool {
	if !c.IsSet(constants.SectionConfigurationGroups) {
		return false
	}
	err := c.loadGroups()
	if err != nil {
		return false
	}
	_, ok := c.groups[groupKey]
	return ok
}

// LoadGroup returns the list of profiles in a group
func (c *Config) LoadGroup(groupKey string) ([]string, error) {
	err := c.loadGroups()
	if err != nil {
		return nil, err
	}

	group, ok := c.groups[groupKey]
	if !ok {
		return nil, fmt.Errorf("group '%s' not found", groupKey)
	}
	return group, nil
}

func (c *Config) loadGroups() error {
	if c.groups == nil {
		groups := map[string][]string{}
		err := c.unmarshalKey(constants.SectionConfigurationGroups, &groups)
		if err != nil {
			return err
		}
		c.groups = groups
	}
	return nil
}

// LoadProfile from configuration
func (c *Config) LoadProfile(profileKey string) (*Profile, error) {
	var err error
	var profile *Profile

	if !c.IsSet(profileKey) {
		return nil, nil
	}

	profile = NewProfile(c, profileKey)
	err = c.unmarshalKey(profileKey, profile)
	if err != nil {
		return nil, err
	}
	if profile.Inherit != "" {
		inherit := profile.Inherit
		// Load inherited profile
		profile, err = c.LoadProfile(inherit)
		if err != nil {
			return nil, err
		}
		if profile == nil {
			return nil, fmt.Errorf("error in profile '%s': parent profile '%s' not found", profileKey, inherit)
		}
		// and reload this profile onto the inherited one
		err = c.unmarshalKey(profileKey, profile)
		if err != nil {
			return nil, err
		}
		// make sure it has the right name
		profile.Name = profileKey
	}
	return profile, nil
}

// unmarshalKey is a wrapper around viper.UnmarshalKey with the right decoder config options
func (c *Config) unmarshalKey(key string, rawVal interface{}) error {
	if c.format == "hcl" {
		return c.viper.UnmarshalKey(key, rawVal, configOptionHCL)
	}
	return c.viper.UnmarshalKey(key, rawVal, configOption)
}

// sliceOfMapsToMapHookFunc merges a slice of maps to a map
func sliceOfMapsToMapHookFunc() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() == reflect.Slice && from.Elem().Kind() == reflect.Map && (to.Kind() == reflect.Struct || to.Kind() == reflect.Map) {
			clog.Debugf("hook: from slice %+v to %+v", from.Elem(), to)
			source, ok := data.([]map[string]interface{})
			if !ok {
				return data, nil
			}
			if len(source) == 0 {
				return data, nil
			}
			if len(source) == 1 {
				return source[0], nil
			}
			// flatten the slice into one map
			convert := make(map[string]interface{})
			for _, mapItem := range source {
				for key, value := range mapItem {
					convert[key] = value
				}
			}
			return convert, nil
		}
		clog.Debugf("default from %+v to %+v", from, to)
		return data, nil
	}
}
