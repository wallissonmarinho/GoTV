package m3u_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wallissonmarinho/GoTV/internal/adapters/m3u"
)

func TestParser_basic(t *testing.T) {
	const input = `#EXTM3U
#EXTINF:-1 tvg-id="c1" tvg-name="One" group-title="News",Channel One
http://example.com/one.m3u8
`
	p := m3u.Parser{}
	ch, err := p.Parse(input)
	require.NoError(t, err)
	require.Len(t, ch, 1)
	require.Equal(t, "Channel One", ch[0].Name)
	require.Equal(t, "c1", ch[0].TVGID)
	require.Equal(t, "News", ch[0].Group)
	require.Equal(t, "http://example.com/one.m3u8", ch[0].URL)
}
