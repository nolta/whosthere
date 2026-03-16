package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSave_PersistsThemeChanges(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Theme.Name = "monokai"

	err := Save(cfg, configPath)
	require.NoError(t, err)

	loaded, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)

	assert.Equal(t, "monokai", loaded.Theme.Name)
}

func TestSave_PersistsMultipleChanges(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Theme.Name = "dracula"
	cfg.Theme.Enabled = false
	cfg.Splash.Enabled = false

	err := Save(cfg, configPath)
	require.NoError(t, err)

	loaded, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)

	assert.Equal(t, "dracula", loaded.Theme.Name)
	assert.False(t, loaded.Theme.Enabled)
	assert.False(t, loaded.Splash.Enabled)
}

func TestSave_OverwritesExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	initialCfg := DefaultConfig()
	initialCfg.Theme.Name = "initial"

	err := Save(initialCfg, configPath)
	require.NoError(t, err)

	updatedCfg := DefaultConfig()
	updatedCfg.Theme.Name = "updated"

	err = Save(updatedCfg, configPath)
	require.NoError(t, err)

	loaded, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)

	assert.Equal(t, "updated", loaded.Theme.Name)
}

func TestSave_NilConfigReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := Save(nil, configPath)
	assert.Error(t, err)
	assert.Equal(t, ErrConfigNil, err)
}

func TestSave_CreatesDirectoryIfNeeded(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "nested", "config.yaml")

	cfg := DefaultConfig()
	cfg.Theme.Name = "test"

	err := Save(cfg, configPath)
	require.NoError(t, err)

	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config file should exist")

	loaded, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)

	assert.Equal(t, "test", loaded.Theme.Name)
}

func TestWriteConfigFile_WithNilConfigWritesDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := DefaultConfig()
	err := Save(cfg, configPath)
	require.NoError(t, err)

	raw, err := os.ReadFile(configPath)
	require.NoError(t, err)

	defaultYAML := GenerateDefaultYAML()
	assert.Equal(t, defaultYAML, string(raw))
}

func TestSave_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	original := DefaultConfig()
	original.Theme.Name = "nord"
	original.Theme.Enabled = true
	original.Splash.Enabled = false
	original.Scanners.MDNS.Enabled = false
	original.Scanners.SSDP.Enabled = true

	err := Save(original, configPath)
	require.NoError(t, err)

	loaded, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)

	assert.Equal(t, "nord", loaded.Theme.Name)
	assert.True(t, loaded.Theme.Enabled)
	assert.False(t, loaded.Splash.Enabled)
	assert.False(t, loaded.Scanners.MDNS.Enabled)
	assert.True(t, loaded.Scanners.SSDP.Enabled)
}

func TestThemePersistence_ChangeAndReload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := DefaultConfig()
	require.Equal(t, DefaultThemeName, cfg.Theme.Name)

	err := Save(cfg, configPath)
	require.NoError(t, err)

	cfg.Theme.Name = "dracula"
	err = Save(cfg, configPath)
	require.NoError(t, err)

	reloaded, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)
	assert.Equal(t, "dracula", reloaded.Theme.Name)

	reloaded.Theme.Name = "monokai"
	err = Save(reloaded, configPath)
	require.NoError(t, err)

	reloaded2, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)
	assert.Equal(t, "monokai", reloaded2.Theme.Name)
}
