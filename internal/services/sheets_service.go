package services

import (
	"context"
	"fmt"
	"time"

	"sheettracer/internal/db"
	"sheettracer/internal/sheets"
)

// SheetsService exposes spreadsheet management to the frontend. It is a
// Wails-bound service.
type SheetsService struct {
	factory         ClientFactory
	spreadsheetRepo *db.SpreadsheetRepo
	tabRepo         *db.TabRepo
	client          *sheets.Client // lazily created on first call
}

func NewSheetsService(
	factory ClientFactory,
	spreadsheetRepo *db.SpreadsheetRepo,
	tabRepo *db.TabRepo,
) *SheetsService {
	return &SheetsService{
		factory:         factory,
		spreadsheetRepo: spreadsheetRepo,
		tabRepo:         tabRepo,
	}
}

func (s *SheetsService) getClient(ctx context.Context) (*sheets.Client, error) {
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

// Add fetches metadata for the given Google Sheets URL, persists it, and
// returns the saved spreadsheet. Returns an error with "no access / not found"
// if Drive returns 404.
func (s *SheetsService) Add(url string, notes string) (*db.Spreadsheet, error) {
	spreadsheetID, err := sheets.ParseID(url)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	ctx := context.Background()

	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	meta, err := client.FetchMetadata(ctx, spreadsheetID)
	if err != nil {
		return nil, err
	}

	sp := &db.Spreadsheet{
		GoogleID:     meta.SpreadsheetID,
		Title:        meta.Title,
		URL:          url,
		Notes:        notes,
		Version:      meta.Version,
		ModifiedTime: meta.ModifiedTime,
		Visibility:   meta.Visibility,
	}

	id, err := s.spreadsheetRepo.UpsertByGoogleID(sp)
	if err != nil {
		return nil, fmt.Errorf("[persist spreadsheet]: %w", err)
	}
	sp.ID = id

	// Persist tabs.
	tabs := make([]db.Tab, len(meta.Tabs))
	for i, t := range meta.Tabs {
		tabs[i] = db.Tab{
			SpreadsheetID: id,
			TabID:         t.SheetID,
			Title:         t.Title,
			Idx:           t.Idx,
		}
	}
	if err := s.tabRepo.UpsertAll(id, tabs); err != nil {
		return nil, fmt.Errorf("[persist tabs]: %w", err)
	}

	return sp, nil
}

// List returns all tracked spreadsheets.
func (s *SheetsService) List() ([]db.Spreadsheet, error) {
	return s.spreadsheetRepo.ListTracked()
}

// Remove deletes a spreadsheet and its tabs.
func (s *SheetsService) Remove(id int64) error {
	return s.spreadsheetRepo.Delete(id)
}

type DbMetadata struct {
	SpreadsheetID string
	Title         string
	Tabs          []db.Tab
	Version       string
	ModifiedTime  string
	Visibility    string
	LastScanAt    *time.Time
}

// FetchDbMetadata returns spreadsheet metadata from the database.
func (s *SheetsService) FetchDbMetadata(spreadsheetID int) (*DbMetadata, error) {
	ss, err := s.spreadsheetRepo.GetSpreadsheetByID(spreadsheetID)
	if err != nil {
		return nil, err
	}

	tt, err := s.tabRepo.ListBySpreadsheet(int64(spreadsheetID))
	if err != nil {
		return nil, err
	}

	return &DbMetadata{
		SpreadsheetID: ss.GoogleID,
		Title:         ss.Title,
		Visibility:    ss.Visibility,
		Version:       ss.Version,
		ModifiedTime:  ss.ModifiedTime,
		LastScanAt:    ss.LastScanAt,
		Tabs:          tt,
	}, nil
}
