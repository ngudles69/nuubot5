package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Credentials contains local datastore and exchange credentials.
type Credentials struct {
	Datastore   DatastoreCredentials   `toml:"datastore"`
	Hyperliquid HyperliquidCredentials `toml:"hyperliquid"`
}

// DatastoreCredentials contains datastore connection secrets.
type DatastoreCredentials struct {
	Kind     string `toml:"kind"`
	Host     string `toml:"host"`
	Port     uint16 `toml:"port"`
	Database string `toml:"database"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

// HyperliquidCredentials contains configured Hyperliquid accounts.
type HyperliquidCredentials struct {
	Accounts []HyperliquidAccountCredentials `toml:"accounts"`
}

// HyperliquidAccountCredentials contains one Hyperliquid account secret.
type HyperliquidAccountCredentials struct {
	Network string `toml:"network"`
	Name    string `toml:"name"`
	Address string `toml:"address"`
	APIKey  string `toml:"api_key"`
}

// Section 1 - Program Flow

// LoadCredentials decodes one credentials file without authenticating accounts.
func LoadCredentials(path string) (Credentials, error) {
	// decode toml
	var credentials Credentials
	var _, err = toml.DecodeFile(path, &credentials)
	if err != nil {
		return credentials, fmt.Errorf("load credentials %s: invalid toml", path)
	}
	return credentials, nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
