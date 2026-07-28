// Package control owns central command and process coordination records.
package control

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Kind identifies one command or process target.
type Kind string

const (
	Bot   Kind = "bot"
	Sweep Kind = "sweep"
)

// CommandStatus identifies durable command progress.
type CommandStatus string

const (
	Requested    CommandStatus = "requested"
	Claimed      CommandStatus = "claimed"
	Acknowledged CommandStatus = "acknowledged"
)

// Outcome identifies one acknowledged command result.
type Outcome string

const (
	Processed Outcome = "processed"
	Skipped   Outcome = "skipped"
	Rejected  Outcome = "rejected"
)

const (
	Start  = "start"
	Pause  = "pause"
	Resume = "resume"
	Stop   = "stop"
)

const (
	ProcessStarting = "starting"
	ProcessRunning  = "running"
	ProcessStopping = "stopping"
	ProcessStopped  = "stopped"
	ProcessError    = "error"
)

// Command contains one durable control request.
type Command struct {
	ID          uint64
	TargetKind  Kind
	TargetID    uint64
	Generation  uint64
	Action      string
	Status      CommandStatus
	Outcome     Outcome
	Detail      string
	RequestedMS uint64
	ClaimedMS   uint64
	CompletedMS uint64
}

// Process identifies one exact supervised process generation.
type Process struct {
	TargetKind  Kind
	TargetID    uint64
	Generation  uint64
	PID         int
	Token       string
	Status      string
	Health      string
	HeartbeatMS uint64
}

// Store owns one short-lived connection to central control data.
type Store struct {
	db      *sql.DB
	started bool
	stopped bool
}

// Section 1 - Program Flow

// Init opens central control storage and prepares its schema.
func (s *Store) Init(path string) error {
	// Step 1: validate Store state
	if path == "" || s.started || s.stopped {
		return fmt.Errorf("initialize control Store: invalid state or path")
	}

	// Step 2: open central database
	var dsn = "file:" + filepath.ToSlash(path) +
		"?_txlock=immediate&_pragma=busy_timeout(30000)&_pragma=foreign_keys(1)"
	var err error
	s.db, err = sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("initialize control Store: open database: %v", err)
	}
	s.db.SetMaxOpenConns(1)

	// Step 3: prepare control schema
	_, err = s.db.Exec(controlDDL)
	if err != nil {
		var closeErr = s.db.Close()
		if closeErr != nil {
			return fmt.Errorf(
				"initialize control Store: prepare schema: %v; close database: %v",
				err,
				closeErr,
			)
		}
		return fmt.Errorf("initialize control Store: prepare schema: %v", err)
	}
	s.started = true
	return nil
}

// Enqueue appends one target command for the current or next generation.
func (s *Store) Enqueue(kind Kind, targetID uint64, action string, nowMS uint64) (Command, error) {
	// Step 1: validate command
	if !s.ready() || !validTarget(kind, targetID) || !validAction(action) || nowMS == 0 {
		return Command{}, fmt.Errorf("enqueue command: invalid Store, target, action, or timestamp")
	}

	// Step 2: verify target
	var exists, err = targetExists(s.db, kind, targetID)
	if err != nil {
		return Command{}, err
	}
	if !exists {
		return Command{}, fmt.Errorf("enqueue command: target does not exist")
	}

	// Step 3: select current generation
	var generation uint64
	err = s.db.QueryRow(`
		SELECT COALESCE((
			SELECT CASE
				WHEN status IN ('starting','running','stopping') THEN generation
				ELSE 0
			END
			FROM process_state
			WHERE target_kind = ? AND target_id = ?
		), 0)
	`, kind, targetID).Scan(&generation)
	if err != nil {
		return Command{}, fmt.Errorf("enqueue command: select generation: %v", err)
	}

	// Step 4: persist requested command
	var result sql.Result
	result, err = s.db.Exec(`
		INSERT INTO control_command (
			target_kind, target_id, generation, action, status, requested_ms
		) VALUES (?, ?, ?, ?, ?, ?)
	`, kind, targetID, generation, action, Requested, nowMS)
	if err != nil {
		return Command{}, fmt.Errorf("enqueue command: insert request: %v", err)
	}
	var id int64
	id, err = result.LastInsertId()
	if err != nil {
		return Command{}, fmt.Errorf("enqueue command: read identity: %v", err)
	}
	return Command{
		ID:          uint64(id),
		TargetKind:  kind,
		TargetID:    targetID,
		Generation:  generation,
		Action:      action,
		Status:      Requested,
		RequestedMS: nowMS,
	}, nil
}

