package sheets

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// spreadsheetIDPattern matches the spreadsheet ID in a Google Sheets URL.
// Handles formats: /d/{id}/, /d/{id}/edit, /d/{id}/edit#gid=0, etc.
var spreadsheetIDPattern = regexp.MustCompile(`/d/([a-zA-Z0-9_-]+)`)

// ParseID extracts the spreadsheet ID from a Google Sheets URL.
// Returns an error if the URL doesn't contain a valid spreadsheet ID.
func ParseID(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty URL")
	}

	// If it looks like just an ID (no slashes, no dots), return as-is.
	if !strings.Contains(raw, "/") && !strings.Contains(raw, ".") {
		return raw, nil
	}

	// Strip fragment (everything after #) before parsing.
	if i := strings.IndexByte(raw, '#'); i >= 0 {
		raw = raw[:i]
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	matches := spreadsheetIDPattern.FindStringSubmatch(u.Path)
	if len(matches) < 2 {
		return "", fmt.Errorf("no spreadsheet ID found in URL")
	}
	return matches[1], nil
}

// Tab represents a sheet tab within a spreadsheet.
type Tab struct {
	SheetID int64
	Title   string
	Idx     int
}

// Metadata holds the result of spreadsheets.get + drive.files.get.
type Metadata struct {
	SpreadsheetID string
	Title         string
	Tabs          []Tab
	Version       string
	ModifiedTime  string
	Visibility    string // "public" | "link-only" | "private" | "unknown"
}

// Client wraps the Google Sheets and Drive APIs.
type Client struct {
	sheetsSvc *sheets.Service
	driveSvc  *drive.Service
}

// NewClient creates a new API client from an OAuth2 token source.
func NewClient(ctx context.Context, src oauth2.TokenSource) (*Client, error) {
	sheetsSvc, err := sheets.NewService(ctx, option.WithTokenSource(src))
	if err != nil {
		return nil, fmt.Errorf("sheets client: %w", err)
	}
	driveSvc, err := drive.NewService(ctx, option.WithTokenSource(src))
	if err != nil {
		return nil, fmt.Errorf("drive client: %w", err)
	}
	return &Client{sheetsSvc: sheetsSvc, driveSvc: driveSvc}, nil
}

// FetchMetadata calls spreadsheets.get for title + tabs, and drive.files.get
// for version, modifiedTime, and permissions (visibility).
// Returns ErrNotFound if Drive returns 404 (no access or doesn't exist).
func (c *Client) FetchMetadata(ctx context.Context, spreadsheetID string) (*Metadata, error) {
	meta := &Metadata{SpreadsheetID: spreadsheetID}

	// spreadsheets.get — title + sheet tabs.
	sp, err := c.sheetsSvc.Spreadsheets.Get(spreadsheetID).
		Fields("properties.title,sheets.properties").
		Context(ctx).
		Do()
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("spreadsheets.get: %w", err)
	}
	meta.Title = sp.Properties.Title
	for i, sh := range sp.Sheets {
		meta.Tabs = append(meta.Tabs, Tab{
			SheetID: sh.Properties.SheetId,
			Title:   sh.Properties.Title,
			Idx:     i,
		})
	}

	// drive.files.get — version, modifiedTime, shared, capabilities.
	f, err := c.driveSvc.Files.Get(spreadsheetID).
		SupportsAllDrives(true).
		Fields("version,modifiedTime,shared,capabilities(canComment,canEdit)").
		Context(ctx).
		Do()
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("drive.files.get: %w", err)
	}
	meta.Version = fmt.Sprintf("%d", f.Version)
	meta.ModifiedTime = f.ModifiedTime
	meta.Visibility = classifyVisibility(f.Shared, f.Capabilities.CanEdit)

	return meta, nil
}

// classifyVisibility derives visibility from shared + capabilities.
func classifyVisibility(shared bool, canEdit bool) string {
	if !shared {
		return "private"
	}
	if canEdit {
		return "editor"
	}
	return "shared"
}

// isNotFound checks if an error is a Google API 404.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	// Google API errors contain "googleapi: Error 404" or "notFound".
	return strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "notFound")
}

// ErrNotFound is returned when Drive returns 404 (no access or doesn't exist).
var ErrNotFound = fmt.Errorf("no access / not found")

func (c *Client) BatchGet(ctx context.Context, spreadsheetID string, tabName string) (*sheets.BatchGetValuesResponse, error) {
	quotedTab := "'" + tabName + "'"
	return c.sheetsSvc.Spreadsheets.Values.BatchGet(spreadsheetID).Ranges(quotedTab).ValueRenderOption("FORMULA").Context(ctx).Do()
}
