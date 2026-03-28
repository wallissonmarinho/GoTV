package ginapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wallissonmarinho/GoTV/internal/adapters/http/ginapi"
	"github.com/wallissonmarinho/GoTV/internal/adapters/state"
)

// Admin /api/v1/rebuild and /api/v1/merge-status.

func TestEndpoint_admin_open_POST_rebuild_202(t *testing.T) {
	mergeDone := make(chan struct{}, 1)
	d := testDeps(&state.MemoryStore{})
	d.ManualMergeDone = mergeDone
	e := newTestEngine(ginapi.Config{AdminAPIKey: ""}, d)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/rebuild", nil))
	require.Equal(t, http.StatusAccepted, w.Code)
	<-mergeDone // after merge + slog in the rebuild goroutine; avoids log bleed into later tests
}

func TestEndpoint_admin_open_GET_merge_status_200(t *testing.T) {
	e := newTestEngine(ginapi.Config{AdminAPIKey: ""}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/merge-status", nil))
	require.Equal(t, http.StatusOK, w.Code)
	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Equal(t, true, got["ok"])
}
