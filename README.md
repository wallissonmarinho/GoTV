# GoTV

HTTP service in Go (hexagonal layout) that registers multiple **M3U** and **EPG (XMLTV)** URLs, merges them on a schedule, optionally validates stream URLs, and serves a unified `playlist.m3u` and `epg.xml`.

## Architecture

- **Core:** `internal/core/domain` (entities), `ports` (interfaces), `merge` (pure playlist/EPG merge rules), `services` (application use cases: `PlaylistMergeService`, `CatalogAdminService`)
- **Composition:** `internal/app` wires catalog, merge service, and store hydration (used from `cmd/gotv`).
- **Adapters:** Gin HTTP (`internal/adapters/http/ginapi`: `router.go`, `handlers_public.go`, `handlers_admin.go`, `handlers_admin_types.go`, `handlers_m3u_sources.go`, `handlers_epg_sources.go`, `handlers_merge.go`, `middleware.go`; handlers depend only on `ports`), HTTP client + stream probe, **M3U/XMLTV** parse + **write** (`internal/adapters/m3u`, `internal/adapters/xmltv`), `storage` (SQLite or PostgreSQL) with **Unit of Work** (`Catalog.WithinTx` + transactional `catalogRepo`), in-memory output cache, **scheduler** (`GOTV_MERGE_INTERVAL`)

The merge service depends on `ports.M3UPlaylistEncoder` / `ports.XMLTVEncoder`; the default implementations are the adapters above (hexagonal “out” side).

### Dependency rules (hexagonal boundaries)

- Packages under `internal/core/...` must **not** import `github.com/gin-gonic/gin`, `internal/adapters/...`, or SQL drivers. They depend only on stdlib, `internal/core`, and neutral libs (e.g. `golang.org/x/text` in `merge`).
- `internal/adapters/...` may import `internal/core/domain` and `internal/core/ports` and **implement** those interfaces (HTTP, DB, parsers, encoders).
- `cmd/gotv` and `internal/app` compose the application: they import adapters and core services and pass concrete implementations into constructors.

**HTTP handlers** (`ginapi`) should only validate/bind requests, map status codes and JSON, and call `ports` — no business rules (those stay in `services` / `merge`).

**Terminology:** Some projects label driving adapters **inbound** and driven ones **outbound**. This repo groups implementations under `internal/adapters` by technology (`http`, `storage`, …), which is equivalent; renaming folders is optional.

## Build

```bash
go build -o bin/gotv ./cmd/gotv
```

## Run (local)

```bash
export GOTV_ADMIN_API_KEY="change-me"   # optional; if unset, admin routes are open
./bin/gotv
```

- Listens on `GOTV_ADDR` (default `:8080`) or **`PORT`** (e.g. Heroku).
- SQLite default: `./data/gotv.db` unless `DATABASE_URL` or `GOTV_SQLITE_DSN` is set.
- PostgreSQL: set `DATABASE_URL` to a `postgres://` or `postgresql://` URL.

## Environment variables

| Variable | Description |
|----------|-------------|
| `PORT` | Listen port (Heroku). |
| `GOTV_ADDR` | Listen address if `PORT` unset (default `:8080`). |
| `DATABASE_URL` | PostgreSQL DSN; if empty, SQLite file under `GOTV_DATA_DIR`. |
| `GOTV_DATA_DIR` | Data directory for SQLite (default `./data`). |
| `GOTV_SQLITE_DSN` | Full SQLite DSN (overrides default file path when `DATABASE_URL` is empty). |
| `GOTV_MERGE_INTERVAL` | Merge job period (default `30m`). |
| `GOTV_ADMIN_API_KEY` or `ADMIN_API_KEY` | Bearer token or `X-Admin-API-Key` for `/api/v1/*`. |
| `GOTV_HTTP_TIMEOUT` | Timeout for fetching playlists/EPG (default `45s`). |
| `GOTV_MAX_BODY_BYTES` | Max download size (default 50MiB). |
| `GOTV_USER_AGENT` | User-Agent for outbound HTTP. |
| `GOTV_PROBE_TIMEOUT` | Per-stream probe timeout (default `8s`). |
| `GOTV_PROBE_STREAMS` | `true` to drop streams that fail probe (default `false`). |
| `GOTV_MAX_PROBES` | Max streams to probe per merge (default `500`). |
| `GOTV_PROBE_WORKERS` | Concurrent probes (default `32`). |
| `GIN_MODE` | Set to `release` for production logging. |
| `OTEL_SDK_DISABLED` | Set to `true` to skip OTLP exporters (stderr JSON logs only). |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Base URL for OTLP HTTP (e.g. `https://otlp-gateway:4318`). When set (or logs/traces-specific URLs below), traces and logs export via OTLP. |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | Optional override for trace export URL. |
| `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` | Optional override for log export URL. |
| `OTEL_EXPORTER_OTLP_HEADERS` | `key=value` pairs comma-separated for auth (same as standard OTel). |
| `OTEL_SERVICE_NAME` | Service name for resources and Gin instrumentation (default `gotv` for HTTP middleware). |
| `OTEL_RESOURCE_ATTRIBUTES` | Extra resource attributes (`key=value,key=value`). |

