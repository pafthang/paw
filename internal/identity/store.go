package identity

import (
	"os"
	"path/filepath"

	"github.com/pafthang/paw/internal/config"
)

type Files struct {
	IdentityFile     string `json:"identity_file"`
	SoulFile         string `json:"soul_file"`
	StyleFile        string `json:"style_file"`
	InstructionsFile string `json:"instructions_file"`
	UserFile         string `json:"user_file"`
}

type SaveResponse struct {
	OK      bool     `json:"ok"`
	Updated []string `json:"updated"`
}

var fileMap = map[string]string{
	"identity_file":     "IDENTITY.md",
	"soul_file":         "SOUL.md",
	"style_file":        "STYLE.md",
	"instructions_file": "INSTRUCTIONS.md",
	"user_file":         "USER.md",
}

func Dir() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "identity"), nil
}

func Load() (Files, error) {
	dir, err := Dir()
	if err != nil {
		return Files{}, err
	}
	read := func(name string) string {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return ""
		}
		return string(data)
	}
	return Files{
		IdentityFile:     read("IDENTITY.md"),
		SoulFile:         read("SOUL.md"),
		StyleFile:        read("STYLE.md"),
		InstructionsFile: read("INSTRUCTIONS.md"),
		UserFile:         read("USER.md"),
	}, nil
}

func Save(files map[string]*string) (SaveResponse, error) {
	dir, err := Dir()
	if err != nil {
		return SaveResponse{}, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return SaveResponse{}, err
	}
	updated := make([]string, 0, len(files))
	for key, value := range files {
		if value == nil {
			continue
		}
		filename, ok := fileMap[key]
		if !ok {
			continue
		}
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte(*value), 0o600); err != nil {
			return SaveResponse{}, err
		}
		updated = append(updated, filename)
	}
	return SaveResponse{OK: true, Updated: updated}, nil
}
