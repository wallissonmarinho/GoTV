-- +goose Up
CREATE TABLE IF NOT EXISTS m3u_sources (
  id BIGSERIAL PRIMARY KEY,
  url TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS epg_sources (
  id BIGSERIAL PRIMARY KEY,
  url TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS merge_snapshot (
  id SMALLINT PRIMARY KEY CHECK (id = 1),
  m3u_text TEXT NOT NULL DEFAULT '',
  epg_text TEXT NOT NULL DEFAULT '',
  ok BOOLEAN NOT NULL DEFAULT FALSE,
  message TEXT NOT NULL DEFAULT '',
  last_error TEXT NOT NULL DEFAULT '',
  channel_count INT NOT NULL DEFAULT 0,
  programme_count INT NOT NULL DEFAULT 0,
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ
);

INSERT INTO merge_snapshot (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS merge_snapshot;
DROP TABLE IF EXISTS epg_sources;
DROP TABLE IF EXISTS m3u_sources;
