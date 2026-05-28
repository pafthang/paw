package filestore

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pafthang/paw/internal/config"
	"github.com/pafthang/paw/internal/db"
	"gorm.io/gorm"
)

func AddFromPath(database *gorm.DB, srcPath string, metadata string) (*db.FileRecord, error) {
	if database == nil {
		return nil, errors.New("database is required")
	}
	srcPath = strings.TrimSpace(srcPath)
	if srcPath == "" {
		return nil, errors.New("path is required")
	}
	if strings.HasPrefix(srcPath, "~") {
		return nil, fmt.Errorf("path %q is not allowed", srcPath)
	}
	abs, err := filepath.Abs(srcPath)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path %q is a directory", srcPath)
	}

	f, err := os.Open(abs)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	hasher := sha256.New()
	head := make([]byte, 512)
	n, _ := io.ReadFull(f, head)
	head = head[:n]
	mime := http.DetectContentType(head)
	if _, err := hasher.Write(head); err != nil {
		return nil, err
	}
	if _, err := io.Copy(hasher, f); err != nil {
		return nil, err
	}
	sum := hex.EncodeToString(hasher.Sum(nil))

	filesDir, err := config.FilesDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filesDir, 0o700); err != nil {
		return nil, err
	}

	base := filepath.Base(abs)
	base = sanitizeName(base)
	if base == "" {
		base = "file"
	}
	destName := fmt.Sprintf("%s_%s", sum[:16], base)
	destPath := filepath.Join(filesDir, destName)

	if err := copyFile(abs, destPath, 0o600); err != nil {
		return nil, err
	}

	rec := db.FileRecord{
		Path:      destName,
		Name:      base,
		MimeType:  mime,
		SizeBytes: info.Size(),
		Sha256:    sum,
		Metadata:  strings.TrimSpace(metadata),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := database.Create(&rec).Error; err != nil {
		_ = os.Remove(destPath)
		return nil, err
	}
	return &rec, nil
}

func Get(database *gorm.DB, id uint) (*db.FileRecord, error) {
	if database == nil {
		return nil, errors.New("database is required")
	}
	if id == 0 {
		return nil, errors.New("id is required")
	}
	var rec db.FileRecord
	if err := database.First(&rec, id).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func Delete(database *gorm.DB, id uint) error {
	if database == nil {
		return errors.New("database is required")
	}
	if id == 0 {
		return errors.New("id is required")
	}
	rec, err := Get(database, id)
	if err != nil {
		return err
	}
	filesDir, err := config.FilesDir()
	if err != nil {
		return err
	}
	target := filepath.Join(filesDir, filepath.Clean(rec.Path))
	_ = os.Remove(target)
	return database.Delete(&db.FileRecord{}, id).Error
}

func List(database *gorm.DB, limit int) ([]db.FileRecord, error) {
	if database == nil {
		return nil, errors.New("database is required")
	}
	if limit <= 0 {
		limit = 50
	}
	var items []db.FileRecord
	if err := database.Order("created_at desc, id desc").Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func Search(database *gorm.DB, query string, limit int) ([]db.FileRecord, error) {
	if database == nil {
		return nil, errors.New("database is required")
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query is required")
	}
	if limit <= 0 {
		limit = 50
	}
	like := "%" + query + "%"
	var items []db.FileRecord
	if err := database.
		Where("name LIKE ? OR path LIKE ? OR metadata LIKE ? OR sha256 LIKE ?", like, like, like, like).
		Order("created_at desc, id desc").
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, string(os.PathSeparator), "_")
	name = strings.ReplaceAll(name, "..", "_")
	return name
}

func copyFile(src, dest string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
