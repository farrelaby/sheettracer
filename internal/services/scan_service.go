package services

import (
	"context"
	"fmt"

	"sheettracer/internal/db"
	"sheettracer/internal/scan"
	"sheettracer/internal/sheets"
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
type ScanResult struct {
	SpreadsheetID string      `json:"spreadsheetId"`
	Title         string      `json:"title"`
	Tabs          []TabResult `json:"tabs"`
}

// ScanService orchestrates formula scanning of tracked spreadsheets.
type ScanService struct {
	factory         ClientFactory
	spreadsheetRepo *db.SpreadsheetRepo
	tabRepo         *db.TabRepo
	edgeRepo        *db.EdgeRepo
	client          *sheets.Client // lazily created on first call
}

func NewScanService(
	factory ClientFactory,
	spreadsheetRepo *db.SpreadsheetRepo,
	tabRepo *db.TabRepo,
	edgeRepo *db.EdgeRepo,
) *ScanService {
	return &ScanService{
		factory:         factory,
		spreadsheetRepo: spreadsheetRepo,
		tabRepo:         tabRepo,
		edgeRepo:        edgeRepo,
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
func (s *ScanService) ScanSpreadsheet(spreadsheetID int) (*ScanResult, error) {
	ctx := context.Background()

	ss, err := s.spreadsheetRepo.GetSpreadsheetByID(spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet: %w", err)
	}

	tt, err := s.tabRepo.ListBySpreadsheet(int64(spreadsheetID))
	if err != nil {
		return nil, fmt.Errorf("get tabs: %w", err)
	}

	client, err := s.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	result := &ScanResult{
		SpreadsheetID: ss.GoogleID,
		Title:         ss.Title,
		Tabs:          make([]TabResult, 0, len(tt)),
	}

	for _, tab := range tt {
		resp, err := client.BatchGet(ctx, ss.GoogleID, tab.Title)
		if err != nil {
			return nil, fmt.Errorf("batch get %s: %w", tab.Title, err)
		}

		tr := TabResult{
			TabName: tab.Title,
			TabID:   tab.TabID,
		}

		if len(resp.ValueRanges) > 0 {
			tr.Range = resp.ValueRanges[0].Range
			tr.Values = resp.ValueRanges[0].Values
		}

		scan.ScanValues(tab, tr.Values)

		result.Tabs = append(result.Tabs, tr)
	}

	return result, nil
}
