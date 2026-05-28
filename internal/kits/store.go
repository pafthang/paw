package kits

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pafthang/paw/internal/config"
	"gopkg.in/yaml.v3"
)

type InstalledKit struct {
	ID          string         `json:"id"`
	Config      map[string]any `json:"config"`
	UserValues  map[string]any `json:"user_values"`
	InstalledAt string         `json:"installed_at"`
	Active      bool           `json:"active"`
}

type CatalogEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Category    string   `json:"category"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	Preview     string   `json:"preview,omitempty"`
	Installed   bool     `json:"installed"`
}

type Store struct {
	baseDir string
}

func NewStore() (*Store, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}
	base := filepath.Join(dir, "kits")
	if err := os.MkdirAll(base, 0o700); err != nil {
		return nil, err
	}
	return &Store{baseDir: base}, nil
}

func (s *Store) List() ([]InstalledKit, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []InstalledKit{}, nil
		}
		return nil, err
	}
	kits := make([]InstalledKit, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		kit, err := s.Get(entry.Name())
		if err != nil {
			continue
		}
		kits = append(kits, *kit)
	}
	sort.Slice(kits, func(i, j int) bool {
		return strings.ToLower(kitName(kits[i].Config)) < strings.ToLower(kitName(kits[j].Config))
	})
	return kits, nil
}

func (s *Store) Get(id string) (*InstalledKit, error) {
	id = slugify(id)
	if id == "" {
		return nil, os.ErrNotExist
	}
	kitDir := filepath.Join(s.baseDir, id)
	yamlPath := filepath.Join(kitDir, "pawkit.yaml")
	metaPath := filepath.Join(kitDir, "meta.json")
	yamlData, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}
	var cfg map[string]any
	if err := yaml.Unmarshal(yamlData, &cfg); err != nil {
		return nil, err
	}
	kit := &InstalledKit{
		ID:         id,
		Config:     cfg,
		UserValues: map[string]any{},
	}
	if metaData, err := os.ReadFile(metaPath); err == nil {
		var meta struct {
			UserValues  map[string]any `json:"user_values"`
			InstalledAt string         `json:"installed_at"`
			Active      bool           `json:"active"`
		}
		if json.Unmarshal(metaData, &meta) == nil {
			kit.UserValues = meta.UserValues
			kit.InstalledAt = meta.InstalledAt
			kit.Active = meta.Active
		}
	}
	if kit.UserValues == nil {
		kit.UserValues = map[string]any{}
	}
	return kit, nil
}

func (s *Store) Install(yamlText string, id string) (*InstalledKit, error) {
	var cfg map[string]any
	if err := yaml.Unmarshal([]byte(yamlText), &cfg); err != nil {
		return nil, err
	}
	if id == "" {
		id = slugify(kitName(cfg))
	} else {
		id = slugify(id)
	}
	if id == "" {
		return nil, fmt.Errorf("kit id must contain at least one alphanumeric character")
	}
	kit := &InstalledKit{
		ID:          id,
		Config:      cfg,
		UserValues:  map[string]any{},
		InstalledAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := s.save(kit, yamlText); err != nil {
		return nil, err
	}
	return kit, nil
}

func (s *Store) Remove(id string) error {
	id = slugify(id)
	if id == "" {
		return os.ErrNotExist
	}
	return os.RemoveAll(filepath.Join(s.baseDir, id))
}

func (s *Store) Activate(id string) error {
	id = slugify(id)
	if id == "" {
		return os.ErrNotExist
	}
	kits, err := s.List()
	if err != nil {
		return err
	}
	found := false
	for _, kit := range kits {
		kit.Active = kit.ID == id
		if kit.Active {
			found = true
		}
		if err := s.saveMeta(&kit); err != nil {
			return err
		}
	}
	if !found {
		return os.ErrNotExist
	}
	return nil
}

func (s *Store) Data(id string) (map[string]any, error) {
	id = slugify(id)
	if id == "" {
		return nil, os.ErrNotExist
	}
	dataDir := filepath.Join(s.baseDir, id, "data")
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	out := map[string]any{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dataDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var value any
		if json.Unmarshal(data, &value) != nil {
			continue
		}
		key := strings.TrimSuffix(entry.Name(), ".json")
		key = strings.ReplaceAll(key, "_", ":")
		out[key] = value
	}
	return out, nil
}

func (s *Store) Catalog() ([]CatalogEntry, error) {
	kits, err := s.List()
	if err != nil {
		return nil, err
	}
	installed := map[string]bool{}
	for _, kit := range kits {
		installed[kit.ID] = true
	}
	return []CatalogEntry{
		{
			ID:          "mission-control",
			Name:        "Mission Control",
			Description: "Coordinate agents, tasks, notifications, and documents.",
			Icon:        "layout-dashboard",
			Category:    "productivity",
			Author:      "PocketPaw",
			Tags:        []string{"agents", "tasks", "documents"},
			Installed:   installed["mission-control"],
		},
	}, nil
}

func (s *Store) InstallCatalog(id string) (*InstalledKit, error) {
	id = slugify(id)
	if id != "mission-control" {
		return nil, os.ErrNotExist
	}
	return s.Install(missionControlYAML, id)
}

func (s *Store) save(kit *InstalledKit, yamlText string) error {
	kitDir := filepath.Join(s.baseDir, kit.ID)
	if err := os.MkdirAll(filepath.Join(kitDir, "data"), 0o700); err != nil {
		return err
	}
	yamlPath := filepath.Join(kitDir, "pawkit.yaml")
	if yamlText == "" {
		data, err := yaml.Marshal(kit.Config)
		if err != nil {
			return err
		}
		yamlText = string(data)
	}
	if err := os.WriteFile(yamlPath, []byte(yamlText), 0o600); err != nil {
		return err
	}
	return s.saveMeta(kit)
}

func (s *Store) saveMeta(kit *InstalledKit) error {
	kitDir := filepath.Join(s.baseDir, kit.ID)
	if err := os.MkdirAll(filepath.Join(kitDir, "data"), 0o700); err != nil {
		return err
	}
	meta := map[string]any{
		"user_values":  kit.UserValues,
		"installed_at": kit.InstalledAt,
		"active":       kit.Active,
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(kitDir, "meta.json"), append(data, '\n'), 0o600)
}

func kitName(cfg map[string]any) string {
	meta, _ := cfg["meta"].(map[string]any)
	name, _ := meta["name"].(string)
	return name
}

var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = slugPattern.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

const missionControlYAML = `meta:
  name: Mission Control
  author: PocketPaw
  version: "1.0.0"
  description: Coordinate agents, tasks, notifications, and documents.
  category: productivity
  tags: [agents, tasks, documents]
  icon: layout-dashboard
  built_in: true
layout:
  columns: 2
  sections:
    - title: Overview
      span: full
      panels:
        - id: stats
          type: metrics-row
          metrics:
            - label: Agents
              source: api:stats
              field: agents.total
            - label: Tasks
              source: api:stats
              field: tasks.total
            - label: Documents
              source: api:stats
              field: documents.total
    - title: Tasks
      span: left
      panels:
        - id: tasks
          type: kanban
          source: api:tasks
    - title: Activity
      span: right
      panels:
        - id: activity
          type: feed
          source: api:activities
workflows: {}
`
