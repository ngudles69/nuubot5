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
}

func TestLoadAppRejectsOldBtRunnerKey(t *testing.T) {
	var sourcePath = filepath.Join("..", "..", "workspace", "config", "config.toml")
	var contents, err = os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read App Config: %v", err)
	}
	var oldConfig = strings.Replace(string(contents), "[btbot]", "[btrunner]", 1)
	var path = filepath.Join(t.TempDir(), "config.toml")
	if err = os.WriteFile(path, []byte(oldConfig), 0o600); err != nil {
		t.Fatalf("write old App Config: %v", err)
	}

	var _, loadErr = LoadApp(path)
	if loadErr == nil {
		t.Fatal("old btrunner Config key was accepted")
	}
	if !strings.Contains(loadErr.Error(), "btrunner") {
		t.Fatalf("error = %q, want old key", loadErr)
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

func writeCredentials(t *testing.T, contents string) string {
	t.Helper()
	var path = filepath.Join(t.TempDir(), "credentials.toml")
	var err = os.WriteFile(path, []byte(contents), 0o600)
	if err != nil {
		t.Fatalf("write credentials failed: %v", err)
	}
	return path
}
