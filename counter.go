// ====================================================================
//  JIRA: PLATFORM-2828 — Fix Race Condition in Analytics Counter Service
// ====================================================================
//  Priority: P1 — Sev2 | Sprint: Sprint 23 | Points: 3
//  Reporter: Data Engineering Team
//  Assignee: You (Intern)
//  Due: End of day — analytics pipeline is producing incorrect data
//  Labels: concurrency, go, production, data-integrity
//
//  DESCRIPTION:
//  The analytics counter service is losing increments under concurrent
//  load. Dashboard shows 847,231 page views but the counter only
//  recorded 801,445 — a 5.4% data loss. This counter feeds into our
//  revenue attribution model, so incorrect counts directly affect
//  billing accuracy.
//
//  The counter struct does not use any synchronization primitives.
//  All read/write operations on shared state are unprotected.
//
//  STEPS TO REPRODUCE:
//  1. Run: go test -race -v counter.go
//  2. TestConcurrentIncrement fails with lost increments
//  3. Race detector reports data races on the 'value' field
//  4. TestConcurrentRegistry panics with "concurrent map writes"
//
//  ACCEPTANCE CRITERIA:
//  - [ ] No data races detected by `go test -race`
//  - [ ] Counter is accurate under 1000 concurrent goroutines
//  - [ ] Registry is safe for concurrent GetOrCreate calls
//  - [ ] Reset operation is atomic (no partial state)
//  - [ ] All 4 test cases pass with -race flag
// ====================================================================
//
//  SLACK THREAD — #data-engineering — Feb 12, 2026:
//  ─────────────────────────────────────────────────
//  @mike.chen (Data Eng) 9:30 AM:
//    "The daily analytics numbers are off again. Page view counter shows
//     5.4% less than what our CDN logs report. This has been getting
//     worse as traffic grows."
//
//  @priya.das (Data Eng Lead) 9:35 AM:
//    "I bet it's a concurrency issue. The counter service handles
//     ~2000 req/s and I don't think the Go code uses mutexes or atomics.
//     Classic read-modify-write race condition."
//
//  @raj.patel (Senior Dev) 9:40 AM:
//    "Just confirmed — no sync primitives anywhere. Also the
//     CounterRegistry uses a plain map which panics on concurrent
//     writes in Go. We need either sync.Mutex or sync.Map."
//
//  @nisha.gupta (Tech Lead) 9:42 AM:
//    "@intern — This is a great learning exercise. Fix the race
//     conditions. Options: sync.Mutex, sync.RWMutex, or sync/atomic.
//     For the map, use sync.RWMutex (sync.Map is overkill here).
//     Run `go test -race` to verify your fix."
//
//  RACE DETECTOR OUTPUT (from CI):
//  ───────────────────────────────
//  WARNING: DATA RACE
//  Write at 0x00c0000b4010 by goroutine 8:
//    counter.(*AtomicCounter).Increment()
//        counter.go:29 +0x4a
//  Previous read at 0x00c0000b4010 by goroutine 7:
//    counter.(*AtomicCounter).Value()
//        counter.go:34 +0x3e
//
//  WARNING: DATA RACE
//  Write at 0x00c0000c2060 by goroutine 15:
//    counter.(*CounterRegistry).GetOrCreate()
//        counter.go:62 +0xbb
//  Previous write at 0x00c0000c2060 by goroutine 12:
//    counter.(*CounterRegistry).GetOrCreate()
//        counter.go:62 +0xbb
//
//  FAIL — 2 data races detected
// ====================================================================

package counter

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// ─── Counter ──────────────────────────────────────────────────────────

// AtomicCounter tracks a single named metric.
// Despite the name, this implementation is NOT atomic.
type AtomicCounter struct {
	name      string
	value     int64
	createdAt time.Time
}

// NewCounter creates a new counter with the given name.
func NewCounter(name string) *AtomicCounter {
	return &AtomicCounter{
		name:      name,
		value:     0,
		createdAt: time.Now(),
	}
}

// Increment increases the counter by 1.
// BUG: Not thread-safe — concurrent goroutines cause lost writes.
func (c *AtomicCounter) Increment() {
	c.value++
}

// Value returns the current counter value.
// BUG: Reading without synchronization — may see stale or torn values.
func (c *AtomicCounter) Value() int64 {
	return c.value
}

// Reset sets the counter back to zero and returns the previous value.
// BUG: Read and write are not atomic — another goroutine could
// increment between the read and the reset.
func (c *AtomicCounter) Reset() int64 {
	prev := c.value
	c.value = 0
	return prev
}

// Name returns the counter's name.
func (c *AtomicCounter) Name() string {
	return c.name
}

// ─── Counter Registry ─────────────────────────────────────────────────

// CounterRegistry manages multiple named counters.
type CounterRegistry struct {
	counters map[string]*AtomicCounter
}

// NewRegistry creates an empty counter registry.
func NewRegistry() *CounterRegistry {
	return &CounterRegistry{
		counters: make(map[string]*AtomicCounter),
	}
}

// GetOrCreate returns an existing counter or creates a new one.
// BUG: Map access is not synchronized — concurrent goroutines will
// panic with "concurrent map read and map write" at high load.
func (r *CounterRegistry) GetOrCreate(name string) *AtomicCounter {
	if c, ok := r.counters[name]; ok {
		return c
	}
	c := NewCounter(name)
	r.counters[name] = c
	return c
}

// GetAll returns all counters and their current values.
func (r *CounterRegistry) GetAll() map[string]int64 {
	result := make(map[string]int64)
	for name, counter := range r.counters {
		result[name] = counter.Value()
	}
	return result
}

// ─── Tests (run with: go test -race -v) ───────────────────────────────

func TestBasicIncrement(t *testing.T) {
	c := NewCounter("basic_test")
	c.Increment()
	c.Increment()
	c.Increment()
	if c.Value() != 3 {
		t.Errorf("Expected 3, got %d", c.Value())
	}
}

func TestConcurrentIncrement(t *testing.T) {
	c := NewCounter("concurrent_test")
	numGoroutines := 1000
	incrementsPerGoroutine := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				c.Increment()
			}
		}()
	}
	wg.Wait()

	expected := int64(numGoroutines * incrementsPerGoroutine)
	actual := c.Value()
	if actual != expected {
		t.Errorf("Expected %d, got %d (lost %d increments — %.1f%% loss)",
			expected, actual, expected-actual,
			float64(expected-actual)/float64(expected)*100)
	}
}

func TestConcurrentResetAndIncrement(t *testing.T) {
	c := NewCounter("reset_test")

	var wg sync.WaitGroup
	wg.Add(2)

	// Writer: increment 10000 times
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			c.Increment()
		}
	}()

	// Resetter: reset 100 times
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			c.Reset()
			time.Sleep(10 * time.Microsecond)
		}
	}()

	wg.Wait()
	fmt.Printf("Final value after concurrent increment+reset: %d\n", c.Value())
}

func TestConcurrentRegistry(t *testing.T) {
	registry := NewRegistry()

	var wg sync.WaitGroup
	wg.Add(100)

	// 100 goroutines all trying to GetOrCreate the same counters
	for i := 0; i < 100; i++ {
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("counter_%d", id%10)
			counter := registry.GetOrCreate(name)
			counter.Increment()
		}(i)
	}

	wg.Wait()

	all := registry.GetAll()
	total := int64(0)
	for _, v := range all {
		total += v
	}

	if total != 100 {
		t.Errorf("Expected total 100 across all counters, got %d", total)
	}
}
