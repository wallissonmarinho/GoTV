package ginapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wallissonmarinho/GoTV/internal/adapters/http/ginapi"
	"github.com/wallissonmarinho/GoTV/internal/adapters/state"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// Routes: /health, /playlist.m3u, /epg.xml (no admin).

func TestEndpoint_GET_health(t *testing.T) {
	e := newTestEngine(ginapi.Config{}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "ok", w.Body.String())
}

func TestEndpoint_HEAD_health(t *testing.T) {
	e := newTestEngine(ginapi.Config{}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodHead, "/health", nil))
	require.Equal(t, http.StatusOK, w.Code)
	require.Empty(t, w.Body.String())
}

func TestEndpoint_GET_playlist_m3u_empty_store_404(t *testing.T) {
	e := newTestEngine(ginapi.Config{}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/playlist.m3u", nil))
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestEndpoint_GET_playlist_m3u_with_data_200(t *testing.T) {
	mem := &state.MemoryStore{}
	mem.Set([]byte("#EXTM3U\n"), nil, ports.OutputMeta{OK: true})
	e := newTestEngine(ginapi.Config{}, testDeps(mem))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/playlist.m3u", nil))
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "#EXTM3U")
}

func TestEndpoint_GET_epg_xml_empty_store_404(t *testing.T) {
	e := newTestEngine(ginapi.Config{}, testDeps(&state.MemoryStore{}))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/epg.xml", nil))
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestEndpoint_GET_epg_xml_with_data_200(t *testing.T) {
	mem := &state.MemoryStore{}
	mem.Set(nil, []byte(`<?xml version="1.0"?><tv></tv>`), ports.OutputMeta{OK: true})
	e := newTestEngine(ginapi.Config{}, testDeps(mem))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/epg.xml", nil))
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "<tv>")
}
