package ginapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wallissonmarinho/GoTV/internal/adapters/http/ginapi"
	"github.com/wallissonmarinho/GoTV/internal/adapters/state"
	"github.com/wallissonmarinho/GoTV/internal/adapters/storage"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
	"github.com/wallissonmarinho/GoTV/internal/core/services"
)

// Admin /api/v1/epg-sources: stubbed httptest + SQLite persistence round-trip.

func TestEndpoint_admin_open_POST_epg_sources_201(t *testing.T) {
	e := newTestEngine(ginapi.Config{AdminAPIKey: ""}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/epg-sources", strings.NewReader(`{"url":"http://epg/guide.xml","label":"E"}`))
	req.Header.Set("Content-Type", "application/json")
	e.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Equal(t, "00000000-0000-4000-8000-000000000002", got["id"])
}

func TestEndpoint_admin_open_GET_epg_sources_200(t *testing.T) {
	e := newTestEngine(ginapi.Config{AdminAPIKey: ""}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/epg-sources", nil))
	require.Equal(t, http.StatusOK, w.Code)
	var list []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Len(t, list, 1)
}

func TestEndpoint_admin_open_DELETE_epg_sources_204(t *testing.T) {
	e := newTestEngine(ginapi.Config{AdminAPIKey: ""}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/epg-sources/00000000-0000-4000-8000-000000000002", nil))
	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestIntegration_POST_epg_sources_persists_then_DELETE_undoes(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "gotv.db")
	cat, err := storage.Open("file:" + dbPath + "?_pragma=busy_timeout(5000)")
	require.NoError(t, err)
	t.Cleanup(func() { _ = cat.Close() })

	admin := services.NewCatalogAdminService(cat, cat, ports.NoopAppLog{})
	e := newTestEngine(ginapi.Config{AdminAPIKey: ""}, depsWithCatalog(admin, &state.MemoryStore{}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/epg-sources",
		strings.NewReader(`{"url":"http://persist.test/epg.xml","label":"epg"}`))
	req.Header.Set("Content-Type", "application/json")
	e.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	id, ok := created["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, id)

	w2 := httptest.NewRecorder()
	e.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/api/v1/epg-sources", nil))
	require.Equal(t, http.StatusOK, w2.Code)
	var list []map[string]any
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &list))
	require.Len(t, list, 1)

	fromDB, err := cat.ListEPGSources(ctx)
	require.NoError(t, err)
	require.Len(t, fromDB, 1)
	require.Equal(t, "http://persist.test/epg.xml", fromDB[0].URL)

	w3 := httptest.NewRecorder()
	e.ServeHTTP(w3, httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/epg-sources/%s", id), nil))
	require.Equal(t, http.StatusNoContent, w3.Code)

	w4 := httptest.NewRecorder()
	e.ServeHTTP(w4, httptest.NewRequest(http.MethodGet, "/api/v1/epg-sources", nil))
	require.Equal(t, http.StatusOK, w4.Code)
	var after []map[string]any
	require.NoError(t, json.Unmarshal(w4.Body.Bytes(), &after))
	require.Empty(t, after)

	fromDB2, err := cat.ListEPGSources(ctx)
	require.NoError(t, err)
	require.Empty(t, fromDB2)
}
