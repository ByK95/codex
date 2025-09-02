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

	jsonStr, err := SnapshotJSON()
	assert.NotEmpty(t, jsonStr)
	assert.NoError(t, err)

	// Parse back to check values
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonStr.(json.RawMessage), &parsed)
	assert.NoError(t, err)

	assert.EqualValues(t, 10, parsed["requests"])
	assert.EqualValues(t, 3.14, parsed["latency"])
	assert.EqualValues(t, true, parsed["ready"])
	assert.EqualValues(t, "2.1.0", parsed["version"])
}

func TestClearAll(t *testing.T) {
	ClearAll()
	AddInt("test_int", 5)
	AddFloat("test_float", 2.34)
	SetBool("test_bool", true)
	SetString("test_str", "abc")

	assert.NotEmpty(t, Default.objects, "Registry should have metrics")

	ClearAll()
	assert.Empty(t, Default.objects, "Registry should be empty after ClearAll")
}

func TestClearPrefix(t *testing.T) {
	ClearAll()
	AddInt("foo_count", 10)
	AddFloat("foo_time", 1.23)
	AddInt("bar_count", 20)

	ClearPrefix("foo_")

	_, fooExists := Default.objects["foo_count"]
	_, fooTimeExists := Default.objects["foo_time"]
	_, barExists := Default.objects["bar_count"]

	assert.False(t, fooExists, "foo_count should be cleared")
	assert.False(t, fooTimeExists, "foo_time should be cleared")
	assert.True(t, barExists, "bar_count should remain")
}

func TestLoadFromJSON(t *testing.T) {
	ClearAll()
	jsonData := `{
		"int_metric": 42,
		"float_metric": 3.14159,
		"bool_metric": true,
		"string_metric": "hello"
	}`

	err := LoadFromJSON(json.RawMessage(jsonData))
	assert.NoError(t, err, "LoadFromJSON should not return error")

	assert.Equal(t, int64(42), GetInt("int_metric"))
	assert.InDelta(t, 3.14, GetFloat("float_metric"), 0.001)
	assert.Equal(t, true, GetBool("bool_metric"))
	assert.Equal(t, "hello", GetString("string_metric"))
}

func TestLoadFromJSONInvalid(t *testing.T) {
	ClearAll()
	err := LoadFromJSON(json.RawMessage(`invalid json`))
	assert.Error(t, err, "Invalid JSON should return error")
}