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

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/wallissonmarinho/GoTV/internal/adapters/http/ginapi"
	"github.com/wallissonmarinho/GoTV/internal/adapters/state"
	"github.com/wallissonmarinho/GoTV/internal/adapters/storage"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
	"github.com/wallissonmarinho/GoTV/internal/core/services"
)

// Admin /api/v1/m3u-sources: stubbed httptest + SQLite persistence round-trip.

func TestEndpoint_admin_open_POST_m3u_sources_201(t *testing.T) {
	e := newTestEngine(ginapi.Config{AdminAPIKey: ""}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/m3u-sources", strings.NewReader(`{"url":"http://x/playlist.m3u","label":"L"}`))
	req.Header.Set("Content-Type", "application/json")
	e.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var got map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	idStr, ok := got["id"].(string)
	require.True(t, ok)
	_, err := uuid.Parse(idStr)
	require.NoError(t, err)
	require.Equal(t, "http://x/playlist.m3u", got["url"])
}

func TestEndpoint_admin_open_GET_m3u_sources_200(t *testing.T) {
	e := newTestEngine(ginapi.Config{AdminAPIKey: ""}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/m3u-sources", nil))
	require.Equal(t, http.StatusOK, w.Code)
	var list []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Len(t, list, 1)
	require.Equal(t, "00000000-0000-4000-8000-00000000000a", list[0]["id"])
}

func TestEndpoint_admin_open_DELETE_m3u_sources_204(t *testing.T) {
	e := newTestEngine(ginapi.Config{AdminAPIKey: ""}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/v1/m3u-sources/00000000-0000-4000-8000-000000000001", nil))
	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestIntegration_POST_m3u_sources_persists_then_DELETE_undoes(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "gotv.db")
	cat, err := storage.Open("file:" + dbPath + "?_pragma=busy_timeout(5000)")
	require.NoError(t, err)
	t.Cleanup(func() { _ = cat.Close() })

	admin := services.NewCatalogAdminService(cat, cat, ports.NoopAppLog{})
	e := newTestEngine(ginapi.Config{AdminAPIKey: ""}, depsWithCatalog(admin, &state.MemoryStore{}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/m3u-sources",
		strings.NewReader(`{"url":"http://persist.test/list.m3u","label":"integ"}`))
	req.Header.Set("Content-Type", "application/json")
	e.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	id, ok := created["id"].(string)
	require.True(t, ok)
	_, perr := uuid.Parse(id)
	require.NoError(t, perr)

	w2 := httptest.NewRecorder()
	e.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/api/v1/m3u-sources", nil))
	require.Equal(t, http.StatusOK, w2.Code)
	var list []map[string]any
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &list))
	require.Len(t, list, 1)
	require.Equal(t, "http://persist.test/list.m3u", list[0]["url"])

	fromDB, err := cat.ListM3USources(ctx)
	require.NoError(t, err)
	require.Len(t, fromDB, 1)
	require.Equal(t, "http://persist.test/list.m3u", fromDB[0].URL)

	w3 := httptest.NewRecorder()
	e.ServeHTTP(w3, httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/m3u-sources/%s", id), nil))
	require.Equal(t, http.StatusNoContent, w3.Code)

	w4 := httptest.NewRecorder()
	e.ServeHTTP(w4, httptest.NewRequest(http.MethodGet, "/api/v1/m3u-sources", nil))
	require.Equal(t, http.StatusOK, w4.Code)
	var after []map[string]any
	require.NoError(t, json.Unmarshal(w4.Body.Bytes(), &after))
	require.Empty(t, after)

	fromDB2, err := cat.ListM3USources(ctx)
	require.NoError(t, err)
	require.Empty(t, fromDB2)
}

func TestIntegration_POST_m3u_sources_duplicate_url_409(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "gotv.db")
	cat, err := storage.Open("file:" + dbPath + "?_pragma=busy_timeout(5000)")
	require.NoError(t, err)
	t.Cleanup(func() { _ = cat.Close() })

	admin := services.NewCatalogAdminService(cat, cat, ports.NoopAppLog{})
	e := newTestEngine(ginapi.Config{AdminAPIKey: ""}, depsWithCatalog(admin, &state.MemoryStore{}))
	body := `{"url":"http://dup.example/list.m3u","label":"one"}`

	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/m3u-sources", strings.NewReader(body)))
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	w2 := httptest.NewRecorder()
	e.ServeHTTP(w2, httptest.NewRequest(http.MethodPost, "/api/v1/m3u-sources", strings.NewReader(body)))
	require.Equal(t, http.StatusConflict, w2.Code, w2.Body.String())

	list, err := cat.ListM3USources(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
}
