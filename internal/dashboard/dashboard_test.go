package dashboard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	s := NewStore(t.TempDir())
	d := &Dashboard{
		Name: "prod overview",
		Panels: []Panel{
			{Title: "cpu", Targets: []string{"stats.cpu"}, TimeFrom: "-1h", DrawMode: "stacked"},
			{Title: "mem", Targets: []string{"stats.mem"}, TimeFrom: "-1h"},
		},
	}
	if err := s.Save(d); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load("prod overview")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != d.Name || len(got.Panels) != 2 {
		t.Errorf("got %+v, want %+v", got, d)
	}
	if got.Panels[0].Title != "cpu" || got.Panels[1].Title != "mem" {
		t.Errorf("panels not preserved: %+v", got.Panels)
	}
	if got.Panels[0].DrawMode != "stacked" {
		t.Errorf("Panels[0].DrawMode = %q, want stacked", got.Panels[0].DrawMode)
	}
	if got.Panels[1].DrawMode != "" {
		t.Errorf("Panels[1].DrawMode = %q, want empty (defaults to line)", got.Panels[1].DrawMode)
	}
}

func TestListEmptyDir(t *testing.T) {
	s := NewStore(filepath.Join(t.TempDir(), "does-not-exist-yet"))
	names, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Errorf("got %v, want empty", names)
	}
}

func TestListSorted(t *testing.T) {
	s := NewStore(t.TempDir())
	for _, name := range []string{"zebra", "alpha", "mango"} {
		if err := s.Save(&Dashboard{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	names, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"alpha", "mango", "zebra"}
	if len(names) != len(want) {
		t.Fatalf("got %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("got %v, want %v", names, want)
			break
		}
	}
}

func TestDelete(t *testing.T) {
	s := NewStore(t.TempDir())
	if err := s.Save(&Dashboard{Name: "temp"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("temp"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load("temp"); err == nil {
		t.Error("expected error loading deleted dashboard")
	}
}

func TestSaveRejectsEmptyName(t *testing.T) {
	s := NewStore(t.TempDir())
	if err := s.Save(&Dashboard{Name: "   "}); err == nil {
		t.Error("expected error for empty/whitespace name")
	}
}

func TestNameSanitizedAgainstPathTraversal(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	if err := s.Save(&Dashboard{Name: "../../etc/passwd"}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly one file written inside store dir, got %d", len(entries))
	}
	for _, e := range entries {
		if filepath.Dir(filepath.Join(dir, e.Name())) != dir {
			t.Errorf("file escaped store dir: %s", e.Name())
		}
	}
}
