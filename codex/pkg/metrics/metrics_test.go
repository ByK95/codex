package metrics

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntCounter(t *testing.T) {
	Default = NewRegistry()

	IncInt("requests")
	AddInt("requests", 4)
	assert.Equal(t, int64(5), GetInt("requests"))

	// Increment again
	IncInt("requests")
	assert.Equal(t, int64(6), GetInt("requests"))
}

func TestFloatCounter(t *testing.T) {
	Default = NewRegistry()

	AddFloat("latency", 2.5)
	AddFloat("latency", 0.5)
	assert.Equal(t, 3.0, GetFloat("latency"))
}

func TestBoolGauge(t *testing.T) {
	Default = NewRegistry()

	SetBool("ready", true)
	assert.True(t, GetBool("ready"))

	SetBool("ready", false)
	assert.False(t, GetBool("ready"))
}

func TestStringGauge(t *testing.T) {
	Default = NewRegistry()

	SetString("version", "1.0.0")
	assert.Equal(t, "1.0.0", GetString("version"))

	SetString("version", "1.0.1")
	assert.Equal(t, "1.0.1", GetString("version"))
}

func TestSnapshotJSON(t *testing.T) {
	Default = NewRegistry()

	AddInt("requests", 10)
	AddFloat("latency", 3.14)
	SetBool("ready", true)
	SetString("version", "2.1.0")

	jsonStr := SnapshotJSON()
	assert.NotEmpty(t, jsonStr)

	// Parse back to check values
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &parsed)
	assert.NoError(t, err)

	assert.EqualValues(t, 10, parsed["requests"])
	assert.EqualValues(t, 3.14, parsed["latency"])
	assert.EqualValues(t, true, parsed["ready"])
	assert.EqualValues(t, "2.1.0", parsed["version"])
}
