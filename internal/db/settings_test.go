package db

import "testing"

func TestSettingsGetSetDelete(t *testing.T) {
	handle := mustReopen(t, openTestDB(t))
	s := NewSettings(handle)

	if _, ok, err := s.Get("missing"); err != nil || ok {
		t.Fatalf("missing key: ok=%v err=%v", ok, err)
	}

	if err := s.Set("foo", "bar"); err != nil {
		t.Fatal(err)
	}
	raw, ok, err := s.Get("foo")
	if err != nil || !ok || len(raw) == 0 {
		t.Fatalf("get failed: raw=%q ok=%v err=%v", raw, ok, err)
	}

	// Verify round-trip via GetJSON.
	var got string
	if ok, err := s.GetJSON("foo", &got); err != nil || !ok || got != "bar" {
		t.Fatalf("round-trip: got=%q ok=%v err=%v", got, ok, err)
	}

	if err := s.Set("foo", "baz"); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.GetJSON("foo", &got); err != nil || !ok || got != "baz" {
		t.Fatalf("upsert failed: got=%q ok=%v err=%v", got, ok, err)
	}

	if err := s.Delete("foo"); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := s.Get("foo"); ok {
		t.Fatal("key still present after delete")
	}
}

func TestSettingsJSON(t *testing.T) {
	handle := mustReopen(t, openTestDB(t))
	s := NewSettings(handle)

	type config struct {
		Width    int  `json:"width"`
		DarkMode bool `json:"dark_mode"`
	}
	cfg := config{Width: 340, DarkMode: true}
	if err := s.Set("appearance", cfg); err != nil {
		t.Fatal(err)
	}

	var out config
	if ok, err := s.GetJSON("appearance", &out); err != nil || !ok {
		t.Fatalf("GetJSON: ok=%v err=%v", ok, err)
	}
	if out.Width != 340 || !out.DarkMode {
		t.Fatalf("got %+v", out)
	}

	// Overwrite with new values.
	cfg.Width = 500
	if err := s.Set("appearance", cfg); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.GetJSON("appearance", &out); err != nil || !ok || out.Width != 500 {
		t.Fatalf("overwrite: got %+v ok=%v err=%v", out, ok, err)
	}

	// Bad stored value surfaces as an error.
	if err := s.Set("bad", []byte("not json")); err != nil {
		t.Fatal(err)
	}
	var dummy any
	if _, err := s.GetJSON("bad", &dummy); err == nil {
		t.Fatal("expected parse error for invalid JSON")
	}
}
