package admins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile_Missing(t *testing.T) {
	l, err := LoadFromFile("/no/such/file.json")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if l == nil || l.Count() != 0 {
		t.Errorf("expected empty list for missing file, got %v", l)
	}
}

func TestLoadFromFile_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "admins.json")
	contents := []byte(`{"admins": ["admin@hse.ru", "  TEACHER@Domain.RU  ", ""]}`)
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}

	l, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile: %v", err)
	}
	if l.Count() != 2 {
		t.Errorf("expected 2 admins, got %d", l.Count())
	}
	cases := map[string]bool{
		"admin@hse.ru":        true,
		"ADMIN@HSE.RU":        true,
		" Admin@hse.ru ":      true,
		"teacher@domain.ru":   true,
		"unknown@example.com": false,
		"":                    false,
	}
	for email, want := range cases {
		if got := l.IsAdmin(email); got != want {
			t.Errorf("IsAdmin(%q) = %v, want %v", email, got, want)
		}
	}
}

func TestLoadFromFile_Malformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.json")
	_ = os.WriteFile(path, []byte("not json"), 0o644)
	if _, err := LoadFromFile(path); err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestNilSafety(t *testing.T) {
	var l *List
	if l.IsAdmin("anyone") {
		t.Error("nil List must answer false")
	}
	if l.Count() != 0 {
		t.Error("nil List must report 0 count")
	}
}
