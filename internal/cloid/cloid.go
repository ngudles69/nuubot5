// Package cloid encodes Nuubot Order identity into Hyperliquid CLOIDs.
package cloid

import (
	"fmt"
	"math/big"
	"strings"
)

// Fields contains one canonical 128-bit Order identity.
type Fields struct {
	BotCycleID uint32
	SymbolID   uint16
	Exchange   uint8
	Network    uint8
	Side       uint8
	ReduceOnly bool
	Purpose    uint8
	TradeNo    uint32
	BatchNo    uint16
	OrderPos   uint16
	TimestampS uint32
}

// Section 1 - Program Flow

// Encode validates and packs one Order identity.
func Encode(fields Fields) (string, error) {
	// validate fixed ranges
	var err = validateFields(fields)
	if err != nil {
		return "", err
	}

	// pack fixed fields
	var raw = new(big.Int)
	appendField(raw, uint64(fields.BotCycleID), 24)
	appendField(raw, uint64(fields.SymbolID), 16)
	appendField(raw, uint64(fields.Exchange), 4)
	appendField(raw, uint64(fields.Network), 2)
	appendField(raw, uint64(fields.Side), 1)
	if fields.ReduceOnly {
		appendField(raw, 1, 1)
	} else {
		appendField(raw, 0, 1)
	}
	appendField(raw, uint64(fields.Purpose), 8)
	appendField(raw, uint64(fields.TradeNo), 21)
	appendField(raw, uint64(fields.BatchNo), 10)
	appendField(raw, uint64(fields.OrderPos), 10)
	appendField(raw, uint64(fields.TimestampS), 31)
	return fmt.Sprintf("0x%032x", raw), nil
}

// Decode validates and unpacks one Hyperliquid CLOID.
func Decode(value string) (Fields, error) {
	// validate exchange shape
	if len(value) != 34 || !strings.HasPrefix(value, "0x") {
		return Fields{}, fmt.Errorf("decode cloid: expected 0x and 32 hexadecimal characters")
	}
	var raw, ok = new(big.Int).SetString(value[2:], 16)
	if !ok {
		return Fields{}, fmt.Errorf("decode cloid: expected 0x and 32 hexadecimal characters")
	}

	// unpack fixed fields
	var fields = Fields{
		BotCycleID: uint32(readField(raw, 104, 24)),
		SymbolID:   uint16(readField(raw, 88, 16)),
		Exchange:   uint8(readField(raw, 84, 4)),
		Network:    uint8(readField(raw, 82, 2)),
		Side:       uint8(readField(raw, 81, 1)),
		ReduceOnly: readField(raw, 80, 1) == 1,
		Purpose:    uint8(readField(raw, 72, 8)),
		TradeNo:    uint32(readField(raw, 51, 21)),
		BatchNo:    uint16(readField(raw, 41, 10)),
		OrderPos:   uint16(readField(raw, 31, 10)),
		TimestampS: uint32(readField(raw, 0, 31)),
	}
	var err = validateFields(fields)
	if err != nil {
		return Fields{}, fmt.Errorf("decode cloid: %w", err)
	}
	return fields, nil
}

// Section 2 - Domain Helpers

func validateFields(fields Fields) error {
	if fields.BotCycleID > 0x00ffffff {
		return fmt.Errorf("validate cloid: botcycle id exceeds 24 bits")
	}
	if fields.Exchange > 0x0f || fields.Network > 0x03 || fields.Side > 1 {
		return fmt.Errorf("validate cloid: exchange, network, or side exceeds its fixed range")
	}
	if fields.TradeNo == 0 || fields.TradeNo > 0x001fffff {
		return fmt.Errorf("validate cloid: trade number must be from 1 to 2097151")
	}
	if fields.BatchNo == 0 || fields.BatchNo > 1000 {
		return fmt.Errorf("validate cloid: batch number must be from 1 to 1000")
	}
	if fields.OrderPos == 0 || fields.OrderPos > 1000 {
		return fmt.Errorf("validate cloid: order position must be from 1 to 1000")
	}
	if fields.TimestampS > 0x7fffffff {
		return fmt.Errorf("validate cloid: timestamp exceeds 31 bits")
	}
	return nil
}

// Section 3 - Generic Helpers

func appendField(raw *big.Int, value uint64, bits uint) {
	raw.Lsh(raw, bits)
	raw.Or(raw, new(big.Int).SetUint64(value))
}

func readField(raw *big.Int, shift uint, bits uint) uint64 {
	var shifted = new(big.Int).Rsh(new(big.Int).Set(raw), shift)
	var mask = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), bits), big.NewInt(1))
	return shifted.And(shifted, mask).Uint64()
}
