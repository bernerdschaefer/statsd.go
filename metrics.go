package main

import (
	"sync"
)

type counters map[string]float64
type gauges map[string]float64
type timers map[string][]float64

type Metrics struct {
	_mu      sync.RWMutex
	_queue   chan func()
	Counters counters
	Gauges   gauges
	Timers   timers
}

// NewMetrics returns a new Metrics containing maps for storing counters,
// gauges, and timers. It also starts a goroutine for processing the internal
// async queue.
func NewMetrics() *Metrics {
	metrics := &Metrics{
		_queue:   make(chan func(), 1000),
		Counters: make(counters),
		Gauges:   make(gauges),
		Timers:   make(timers),
	}

	go metrics.processQueue()

	return metrics
}

func (m *Metrics) processQueue() {
	for op := range m._queue {
		m.withLock(op)
	}
}

// Wraps the provided function inside of an exclusive write lock.
func (m *Metrics) withLock(op func()) {
	m._mu.Lock()
	defer m._mu.Unlock()

	op()
}

// Adds the provided function to the internal operation queue.
func (m *Metrics) queue(op func()) {
	m._queue <- op
}

// Updates the named counter asynchronously.
func (m *Metrics) UpdateCounter(key string, increment float64, sampleRate float64) {
	m.queue(func() {
		v := m.Counters[key]
		m.Counters[key] = v + (increment / sampleRate)
	})
}

// Updates the named gauge asynchronously.
func (m *Metrics) UpdateGauge(key string, value float64) {
	m.queue(func() {
		m.Gauges[key] = value
	})
}

// Updates the named timer asynchronously.
func (m *Metrics) UpdateTimer(key string, timing float64) {
	m.queue(func() {
		m.Timers[key] = append(m.Timers[key], timing)
	})
}

// Deletes the named counter asynchronously.
func (m *Metrics) DeleteCounter(key string) {
	m.queue(func() {
		delete(m.Counters, key)
	})
}

// Deletes the named gauge asynchronously.
func (m *Metrics) DeleteGauge(key string) {
	m.queue(func() {
		delete(m.Gauges, key)
	})
}

// Deletes the named timer asynchronously.
func (m *Metrics) DeleteTimer(key string) {
	m.queue(func() {
		delete(m.Timers, key)
	})
}

// Returns the current counter, gauge, and timer maps immediately.
func (m *Metrics) Read() (counters, gauges, timers) {
	m._mu.RLock()
	defer m._mu.RUnlock()

	return m.read()
}

// Returns a copy of the current counter, gauge, and timer maps.
func (m *Metrics) read() (counters, gauges, timers) {
	counters := make(counters)
	for k, v := range m.Counters {
		counters[k] = v
	}

	gauges := make(gauges)
	for k, v := range m.Gauges {
		gauges[k] = v
	}

	timers := make(timers)
	for k, v := range m.Timers {
		timers[k] = v
	}

	return counters, gauges, timers
}

// Resets the metrics and returns the current counter, gauge, and timer maps
// immediately.
func (m *Metrics) Flush() *Metrics {
	m._mu.Lock()
	defer m._mu.Unlock()

	counters, gauges, timers := m.read()
	m.reset()

	return &Metrics{
		Counters : counters,
		Gauges: gauges,
		Timers: timers,
	}
}

// Resets the counter, gauge, and timer maps.
func (m *Metrics) reset() {
	m.Counters = make(counters)
	m.Gauges = make(gauges)
	m.Timers = make(timers)
}
