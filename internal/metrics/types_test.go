package metrics

import (
	"sync"
	"testing"

	"github.com/tanay13/GlitchMesh/internal/constants"
)

// TestFaultMetrics_Increment verifies that incrementing a metric
// accumulates the value correctly over multiple calls.
func TestFaultMetrics_Increment(t *testing.T) {
	fm := NewFaultMetrics()

	fm.Increment(constants.LATENCY, 1)
	fm.Increment(constants.LATENCY, 1)
	fm.Increment(constants.LATENCY, 3)

	snap := fm.Snapshot()
	if snap.LatencyCount != 5 {
		t.Errorf("LatencyCount = %d, want 5", snap.LatencyCount)
	}
}

// TestFaultMetrics_Set verifies that Set overwrites (not adds to) a metric value.
func TestFaultMetrics_Set(t *testing.T) {
	fm := NewFaultMetrics()

	fm.Increment(constants.ERROR, 10)
	fm.Set(constants.ERROR, 3) // should overwrite, not add

	snap := fm.Snapshot()
	if snap.ErrorCount != 3 {
		t.Errorf("ErrorCount = %d, want 3 (Set should overwrite)", snap.ErrorCount)
	}
}

// TestFaultMetrics_Snapshot verifies that Snapshot returns a consistent view
// of all metric counters at once.
func TestFaultMetrics_Snapshot(t *testing.T) {
	fm := NewFaultMetrics()

	fm.Increment(constants.LATENCY, 2)
	fm.Increment(constants.ERROR, 5)
	fm.Increment(constants.TIMEOUT, 1)
	fm.Increment(constants.TOTAL_FAULTS_INJECTED, 8)

	snap := fm.Snapshot()

	if snap.LatencyCount != 2 {
		t.Errorf("LatencyCount = %d, want 2", snap.LatencyCount)
	}
	if snap.ErrorCount != 5 {
		t.Errorf("ErrorCount = %d, want 5", snap.ErrorCount)
	}
	if snap.TimeoutCount != 1 {
		t.Errorf("TimeoutCount = %d, want 1", snap.TimeoutCount)
	}
	if snap.TotalFaultsInjected != 8 {
		t.Errorf("TotalFaultsInjected = %d, want 8", snap.TotalFaultsInjected)
	}
}

// TestFaultMetrics_GetReturnsCopy verifies that Get() returns a copy of
// the internal map, so callers can't accidentally mutate the metrics.
func TestFaultMetrics_GetReturnsCopy(t *testing.T) {
	fm := NewFaultMetrics()
	fm.Increment(constants.LATENCY, 5)

	// Get a copy and mutate it
	got := fm.Get()
	got[constants.LATENCY] = 999

	// Original should be untouched
	snap := fm.Snapshot()
	if snap.LatencyCount != 5 {
		t.Errorf("Original mutated! LatencyCount = %d, want 5", snap.LatencyCount)
	}
}

// TestFaultMetrics_ConcurrentAccess hammers the metrics from 100 goroutines
// simultaneously. Run with -race to detect data races.
// This is a regression test for the race condition we fixed by adding sync.RWMutex.
func TestFaultMetrics_ConcurrentAccess(t *testing.T) {
	fm := NewFaultMetrics()
	var wg sync.WaitGroup

	goroutines := 100

	// Each goroutine increments every counter once and reads a snapshot
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fm.Increment(constants.LATENCY, 1)
			fm.Increment(constants.ERROR, 1)
			fm.Increment(constants.TIMEOUT, 1)
			fm.Increment(constants.TOTAL_FAULTS_INJECTED, 3)
			_ = fm.Snapshot() // concurrent read while others write
			_ = fm.Get()     // concurrent read returning copy
		}()
	}

	wg.Wait()

	// After all goroutines finish, counts should be exact
	snap := fm.Snapshot()
	if snap.LatencyCount != goroutines {
		t.Errorf("LatencyCount = %d, want %d", snap.LatencyCount, goroutines)
	}
	if snap.ErrorCount != goroutines {
		t.Errorf("ErrorCount = %d, want %d", snap.ErrorCount, goroutines)
	}
	if snap.TimeoutCount != goroutines {
		t.Errorf("TimeoutCount = %d, want %d", snap.TimeoutCount, goroutines)
	}
	if snap.TotalFaultsInjected != goroutines*3 {
		t.Errorf("TotalFaultsInjected = %d, want %d", snap.TotalFaultsInjected, goroutines*3)
	}
}

// TestFaultMetrics_InitialValues verifies all counters start at zero.
func TestFaultMetrics_InitialValues(t *testing.T) {
	fm := NewFaultMetrics()
	snap := fm.Snapshot()

	if snap.LatencyCount != 0 || snap.ErrorCount != 0 || snap.TimeoutCount != 0 || snap.TotalFaultsInjected != 0 {
		t.Errorf("initial values should all be 0, got latency=%d error=%d timeout=%d total=%d",
			snap.LatencyCount, snap.ErrorCount, snap.TimeoutCount, snap.TotalFaultsInjected)
	}
}
