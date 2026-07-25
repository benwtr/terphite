// Package dashboard persists named sets of saved graphs ("panels") so they
// can be viewed together as a grid instead of one at a time.
package dashboard

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Panel is a saved snapshot of a graph: the metrics it plots and the time
// range it was viewed over.
type Panel struct {
	Title    string   `json:"title"`
	Targets  []string `json:"targets"`
	TimeFrom string   `json:"timeFrom"`
}

// Dashboard is a named collection of panels.
type Dashboard struct {
	Name   string  `json:"name"`
	Panels []Panel `json:"panels"`
}

// Store persists dashboards as one JSON file per dashboard under Dir.
type Store struct {
	Dir string
}

// DefaultDir returns the standard on-disk location for saved dashboards:
// $XDG_CONFIG_HOME/terphite/dashboards (or the OS equivalent).
func DefaultDir() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, "terphite", "dashboards"), nil
}

// NewStore returns a Store rooted at dir.
func NewStore(dir string) *Store {
	return &Store{Dir: dir}
}

// sanitizeName maps a dashboard name to a safe filename component, so a
// user-supplied name can never escape the store directory.
func sanitizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("dashboard: name must not be empty")
	}
	var sb strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == ' ':
			sb.WriteRune(r)
		default:
			sb.WriteRune('_')
		}
	}
	return sb.String(), nil
}

func (s *Store) filePath(name string) (string, error) {
	safe, err := sanitizeName(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.Dir, safe+".json"), nil
}

func (s *Store) loadFile(path string) (*Dashboard, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d Dashboard
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// List returns the names of all saved dashboards, sorted alphabetically. An
// empty (not-yet-created) store directory returns an empty list, not an error.
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.Dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		d, err := s.loadFile(filepath.Join(s.Dir, e.Name()))
		if err != nil {
			continue
		}
		names = append(names, d.Name)
	}
	sort.Strings(names)
	return names, nil
}

// Load loads the dashboard with the given name.
func (s *Store) Load(name string) (*Dashboard, error) {
	path, err := s.filePath(name)
	if err != nil {
		return nil, err
	}
	return s.loadFile(path)
}

// Save persists d, creating the store directory and/or overwriting an
// existing file for the same name as needed.
func (s *Store) Save(d *Dashboard) error {
	path, err := s.filePath(d.Name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Delete removes the dashboard with the given name.
func (s *Store) Delete(name string) error {
	path, err := s.filePath(name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
