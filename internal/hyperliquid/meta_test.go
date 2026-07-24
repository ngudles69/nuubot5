package hyperliquid

import "testing"

// Section 1 - Program Flow

func TestDecodePerpetualMeta(t *testing.T) {
	var payload = []byte(`{
		"universe":[{
			"name":"BTC",
			"szDecimals":5,
			"maxLeverage":40,
			"marginTableId":20
		}],
		"marginTables":[[
			20,
			{"description":"","marginTiers":[
				{"lowerBound":"0.0","maxLeverage":40}
			]}
		]],
		"collateralToken":0
	}`)
	var actual, err = DecodePerpetualMeta(payload)
	if err != nil {
		t.Fatalf("decode perpetual meta: %v", err)
	}
	if len(actual.Universe) != 1 || actual.Universe[0].Name != "BTC" ||
		actual.Universe[0].SizeDecimals != 5 ||
		actual.Universe[0].MaxLeverage != 40 {
		t.Fatalf("unexpected universe: %+v", actual.Universe)
	}
	if len(actual.MarginTables) != 1 ||
		len(actual.MarginTables[0].Tiers) != 1 ||
		actual.MarginTables[0].Tiers[0].LowerBound != "0.0" {
		t.Fatalf("unexpected margin tables: %+v", actual.MarginTables)
	}
}

func TestDecodePerpetualMetaRejectsMalformedTables(t *testing.T) {
	var payload = []byte(`{
		"universe":[{"name":"BTC","szDecimals":5,"maxLeverage":40}],
		"marginTables":[[20]]
	}`)
	var _, err = DecodePerpetualMeta(payload)
	if err == nil {
		t.Fatal("actual error nil, expected malformed margin table rejection")
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
