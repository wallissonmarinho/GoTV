-- +goose Up
ALTER TABLE epg_sources RENAME TO epg_sources_legacy;
CREATE TABLE epg_sources (
  id UUID PRIMARY KEY,
  url TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL,
  CONSTRAINT epg_sources_url_key UNIQUE (url)
);
INSERT INTO epg_sources (id, url, label, created_at)
SELECT gen_random_uuid(), url, label, created_at FROM epg_sources_legacy;
DROP TABLE epg_sources_legacy;

-- +goose Down
ALTER TABLE epg_sources RENAME TO epg_sources_legacy;
CREATE TABLE epg_sources (
  id BIGSERIAL PRIMARY KEY,
  url TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
);
INSERT INTO epg_sources (url, label, created_at)
SELECT url, label, created_at FROM epg_sources_legacy;
DROP TABLE epg_sources_legacy;
