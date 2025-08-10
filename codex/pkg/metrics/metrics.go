package metrics

import (
	"encoding/json"
	"fmt"
	"sync"
)

// ---- Metric interfaces/types ----

type Metric interface{}

// IntCounter
type IntCounter struct {
	mu sync.Mutex
	v  int64
}

func (c *IntCounter) Add(delta int64) {
	c.mu.Lock()
	c.v += delta
	c.mu.Unlock()
}

func (c *IntCounter) Inc() {
	c.Add(1)
}

func (c *IntCounter) Get() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.v
}

// FloatCounter
type FloatCounter struct {
	mu sync.Mutex
	v  float64
}

func (c *FloatCounter) Add(delta float64) {
	c.mu.Lock()
	c.v += delta
	c.mu.Unlock()
}

func (c *FloatCounter) Get() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.v
}

// BoolGauge
type BoolGauge struct {
	mu sync.Mutex
	v  bool
}

func (g *BoolGauge) Set(val bool) {
	g.mu.Lock()
	g.v = val
	g.mu.Unlock()
}

func (g *BoolGauge) Get() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.v
}

// StringGauge
type StringGauge struct {
	mu sync.Mutex
	v  string
}

func (g *StringGauge) Set(val string) {
	g.mu.Lock()
	g.v = val
	g.mu.Unlock()
}

func (g *StringGauge) Get() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.v
}

// ---- Registry ----

type Registry struct {
	mu      sync.RWMutex
	objects map[string]Metric
}

func NewRegistry() *Registry {
	return &Registry{objects: make(map[string]Metric)}
}

// internal helpers
func (r *Registry) getOrCreateInt(name string) *IntCounter {
	r.mu.RLock()
	if m, ok := r.objects[name]; ok {
		r.mu.RUnlock()
		if ic, ok := m.(*IntCounter); ok {
			return ic
		}
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.objects[name]; ok {
		if ic, ok := m.(*IntCounter); ok {
			return ic
		}
	}
	ic := &IntCounter{}
	r.objects[name] = ic
	return ic
}

func (r *Registry) getOrCreateFloat(name string) *FloatCounter {
	r.mu.RLock()
	if m, ok := r.objects[name]; ok {
		r.mu.RUnlock()
		if fc, ok := m.(*FloatCounter); ok {
			return fc
		}
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.objects[name]; ok {
		if fc, ok := m.(*FloatCounter); ok {
			return fc
		}
	}
	fc := &FloatCounter{}
	r.objects[name] = fc
	return fc
}

func (r *Registry) getOrCreateBool(name string) *BoolGauge {
	r.mu.RLock()
	if m, ok := r.objects[name]; ok {
		r.mu.RUnlock()
		if bg, ok := m.(*BoolGauge); ok {
			return bg
		}
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.objects[name]; ok {
		if bg, ok := m.(*BoolGauge); ok {
			return bg
		}
	}
	bg := &BoolGauge{}
	r.objects[name] = bg
	return bg
}

func (r *Registry) getOrCreateString(name string) *StringGauge {
	r.mu.RLock()
	if m, ok := r.objects[name]; ok {
		r.mu.RUnlock()
		if sg, ok := m.(*StringGauge); ok {
			return sg
		}
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.objects[name]; ok {
		if sg, ok := m.(*StringGauge); ok {
			return sg
		}
	}
	sg := &StringGauge{}
	r.objects[name] = sg
	return sg
}

// ---- Public Go API ----

var Default = NewRegistry()

func IncInt(name string)                 { Default.getOrCreateInt(name).Inc() }
func AddInt(name string, delta int64)    { Default.getOrCreateInt(name).Add(delta) }
func GetInt(name string) int64           { return Default.getOrCreateInt(name).Get() }

func AddFloat(name string, delta float64) { Default.getOrCreateFloat(name).Add(delta) }
func GetFloat(name string) float64        { return Default.getOrCreateFloat(name).Get() }

func SetBool(name string, v bool)         { Default.getOrCreateBool(name).Set(v) }
func GetBool(name string) bool            { return Default.getOrCreateBool(name).Get() }

func SetString(name, v string)            { Default.getOrCreateString(name).Set(v) }
func GetString(name string) string        { return Default.getOrCreateString(name).Get() }

// ---- Snapshot JSON ----

func (r *Registry) SnapshotJSON() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data := make(map[string]interface{}, len(r.objects))
	for name, m := range r.objects {
		switch v := m.(type) {
		case *IntCounter:
			data[name] = v.Get()
		case *FloatCounter:
			data[name] = v.Get()
		case *BoolGauge:
			data[name] = v.Get()
		case *StringGauge:
			data[name] = v.Get()
		default:
			data[name] = fmt.Sprintf("%v", v)
		}
	}

	b, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func SnapshotJSON() string {
	return Default.SnapshotJSON()
}
