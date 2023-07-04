package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var rightConfig = `
{
	"port": ":3000",
	"version": "v1",
	"db_file": "pkid.db"
}
	`

func TestReadConfFile(t *testing.T) {
	t.Run("read config file ", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(rightConfig), 0644)
		assert.NoError(t, err)

		data, err := ReadConfFile(configPath)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

	})

	t.Run("change permissions of file", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(rightConfig), fs.FileMode(os.O_RDONLY))
		assert.NoError(t, err)

		data, err := ReadConfFile(configPath)
		assert.Error(t, err)
		assert.Empty(t, data)

	})

	t.Run("no file exists", func(t *testing.T) {
		data, err := ReadConfFile("./testing.json")
		assert.Error(t, err)
		assert.Empty(t, data)

	})

}

func TestParseConf(t *testing.T) {

	t.Run("can't unmarshal", func(t *testing.T) {
		config := `{testing}`

		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err)

	})

	t.Run("parse config file", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(rightConfig), 0644)
		assert.NoError(t, err)

		got, err := ReadConfFile(configPath)
		assert.NoError(t, err)

		expected := Configuration{
			Port: ":3000",
		}

		assert.NoError(t, err)
		assert.Equal(t, got.Port, expected.Port)
	})

	t.Run("no file", func(t *testing.T) {
		_, err := ReadConfFile("config.json")
		assert.Error(t, err)

	})

	t.Run("no port configuration", func(t *testing.T) {
		config :=
			`
{
	"port": ""
}
	`

		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err, "port configuration is required")
	})

	t.Run("no version configuration", func(t *testing.T) {
		config :=
			`
{
	"port": ":3000",
	"version": ""
}
	`

		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err, "version is required")

	})

	t.Run("no db configuration", func(t *testing.T) {
		config :=
			`
{
	"port": ":3000",
	"version": "v1",
	"db_file": ""
}
	`

		dir := t.TempDir()
		configPath := filepath.Join(dir, "/config.json")

		err := os.WriteFile(configPath, []byte(config), 0644)
		assert.NoError(t, err)

		_, err = ReadConfFile(configPath)
		assert.Error(t, err, "db file is required")
	})
}
