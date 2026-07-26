package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Section 1 - Program Flow

func TestParseInput(t *testing.T) {
	var sweepID, botID, profilePrefix, err = parseInput([]string{"6", "9"})
	if err != nil {
		t.Fatal(err)
	}
	if sweepID != 6 {
		t.Fatalf("actual sweep ID %d, expected 6", sweepID)
	}
	if botID != 9 {
		t.Fatalf("actual bot ID %d, expected 9", botID)
	}
	if profilePrefix != "" {
		t.Fatalf("actual profile prefix %q, expected empty", profilePrefix)
	}
}

func TestParseInputAcceptsProfilePrefix(t *testing.T) {
	var sweepID, botID, profilePrefix, err = parseInput([]string{"6", "9", "-pp", "profiles/run-001"})
	if err != nil {
		t.Fatal(err)
	}
	if sweepID != 6 || botID != 9 {
		t.Fatalf("actual identity %d/%d, expected 6/9", sweepID, botID)
	}
	if profilePrefix != "profiles/run-001" {
		t.Fatalf("actual profile prefix %q, expected profiles/run-001", profilePrefix)
	}
}

func TestParseInputRejectsInvalidInput(t *testing.T) {
	tests := [][]string{
		nil,
		{"0", "9"},
		{"6", "invalid"},
		{"6", "9", "-pp"},
		{"6", "9", "invalid", "prefix"},
		{"6", "9", "-pp", ""},
		{"6", "9", "-pp", "prefix", "extra"},
	}
	for _, args := range tests {
		_, _, _, err := parseInput(args)
		if err == nil {
			t.Fatalf("actual error nil for %v, expected error", args)
		}
	}
}

func TestPerformanceProfileLifecycle(t *testing.T) {
	var prefix = filepath.Join(t.TempDir(), "run-001")
	var profile = performanceProfile{prefix: prefix}
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
