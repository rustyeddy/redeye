package redeye

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigReturnsGlobalPointer(t *testing.T) {
	require.Same(t, Config, GetConfig())
}

func TestConfigurationSave(t *testing.T) {
	cfg := &Configuration{
		HTTPAddr: "127.0.0.1:8080",
		HTMLPath: "/tmp/html",
	}

	path := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, cfg.Save(path))

	buf, err := os.ReadFile(path)
	require.NoError(t, err)

	var got Configuration
	require.NoError(t, json.Unmarshal(buf, &got))
	assert.Equal(t, cfg.HTTPAddr, got.HTTPAddr)
	assert.Equal(t, cfg.HTMLPath, got.HTMLPath)
}

func TestConfigurationSaveError(t *testing.T) {
	cfg := &Configuration{}
	err := cfg.Save(filepath.Join(t.TempDir(), "missing", "config.json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save file")
}

func TestConfigurationLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"addr":"127.0.0.1:9000","video-device":5}`), 0o644))

	cfg := &Configuration{}
	require.NoError(t, cfg.Load(path))
	assert.Equal(t, "127.0.0.1:9000", cfg.HTTPAddr)
	assert.Equal(t, 5, cfg.VideoDevice)
}

func TestConfigurationLoadError(t *testing.T) {
	cfg := &Configuration{}
	err := cfg.Load(filepath.Join(t.TempDir(), "missing.json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestConfigurationLoadDefault(t *testing.T) {
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(origWD) })

	tempWD := t.TempDir()
	require.NoError(t, os.Chdir(tempWD))

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	localPath := filepath.Join(tempWD, "redeye.json")
	homePath := filepath.Join(homeDir, ".redeye.json")
	require.NoError(t, os.WriteFile(localPath, []byte(`{"addr":"127.0.0.1:10000"}`), 0o644))
	require.NoError(t, os.WriteFile(homePath, []byte(`{"addr":"127.0.0.1:20000"}`), 0o644))

	cfg := &Configuration{}
	require.NoError(t, cfg.LoadDefault())
	assert.Equal(t, "127.0.0.1:10000", cfg.HTTPAddr)
}

func TestConfigurationLoadDefaultFallsBackToHome(t *testing.T) {
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(origWD) })

	tempWD := t.TempDir()
	require.NoError(t, os.Chdir(tempWD))

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	homePath := filepath.Join(homeDir, ".redeye.json")
	require.NoError(t, os.WriteFile(homePath, []byte(`{"addr":"127.0.0.1:20000"}`), 0o644))

	cfg := &Configuration{}
	require.NoError(t, cfg.LoadDefault())
	assert.Equal(t, "127.0.0.1:20000", cfg.HTTPAddr)
}

func TestConfigurationServeHTTP(t *testing.T) {
	cfg := Configuration{HTTPAddr: "0.0.0.0:8080", Debug: true}
	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	rr := httptest.NewRecorder()

	cfg.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var got Configuration
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
	assert.Equal(t, cfg.HTTPAddr, got.HTTPAddr)
	assert.Equal(t, cfg.Debug, got.Debug)
}
