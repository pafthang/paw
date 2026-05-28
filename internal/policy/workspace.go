package policy

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func ResolveWorkspace(workspace string) (string, error) {
	ws := strings.TrimSpace(workspace)
	if ws == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		ws = cwd
	}
	ws, err := filepath.Abs(ws)
	if err != nil {
		return "", err
	}
	ws, err = filepath.EvalSymlinks(ws)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(ws)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("workspace must be a directory")
	}
	return ws, nil
}

func ResolveInWorkspace(workspaceAbs string, userPath string) (string, error) {
	p := strings.TrimSpace(userPath)
	if p == "" {
		return "", errors.New("path is required")
	}
	if filepath.IsAbs(p) {
		p = filepath.Clean(p)
	} else {
		p = filepath.Clean(filepath.Join(workspaceAbs, p))
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	// Avoid requiring the target file to exist (writes may create it). Instead,
	// resolve symlinks for the closest existing parent directory.
	dir := abs
	for {
		info, statErr := os.Stat(dir)
		if statErr == nil && info.IsDir() {
			resolvedDir, err := filepath.EvalSymlinks(dir)
			if err != nil {
				return "", err
			}
			rel, err := filepath.Rel(dir, abs)
			if err != nil {
				return "", err
			}
			return filepath.Join(resolvedDir, rel), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return abs, nil
}
