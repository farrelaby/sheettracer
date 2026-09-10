package db

import (
	"testing"
	"time"
)

func TestScanRunLifecycle(t *testing.T) {
	path := openTestDB(t)
	handle := mustReopen(t, path)
	runs := NewScanRunRepo(handle)

	start := time.Now().UTC().Truncate(time.Second)
	id, err := runs.Add(TriggerManual, start)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if id == 0 {
		t.Fatal("Add returned id 0")
	}

	var run ScanRun
	if err := handle.Get(&run, `SELECT * FROM scan_runs WHERE id = ?`, id); err != nil {
		t.Fatalf("select run: %v", err)
	}
	if run.TriggeredBy != TriggerManual || run.Status != ScanStatusRunning {
		t.Fatalf("got %+v, want manual/running", run)
	}
	if run.FinishedAt != nil {
		t.Fatalf("FinishedAt = %v, want NULL", run.FinishedAt)
	}

	finish := start.Add(time.Minute)
	if err := runs.Update(id, finish, ScanStatusOK, 3, 1, nil); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := handle.Get(&run, `SELECT * FROM scan_runs WHERE id = ?`, id); err != nil {
		t.Fatalf("select finished run: %v", err)
	}
	if run.Status != ScanStatusOK || run.TabsScanned != 3 || run.TabsSkipped != 1 {
		t.Fatalf("got %+v, want ok/3/1", run)
	}
	if run.FinishedAt == nil || !run.FinishedAt.Equal(finish) {
		t.Fatalf("FinishedAt = %v, want %v", run.FinishedAt, finish)
	}

	msg := "boom"
	if err := runs.Update(id, finish, ScanStatusError, 0, 0, &msg); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if err := handle.Get(&run, `SELECT * FROM scan_runs WHERE id = ?`, id); err != nil {
		t.Fatalf("select errored run: %v", err)
	}
	if run.Error == nil || *run.Error != msg {
		t.Fatalf("Error = %v, want %q", run.Error, msg)
	}
}

func TestUpdateLastScanAt(t *testing.T) {
	path := openTestDB(t)
	handle := mustReopen(t, path)
	spreadsheets := NewSpreadsheetRepo(handle)

	id, err := spreadsheets.UpsertByGoogleID(&Spreadsheet{
		GoogleID:   "abc123",
		Title:      "Budget",
		URL:        "https://docs.google.com/spreadsheets/d/abc123",
		Visibility: "private",
	})
	if err != nil {
		t.Fatalf("UpsertByGoogleID: %v", err)
	}

	got, err := spreadsheets.GetByGoogleID("abc123")
	if err != nil {
		t.Fatalf("GetByGoogleID: %v", err)
	}
	if got.LastScanAt != nil {
		t.Fatalf("LastScanAt = %v, want NULL before first scan", got.LastScanAt)
	}

	stamp := time.Now().UTC().Truncate(time.Second)
	if err := spreadsheets.UpdateLastScanAt(id, stamp); err != nil {
		t.Fatalf("UpdateLastScanAt: %v", err)
	}
	got, err = spreadsheets.GetByGoogleID("abc123")
	if err != nil {
		t.Fatalf("GetByGoogleID after stamp: %v", err)
	}
	if got.LastScanAt == nil || !got.LastScanAt.Equal(stamp) {
		t.Fatalf("LastScanAt = %v, want %v", got.LastScanAt, stamp)
	}
}
