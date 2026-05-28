package skills

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func LoadFromFile(path string) (Skill, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Skill{}, errors.New("path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, err
	}
	var s Skill
	if err := yaml.Unmarshal(data, &s); err != nil {
		return Skill{}, fmt.Errorf("invalid YAML: %w", err)
	}
	if err := Validate(s); err != nil {
		return Skill{}, err
	}
	return s, nil
}

func DefaultSkillPath(skillName string) (string, error) {
	name := strings.TrimSpace(skillName)
	if name == "" {
		return "", errors.New("skill name is required")
	}
	if strings.Contains(name, string(os.PathSeparator)) {
		return "", errors.New("invalid skill name")
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".pocketpaw", "skills", name, "skill.yaml"), nil
}

func Validate(s Skill) error {
	if strings.TrimSpace(s.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(s.Description) == "" {
		return errors.New("description is required")
	}
	if strings.TrimSpace(s.Version) == "" {
		return errors.New("version is required")
	}
	if strings.TrimSpace(s.Prompts.System) == "" && len(s.Commands) == 0 {
		return errors.New("prompts.system or commands is required")
	}
	for i, c := range s.Commands {
		if strings.TrimSpace(c.Name) == "" {
			return fmt.Errorf("commands[%d].name is required", i)
		}
		if strings.TrimSpace(c.Tool) == "" {
			return fmt.Errorf("commands[%d].tool is required", i)
		}
	}
	return nil
}
