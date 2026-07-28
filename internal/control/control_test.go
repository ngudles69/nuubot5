package control

import (
	"database/sql"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

// Section 1 - Program Flow

func TestStoreClaimsNewestAndSkipsSupersededCommands(t *testing.T) {
	var store Store
	var path = filepath.Join(t.TempDir(), "nuubot.db")
	if err := store.Init(path); err != nil {
		t.Fatal(err)
	}
	defer store.Stop()
	prepareControlTargets(t, store.db)
	var process, err = store.RegisterProcess(Bot, 9, 100, "worker-1", 1000, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{Pause, Resume, Stop} {
		if _, err = store.Enqueue(Bot, 9, action, 1001); err != nil {
			t.Fatal(err)
		}
	}
	var command Command
	var found bool
	command, found, err = store.ClaimLatest(process, 1002)
	if err != nil {
		t.Fatal(err)
	}
	if !found || command.Action != Stop || command.ID != 3 {
		t.Fatalf("claimed command = %+v found=%t, want latest Stop", command, found)
	}
	if err = store.Acknowledge(process, command.ID, Processed, "stop requested", 1003); err != nil {
		t.Fatal(err)
	}
	var skipped, processed int
	err = store.db.QueryRow(`
		SELECT
			SUM(CASE WHEN outcome = 'skipped' THEN 1 ELSE 0 END),
			SUM(CASE WHEN outcome = 'processed' THEN 1 ELSE 0 END)
		FROM control_command
	`).Scan(&skipped, &processed)
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 2 || processed != 1 {
		t.Fatalf("skipped=%d processed=%d, want 2/1", skipped, processed)
	}
}

func TestStoreAllowsOneConcurrentProcessGeneration(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "nuubot.db")
	var bootstrap Store
	if err := bootstrap.Init(path); err != nil {
		t.Fatal(err)
	}
	prepareControlTargets(t, bootstrap.db)
	if err := bootstrap.Stop(); err != nil {
		t.Fatal(err)
	}
	var winners atomic.Int64
	var wait sync.WaitGroup
	for index := 1; index <= 8; index++ {
		wait.Add(1)
		go func(pid int) {
			defer wait.Done()
			var store Store
			if err := store.Init(path); err != nil {
				t.Errorf("initialize Store: %v", err)
				return
			}
			defer store.Stop()
			if _, err := store.RegisterProcess(
				Bot,
				9,
				pid,
				"worker-"+string(rune('a'+pid)),
				1000,
				1,
			); err == nil {
				winners.Add(1)
			}
		}(index)
	}
	wait.Wait()
	if winners.Load() != 1 {
		t.Fatalf("registration winners = %d, want 1", winners.Load())
	}
}

func TestStoreRejectsStaleProcessAcknowledgement(t *testing.T) {
	var store Store
	var path = filepath.Join(t.TempDir(), "nuubot.db")
	if err := store.Init(path); err != nil {
		t.Fatal(err)
	}
	defer store.Stop()
	prepareControlTargets(t, store.db)
	var first, err = store.RegisterProcess(Bot, 9, 100, "worker-1", 1000, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.UpdateProcess(first, ProcessStopped, "stopped", 1001); err != nil {
		t.Fatal(err)
	}
	var queued Command
	queued, err = store.Enqueue(Bot, 9, Stop, 1002)
	if err != nil {
		t.Fatal(err)
	}
	if queued.Generation != 0 {
		t.Fatalf("queued generation = %d, want next generation", queued.Generation)
	}
	var second Process
	second, err = store.RegisterProcess(Bot, 9, 101, "worker-2", 1003, 1)
	if err != nil {
		t.Fatal(err)
	}
	var command Command
	var found bool
	command, found, err = store.ClaimLatest(second, 1004)
	if err != nil || !found {
		t.Fatalf("claim command found=%t error=%v", found, err)
	}
	if err = store.Acknowledge(first, command.ID, Processed, "stale", 1005); err == nil {
		t.Fatal("stale process acknowledged current command")
	}
	if err = store.Acknowledge(second, command.ID, Processed, "stopped", 1006); err != nil {
		t.Fatal(err)
	}
}

// Section 2 - Domain Helpers

func prepareControlTargets(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		CREATE TABLE bot (
			bot_id INTEGER PRIMARY KEY,
			status TEXT NOT NULL,
			process_pid INTEGER,
			process_create_time TEXT,
			updated_at TEXT NOT NULL,
			started_at TEXT,
			ended_at TEXT
		);
		CREATE TABLE sweep (
			sweep_id INTEGER PRIMARY KEY,
			status TEXT NOT NULL,
			process_status TEXT NOT NULL,
			process_pid INTEGER,
			process_create_time TEXT,
			updated_at TEXT NOT NULL,
			started_at TEXT,
			ended_at TEXT
		);
		INSERT INTO bot (bot_id, status, updated_at) VALUES (9, 'configured', 'initial');
		INSERT INTO sweep (sweep_id, status, process_status, updated_at)
		VALUES (6, 'configured', 'stopped', 'initial');
	`)
	if err != nil {
		t.Fatal(err)
	}
}

// Section 3 - Generic Helpers
