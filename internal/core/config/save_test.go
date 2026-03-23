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

func TestSave_TargetSubnets_RoundTrip_InlineArray(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := DefaultConfig()
	cfg.TargetSubnets = []string{"10.0.0.0/24", "192.168.1.0/24"}

	err := Save(cfg, configPath)
	require.NoError(t, err)

	loaded, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.0/24", "192.168.1.0/24"}, loaded.TargetSubnets)
}

func TestSave_TargetSubnets_RoundTrip_DashArraySyntax(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
target_subnets:
  - "10.0.0.0/24"
  - "192.168.1.0/24"
scanners:
  mdns:
    enabled: true
  ssdp:
    enabled: true
  arp:
    enabled: true
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	loaded, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.0/24", "192.168.1.0/24"}, loaded.TargetSubnets)
}

func TestSave_TargetSubnets_PreservedAfterThemeChange(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := DefaultConfig()
	cfg.TargetSubnets = []string{"10.0.0.0/24", "172.16.0.0/16"}

	err := Save(cfg, configPath)
	require.NoError(t, err)

	loaded, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.0/24", "172.16.0.0/16"}, loaded.TargetSubnets)

	loaded.Theme.Name = "dracula"
	err = Save(loaded, configPath)
	require.NoError(t, err)

	reloaded, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)
	assert.Equal(t, "dracula", reloaded.Theme.Name)
	assert.Equal(t, []string{"10.0.0.0/24", "172.16.0.0/16"}, reloaded.TargetSubnets)
}

func TestSave_TargetSubnets_DashSyntaxSurvivedAfterThemeChange(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
target_subnets:
  - "10.0.0.0/24"
  - "192.168.1.0/24"
scanners:
  mdns:
    enabled: true
  ssdp:
    enabled: true
  arp:
    enabled: true
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	loaded, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.0/24", "192.168.1.0/24"}, loaded.TargetSubnets)

	loaded.Theme.Name = "monokai"
	err = Save(loaded, configPath)
	require.NoError(t, err)

	reloaded, err := LoadForMode(ModeApp, &Flags{ConfigFile: configPath})
	require.NoError(t, err)
	assert.Equal(t, "monokai", reloaded.Theme.Name)
	assert.Equal(t, []string{"10.0.0.0/24", "192.168.1.0/24"}, reloaded.TargetSubnets)
}

func TestSave_TargetSubnets_EmptyNotWrittenActive(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := DefaultConfig()
	err := Save(cfg, configPath)
	require.NoError(t, err)

	raw, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(raw), "# target_subnets:")
	assert.NotContains(t, string(raw), "\ntarget_subnets:")
}

func TestSave_TargetSubnets_SetValueWrittenUncommented(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := DefaultConfig()
	cfg.TargetSubnets = []string{"10.0.0.0/24"}

	err := Save(cfg, configPath)
	require.NoError(t, err)

	raw, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `target_subnets: ["10.0.0.0/24"]`)
	assert.NotContains(t, string(raw), "# target_subnets:")
}
