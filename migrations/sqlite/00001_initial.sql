-- +goose Up
CREATE TABLE IF NOT EXISTS m3u_sources (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  url TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS epg_sources (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  url TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS merge_snapshot (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  m3u_text TEXT NOT NULL DEFAULT '',
  epg_text TEXT NOT NULL DEFAULT '',
  ok INTEGER NOT NULL DEFAULT 0,
  message TEXT NOT NULL DEFAULT '',
  last_error TEXT NOT NULL DEFAULT '',
  channel_count INTEGER NOT NULL DEFAULT 0,
  programme_count INTEGER NOT NULL DEFAULT 0,
  started_at TEXT,
  finished_at TEXT
);

INSERT OR IGNORE INTO merge_snapshot (id) VALUES (1);

-- +goose Down
DROP TABLE IF EXISTS merge_snapshot;
DROP TABLE IF EXISTS epg_sources;
DROP TABLE IF EXISTS m3u_sources;
