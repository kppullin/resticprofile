package main

import (
	"errors"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Change default commands for testing ones
	ownCommands = []ownCommand{
		{
			name:              "first",
			description:       "first first",
			action:            firstCommand,
			needConfiguration: false,
		},
		{
			name:              "second",
			description:       "second second",
			action:            secondCommand,
			needConfiguration: true,
		},
		{
			name:              "third",
			description:       "third third",
			action:            thirdCommand,
			needConfiguration: false,
			hide:              true,
		},
	}
}

func firstCommand(_ *config.Config, _ commandLineFlags, _ []string) error {
	return errors.New("first")
}

func secondCommand(_ *config.Config, _ commandLineFlags, _ []string) error {
	return errors.New("second")
}

func thirdCommand(_ *config.Config, _ commandLineFlags, _ []string) error {
	return errors.New("third")
}

func ExampleDisplayOwnCommands() {
	displayOwnCommands()
	// Output:
	//    first    first first
	//    second   second second
}

func TestIsOwnCommand(t *testing.T) {
	assert.True(t, isOwnCommand("first", false))
	assert.True(t, isOwnCommand("second", true))
	assert.True(t, isOwnCommand("third", false))
	assert.False(t, isOwnCommand("another one", true))
}

func TestRunOwnCommand(t *testing.T) {
	assert.EqualError(t, runOwnCommand(nil, "first", commandLineFlags{}, nil), "first")
	assert.EqualError(t, runOwnCommand(nil, "second", commandLineFlags{}, nil), "second")
	assert.EqualError(t, runOwnCommand(nil, "third", commandLineFlags{}, nil), "third")
	assert.EqualError(t, runOwnCommand(nil, "another one", commandLineFlags{}, nil), "command not found: another one")
}
