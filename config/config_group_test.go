package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testGroupData struct {
	format string
	config string
}

func TestLoadGroup(t *testing.T) {
	testData := []testGroupData{
		{
			"toml",
			`
[groups]
test = ["first", "second", "third"]
`,
		},
		{
			"json",
			`{ "groups": { "test": ["first", "second", "third"] } }`,
		},
		{
			"yaml",
			`
groups:
  test:
  - first
  - second
  - third
`,
		},
		{
			"hcl",
			`
groups = {
	"test" = ["first", "second", "third"]
}
`,
		},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(testItem.format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			assert.False(t, c.HasGroup("my-group"))
			assert.True(t, c.HasGroup("test"))

			group, err := c.LoadGroup("test")
			require.NoError(t, err)
			assert.Equal(t, []string{"first", "second", "third"}, group)

			_, err = c.LoadGroup("my-group")
			assert.Error(t, err)
		})
	}
}
