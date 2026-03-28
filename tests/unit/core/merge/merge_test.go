package merge_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/merge"
)

func TestMergeChannelsByURLOrder_dedupesURL(t *testing.T) {
	a := []domain.Channel{{Name: "A", URL: "http://x/1", TVGID: "id1"}}
	b := []domain.Channel{{Name: "B", URL: "http://x/1", TVGID: "id2"}}
	out := merge.MergeChannelsByURLOrder([][]domain.Channel{a, b})
	require.Len(t, out, 1)
	require.Equal(t, "A", out[0].Name)
}

func TestRemapEPG_alignsProgrammeToMergedTvgID(t *testing.T) {
	perM3U := [][]domain.Channel{{
		{Name: "Globo", URL: "http://s/globo", TVGID: "globo.hd"},
	}}
	epg := &domain.EPGData{
		Channels: []domain.EPGChannel{
			{ID: "globo.hd", DisplayNames: []string{"Globo HD"}},
		},
		Programmes: []domain.EPGProgramme{
			{Channel: "globo.hd", Start: "20240101120000 +0000", Stop: "20240101130000 +0000", Titles: []string{"News"}},
		},
	}
	merged := merge.MergeChannelsByURLOrder(perM3U)
	out := merge.RemapEPG(merged, perM3U, []*domain.EPGData{epg})
	require.Len(t, out.Programmes, 1)
	require.Equal(t, "globo.hd", out.Programmes[0].Channel)
}

func TestRemapEPG_matchesPlaylistIDWithFeedSuffix(t *testing.T) {
	perM3U := [][]domain.Channel{{
		{Name: "Rede TV! SP", URL: "http://s/rtv", TVGID: "RedeTV.br@SD"},
	}}
	epg := &domain.EPGData{
		Channels: []domain.EPGChannel{
			{ID: "RedeTV.br", DisplayNames: []string{"BR - RedeTV!"}},
		},
		Programmes: []domain.EPGProgramme{
			{Channel: "RedeTV.br", Start: "20240101120000 +0000", Stop: "20240101130000 +0000", Titles: []string{"X"}},
		},
	}
	merged := merge.MergeChannelsByURLOrder(perM3U)
	out := merge.RemapEPG(merged, perM3U, []*domain.EPGData{epg})
	require.Len(t, out.Programmes, 1)
	require.Equal(t, "RedeTV.br@SD", out.Programmes[0].Channel)
}

func TestRemapEPG_matchesAccentDifferenceInID(t *testing.T) {
	perM3U := [][]domain.Channel{{
		{Name: "Record TV Goias", URL: "http://s/recgo", TVGID: "RecordTVGoias.br@SD"},
	}}
	epg := &domain.EPGData{
		Channels: []domain.EPGChannel{
			{ID: "RecordTVGoiás.br", DisplayNames: []string{"BR - Record TV Goiás"}},
		},
		Programmes: []domain.EPGProgramme{
			{Channel: "RecordTVGoiás.br", Start: "20240101120000 +0000", Stop: "20240101130000 +0000", Titles: []string{"Y"}},
		},
	}
	merged := merge.MergeChannelsByURLOrder(perM3U)
	out := merge.RemapEPG(merged, perM3U, []*domain.EPGData{epg})
	require.Len(t, out.Programmes, 1)
	require.Equal(t, "RecordTVGoias.br@SD", out.Programmes[0].Channel)
}

func TestRemapEPG_matchesAliasSaoPauloToSP(t *testing.T) {
	perM3U := [][]domain.Channel{{
		{Name: "Record TV SP", URL: "http://s/recsp", TVGID: "RecordTVSaoPaulo.br@SD"},
	}}
	epg := &domain.EPGData{
		Channels: []domain.EPGChannel{
			{ID: "RecordTVSP.br", DisplayNames: []string{"BR - RecordTV SP"}},
		},
		Programmes: []domain.EPGProgramme{
			{Channel: "RecordTVSP.br", Start: "20240101120000 +0000", Stop: "20240101130000 +0000", Titles: []string{"Z"}},
		},
	}
	merged := merge.MergeChannelsByURLOrder(perM3U)
	out := merge.RemapEPG(merged, perM3U, []*domain.EPGData{epg})
	require.Len(t, out.Programmes, 1)
	require.Equal(t, "RecordTVSaoPaulo.br@SD", out.Programmes[0].Channel)
}

func TestMergeIndependentEPGs_dedupes(t *testing.T) {
	a := &domain.EPGData{
		Channels: []domain.EPGChannel{{ID: "c1", DisplayNames: []string{"One"}}},
		Programmes: []domain.EPGProgramme{
			{Channel: "c1", Start: "1", Stop: "2", Titles: []string{"T"}},
		},
	}
	b := &domain.EPGData{
		Channels: []domain.EPGChannel{{ID: "c1", DisplayNames: []string{"Dup"}}},
		Programmes: []domain.EPGProgramme{
			{Channel: "c1", Start: "1", Stop: "2", Titles: []string{"T"}},
		},
	}
	out := merge.MergeIndependentEPGs([]*domain.EPGData{a, b})
	require.Len(t, out.Channels, 1)
	require.Len(t, out.Programmes, 1)
}
