package ginapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wallissonmarinho/GoTV/internal/adapters/http/ginapi"
	"github.com/wallissonmarinho/GoTV/internal/adapters/state"
)

// Admin API key middleware on /api/v1/*.

func TestEndpoint_admin_key_set_unauthorized_without_header(t *testing.T) {
	e := newTestEngine(ginapi.Config{AdminAPIKey: "secret"}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/m3u-sources", nil))
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEndpoint_admin_bearer_authorized(t *testing.T) {
	e := newTestEngine(ginapi.Config{AdminAPIKey: "secret"}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/m3u-sources", nil)
	req.Header.Set("Authorization", "Bearer secret")
	e.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}
