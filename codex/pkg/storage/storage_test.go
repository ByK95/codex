package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReloadAll(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "data.json")

	// Prepare storage manager
	SetStorageManagerPath(filename)
	sm := SM()

	var loadedValue string
	sm.BindFuncs("testKey",
		func(raw json.RawMessage) error {
			return json.Unmarshal(raw, &loadedValue)
		},
		nil,
	)

	// Write initial file with testKey
	initial := map[string]any{"testKey": "hello"}
	data, _ := json.Marshal(initial)
	assert.NoError(t, os.WriteFile(filename, data, 0644))

	// Reload
	assert.NoError(t, sm.ReloadAll())
	assert.Equal(t, "hello", loadedValue)

	// Overwrite file with different value
	updated := map[string]any{"testKey": "world"}
	data, _ = json.Marshal(updated)
	assert.NoError(t, os.WriteFile(filename, data, 0644))

	// Reload again
	assert.NoError(t, sm.ReloadAll())
	assert.Equal(t, "world", loadedValue)
}
