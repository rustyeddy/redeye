package redeye

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigHandlerGet(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAPIRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/camera/config", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	var body map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
}

func TestConfigHandlerPut(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAPIRoutes(mux)

	body := strings.NewReader(`{"pipeline":"resize"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/camera/config", body)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var cfg Configuration
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&cfg))
	assert.Equal(t, "resize", cfg.Pipeline)
}

func TestConfigHandlerRejectsUnsupportedMethod(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAPIRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/camera/config", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Equal(t, "GET, PUT", rr.Header().Get("Allow"))
}

func TestHealthEndpointReturnsOK(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAPIRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	assert.Equal(t, "ok", body["status"])
}

func TestAPIRoutesRejectUnsupportedMethods(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAPIRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/camera/snap", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Equal(t, http.MethodPost, rr.Header().Get("Allow"))
}

// --- snap handler ---

// mockSnapper records the last filename passed to Snap and returns a preset error.
type mockSnapper struct {
	calledWith string
	returnErr  error
}

func (m *mockSnapper) Snap(file string) error {
	m.calledWith = file
	return m.returnErr
}

func snapMux(t *testing.T, snapper Snapper) *http.ServeMux {
	t.Helper()
	// Each test gets a fresh mux and snapper so tests don't share global state.
	mux := http.NewServeMux()
	prev := activeSnapper.Load()
	SetSnapper(snapper)
	t.Cleanup(func() { activeSnapper.Store(prev) })
	RegisterAPIRoutes(mux)
	return mux
}

func TestSnapHandlerNoSource(t *testing.T) {
	mux := snapMux(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/camera/snap", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusServiceUnavailable, rr.Code)
	var body map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	assert.Contains(t, body["error"], "no active camera source")
}

func TestSnapHandlerRejectsNonPost(t *testing.T) {
	mux := snapMux(t, &mockSnapper{})

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		req := httptest.NewRequest(method, "/api/camera/snap", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "method %s", method)
		assert.Equal(t, http.MethodPost, rr.Header().Get("Allow"))
	}
}

func TestSnapHandlerUsesQueryParamFilename(t *testing.T) {
	ms := &mockSnapper{}
	mux := snapMux(t, ms)

	req := httptest.NewRequest(http.MethodPost, "/api/camera/snap?file=myshot.jpg", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "myshot.jpg", ms.calledWith)

	var body map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	assert.Equal(t, "myshot.jpg", body["file"])
}

func TestSnapHandlerGeneratesDefaultFilename(t *testing.T) {
	ms := &mockSnapper{}
	mux := snapMux(t, ms)

	req := httptest.NewRequest(http.MethodPost, "/api/camera/snap", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, ms.calledWith, "snapshot-")
	assert.Contains(t, ms.calledWith, ".jpg")

	var body map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	assert.Equal(t, ms.calledWith, body["file"])
}

func TestSnapHandlerPropagatesSnapError(t *testing.T) {
	ms := &mockSnapper{returnErr: fmt.Errorf("disk full")}
	mux := snapMux(t, ms)

	req := httptest.NewRequest(http.MethodPost, "/api/camera/snap?file=x.jpg", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	var body map[string]string
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&body))
	assert.Contains(t, body["error"], "disk full")
}
