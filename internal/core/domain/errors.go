package domain

import "errors"

// ErrDuplicateM3USourceURL is returned when an M3U source with the same URL already exists.
var ErrDuplicateM3USourceURL = errors.New("m3u source with this url already exists")

// ErrInvalidSourceURL is returned when a source URL is missing scheme/host or is not http/https.
var ErrInvalidSourceURL = errors.New("source url must be absolute http or https")

// ErrInvalidM3USourceURLSuffix is returned when the URL path does not end with .m3u or .m3u8.
var ErrInvalidM3USourceURLSuffix = errors.New("m3u source url path must end with .m3u or .m3u8")

// ErrInvalidEPGSourceURLSuffix is returned when the URL path does not end with .xml or .xml.gz.
var ErrInvalidEPGSourceURLSuffix = errors.New("epg source url path must end with .xml or .xml.gz")
