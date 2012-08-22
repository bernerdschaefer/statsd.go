package main

import (
	"reflect"
	"testing"
)

func waitForOperations(metrics *Metrics) {
	done := make(chan bool)
	metrics.queue(func() {
		done <- true
	})
	<-done
}

func TestUpdateCounter(t *testing.T) {
	metrics := NewMetrics()

	metrics.UpdateCounter("a", 1, 1)
	waitForOperations(metrics)
	a := metrics.Counters["a"]
	if a != 1 {
		t.Error("Counter 'a' expected to equal 1, was:", a)
	}

	metrics.UpdateCounter("a", 1, 1)
	waitForOperations(metrics)
	a = metrics.Counters["a"]
	if a != 2 {
		t.Error("Counter 'a' expected to equal 2, was:", a)
	}

	// with sample rate
	metrics.UpdateCounter("a", 1, 0.5)
	waitForOperations(metrics)
	a = metrics.Counters["a"]
	if a != 4 {
		t.Error("Counter 'a' expected to equal 4, was:", a)
	}
}

func TestUpdateGauge(t *testing.T) {
	metrics := NewMetrics()

	metrics.UpdateGauge("a", 1)
	waitForOperations(metrics)

	a := metrics.Gauges["a"]
	if a != 1 {
		t.Error("Gauge 'a' expected to equal 1, was:", a)
	}

	metrics.UpdateGauge("a", 2.5)
	waitForOperations(metrics)

	a = metrics.Gauges["a"]
	if a != 2.5 {
		t.Error("Gauge 'a' expected to equal 2.5, was:", a)
	}
}

func TestUpdateTimer(t *testing.T) {
	metrics := NewMetrics()

	metrics.UpdateTimer("a", 1)
	waitForOperations(metrics)

	a := metrics.Timers["a"]
	if !reflect.DeepEqual(a, []float64{1}) {
		t.Error("Expected timer 'a' to be [1], was:", a)
	}

	metrics.UpdateTimer("a", 3)
	waitForOperations(metrics)

	a = metrics.Timers["a"]
	if !reflect.DeepEqual(a, []float64{1, 3}) {
		t.Error("Expected timer 'a' to be [1, 3], was:", a)
	}
}

func TestDeleteCounter(t *testing.T) {
	metrics := NewMetrics()

	metrics.UpdateCounter("a", 1, 1)
	metrics.DeleteCounter("a")
	waitForOperations(metrics)

	if _, ok := metrics.Counters["a"]; ok {
		t.Error("Counter was not deleted")
	}
}

func TestDeleteGauge(t *testing.T) {
	metrics := NewMetrics()

	metrics.UpdateGauge("a", 1)
	metrics.DeleteGauge("a")
	waitForOperations(metrics)

	if _, ok := metrics.Gauges["a"]; ok {
		t.Error("Gauge was not deleted")
	}
}

func TestDeleteTimer(t *testing.T) {
	metrics := NewMetrics()

	metrics.UpdateTimer("a", 1)
	metrics.DeleteTimer("a")
	waitForOperations(metrics)

	if _, ok := metrics.Timers["a"]; ok {
		t.Error("Timer was not deleted")
	}
}

func TestRead(t *testing.T) {
	metrics := NewMetrics()

	metrics.UpdateCounter("a", 1, 1)
	metrics.UpdateGauge("a", 1)
	metrics.UpdateTimer("a", 1)
	counters, gauges, timers := metrics.Read()

	if _, ok := counters["a"]; ok {
		t.Error("read waited for counter update")
	}

	if _, ok := gauges["a"]; ok {
		t.Error("read waited for gauge update")
	}

	if _, ok := timers["a"]; ok {
		t.Error("read waited for timer update")
	}

	waitForOperations(metrics)

	counters, gauges, timers = metrics.Read()

	if counters["a"] != 1 {
		t.Error("Counter not updated")
	}

	if gauges["a"] != 1 {
		t.Error("Gauge not updated")
	}

	if !reflect.DeepEqual(timers["a"], []float64{1}) {
		t.Error("Timer not updated")
	}
}

func TestFlush(t *testing.T) {
	metrics := NewMetrics()

	metrics.UpdateCounter("a", 1, 1)
	metrics.UpdateGauge("a", 1)
	metrics.UpdateTimer("a", 1)
	flushedMetrics := metrics.Flush()
	counters := flushedMetrics.Counters
	gauges := flushedMetrics.Gauges
	timers := flushedMetrics.Timers

	if _, ok := counters["a"]; ok {
		t.Error("flush waited for counter update")
	}

	if _, ok := gauges["a"]; ok {
		t.Error("flush waited for gauge update")
	}

	if _, ok := timers["a"]; ok {
		t.Error("flush waited for timer update")
	}

	waitForOperations(metrics)

	flushedMetrics = metrics.Flush()
	counters = flushedMetrics.Counters
	gauges = flushedMetrics.Gauges
	timers = flushedMetrics.Timers

	if counters["a"] != 1 {
		t.Error("Counter not updated")
	}

	if gauges["a"] != 1 {
		t.Error("Gauge not updated")
	}

	if !reflect.DeepEqual(timers["a"], []float64{1}) {
		t.Error("Timer not updated")
	}

	if _, ok := metrics.Counters["a"]; ok {
		t.Error("flush did not clear Counters")
	}

	if _, ok := metrics.Gauges["a"]; ok {
		t.Error("flush did not clear Gauges")
	}

	if _, ok := metrics.Timers["a"]; ok {
		t.Error("flush did not clear Timers")
	}
}

func BenchmarkUpdateCounter(b *testing.B) {
	metrics := NewMetrics()

	for i := 0; i < b.N; i++ {
		metrics.UpdateCounter("a", 1, 1)
	}

	waitForOperations(metrics)
	a := metrics.Counters["a"]

	if a != float64(b.N) {
		b.Error("Expected counter 'a' to equal", b.N, "but was:", a)
	}
}