## Database migrations

Schema is managed with [goose](https://github.com/pressly/goose) SQL files embedded in the binary (`migrations/postgres`, `migrations/sqlite`). On startup the app applies pending **up** migrations automatically.

CLI (same binary as the server):

```bash
./bin/gotv migrate up        # apply all pending
./bin/gotv migrate down [N]  # roll back N steps (default 1)
./bin/gotv migrate version   # print highest applied version
./bin/gotv migrate status    # list files vs applied state
./bin/gotv migrate goto V    # migrate up or down to version V (0 = goose baseline only)
```

Uses the same `DATABASE_URL` / SQLite DSN as the server. **Down** drops tables; use only in dev/staging or with backups.

Goose stores versions in `goose_db_version` (per dialect).

## Observability

- **Traces:** Gin requests get a span via `otelgin`; merge runs use spans `merge.scheduled`, `merge.initial`, and `merge.manual` (rebuild).
- **Logs:** `log/slog` with JSON to stderr; when an OTLP endpoint env var is set, the same records are bridged to the OpenTelemetry Logs API (`otelslog`), including **trace_id** / **span_id** when a span is active in context.

Local dev without a collector: leave OTLP env vars unset (stderr only), or set `OTEL_SDK_DISABLED=true`.

## API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | no | Liveness. |
| GET | `/playlist.m3u` | no | Last merged M3U. |
| GET | `/epg.xml` | no | Last merged EPG. |
| POST | `/api/v1/m3u-sources` | yes | Body: `{"url":"...","label":"..."}`. |
| GET | `/api/v1/m3u-sources` | yes | List M3U sources. |
| DELETE | `/api/v1/m3u-sources/:id` | yes | Remove source. |
| POST | `/api/v1/epg-sources` | yes | Body: `{"url":"...","label":"..."}`. |
| GET | `/api/v1/epg-sources` | yes | List EPG sources. |
| DELETE | `/api/v1/epg-sources/:id` | yes | Remove source. |
| POST | `/api/v1/rebuild` | yes | Triggers merge asynchronously (`202`). |
| GET | `/api/v1/merge-status` | yes | Last merge metadata from DB. |

Admin auth: `Authorization: Bearer <key>` or header `X-Admin-API-Key: <key>`.

## Merge job

An in-process ticker runs `PlaylistMergeService.Run` every `GOTV_MERGE_INTERVAL`. Manual `POST /api/v1/rebuild` uses the same path (serialized with a mutex). A merge also runs once shortly after process start.

## Deploy (Heroku-style)

1. Set buildpack: **Heroku Go** (or build `bin/gotv` in CI and use a slug with `Procfile`).
2. Point the build at the main package: `heroku config:set GO_INSTALL_PACKAGE_SPEC=./cmd/gotv`
3. Configure `DATABASE_URL` (Heroku Postgres or external Neon/Supabase).
4. Set `GOTV_ADMIN_API_KEY`, `GOTV_MERGE_INTERVAL`, and optionally `GIN_MODE=release`.
5. Run `./bin/gotv migrate up` in a **release phase** (or before start) so schema is applied before traffic; the web process also runs pending migrations on boot.

Free Heroku dynos are discontinued; use Eco or another PaaS with the same `PORT` + `Procfile` pattern.

## GitHub Actions

On push of a tag matching `v*`, `.github/workflows/release-deploy.yml` runs tests and can deploy to Heroku when secrets are set: `HEROKU_API_KEY`, `HEROKU_APP_NAME`, `HEROKU_EMAIL`.

## Test

All test files live under `tests/` (no `*_test.go` next to production code under `internal/`).

```bash
go test ./...
```

Layout:

- `tests/unit/core/` — domain merge rules, `CatalogAdminService`
- `tests/unit/adapters/` — M3U parser, storage `WithinTx`
- `tests/integration/http/ginapi/` — HTTP tests (`httptest` + `ginapi.Register`): vários `*_test.go` por rota/recurso (`public_routes_test.go`, `m3u_sources_test.go`, …), stubs em `helpers_test.go`, e fluxos com **SQLite temporário** (POST→GET→DELETE) onde faz sentido. Verbose: `go test ./tests/integration/http/ginapi -v`
</think>
<｜tool▁calls▁begin｜><｜tool▁call▁begin｜>
Shell
