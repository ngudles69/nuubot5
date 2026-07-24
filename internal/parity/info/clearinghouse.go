package info

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nuubot/internal/hyperliquid"
)

// Section 1 - Program Flow

// ClearinghouseState captures one live clearinghouse payload and its translation.
func ClearinghouseState(
	ctx context.Context,
	client clearinghouseClient,
	target Target,
	capture string,
) (Result, error) {
	// build evidence path
	var evidenceDir = filepath.Join(
		target.EvidenceDir,
		"info",
		"clearinghouse-state",
		capture,
		target.Network,
		target.Account,
	)

	// reserve evidence directory
	var err = reserveEvidenceDir(evidenceDir)
	if err != nil {
		return Result{}, fmt.Errorf("probe clearinghouse state: %w", err)
	}

	// request clearinghouse payload
	var started = time.Now()
	var response hyperliquid.Response
	response, err = client.ClearinghouseStatePayload(ctx, target.Address)
	var duration = time.Since(started)
	if err != nil {
		var evidenceErr = writeClearinghouseFailure(
			evidenceDir,
			target,
			response,
			duration,
			err,
		)
		if evidenceErr != nil {
			err = errors.Join(err, evidenceErr)
		}
		return Result{}, fmt.Errorf("probe clearinghouse state: %w", err)
	}

	// preserve raw payload
	err = writeRawPayload(evidenceDir, response.Payload)
	if err != nil {
		return Result{}, fmt.Errorf("probe clearinghouse state: %w", err)
	}

	// decode clearinghouse payload
	var state hyperliquid.AccountState
	state, err = hyperliquid.DecodeClearinghouseState(response.Payload)
	if err != nil {
		var evidenceErr = writeClearinghouseReport(
			evidenceDir,
			target,
			response,
			duration,
			"raw.json",
			err,
		)
		if evidenceErr != nil {
			err = errors.Join(err, evidenceErr)
		}
		return Result{}, fmt.Errorf("probe clearinghouse state: %w", err)
	}

	// write clearinghouse evidence
	err = writeClearinghouseSuccess(
		evidenceDir,
		target,
		response,
		state,
		duration,
	)
	if err != nil {
		return Result{}, fmt.Errorf("probe clearinghouse state: %w", err)
	}
	return Result{
		EvidenceDir:       evidenceDir,
		Duration:          duration,
		Positions:         len(state.Positions),
		Equity:            state.Margin.Equity,
		MarginUsed:        state.Margin.MarginUsed,
		MaintenanceMargin: state.MaintenanceMargin,
		Withdrawable:      state.Withdrawable,
	}, nil
}

// Section 2 - Domain Helpers

func reserveEvidenceDir(evidenceDir string) error {
	var err = os.MkdirAll(filepath.Dir(evidenceDir), 0o755)
	if err != nil {
		return fmt.Errorf("create evidence parent %s: %w", evidenceDir, err)
	}
	err = os.Mkdir(evidenceDir, 0o755)
	if err != nil {
		return fmt.Errorf("create evidence directory %s: %w", evidenceDir, err)
	}
	return nil
}

func writeClearinghouseFailure(
	evidenceDir string,
	target Target,
	response hyperliquid.Response,
	duration time.Duration,
	probeErr error,
) error {
	var rawPayload string
	if response.Payload != nil {
		var err = writeRawPayload(evidenceDir, response.Payload)
		if err != nil {
			return err
		}
		rawPayload = "raw.json"
	}
	return writeClearinghouseReport(
		evidenceDir,
		target,
		response,
		duration,
		rawPayload,
		probeErr,
	)
}

func writeClearinghouseSuccess(
	evidenceDir string,
	target Target,
	response hyperliquid.Response,
	state hyperliquid.AccountState,
	duration time.Duration,
) error {
	var normalizedPath = filepath.Join(evidenceDir, "normalized.json")
	var err = writeJSON(normalizedPath, state)
	if err != nil {
		return err
	}
	return writeClearinghouseReport(
		evidenceDir,
		target,
		response,
		duration,
		"raw.json",
		nil,
	)
}

func writeRawPayload(evidenceDir string, payload []byte) error {
	var rawPath = filepath.Join(evidenceDir, "raw.json")
	var err = os.WriteFile(rawPath, payload, 0o644)
	if err != nil {
		return fmt.Errorf("write raw payload %s: %w", rawPath, err)
	}
	return nil
}

func writeClearinghouseReport(
	evidenceDir string,
	target Target,
	response hyperliquid.Response,
	duration time.Duration,
	rawPayload string,
	probeErr error,
) error {
	var errorText string
	var normalized string
	if probeErr != nil {
		errorText = probeErr.Error()
	} else {
		normalized = "normalized.json"
	}
	var reportPath = filepath.Join(evidenceDir, "report.json")
	return writeJSON(reportPath, report{
		Network:    target.Network,
		Account:    target.Account,
		Operation:  "clearinghouse-state",
		CapturedAt: time.Now().UTC().Format(time.RFC3339Nano),
		StatusCode: response.StatusCode,
		DurationMS: duration.Milliseconds(),
		RawPayload: rawPayload,
		Normalized: normalized,
		Error:      errorText,
	})
}

// Section 3 - Generic Helpers

func writeJSON(path string, value any) error {
	var payload, err = json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode json %s: %v", path, err)
	}
	payload = append(payload, '\n')
	err = os.WriteFile(path, payload, 0o644)
	if err != nil {
		return fmt.Errorf("write json %s: %w", path, err)
	}
	return nil
}
