package hyperliquid

import (
	"context"
	"encoding/json"
	"fmt"
)

// PerpetualAsset contains one admitted Hyperliquid perpetual instrument.
type PerpetualAsset struct {
	Name          string
	SizeDecimals  uint32
	MaxLeverage   uint32
	MarginTableID uint32
	OnlyIsolated  bool
	IsDelisted    bool
	MarginMode    string
	Raw           string
}

// MarginTier contains one admitted Hyperliquid margin tier.
type MarginTier struct {
	LowerBound  string
	MaxLeverage uint32
}

// MarginTable contains one admitted Hyperliquid margin table.
type MarginTable struct {
	ID          uint32
	Description string
	Tiers       []MarginTier
}

// PerpetualMeta contains one complete admitted perpetual Meta response.
type PerpetualMeta struct {
	Universe        []PerpetualAsset
	MarginTables    []MarginTable
	CollateralToken uint32
}

type perpetualAssetResponse struct {
	Name          string `json:"name"`
	SizeDecimals  uint32 `json:"szDecimals"`
	MaxLeverage   uint32 `json:"maxLeverage"`
	MarginTableID uint32 `json:"marginTableId"`
	OnlyIsolated  bool   `json:"onlyIsolated"`
	IsDelisted    bool   `json:"isDelisted"`
	MarginMode    string `json:"marginMode"`
}

type marginTableResponse struct {
	Description string               `json:"description"`
	MarginTiers []marginTierResponse `json:"marginTiers"`
}

type marginTierResponse struct {
	LowerBound  string `json:"lowerBound"`
	MaxLeverage uint32 `json:"maxLeverage"`
}

// Section 1 - Program Flow

// PerpetualMeta reads the complete default perpetual Meta dataset.
func (c *Client) PerpetualMeta(ctx context.Context) (PerpetualMeta, error) {
	// request perpetual meta payload
	var response, err = c.PerpetualMetaPayload(ctx)
	if err != nil {
		return PerpetualMeta{}, err
	}

	// decode perpetual meta payload
	var result PerpetualMeta
	result, err = DecodePerpetualMeta(response.Payload)
	if err != nil {
		return PerpetualMeta{}, fmt.Errorf("read Hyperliquid perpetual meta: %w", err)
	}
	return result, nil
}

// PerpetualMetaPayload reads one raw default perpetual Meta payload.
func (c *Client) PerpetualMetaPayload(ctx context.Context) (Response, error) {
	// post request payload
	var response, err = c.Post(ctx, "/info", []byte(`{"type":"meta"}`))
	if err != nil {
		return response, fmt.Errorf("read Hyperliquid perpetual meta: %w", err)
	}
	return response, nil
}

// Section 2 - Domain Helpers

// DecodePerpetualMeta validates and translates one raw perpetual Meta payload.
func DecodePerpetualMeta(payload []byte) (PerpetualMeta, error) {
	// decode response payload
	var response struct {
		Universe        []json.RawMessage `json:"universe"`
		MarginTables    []json.RawMessage `json:"marginTables"`
		CollateralToken uint32            `json:"collateralToken"`
	}
	var err = json.Unmarshal(payload, &response)
	if err != nil {
		return PerpetualMeta{}, fmt.Errorf("decode perpetual meta payload: %v", err)
	}
	if len(response.Universe) == 0 {
		return PerpetualMeta{}, fmt.Errorf("validate perpetual meta payload: universe is empty")
	}
	if response.MarginTables == nil {
		return PerpetualMeta{}, fmt.Errorf("validate perpetual meta payload: marginTables must be an array")
	}

	// translate universe
	var universe = make([]PerpetualAsset, 0, len(response.Universe))
	var names = make(map[string]struct{}, len(response.Universe))
	for index, raw := range response.Universe {
		var asset perpetualAssetResponse
		err = json.Unmarshal(raw, &asset)
		if err != nil {
			return PerpetualMeta{}, fmt.Errorf(
				"decode perpetual meta payload: universe[%d]: %v",
				index,
				err,
			)
		}
		if asset.Name == "" || asset.SizeDecimals > 6 || asset.MaxLeverage == 0 {
			return PerpetualMeta{}, fmt.Errorf(
				"validate perpetual meta payload: invalid universe[%d]",
				index,
			)
		}
		if _, exists := names[asset.Name]; exists {
			return PerpetualMeta{}, fmt.Errorf(
				"validate perpetual meta payload: duplicate symbol %q",
				asset.Name,
			)
		}
		names[asset.Name] = struct{}{}
		universe = append(universe, PerpetualAsset{
			Name:          asset.Name,
			SizeDecimals:  asset.SizeDecimals,
			MaxLeverage:   asset.MaxLeverage,
			MarginTableID: asset.MarginTableID,
			OnlyIsolated:  asset.OnlyIsolated,
			IsDelisted:    asset.IsDelisted,
			MarginMode:    asset.MarginMode,
			Raw:           string(raw),
		})
	}

	// translate margin tables
	var tables = make([]MarginTable, 0, len(response.MarginTables))
	var tableIDs = make(map[uint32]struct{}, len(response.MarginTables))
	for index, raw := range response.MarginTables {
		var tuple []json.RawMessage
		err = json.Unmarshal(raw, &tuple)
		if err != nil || len(tuple) != 2 {
			return PerpetualMeta{}, fmt.Errorf(
				"validate perpetual meta payload: invalid marginTables[%d]",
				index,
			)
		}
		var id uint32
		err = json.Unmarshal(tuple[0], &id)
		if err != nil {
			return PerpetualMeta{}, fmt.Errorf(
				"decode perpetual meta payload: marginTables[%d] id: %v",
				index,
				err,
			)
		}
		if _, exists := tableIDs[id]; exists {
			return PerpetualMeta{}, fmt.Errorf(
				"validate perpetual meta payload: duplicate margin table %d",
				id,
			)
		}
		var table marginTableResponse
		err = json.Unmarshal(tuple[1], &table)
		if err != nil || len(table.MarginTiers) == 0 {
			return PerpetualMeta{}, fmt.Errorf(
				"validate perpetual meta payload: invalid marginTables[%d] data",
				index,
			)
		}
		var tiers = make([]MarginTier, 0, len(table.MarginTiers))
		for tierIndex, tier := range table.MarginTiers {
			var lowerBound, parseErr = parseDecimal("margin tier lowerBound", tier.LowerBound)
			if parseErr != nil || lowerBound.IsNegative() || tier.MaxLeverage == 0 {
				return PerpetualMeta{}, fmt.Errorf(
					"validate perpetual meta payload: invalid marginTables[%d] tier %d",
					index,
					tierIndex,
				)
			}
			tiers = append(tiers, MarginTier{
				LowerBound:  tier.LowerBound,
				MaxLeverage: tier.MaxLeverage,
			})
		}
		tableIDs[id] = struct{}{}
		tables = append(tables, MarginTable{
			ID:          id,
			Description: table.Description,
			Tiers:       tiers,
		})
	}
	return PerpetualMeta{
		Universe:        universe,
		MarginTables:    tables,
		CollateralToken: response.CollateralToken,
	}, nil
}

// Section 3 - Generic Helpers
