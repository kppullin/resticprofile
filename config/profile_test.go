package config

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoProfile(t *testing.T) {
	testConfig := ""
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, profile)
}

func TestEmptyProfile(t *testing.T) {
	testConfig := `[profile]
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, "profile", profile.Name)
}

func TestNoInitializeValue(t *testing.T) {
	testConfig := `[profile]
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, false, profile.Initialize)
}

func TestInitializeValueFalse(t *testing.T) {
	testConfig := `[profile]
initialize = false
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, false, profile.Initialize)
}

func TestInitializeValueTrue(t *testing.T) {
	testConfig := `[profile]
initialize = true
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, true, profile.Initialize)
}

func TestInheritedInitializeValueTrue(t *testing.T) {
	testConfig := `[parent]
initialize = true

[profile]
inherit = "parent"
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, true, profile.Initialize)
}

func TestOverriddenInitializeValueFalse(t *testing.T) {
	testConfig := `[parent]
initialize = true

[profile]
initialize = false
inherit = "parent"
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, false, profile.Initialize)
}

func TestUnknownParent(t *testing.T) {
	testConfig := `[profile]
inherit = "parent"
`
	_, err := getProfile(testConfig, "profile")
	assert.Error(t, err)
}

func TestMultiInheritance(t *testing.T) {
	testConfig := `
[grand-parent]
repository = "grand-parent"
first-value = 1
override-value = 1

[parent]
inherit = "grand-parent"
initialize = true
repository = "parent"
second-value = 2
override-value = 2
quiet = true

[profile]
inherit = "parent"
third-value = 3
verbose = true
quiet = false
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, "profile", profile.Name)
	assert.Equal(t, "parent", profile.Repository)
	assert.Equal(t, true, profile.Initialize)
	assert.Equal(t, int64(1), profile.OtherFlags["first-value"])
	assert.Equal(t, int64(2), profile.OtherFlags["second-value"])
	assert.Equal(t, int64(3), profile.OtherFlags["third-value"])
	assert.Equal(t, int64(2), profile.OtherFlags["override-value"])
	assert.Equal(t, false, profile.Quiet)
	assert.Equal(t, true, profile.Verbose)
}

func TestProfileCommonFlags(t *testing.T) {
	assert := assert.New(t)
	testConfig := `
[profile]
quiet = true
verbose = false
repository = "test"
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(profile)

	flags := profile.GetCommonFlags()
	assert.NotNil(flags)
	assert.Contains(flags, "quiet")
	assert.NotContains(flags, "verbose")
	assert.Contains(flags, "repo")
}

func TestProfileOtherFlags(t *testing.T) {
	assert := assert.New(t)
	testConfig := `
[profile]
bool-true = true
bool-false = false
string = "test"
zero = 0
empty = ""
float = 4.2
int = 42
# comment
array0 = []
array1 = [1]
array2 = ["one", "two"]
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(profile)

	flags := profile.GetCommonFlags()
	assert.NotNil(flags)
	assert.Contains(flags, "bool-true")
	assert.NotContains(flags, "bool-false")
	assert.Contains(flags, "string")
	assert.NotContains(flags, "zero")
	assert.NotContains(flags, "empty")
	assert.Contains(flags, "float")
	assert.Contains(flags, "int")
	assert.NotContains(flags, "array0")
	assert.Contains(flags, "array1")
	assert.Contains(flags, "array2")

	assert.Equal([]string{}, flags["bool-true"])
	assert.Equal([]string{"test"}, flags["string"])
	assert.Equal([]string{fmt.Sprintf("%f", 4.2)}, flags["float"])
	assert.Equal([]string{"42"}, flags["int"])
	assert.Equal([]string{"1"}, flags["array1"])
	assert.Equal([]string{"one", "two"}, flags["array2"])
}

func TestHostInProfile(t *testing.T) {
	assert := assert.New(t)
	testConfig := `
[profile]
initialize = true
[profile.backup]
host = true
[profile.snapshots]
host = "ConfigHost"
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(profile)

	profile.SetHost("TestHost")

	flags := profile.GetCommandFlags(constants.CommandBackup)
	assert.NotNil(flags)
	assert.Contains(flags, "host")
	assert.Equal([]string{"TestHost"}, flags["host"])

	flags = profile.GetCommandFlags(constants.CommandSnapshots)
	assert.NotNil(flags)
	assert.Contains(flags, "host")
	assert.Equal([]string{"ConfigHost"}, flags["host"])
}

func TestKeepPathInRetention(t *testing.T) {
	assert := assert.New(t)
	root, err := filepath.Abs("/")
	require.NoError(t, err)
	root = filepath.ToSlash(root)
	testConfig := `
[profile]
initialize = true

[profile.backup]
source = "` + root + `"

[profile.retention]
host = false
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(profile)

	flags := profile.GetRetentionFlags()
	assert.NotNil(flags)
	assert.Contains(flags, "path")
	assert.Equal([]string{root}, flags["path"])
}

func TestReplacePathInRetention(t *testing.T) {
	assert := assert.New(t)
	testConfig := `
[profile]
initialize = true

[profile.backup]
source = "/some_other_path"

[profile.retention]
path = "/"
`
	profile, err := getProfile(testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(profile)

	flags := profile.GetRetentionFlags()
	assert.NotNil(flags)
	assert.Contains(flags, "path")
	assert.Equal([]string{"/"}, flags["path"])
}

func getProfile(configString, profileKey string) (*Profile, error) {
	c, err := Load(bytes.NewBufferString(configString), "toml")
	if err != nil {
		return nil, err
	}

	profile, err := c.GetProfile(profileKey)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func TestForgetCommandFlags(t *testing.T) {
	testData := []testTemplate{
		{"toml", `
[profile]
initialize = true

[profile.backup]
source = "/"

[profile.forget]
keep-daily = 1
`},
		{"json", `
{
  "profile": {
    "backup": {"source": "/"},
    "forget": {"keep-daily": 1}
  }
}`},
		{"yaml", `---
profile:
  backup:
    source: "/"
  forget:
    keep-daily: 1
`},
		{"hcl", `
"profile" = {
	backup = {
		source = "/"
	}
	forget = {
		keep-daily = 1
	}
}
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			profile := getConfigProfile(t, format, testConfig, "profile")

			assert.NotNil(t, profile)
			assert.NotNil(t, profile.Forget)
			assert.NotEmpty(t, profile.Forget["keep-daily"])
		})
	}
}

func getConfigProfile(t *testing.T, configFormat, configString, profileKey string) *Profile {
	c, err := Load(bytes.NewBufferString(configString), configFormat)
	require.NoError(t, err)

	profile, err := c.GetProfile(profileKey)
	require.NoError(t, err)
	return profile
}
