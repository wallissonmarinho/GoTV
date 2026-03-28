package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/merge"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// PlaylistMergeDeps bundles dependencies for NewPlaylistMergeService.
type PlaylistMergeDeps struct {
	Repo         ports.CatalogRepository
	UoW          ports.UnitOfWork
	HTTP         ports.HTTPGetter
	M3U          ports.M3UParser
	XMLTV        ports.XMLTVParser
	Probe        ports.StreamProbe
	Clock        ports.Clock
	Store        ports.OutputStore
	M3UOut       ports.M3UPlaylistEncoder
	EPGOut       ports.XMLTVEncoder
	ProbeStreams bool
	MaxProbes    int
	ProbeWorkers int
}

// PlaylistMergeService implements ports.MergeRunner: fetch sources, merge, encode, persist snapshot and cache.
type PlaylistMergeService struct {
	runMu        sync.Mutex
	Repo         ports.CatalogRepository
	UoW          ports.UnitOfWork
	HTTP         ports.HTTPGetter
	M3U          ports.M3UParser
	XMLTV        ports.XMLTVParser
	Probe        ports.StreamProbe
	Clock        ports.Clock
	Store        ports.OutputStore
	M3UOut       ports.M3UPlaylistEncoder
	EPGOut       ports.XMLTVEncoder
	ProbeStreams bool
	MaxProbes    int
	ProbeWorkers int
}

// NewPlaylistMergeService returns a configured merge runner (snapshot writes use UoW when non-nil).
func NewPlaylistMergeService(d PlaylistMergeDeps) *PlaylistMergeService {
	return &PlaylistMergeService{
		Repo:         d.Repo,
		UoW:          d.UoW,
		HTTP:         d.HTTP,
		M3U:          d.M3U,
		XMLTV:        d.XMLTV,
		Probe:        d.Probe,
		Clock:        d.Clock,
		Store:        d.Store,
		M3UOut:       d.M3UOut,
		EPGOut:       d.EPGOut,
		ProbeStreams: d.ProbeStreams,
		MaxProbes:    d.MaxProbes,
		ProbeWorkers: d.ProbeWorkers,
	}
}

// Run executes one full merge pass (used by cron and optional manual trigger).
func (s *PlaylistMergeService) Run(ctx context.Context) domain.MergeResult {
	s.runMu.Lock()
	defer s.runMu.Unlock()

	started := s.Clock.Now()
	var errs []string

	m3uSources, err := s.Repo.ListM3USources(ctx)
	if err != nil {
		r := domain.MergeResult{
			OK: false, Message: "list m3u sources: " + err.Error(),
			StartedAt: started, FinishedAt: s.Clock.Now(),
		}
		s.persistFailure(ctx, started, r.Message, errs)
		return r
	}
	epgSources, err := s.Repo.ListEPGSources(ctx)
	if err != nil {
		r := domain.MergeResult{
			OK: false, Message: "list epg sources: " + err.Error(),
			StartedAt: started, FinishedAt: s.Clock.Now(),
		}
		s.persistFailure(ctx, started, r.Message, errs)
		return r
	}

	if len(m3uSources) == 0 {
		emptyM3U := []byte("#EXTM3U\n")
		emptyEPG := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n<tv></tv>\n")
		msg := "no m3u sources registered"
		meta := ports.OutputMeta{OK: true, Message: msg}
		s.Store.Set(emptyM3U, emptyEPG, meta)
		_ = s.saveSnapshotTx(ctx, domain.MergeSnapshot{
			M3U: emptyM3U, EPG: emptyEPG, OK: true, Message: msg,
			StartedAt: started, FinishedAt: s.Clock.Now(),
		})
		return domain.MergeResult{OK: true, Message: msg, StartedAt: started, FinishedAt: s.Clock.Now()}
	}

	var batches [][]domain.Channel
	for _, src := range m3uSources {
		body, err := s.HTTP.Get(ctx, src.URL)
		if err != nil {
			errs = append(errs, fmt.Sprintf("m3u source %d (%s): %v", src.ID, src.Label, err))
			batches = append(batches, nil)
			continue
		}
		chans, perr := s.M3U.Parse(sanitizeM3UText(string(body)))
		if perr != nil {
			errs = append(errs, fmt.Sprintf("m3u source %d parse: %v", src.ID, perr))
			batches = append(batches, nil)
			continue
		}
		if s.ProbeStreams && s.Probe != nil && len(chans) > 0 {
			chans = s.probeChannelsPool(ctx, chans, s.MaxProbes, probeConcurrency(s.ProbeWorkers))
		}
		batches = append(batches, chans)
	}

	merged := merge.MergeChannelsByURLOrder(batches)

	var epgs []*domain.EPGData
	for _, src := range epgSources {
		body, err := s.HTTP.Get(ctx, src.URL)
		if err != nil {
			errs = append(errs, fmt.Sprintf("epg source %d (%s): %v", src.ID, src.Label, err))
			continue
		}
		ed, xerr := s.XMLTV.Parse(body)
		if xerr != nil {
			errs = append(errs, fmt.Sprintf("epg source %d parse: %v", src.ID, xerr))
			continue
		}
		epgs = append(epgs, ed)
	}
	mergedEPG := merge.MergeIndependentEPGs(epgs)
	outEPG := merge.BuildOutputEPGForPlaylist(merged, mergedEPG)

	m3uBytes := s.M3UOut.Build(merged)
	epgBytes := s.EPGOut.Build(outEPG)
	finished := s.Clock.Now()
	msg := fmt.Sprintf("merged %d channels, %d programmes", len(merged), len(outEPG.Programmes))
	meta := ports.OutputMeta{
		OK:             true,
		Message:        msg,
		ChannelCount:   len(merged),
		ProgrammeCount: len(outEPG.Programmes),
	}
	s.Store.Set(m3uBytes, epgBytes, meta)
	lastErr := ""
	if len(errs) > 0 {
		lastErr = strings.Join(errs, "; ")
	}
	_ = s.saveSnapshotTx(ctx, domain.MergeSnapshot{
		M3U: m3uBytes, EPG: epgBytes, OK: true, Message: msg, LastError: lastErr,
		ChannelCount: len(merged), ProgrammeCount: len(outEPG.Programmes),
		StartedAt: started, FinishedAt: finished,
	})

	return domain.MergeResult{
		OK: true, Message: msg, Errors: errs,
		ChannelCount: len(merged), ProgrammeCount: len(outEPG.Programmes),
		StartedAt: started, FinishedAt: finished,
	}
}

