package services

import (
	"context"
	"fmt"

	"sheettracer/internal/db"
	"sheettracer/internal/oauth"
	"sheettracer/internal/sheets"

	"golang.org/x/oauth2"
)

// SheetsService exposes spreadsheet management to the frontend. It is a
// Wails-bound service.
type SheetsService struct {
	client         *oauth.Client
	store          oauth.TokenStore
	spreadsheetRepo *db.SpreadsheetRepo
	sheetRepo      *db.SheetRepo
}

func NewSheetsService(
	client *oauth.Client,
	store oauth.TokenStore,
	spreadsheetRepo *db.SpreadsheetRepo,
	sheetRepo *db.SheetRepo,
) *SheetsService {
	return &SheetsService{
		client:          client,
		store:           store,
		spreadsheetRepo: spreadsheetRepo,
		sheetRepo:       sheetRepo,
	}
}

// Add fetches metadata for the given Google Sheets URL, persists it, and
// returns the saved spreadsheet. Returns an error with "no access / not found"
// if Drive returns 404.
func (s *SheetsService) Add(url string, notes string) (*db.Spreadsheet, error) {
	spreadsheetID, err := sheets.ParseID(url)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	src, err := s.tokenSource()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	client, err := sheets.NewClient(ctx, src)
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
		return nil, fmt.Errorf("persist spreadsheet: %w", err)
	}
	sp.ID = id

	// Persist tabs.
	tabs := make([]db.Sheet, len(meta.Tabs))
	for i, t := range meta.Tabs {
		tabs[i] = db.Sheet{
			SpreadsheetID: id,
			TabID:         t.TabID,
			Title:         t.Title,
			Idx:           t.Idx,
		}
	}
	if err := s.sheetRepo.UpsertAll(id, tabs); err != nil {
		return nil, fmt.Errorf("persist tabs: %w", err)
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

// tokenSource returns a fresh OAuth2 token source, or an error if not connected.
func (s *SheetsService) tokenSource() (oauth2.TokenSource, error) {
	tok, err := s.store.Get()
	if err != nil {
		return nil, err
	}
	if tok == nil {
		return nil, fmt.Errorf("not connected")
	}
	return s.client.Source(context.Background(), tok, s.store), nil
}

// FetchMetadata is a thin wrapper for testing — calls the Google API directly.
func (s *SheetsService) FetchMetadata(spreadsheetID string) (*sheets.Metadata, error) {
	src, err := s.tokenSource()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	client, err := sheets.NewClient(ctx, src)
	if err != nil {
		return nil, err
	}
	return client.FetchMetadata(ctx, spreadsheetID)
}
