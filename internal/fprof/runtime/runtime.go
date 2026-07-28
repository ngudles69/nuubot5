// Package fprofruntime owns dependency-free generated function instrumentation.
package fprofruntime

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"
)

const (
	// ModeStructural records exact calls without timestamps.
	ModeStructural = "structural"
	// ModeTimed records exact calls plus flat and cumulative elapsed time.
	ModeTimed = "timed"
)

// Function contains one exact instrumented function result.
type Function struct {
	Name  string `json:"name"`
	Calls uint64 `json:"calls"`
	Flat  int64  `json:"flat_ns"`
	Cum   int64  `json:"cum_ns"`
}

// RuntimeReport contains one complete instrumented-process profile.
type RuntimeReport struct {
	Mode       string     `json:"mode"`
	RootNS     int64      `json:"root_ns"`
	Functions  []Function `json:"functions"`
	GCRuns     uint32     `json:"gc_runs"`
	GCPauseNS  uint64     `json:"gc_pause_ns"`
	HeapBytes  uint64     `json:"heap_bytes"`
	TotalAlloc uint64     `json:"total_alloc_bytes"`
	StackError string     `json:"stack_error,omitempty"`
	OpenFrames int        `json:"open_frames"`
	Output     string     `json:"-"`
}

type functionStats struct {
	calls uint64
	flat  int64
	cum   int64
}

type frame struct {
	name    string
	started time.Time
	child   time.Duration
}

type collector struct {
	mu         sync.Mutex
	mode       string
	output     string
	functions  map[string]*functionStats
	stack      []*frame
	root       time.Duration
	stackError string
}

var process = collector{
	mode:      os.Getenv("NUUBOT_FPROF_MODE"),
	output:    os.Getenv("NUUBOT_FPROF_OUTPUT"),
	functions: make(map[string]*functionStats),
}

// Section 1 - Program Flow

// Enter records one exact function entry and returns its exit handler.
func Enter(name string) func() {
	if process.mode != ModeStructural && process.mode != ModeTimed {
		return func() {}
	}

	// Step 1: record exact function call
	process.mu.Lock()
	var stats = process.functions[name]
	if stats == nil {
		stats = &functionStats{}
		process.functions[name] = stats
	}
	stats.calls++
	if process.mode == ModeStructural {
		process.mu.Unlock()
		return func() {}
	}

	// Step 2: push timed function frame
	var current = &frame{name: name, started: time.Now()}
	process.stack = append(process.stack, current)
	process.mu.Unlock()

	// Step 3: return timed function exit
	return func() {
		process.exit(current)
	}
}

// Write writes one complete process profile to the configured output.
func Write() {
	if process.mode != ModeStructural && process.mode != ModeTimed {
		return
	}
	var err = writeReport(process.output)
	if err != nil {
		panic(fmt.Sprintf("write function profile: %v", err))
	}
}

// Section 2 - Domain Helpers

func (c *collector) exit(current *frame) {
	var elapsed = time.Since(current.started)
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.stack) == 0 || c.stack[len(c.stack)-1] != current {
		if c.stackError == "" {
			c.stackError = fmt.Sprintf("non-LIFO exit: %s", current.name)
		}
		return
	}
	c.stack = c.stack[:len(c.stack)-1]
	var flat = elapsed - current.child
	if flat < 0 {
		flat = 0
	}
	var stats = c.functions[current.name]
	stats.flat += flat.Nanoseconds()
	stats.cum += elapsed.Nanoseconds()
	if len(c.stack) == 0 {
		c.root += elapsed
		return
	}
	c.stack[len(c.stack)-1].child += elapsed
}

func writeReport(path string) error {
	if path == "" {
		return fmt.Errorf("NUUBOT_FPROF_OUTPUT is required")
	}

	// Step 1: collect function results
	process.mu.Lock()
	var report = RuntimeReport{
		Mode:       process.mode,
		RootNS:     process.root.Nanoseconds(),
		StackError: process.stackError,
		OpenFrames: len(process.stack),
		Output:     path,
	}
	for name, stats := range process.functions {
		report.Functions = append(report.Functions, Function{
			Name:  name,
			Calls: stats.calls,
			Flat:  stats.flat,
			Cum:   stats.cum,
		})
	}
	process.mu.Unlock()
	sort.Slice(report.Functions, func(left int, right int) bool {
		return report.Functions[left].Name < report.Functions[right].Name
	})

	// Step 2: collect runtime memory results
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	report.GCRuns = memory.NumGC
	report.GCPauseNS = memory.PauseTotalNs
	report.HeapBytes = memory.HeapAlloc
	report.TotalAlloc = memory.TotalAlloc

	// Step 3: write runtime report
	var file, err = os.Create(path)
	if err != nil {
		return err
	}
	var encoder = json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	var writeErr = encoder.Encode(report)
	var closeErr = file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

// Section 3 - Generic Helpers
