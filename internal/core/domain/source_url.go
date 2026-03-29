package domain

import (
	"net/url"
	"strings"
)

// ValidateM3USourceURL returns an error if raw is not an absolute http(s) URL whose path ends with .m3u or .m3u8.
func ValidateM3USourceURL(raw string) error {
	path, err := sourceURLPath(raw)
	if err != nil {
		return err
	}
	if strings.HasSuffix(path, ".m3u") || strings.HasSuffix(path, ".m3u8") {
		return nil
	}
	return ErrInvalidM3USourceURLSuffix
}

// ValidateEPGSourceURL returns an error if raw is not an absolute http(s) URL whose path ends with .xml or .xml.gz.
func ValidateEPGSourceURL(raw string) error {
	path, err := sourceURLPath(raw)
	if err != nil {
		return err
	}
	if strings.HasSuffix(path, ".xml.gz") || strings.HasSuffix(path, ".xml") {
		return nil
	}
	return ErrInvalidEPGSourceURLSuffix
}

func sourceURLPath(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", ErrInvalidSourceURL
	}
	return strings.ToLower(u.Path), nil
}
