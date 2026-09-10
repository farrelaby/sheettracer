package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"sheettracer/internal/db"
	"sheettracer/internal/scan"
	"sheettracer/internal/sheets"

	"golang.org/x/sync/errgroup"
)

// ClientFactory creates a sheets.Client using the caller's context.
type ClientFactory func(ctx context.Context) (*sheets.Client, error)

// TabResult holds the raw BatchGet response for one tab.
type TabResult struct {
	TabName string          `json:"tabName"`
	TabID   int64           `json:"tabId"`
	Range   string          `json:"range"`
	Values  [][]interface{} `json:"values"`
}

// ScanResult holds all tab results for a spreadsheet.
// type ScanResult struct {
// 	SpreadsheetID string      `json:"spreadsheetId"`
// 	Title         string      `json:"title"`
// 	Tabs          []TabResult `json:"tabs"`
// }

type ScanSummary struct {
	SpreadsheetID string `json:"spreadsheetId"`
	Title         string `json:"title"`
	TabsScanned   int    `json:"tabsScanned"`
	EdgesFound    int    `json:"edgesFound"`
}

// ScanService orchestrates formula scanning of tracked spreadsheets.
type ScanService struct {
	factory         ClientFactory
	spreadsheetRepo *db.SpreadsheetRepo
	tabRepo         *db.TabRepo
	edgeRepo        *db.EdgeRepo
	scanRunRepo     *db.ScanRunRepo
	scanCacheRepo   *db.ScanCacheRepo
	client          *sheets.Client // lazily created on first call
}

func NewScanService(
	factory ClientFactory,
	spreadsheetRepo *db.SpreadsheetRepo,
	tabRepo *db.TabRepo,
	edgeRepo *db.EdgeRepo,
	scanRunRepo *db.ScanRunRepo,
	scanCacheRepo *db.ScanCacheRepo,
) *ScanService {
	return &ScanService{
		factory:         factory,
		spreadsheetRepo: spreadsheetRepo,
		tabRepo:         tabRepo,
		edgeRepo:        edgeRepo,
		scanRunRepo:     scanRunRepo,
		scanCacheRepo:   scanCacheRepo,
	}
}

func (s *ScanService) getClient(ctx context.Context) (*sheets.Client, error) {
	if s.client != nil {
		return s.client, nil
	}
	c, err := s.factory(ctx)
	if err != nil {
		return nil, err
	}
	s.client = c
	return c, nil
}

// ScanSpreadsheet fetches formula data from all tabs of a spreadsheet.
// Returns raw cell values so we can inspect the BatchGet response structure.
func (s *ScanService) ScanSpreadsheet(spreadsheetID int, trigger db.TriggeredBy) (*ScanSummary, error) {
	ctx := context.Background()

	// The scan_runs row is an internal audit log (per-sheet recency comes
	// from spreadsheets.last_scan_at). If logging fails the scan still
	// proceeds; finishRun becomes a no-op.
	runID, err := s.scanRunRepo.Add(trigger, time.Now())
	trackRun := err == nil
	if err != nil {
		fmt.Printf("[ScanSpreadsheet] scan run log start failed: %v\n", err)
	}
	finishRun := func(status db.ScanStatus, scanned, skipped int, errMsg *string) {
		if !trackRun {
			return
		}
		if err := s.scanRunRepo.Update(runID, time.Now(), status, scanned, skipped, errMsg); err != nil {
			fmt.Printf("[ScanSpreadsheet] scan run log finish failed: %v\n", err)
		}
	}
	fail := func(status db.ScanStatus, err error) (*ScanSummary, error) {
		msg := err.Error()
		finishRun(status, 0, 0, &msg)
		return nil, err
	}

	ss, err := s.spreadsheetRepo.GetSpreadsheetByID(spreadsheetID)
	if err != nil {
		return fail(db.ScanStatusError, fmt.Errorf("get spreadsheet: %w", err))
	}

	tt, err := s.tabRepo.ListBySpreadsheet(int64(spreadsheetID))
	if err != nil {
		return fail(db.ScanStatusError, fmt.Errorf("get tabs: %w", err))
	}

	client, err := s.getClient(ctx)
	if err != nil {
		return fail(db.ScanStatusError, fmt.Errorf("create client: %w", err))
	}

	scanSummary := &ScanSummary{
		SpreadsheetID: ss.GoogleID,
		Title:         ss.Title,
		TabsScanned:   len(tt),
	}

	start := time.Now()

	g, gctx := errgroup.WithContext(ctx)
	const maxWorker = 3
	g.SetLimit(maxWorker)
	// results := make([]TabResult, len(tt))
	edgeCount := 0
	var edges []db.Edge

	var mux sync.Mutex

	for _, tab := range tt {

		g.Go(func() error {
			resp, err := client.BatchGet(gctx, ss.GoogleID, tab.Title)
			if err != nil {
				return fmt.Errorf("batch get %s: %w", tab.Title, err)
			}

			tr := TabResult{
				TabName: tab.Title,
				TabID:   tab.TabID,
			}

			if len(resp.ValueRanges) > 0 {
				tr.Range = resp.ValueRanges[0].Range
				tr.Values = resp.ValueRanges[0].Values
			}

			mux.Lock()
			e := scan.ScanValues(tab, tr.Values)
			edgeCount += len(e)
			edges = append(edges, e...)
			mux.Unlock()

			return nil

		})
	}

	if err := g.Wait(); err != nil {
		return fail(db.ScanStatusError, err)
	}

	if err := s.edgeRepo.UpsertAll(int64(spreadsheetID), edges); err != nil {
		return fail(db.ScanStatusError, fmt.Errorf("upsert edges for %w", err))
	}

	// Stamp per-sheet recency for the sidebar/inspector "last scan" display.
	// A stamp failure shouldn't fail an otherwise good scan, so just log it.
	if err := s.spreadsheetRepo.UpdateLastScanAt(int64(spreadsheetID), time.Now()); err != nil {
		fmt.Printf("[ScanSpreadsheet] update last_scan_at: %v\n", err)
	}

	scanSummary.EdgesFound = edgeCount

	finishRun(db.ScanStatusOK, scanSummary.TabsScanned, 0, nil)

	fmt.Printf("[ScanSpreadsheet] took %v\n", time.Since(start))

	return scanSummary, nil
}
