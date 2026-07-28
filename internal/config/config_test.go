package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Section 1 - Program Flow

func TestLoadIsIdempotent(t *testing.T) {
	var path = filepath.Join("..", "..", "workspace", "config", "config.toml")
	var first, err = LoadApp(path)
	if err != nil {
		t.Fatalf("first load failed: %v", err)
	}
	var second App
	second, err = LoadApp(path)
	if err != nil {
		t.Fatalf("second load failed: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("second config differs from first")
	}
	if first.Hyperliquid.MinOrderNotionalUSDC != 11 {
		t.Fatalf(
			"minimum order notional actual %d, expected 11",
			first.Hyperliquid.MinOrderNotionalUSDC,
		)
	}
	var expectedBacktest = Runtime{
		ControllerIntervalMS:    1000,
		ReconIntervalMS:         10000,
		ReconSweepIntervalMS:    60000,
		TelemetryIntervalMS:     10000,
		TelemetryWriteOnCollect: false,
	}
	if first.Backtest != expectedBacktest {
		t.Fatalf("unexpected Backtest Config: %+v", first.Backtest)
	}
	var expectedLive = expectedBacktest
	expectedLive.TelemetryWriteOnCollect = true
	if first.Live != expectedLive {
		t.Fatalf("unexpected Live Config: %+v", first.Live)
	}
}

func TestLoadAppRejectsOldControllerKey(t *testing.T) {
	var sourcePath = filepath.Join("..", "..", "workspace", "config", "config.toml")
	var contents, err = os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read App Config: %v", err)
	}
	var oldConfig = strings.Replace(string(contents), "[live]", "[controller]", 1)
	var path = filepath.Join(t.TempDir(), "config.toml")
	if err = os.WriteFile(path, []byte(oldConfig), 0o600); err != nil {
		t.Fatalf("write old App Config: %v", err)
	}

	var _, loadErr = LoadApp(path)
	if loadErr == nil {
		t.Fatal("old Controller Config key was accepted")
	}
	if !strings.Contains(loadErr.Error(), "controller") {
		t.Fatalf("error = %q, want old key", loadErr)
	}
}

func TestLoadAppRejectsReconAtControllerCadence(t *testing.T) {
	var path = writeAppConfig(t, func(contents string) string {
		return strings.Replace(contents, "recon_interval_ms = 10000", "recon_interval_ms = 1000", 1)
	})
	var _, err = LoadApp(path)
	if err == nil || !strings.Contains(err.Error(), "recon_interval_ms") {
		t.Fatalf("error = %v, want Recon interval rejection", err)
	}
}

func TestLoadAppRejectsSweepAtReconCadence(t *testing.T) {
	var path = writeAppConfig(t, func(contents string) string {
		return strings.Replace(contents, "recon_sweep_interval_ms = 60000", "recon_sweep_interval_ms = 10000", 1)
	})
	var _, err = LoadApp(path)
	if err == nil || !strings.Contains(err.Error(), "recon_sweep_interval_ms") {
		t.Fatalf("error = %v, want Recon sweep interval rejection", err)
	}
}

func TestLoadAppRejectsZeroProcessPoll(t *testing.T) {
	var path = writeAppConfig(t, func(contents string) string {
		return strings.Replace(contents, "poll_seconds = 10", "poll_seconds = 0", 1)
	})
	var _, err = LoadApp(path)
	if err == nil || !strings.Contains(err.Error(), "process.poll_seconds") {
		t.Fatalf("error = %v, want process poll rejection", err)
	}
}

func TestLoadAppRejectsUnresponsiveAtPollCadence(t *testing.T) {
	var path = writeAppConfig(t, func(contents string) string {
		return strings.Replace(
			contents,
			"unresponsive_seconds = 30",
			"unresponsive_seconds = 10",
			1,
		)
	})
	var _, err = LoadApp(path)
	if err == nil || !strings.Contains(err.Error(), "process.unresponsive_seconds") {
		t.Fatalf("error = %v, want unresponsive cadence rejection", err)
	}
}

func TestLoadAppRejectsLiveWithoutWriteOnCollect(t *testing.T) {
	var path = writeAppConfig(t, func(contents string) string {
		return strings.Replace(
			contents,
			"telemetry_write_on_collect = true",
			"telemetry_write_on_collect = false",
			1,
		)
	})
	var _, err = LoadApp(path)
	if err == nil || !strings.Contains(err.Error(), "live.telemetry_write_on_collect") {
		t.Fatalf("error = %v, want Live telemetry policy rejection", err)
	}
}

func TestLoadAppRejectsBacktestWriteOnCollect(t *testing.T) {
	var path = writeAppConfig(t, func(contents string) string {
		return strings.Replace(
			contents,
			"telemetry_write_on_collect = false",
			"telemetry_write_on_collect = true",
			1,
		)
	})
	var _, err = LoadApp(path)
	if err == nil || !strings.Contains(err.Error(), "backtest.telemetry_write_on_collect") {
		t.Fatalf("error = %v, want Backtest telemetry policy rejection", err)
	}
}

func TestLoadCredentialsIsIdempotent(t *testing.T) {
	var path = writeCredentials(t, `
[datastore]
kind = "test"
host = "127.0.0.1"
port = 1
database = "test"
user = "test"
password = "test"

[[hyperliquid.accounts]]
network = "testnet"
name = "test"
address = "test"
api_key = "test"
`)
	var first, err = LoadCredentials(path)
	if err != nil {
		t.Fatalf("first load failed: %v", err)
	}
	var second Credentials
	second, err = LoadCredentials(path)
	if err != nil {
		t.Fatalf("second load failed: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("second credentials differ from first")
	}
}

func TestLoadCredentialsRejectsMalformedTOML(t *testing.T) {
	var path = writeCredentials(t, "[datastore")
	var _, err = LoadCredentials(path)
	if err == nil {
		t.Fatalf("malformed credentials loaded")
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers

func writeAppConfig(t *testing.T, change func(string) string) string {
	t.Helper()
	var sourcePath = filepath.Join("..", "..", "workspace", "config", "config.toml")
	var contents, err = os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read App Config: %v", err)
	}
	var path = filepath.Join(t.TempDir(), "config.toml")
	if err = os.WriteFile(path, []byte(change(string(contents))), 0o600); err != nil {
		t.Fatalf("write App Config: %v", err)
	}
	return path
}

func writeCredentials(t *testing.T, contents string) string {
	t.Helper()
	var path = filepath.Join(t.TempDir(), "credentials.toml")
	var err = os.WriteFile(path, []byte(contents), 0o600)
	if err != nil {
		t.Fatalf("write credentials failed: %v", err)
	}
	return path
}
