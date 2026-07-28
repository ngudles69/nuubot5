package runharness

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Section 1 - Program Flow

func TestProfileLifecycle(t *testing.T) {
	var prefix = filepath.Join(t.TempDir(), "run-001")
	var profile = NewProfile(prefix)
	var err = profile.Start()
	if err != nil {
		t.Fatal(err)
	}

	var values = make([]byte, 1024*1024)
	for index := range values {
		values[index] = byte(index)
	}
	time.Sleep(25 * time.Millisecond)

	err = profile.Stop()
	if err != nil {
		t.Fatal(err)
	}
	err = profile.Stop()
	if err != nil {
		t.Fatalf("second stop failed: %v", err)
	}

	var suffixes = []string{
		".cpu.pprof",
		".trace",
		".heap.pprof",
		".allocs.pprof",
		".block.pprof",
		".mutex.pprof",
	}
	for _, suffix := range suffixes {
		var info os.FileInfo
		info, err = os.Stat(prefix + suffix)
		if err != nil {
			t.Errorf("stat %s: %v", suffix, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("profile %s is empty", suffix)
		}
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
