package skills

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pafthang/paw/internal/config"
)

type LoadError struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

type LoadReport struct {
	Root   string      `json:"root"`
	Skills []Skill     `json:"skills"`
	Errors []LoadError `json:"errors,omitempty"`
}

func SkillsRoot() (string, error) {
	return config.SkillsDir()
}

func LoadAll() (LoadReport, error) {
	root, err := SkillsRoot()
	if err != nil {
		return LoadReport{}, err
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return LoadReport{}, err
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return LoadReport{}, err
	}

	report := LoadReport{Root: root}
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		name := ent.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		path := filepath.Join(root, name, "skill.yaml")
		s, err := LoadFromFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			report.Errors = append(report.Errors, LoadError{Path: path, Error: err.Error()})
			continue
		}
		report.Skills = append(report.Skills, s)
	}

	sort.Slice(report.Skills, func(i, j int) bool {
		return strings.ToLower(report.Skills[i].Name) < strings.ToLower(report.Skills[j].Name)
	})
	return report, nil
}

func LoadByName(name string) (Skill, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Skill{}, errors.New("name is required")
	}
	root, err := SkillsRoot()
	if err != nil {
		return Skill{}, err
	}
	if strings.Contains(name, string(os.PathSeparator)) {
		return Skill{}, fmt.Errorf("invalid name %q", name)
	}
	return LoadFromFile(filepath.Join(root, name, "skill.yaml"))
}
