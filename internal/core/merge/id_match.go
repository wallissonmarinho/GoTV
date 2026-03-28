package merge

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var accentFold = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

func foldAccents(s string) string {
	out, _, err := transform.String(accentFold, s)
	if err != nil {
		return s
	}
	return out
}

func normalizeEPGIDKey(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if i := strings.IndexByte(s, '@'); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	s = foldAccents(s)
	return strings.ToLower(s)
}

var epgIDAliasToCanonical = map[string]string{
	"recordtvsaopaulo.br": "recordtvsp.br",
}

func canonicalEPGIDKey(raw string) string {
	k := normalizeEPGIDKey(raw)
	for {
		next, ok := epgIDAliasToCanonical[k]
		if !ok || next == k {
			return k
		}
		k = next
	}
}