func (s *PlaylistMergeService) persistFailure(ctx context.Context, started time.Time, msg string, errs []string) {
	finished := s.Clock.Now()
	lastErr := msg
	if len(errs) > 0 {
		lastErr = msg + "; " + strings.Join(errs, "; ")
	}
	_ = s.withWritableRepo(ctx, func(ctx context.Context, repo ports.CatalogRepository) error {
		prev, err := repo.LoadSnapshot(ctx)
		if err != nil {
			return err
		}
		return repo.SaveSnapshot(ctx, domain.MergeSnapshot{
			M3U: prev.M3U, EPG: prev.EPG,
			OK: false, Message: msg, LastError: lastErr,
			ChannelCount: prev.ChannelCount, ProgrammeCount: prev.ProgrammeCount,
			StartedAt: started, FinishedAt: finished,
		})
	})
}

func (s *PlaylistMergeService) saveSnapshotTx(ctx context.Context, snap domain.MergeSnapshot) error {
	return s.withWritableRepo(ctx, func(ctx context.Context, repo ports.CatalogRepository) error {
		return repo.SaveSnapshot(ctx, snap)
	})
}

func (s *PlaylistMergeService) withWritableRepo(ctx context.Context, fn func(ctx context.Context, repo ports.CatalogRepository) error) error {
	if s.UoW != nil {
		return s.UoW.WithinTx(ctx, fn)
	}
	return fn(ctx, s.Repo)
}

func (s *PlaylistMergeService) probeChannelsPool(ctx context.Context, chans []domain.Channel, max int, workers int) []domain.Channel {
	if max <= 0 {
		max = len(chans)
	}
	if workers <= 0 {
		workers = 32
	}
	type job struct{ idx int }
	jobs := make(chan job)
	results := make([]domain.Channel, len(chans))
	copy(results, chans)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				ok, inc := s.Probe.Probe(ctx, results[j.idx].URL)
				if !ok && !inc {
					results[j.idx].URL = ""
				}
			}
		}()
	}

	sent := 0
	for i := range chans {
		if sent >= max {
			break
		}
		jobs <- job{idx: i}
		sent++
	}
	close(jobs)
	wg.Wait()

	filtered := results[:0]
	for _, c := range results {
		if strings.TrimSpace(c.URL) != "" {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func probeConcurrency(n int) int {
	if n <= 0 {
		return 32
	}
	return n
}

func sanitizeM3UText(s string) string {
	s = strings.TrimPrefix(s, "\ufeff")
	return strings.ReplaceAll(s, "\r\n", "\n")
}

var _ ports.MergeRunner = (*PlaylistMergeService)(nil)
