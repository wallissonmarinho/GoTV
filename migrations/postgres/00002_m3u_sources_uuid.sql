-- +goose Up
ALTER TABLE m3u_sources RENAME TO m3u_sources_legacy;
CREATE TABLE m3u_sources (
  id UUID PRIMARY KEY,
  url TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  CONSTRAINT m3u_sources_url_key UNIQUE (url)
);
INSERT INTO m3u_sources (id, url, label, created_at)
SELECT gen_random_uuid(), url, label, created_at FROM m3u_sources_legacy;
DROP TABLE m3u_sources_legacy;

-- +goose Down
ALTER TABLE m3u_sources RENAME TO m3u_sources_legacy;
CREATE TABLE m3u_sources (
  id BIGSERIAL PRIMARY KEY,
  url TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);
INSERT INTO m3u_sources (url, label, created_at)
SELECT url, label, created_at FROM m3u_sources_legacy;
DROP TABLE m3u_sources_legacy;
