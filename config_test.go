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
