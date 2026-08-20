package calib

import (
	"encoding/json"
	"os"
	"sync"
)

// FileStore persists calibrations to JSON file on disk.
type FileStore struct {
	mu   sync.RWMutex
	path string
	data map[string]Calibration
}

// NewFileStore loads or creates file-backed store.
func NewFileStore(path string) (*FileStore, error) {
	fs := &FileStore{path: path, data: make(map[string]Calibration)}
	if err := fs.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return fs, nil
}

func (fs *FileStore) load() error {
	b, err := os.ReadFile(fs.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &fs.data)
}

// Save writes current data to disk.
func (fs *FileStore) Save() error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	b, err := json.MarshalIndent(fs.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.path, b, 0644)
}

// Get returns calibration for hook.
func (fs *FileStore) Get(hookID string) (Calibration, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	c, ok := fs.data[hookID]
	return c, ok
}

// Put stores calibration and persists.
func (fs *FileStore) Put(hookID string, c Calibration) error {
	if err := c.Validate(); err != nil {
		return err
	}
	fs.mu.Lock()
	fs.data[hookID] = c.Clone()
	fs.mu.Unlock()
	return fs.Save()
}

// ImportInto copies all entries into memory Store.
func (fs *FileStore) ImportInto(store *Store) error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	for id, c := range fs.data {
		if err := store.Put(id, c); err != nil {
			return err
		}
	}
	return nil
}

// ExportFrom copies memory Store to file store.
func ExportFrom(store *Store, fs *FileStore) error {
	snap := store.List()
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.data = snap
	return fs.Save()
}

// SeedDemo loads example H7 calibration into store.
func SeedDemo(store *Store) error {
	return store.Put("H7", Calibration{
		Tare:      1000,
		Span:      0.05,
		MaxLoadKg: 45000,
	})
}

// DemoHooks returns standard demo hook identifiers.
func DemoHooks() []string {
	return []string{"H1", "H2", "H3", "H4", "H5", "H6", "H7", "H8"}
}

// BulkSeed creates default calibrations for demo hooks.
func BulkSeed(store *Store, tare int64, span, maxLoad float64) error {
	for _, id := range DemoHooks() {
		if err := store.Put(id, Calibration{Tare: tare, Span: span, MaxLoadKg: maxLoad}); err != nil {
			return err
		}
	}
	return nil
}
