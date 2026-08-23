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
	v, ok, err := s.Get("foo")
	if err != nil || !ok || v != "bar" {
		t.Fatalf("got %q ok=%v err=%v", v, ok, err)
	}

	if err := s.Set("foo", "baz"); err != nil {
		t.Fatal(err)
	}
	if v, _, _ := s.Get("foo"); v != "baz" {
		t.Fatalf("upsert failed: %q", v)
	}

	if err := s.Delete("foo"); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := s.Get("foo"); ok {
		t.Fatal("key still present after delete")
	}
}

func TestSettingsTypedHelpers(t *testing.T) {
	handle := mustReopen(t, openTestDB(t))
	s := NewSettings(handle)

	if err := s.SetBool("flag", true); err != nil {
		t.Fatal(err)
	}
	if b, ok, err := s.GetBool("flag"); err != nil || !ok || !b {
		t.Fatalf("bool: got %v ok=%v err=%v", b, ok, err)
	}

	if err := s.SetInt("width", 340); err != nil {
		t.Fatal(err)
	}
	if n, ok, err := s.GetInt("width"); err != nil || !ok || n != 340 {
		t.Fatalf("int: got %d ok=%v err=%v", n, ok, err)
	}

	type nested struct{ A int `json:"a"` }
	if err := s.SetJSON("obj", nested{A: 7}); err != nil {
		t.Fatal(err)
	}
	var out nested
	if ok, err := s.GetJSON("obj", &out); err != nil || !ok || out.A != 7 {
		t.Fatalf("json: got %+v ok=%v err=%v", out, ok, err)
	}

	// Bad stored value surfaces as a typed error, key still exists.
	if err := s.Set("flag", "not-a-bool"); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := s.GetBool("flag"); err == nil || !ok {
		t.Fatalf("expected parse error, got ok=%v err=%v", ok, err)
	}
}