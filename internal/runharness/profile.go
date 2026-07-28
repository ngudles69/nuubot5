// Package runharness owns shared whole-Run process infrastructure.
package runharness

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
)

// Profile owns optional whole-Run Go profiling.
type Profile struct {
	prefix                string
	cpuFile               *os.File
	traceFile             *os.File
	previousMutexFraction int
	cpuStarted            bool
	traceStarted          bool
	started               bool
	stopped               bool
}

// Section 1 - Program Flow

// NewProfile creates one optional whole-Run Profile.
func NewProfile(prefix string) *Profile {
	return &Profile{prefix: prefix}
}

// Start starts every configured whole-Run profile.
func (p *Profile) Start() error {
	// Step 1: skip disabled profiling
	if p.prefix == "" {
		return nil
	}

	// Step 2: validate Profile state
	if p.started || p.stopped {
		return fmt.Errorf("performance profile cannot start from current state")
	}

	// Step 3: create profile outputs
	var err error
	p.cpuFile, err = os.Create(p.prefix + ".cpu.pprof")
	if err != nil {
		return fmt.Errorf("create CPU profile: %w", err)
	}
	p.traceFile, err = os.Create(p.prefix + ".trace")
	if err != nil {
		return errors.Join(fmt.Errorf("create runtime trace: %w", err), p.cpuFile.Close())
	}

	// Step 4: start runtime profiles
	p.previousMutexFraction = runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)
	err = pprof.StartCPUProfile(p.cpuFile)
	if err != nil {
		runtime.SetBlockProfileRate(0)
		runtime.SetMutexProfileFraction(p.previousMutexFraction)
		return errors.Join(fmt.Errorf("start CPU profile: %w", err), p.traceFile.Close(), p.cpuFile.Close())
	}
	p.cpuStarted = true
	err = trace.Start(p.traceFile)
	if err != nil {
		pprof.StopCPUProfile()
		p.cpuStarted = false
		runtime.SetBlockProfileRate(0)
		runtime.SetMutexProfileFraction(p.previousMutexFraction)
		return errors.Join(fmt.Errorf("start runtime trace: %w", err), p.traceFile.Close(), p.cpuFile.Close())
	}
	p.traceStarted = true
	p.started = true
	return nil
}

// Stop stops every configured whole-Run profile and writes runtime profiles.
func (p *Profile) Stop() error {
	// Step 1: skip disabled or stopped profiling
	if p.prefix == "" || p.stopped {
		return nil
	}

	// Step 2: validate Profile state
	if !p.started {
		return fmt.Errorf("performance profile is not started")
	}
	p.stopped = true

	// Step 3: stop active profiles
	if p.traceStarted {
		trace.Stop()
		p.traceStarted = false
	}
	if p.cpuStarted {
		pprof.StopCPUProfile()
		p.cpuStarted = false
	}
	var err = errors.Join(p.traceFile.Close(), p.cpuFile.Close())
	runtime.SetBlockProfileRate(0)
	runtime.SetMutexProfileFraction(p.previousMutexFraction)

	// Step 4: write runtime profiles
	runtime.GC()
	for _, name := range []string{"heap", "allocs", "block", "mutex"} {
		var writeErr = writeRuntimeProfile(p.prefix, name)
		err = errors.Join(err, writeErr)
	}
	return err
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers

func writeRuntimeProfile(prefix string, name string) error {
	var profile = pprof.Lookup(name)
	if profile == nil {
		return fmt.Errorf("runtime profile %q is unavailable", name)
	}
	var file, err = os.Create(prefix + "." + name + ".pprof")
	if err != nil {
		return fmt.Errorf("create %s profile: %w", name, err)
	}
	var writeErr = profile.WriteTo(file, 0)
	var closeErr = file.Close()
	if writeErr != nil {
		return errors.Join(fmt.Errorf("write %s profile: %w", name, writeErr), closeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close %s profile: %w", name, closeErr)
	}
	return nil
}
