package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Cache struct {
	baseDir string
}

type Entry struct {
	Key         string    `json:"key"`
	Data        []byte    `json:"data"`
	ExpiresAt   time.Time `json:"expires_at"`
	GeneratedAt time.Time `json:"generated_at"`
}

func New(baseDir string) *Cache {
	return &Cache{
		baseDir: baseDir,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	path := c.path(key)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		os.Remove(path)
		return nil, false
	}

	return entry.Data, true
}

func (c *Cache) Set(key string, data []byte, ttl time.Duration) error {
	entry := Entry{
		Key:         key,
		Data:        data,
		ExpiresAt:   time.Now().Add(ttl),
		GeneratedAt: time.Now(),
	}

	content, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling cache entry: %w", err)
	}

	dir := filepath.Join(c.baseDir, filepath.Dir(key))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}

	return os.WriteFile(c.path(key), content, 0644)
}

func (c *Cache) path(key string) string {
	return filepath.Join(c.baseDir, key+".json")
}

func Fingerprint(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func HardwareFingerprint(gpuName string, vramTotal int64, ramTotal int64, cpuCores int) string {
	input := fmt.Sprintf("%s|%d|%d|%d", gpuName, vramTotal, ramTotal, cpuCores)
	return Fingerprint(input)
}
