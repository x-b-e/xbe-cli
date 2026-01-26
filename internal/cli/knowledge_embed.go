package cli

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "embed"
)

const embeddedKnowledgeDBName = "knowledge.sqlite"

//go:embed knowledge_db/knowledge.sqlite
var embeddedKnowledgeDB []byte

var (
	embeddedKnowledgeDBOnce sync.Once
	embeddedKnowledgeDBPath string
	embeddedKnowledgeDBErr  error
)

func ensureEmbeddedKnowledgeDB() (string, error) {
	embeddedKnowledgeDBOnce.Do(func() {
		if len(embeddedKnowledgeDB) == 0 {
			embeddedKnowledgeDBErr = errors.New("embedded knowledge database missing")
			return
		}
		cacheDir, err := os.UserCacheDir()
		if err != nil || cacheDir == "" {
			cacheDir = os.TempDir()
		}
		dir := filepath.Join(cacheDir, "xbe", "knowledge")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			embeddedKnowledgeDBErr = fmt.Errorf("prepare knowledge cache: %w", err)
			return
		}

		path := filepath.Join(dir, embeddedKnowledgeDBName)
		hash := fmt.Sprintf("%x", sha256.Sum256(embeddedKnowledgeDB))
		hashPath := path + ".sha256"

		if fileUpToDate(path, hashPath, hash, len(embeddedKnowledgeDB)) {
			embeddedKnowledgeDBPath = path
			return
		}

		tmpPath := path + ".tmp"
		if err := os.WriteFile(tmpPath, embeddedKnowledgeDB, 0o644); err != nil {
			embeddedKnowledgeDBErr = fmt.Errorf("write embedded knowledge database: %w", err)
			return
		}
		if err := os.WriteFile(hashPath, []byte(hash), 0o644); err != nil {
			_ = os.Remove(tmpPath)
			embeddedKnowledgeDBErr = fmt.Errorf("write knowledge database hash: %w", err)
			return
		}
		if err := os.Rename(tmpPath, path); err != nil {
			_ = os.Remove(tmpPath)
			embeddedKnowledgeDBErr = fmt.Errorf("activate embedded knowledge database: %w", err)
			return
		}
		embeddedKnowledgeDBPath = path
	})
	return embeddedKnowledgeDBPath, embeddedKnowledgeDBErr
}

func fileUpToDate(path, hashPath, expectedHash string, expectedSize int) bool {
	info, err := os.Stat(path)
	if err != nil || info.Size() != int64(expectedSize) {
		return false
	}
	data, err := os.ReadFile(hashPath)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == expectedHash
}
