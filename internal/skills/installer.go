package skills

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type InstallOptions struct {
	Force bool
}

type UninstallOptions struct {
	Yes bool
}

func InstallFromDir(srcDir string, opts InstallOptions) (Skill, error) {
	srcDir = strings.TrimSpace(srcDir)
	if srcDir == "" {
		return Skill{}, errors.New("path is required")
	}
	info, err := os.Stat(srcDir)
	if err != nil {
		return Skill{}, err
	}
	if !info.IsDir() {
		return Skill{}, fmt.Errorf("path %q is not a directory", srcDir)
	}
	srcSkillPath := filepath.Join(srcDir, "skill.yaml")
	s, err := LoadFromFile(srcSkillPath)
	if err != nil {
		return Skill{}, err
	}
	root, err := SkillsRoot()
	if err != nil {
		return Skill{}, err
	}
	destDir := filepath.Join(root, s.Name)
	if _, err := os.Stat(destDir); err == nil && !opts.Force {
		return Skill{}, fmt.Errorf("skill %q already exists (use --force to overwrite)", s.Name)
	}
	if opts.Force {
		_ = os.RemoveAll(destDir)
	}
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return Skill{}, err
	}
	if err := copyDir(srcDir, destDir); err != nil {
		_ = os.RemoveAll(destDir)
		return Skill{}, err
	}
	// validate installed copy
	_, err = LoadFromFile(filepath.Join(destDir, "skill.yaml"))
	if err != nil {
		_ = os.RemoveAll(destDir)
		return Skill{}, err
	}
	return s, nil
}

func Uninstall(name string, opts UninstallOptions) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name is required")
	}
	if !opts.Yes {
		return errors.New("refusing to uninstall without --yes")
	}
	root, err := SkillsRoot()
	if err != nil {
		return err
	}
	if strings.Contains(name, string(os.PathSeparator)) {
		return fmt.Errorf("invalid name %q", name)
	}
	return os.RemoveAll(filepath.Join(root, name))
}

func copyDir(src, dest string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		srcInfo, err := d.Info()
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return err
		}
		return out.Close()
	})
}
