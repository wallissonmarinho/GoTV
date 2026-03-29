-- +goose Up
ALTER TABLE m3u_sources RENAME TO m3u_sources_legacy;
CREATE TABLE m3u_sources (
  id TEXT NOT NULL PRIMARY KEY,
  url TEXT NOT NULL UNIQUE,
  label TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
INSERT INTO m3u_sources (id, url, label, created_at)
SELECT
  lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(lower(hex(randomblob(2))), 3) || '-' ||
        substr('89ab', abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))), 3) || '-' || lower(hex(randomblob(6)))),
  url, label, created_at
FROM m3u_sources_legacy;
DROP TABLE m3u_sources_legacy;

-- +goose Down
ALTER TABLE m3u_sources RENAME TO m3u_sources_legacy;
CREATE TABLE m3u_sources (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  url TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
INSERT INTO m3u_sources (url, label, created_at)
SELECT url, label, created_at FROM m3u_sources_legacy;
DROP TABLE m3u_sources_legacy;