// RegisterProcess atomically claims one new process generation.
func (s *Store) RegisterProcess(
	kind Kind,
	targetID uint64,
	pid int,
	token string,
	nowMS,
	staleBeforeMS uint64,
) (Process, error) {
	// Step 1: validate process identity
	if !s.ready() || !validTarget(kind, targetID) || pid <= 0 || token == "" || nowMS == 0 {
		return Process{}, fmt.Errorf("register process: invalid Store, target, identity, or timestamp")
	}

	// Step 2: verify target
	var exists, err = targetExists(s.db, kind, targetID)
	if err != nil {
		return Process{}, err
	}
	if !exists {
		return Process{}, fmt.Errorf("register process: target does not exist")
	}

	// Step 3: reserve generation transaction
	var tx *sql.Tx
	tx, err = s.db.Begin()
	if err != nil {
		return Process{}, fmt.Errorf("register process: begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Step 4: inspect previous generation
	var generation uint64
	var status string
	var heartbeatMS uint64
	err = tx.QueryRow(`
		SELECT generation, status, heartbeat_ms
		FROM process_state
		WHERE target_kind = ? AND target_id = ?
	`, kind, targetID).Scan(&generation, &status, &heartbeatMS)
	if err != nil && err != sql.ErrNoRows {
		return Process{}, fmt.Errorf("register process: inspect generation: %v", err)
	}
	if err == nil && activeProcessStatus(status) && heartbeatMS >= staleBeforeMS {
		return Process{}, fmt.Errorf("register process: target already has active generation %d", generation)
	}
	generation++

	// Step 5: publish process generation
	_, err = tx.Exec(`
		INSERT INTO process_state (
			target_kind, target_id, generation, pid, process_token,
			status, health, heartbeat_ms, updated_ms
		) VALUES (?, ?, ?, ?, ?, ?, 'healthy', ?, ?)
		ON CONFLICT(target_kind, target_id) DO UPDATE SET
			generation = excluded.generation,
			pid = excluded.pid,
			process_token = excluded.process_token,
			status = excluded.status,
			health = excluded.health,
			heartbeat_ms = excluded.heartbeat_ms,
			last_command_id = 0,
			updated_ms = excluded.updated_ms
	`, kind, targetID, generation, pid, token, ProcessStarting, nowMS, nowMS)
	if err != nil {
		return Process{}, fmt.Errorf("register process: publish generation: %v", err)
	}
	err = updateTargetProcess(tx, kind, targetID, pid, token, ProcessStarting, nowMS)
	if err != nil {
		return Process{}, err
	}
	err = tx.Commit()
	if err != nil {
		return Process{}, fmt.Errorf("register process: commit generation: %v", err)
	}
	return Process{
		TargetKind:  kind,
		TargetID:    targetID,
		Generation:  generation,
		PID:         pid,
		Token:       token,
		Status:      ProcessStarting,
		Health:      "healthy",
		HeartbeatMS: nowMS,
	}, nil
}

// ClaimLatest acknowledges superseded requests and claims the newest command.
func (s *Store) ClaimLatest(process Process, nowMS uint64) (Command, bool, error) {
	// Step 1: validate process claim
	if !s.ready() || !validProcess(process) || nowMS == 0 {
		return Command{}, false, fmt.Errorf("claim command: invalid Store, process, or timestamp")
	}

	// Step 2: begin command transaction
	var tx, err = s.db.Begin()
	if err != nil {
		return Command{}, false, fmt.Errorf("claim command: begin transaction: %v", err)
	}
	defer tx.Rollback()
	if err = verifyProcess(tx, process); err != nil {
		return Command{}, false, err
	}

	// Step 3: select newest requested command
	var command Command
	err = tx.QueryRow(`
		SELECT command_id, target_kind, target_id, generation, action,
		       status, requested_ms
		FROM control_command
		WHERE target_kind = ? AND target_id = ? AND status = ?
		  AND (generation = 0 OR generation = ?)
		ORDER BY command_id DESC
		LIMIT 1
	`, process.TargetKind, process.TargetID, Requested, process.Generation).Scan(
		&command.ID,
		&command.TargetKind,
		&command.TargetID,
		&command.Generation,
		&command.Action,
		&command.Status,
		&command.RequestedMS,
	)
	if err == sql.ErrNoRows {
		return Command{}, false, nil
	}
	if err != nil {
		return Command{}, false, fmt.Errorf("claim command: select request: %v", err)
	}

	// Step 4: skip superseded commands
	_, err = tx.Exec(`
		UPDATE control_command
		SET status = ?, outcome = ?, detail = ?, completed_ms = ?
		WHERE target_kind = ? AND target_id = ? AND status = ?
		  AND command_id < ? AND (generation = 0 OR generation = ?)
	`,
		Acknowledged,
		Skipped,
		fmt.Sprintf("superseded by command %d", command.ID),
		nowMS,
		process.TargetKind,
		process.TargetID,
		Requested,
		command.ID,
		process.Generation,
	)
	if err != nil {
		return Command{}, false, fmt.Errorf("claim command: skip superseded requests: %v", err)
	}

	// Step 5: claim newest command
	var result sql.Result
	result, err = tx.Exec(`
		UPDATE control_command
		SET generation = ?, status = ?, claimant = ?, claimed_ms = ?
		WHERE command_id = ? AND status = ?
	`, process.Generation, Claimed, process.Token, nowMS, command.ID, Requested)
	if err != nil {
		return Command{}, false, fmt.Errorf("claim command: reserve request: %v", err)
	}
	var changed int64
	changed, err = result.RowsAffected()
	if err != nil || changed != 1 {
		return Command{}, false, fmt.Errorf("claim command: request reservation lost")
	}
	err = tx.Commit()
	if err != nil {
		return Command{}, false, fmt.Errorf("claim command: commit request: %v", err)
	}
	command.Generation = process.Generation
	command.Status = Claimed
	command.ClaimedMS = nowMS
	return command, true, nil
}

// Acknowledge completes one claimed command.
func (s *Store) Acknowledge(
	process Process,
	commandID uint64,
	outcome Outcome,
	detail string,
	nowMS uint64,
) error {
	if !s.ready() || !validProcess(process) || commandID == 0 ||
		!validOutcome(outcome) || nowMS == 0 {
		return fmt.Errorf("acknowledge command: invalid input")
	}
	var tx, err = s.db.Begin()
	if err != nil {
		return fmt.Errorf("acknowledge command: begin transaction: %v", err)
	}
	defer tx.Rollback()
	var result sql.Result
	result, err = tx.Exec(`
		UPDATE control_command
		SET status = ?, outcome = ?, detail = ?, completed_ms = ?
		WHERE command_id = ? AND generation = ? AND status = ? AND claimant = ?
	`, Acknowledged, outcome, detail, nowMS, commandID, process.Generation, Claimed, process.Token)
	if err != nil {
		return fmt.Errorf("acknowledge command: update request: %v", err)
	}
	var changed int64
	changed, err = result.RowsAffected()
	if err != nil || changed != 1 {
		return fmt.Errorf("acknowledge command: stale or unknown claim")
	}
	_, err = tx.Exec(`
		UPDATE process_state SET last_command_id = ?, updated_ms = ?
		WHERE target_kind = ? AND target_id = ? AND generation = ? AND process_token = ?
	`, commandID, nowMS, process.TargetKind, process.TargetID, process.Generation, process.Token)
	if err != nil {
		return fmt.Errorf("acknowledge command: update process: %v", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("acknowledge command: commit transaction: %v", err)
	}
	return nil
}

// UpdateProcess publishes lifecycle and health for one exact generation.
func (s *Store) UpdateProcess(process Process, status, health string, nowMS uint64) error {
	if !s.ready() || !validProcess(process) || !validProcessStatus(status) || health == "" || nowMS == 0 {
		return fmt.Errorf("update process: invalid input")
	}
	var tx, err = s.db.Begin()
	if err != nil {
		return fmt.Errorf("update process: begin transaction: %v", err)
	}
	defer tx.Rollback()
	var result sql.Result
	result, err = tx.Exec(`
		UPDATE process_state
		SET status = ?, health = ?, heartbeat_ms = ?, updated_ms = ?
		WHERE target_kind = ? AND target_id = ? AND generation = ? AND process_token = ?
	`, status, health, nowMS, nowMS, process.TargetKind, process.TargetID, process.Generation, process.Token)
	if err != nil {
		return fmt.Errorf("update process: publish state: %v", err)
	}
	var changed int64
	changed, err = result.RowsAffected()
	if err != nil || changed != 1 {
		return fmt.Errorf("update process: stale process generation")
	}
	if err = updateTargetStatus(tx, process.TargetKind, process.TargetID, status, nowMS); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("update process: commit transaction: %v", err)
	}
	return nil
}

// Heartbeat refreshes one exact process generation.
func (s *Store) Heartbeat(process Process, nowMS uint64) error {
	if !s.ready() || !validProcess(process) || nowMS == 0 {
		return fmt.Errorf("heartbeat process: invalid input")
	}
	var result, err = s.db.Exec(`
		UPDATE process_state
		SET heartbeat_ms = ?, health = 'healthy', updated_ms = ?
		WHERE target_kind = ? AND target_id = ? AND generation = ? AND process_token = ?
	`, nowMS, nowMS, process.TargetKind, process.TargetID, process.Generation, process.Token)
	if err != nil {
		return fmt.Errorf("heartbeat process: publish heartbeat: %v", err)
	}
	var changed int64
	changed, err = result.RowsAffected()
	if err != nil || changed != 1 {
		return fmt.Errorf("heartbeat process: stale process generation")
	}
	return nil
}

// Stop closes the central control connection.
func (s *Store) Stop() error {
	if s.stopped {
		return nil
	}
	s.stopped = true
	if !s.started {
		return nil
	}
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("stop control Store: %v", err)
	}
	return nil
}

// Section 2 - Domain Helpers

func targetExists(db *sql.DB, kind Kind, targetID uint64) (bool, error) {
	var query string
	if kind == Bot {
		query = `SELECT COUNT(*) FROM bot WHERE bot_id = ?`
	} else {
		query = `SELECT COUNT(*) FROM sweep WHERE sweep_id = ?`
	}
	var count int
	if err := db.QueryRow(query, targetID).Scan(&count); err != nil {
		return false, fmt.Errorf("read control target: %v", err)
	}
	return count == 1, nil
}

func updateTargetProcess(
	tx *sql.Tx,
	kind Kind,
	targetID uint64,
	pid int,
	token,
	status string,
	nowMS uint64,
) error {
	var timestamp = time.UnixMilli(int64(nowMS)).UTC().Format(time.RFC3339Nano)
	var err error
	if kind == Bot {
		_, err = tx.Exec(`
			UPDATE bot
			SET status = ?, process_pid = ?, process_create_time = ?,
			    updated_at = ?, started_at = COALESCE(started_at, ?), ended_at = NULL
			WHERE bot_id = ?
		`, status, pid, token, timestamp, timestamp, targetID)
	} else {
		_, err = tx.Exec(`
			UPDATE sweep
			SET status = ?, process_status = ?, process_pid = ?, process_create_time = ?,
			    updated_at = ?, started_at = COALESCE(started_at, ?), ended_at = NULL
			WHERE sweep_id = ?
		`, status, status, pid, token, timestamp, timestamp, targetID)
	}
	if err != nil {
		return fmt.Errorf("register process: update target lifecycle: %v", err)
	}
	return nil
}

func updateTargetStatus(
	tx *sql.Tx,
	kind Kind,
	targetID uint64,
	status string,
	nowMS uint64,
) error {
	var timestamp = time.UnixMilli(int64(nowMS)).UTC().Format(time.RFC3339Nano)
	var terminal = status == ProcessStopped || status == ProcessError
	var err error
	if kind == Bot {
		_, err = tx.Exec(`
			UPDATE bot
			SET status = ?, updated_at = ?,
			    ended_at = CASE WHEN ? THEN ? ELSE ended_at END
			WHERE bot_id = ?
		`, status, timestamp, terminal, timestamp, targetID)
	} else {
		_, err = tx.Exec(`
			UPDATE sweep
			SET status = ?, process_status = ?, updated_at = ?,
			    ended_at = CASE WHEN ? THEN ? ELSE ended_at END
			WHERE sweep_id = ?
		`, status, status, timestamp, terminal, timestamp, targetID)
	}
	if err != nil {
		return fmt.Errorf("update process: update target lifecycle: %v", err)
	}
	return nil
}

func verifyProcess(tx *sql.Tx, process Process) error {
	var count int
	var err = tx.QueryRow(`
		SELECT COUNT(*) FROM process_state
		WHERE target_kind = ? AND target_id = ? AND generation = ? AND process_token = ?
	`, process.TargetKind, process.TargetID, process.Generation, process.Token).Scan(&count)
	if err != nil {
		return fmt.Errorf("claim command: verify process: %v", err)
	}
	if count != 1 {
		return fmt.Errorf("claim command: stale process generation")
	}
	return nil
}

func (s *Store) ready() bool {
	return s.started && !s.stopped
}

func validTarget(kind Kind, targetID uint64) bool {
	return (kind == Bot || kind == Sweep) && targetID != 0
}

func validAction(action string) bool {
	switch action {
	case Start, Pause, Resume, Stop:
		return true
	default:
		return false
	}
}

func validOutcome(outcome Outcome) bool {
	return outcome == Processed || outcome == Skipped || outcome == Rejected
}

func validProcess(process Process) bool {
	return validTarget(process.TargetKind, process.TargetID) &&
		process.Generation != 0 && process.PID > 0 && process.Token != ""
}

func validProcessStatus(status string) bool {
	switch status {
	case ProcessStarting, ProcessRunning, ProcessStopping, ProcessStopped, ProcessError:
		return true
	default:
		return false
	}
}

func activeProcessStatus(status string) bool {
	return status == ProcessStarting || status == ProcessRunning || status == ProcessStopping
}

// Section 3 - Generic Helpers

const controlDDL = `
CREATE TABLE IF NOT EXISTS control_command (
	command_id INTEGER PRIMARY KEY AUTOINCREMENT,
	target_kind TEXT NOT NULL CHECK(target_kind IN ('bot','sweep')),
	target_id INTEGER NOT NULL,
	generation INTEGER NOT NULL,
	action TEXT NOT NULL CHECK(action IN ('start','pause','resume','stop')),
	status TEXT NOT NULL CHECK(status IN ('requested','claimed','acknowledged')),
	outcome TEXT NOT NULL DEFAULT '' CHECK(outcome IN ('','processed','skipped','rejected')),
	detail TEXT NOT NULL DEFAULT '',
	claimant TEXT NOT NULL DEFAULT '',
	requested_ms INTEGER NOT NULL,
	claimed_ms INTEGER NOT NULL DEFAULT 0,
	completed_ms INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS control_command_target_status
	ON control_command(target_kind, target_id, status, command_id);
CREATE TABLE IF NOT EXISTS process_state (
	target_kind TEXT NOT NULL CHECK(target_kind IN ('bot','sweep')),
	target_id INTEGER NOT NULL,
	generation INTEGER NOT NULL,
	pid INTEGER NOT NULL,
	process_token TEXT NOT NULL,
	status TEXT NOT NULL CHECK(status IN ('starting','running','stopping','stopped','error')),
	health TEXT NOT NULL,
	heartbeat_ms INTEGER NOT NULL,
	last_command_id INTEGER NOT NULL DEFAULT 0,
	updated_ms INTEGER NOT NULL,
	PRIMARY KEY(target_kind, target_id)
);
`
