package redeye

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIRoutesReturnNotImplementedForSupportedMethods(t *testing.T) {
	mux := http.NewServeMux()
	registerAPIRoutes(mux)

	for path, methods := range apiRouteMethods {
		for _, method := range methods {
			req := httptest.NewRequest(method, path, nil)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			require.Equal(t, http.StatusNotImplemented, rr.Code, "%s %s", method, path)
			assert.Contains(t, rr.Body.String(), "not implemented")
			assert.Contains(t, rr.Body.String(), path)
		}
	}
}

func TestAPIRoutesRejectUnsupportedMethods(t *testing.T) {
	mux := http.NewServeMux()
	registerAPIRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/camera/play", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Equal(t, http.MethodPost, rr.Header().Get("Allow"))
}
