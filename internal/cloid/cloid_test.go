package cloid

import "testing"

// Section 1 - Program Flow

func TestEncodeDecodeRoundTrip(t *testing.T) {
	var fields = Fields{
		BotCycleID: 1,
		SymbolID:   2,
		Exchange:   1,
		Network:    2,
		Side:       1,
		ReduceOnly: true,
		Purpose:    3,
		TradeNo:    4,
		BatchNo:    5,
		OrderLevel: 6,
		TimestampS: 7,
	}
	var encoded, err = Encode(fields)
	if err != nil {
		t.Fatalf("encode cloid: %v", err)
	}
	var expected = "0x00000100021b030000200a0300000007"
	if encoded != expected {
		t.Fatalf("actual cloid %q, expected %q", encoded, expected)
	}
	var decoded Fields
	decoded, err = Decode(encoded)
	if err != nil {
		t.Fatalf("decode cloid: %v", err)
	}
	if decoded != fields {
		t.Fatalf("actual fields %+v, expected %+v", decoded, fields)
	}
}

func TestCLOIDRejectsInvalidIdentity(t *testing.T) {
	var fields = Fields{TradeNo: 1, BatchNo: 1}
	fields.BatchNo = 1001
	var _, err = Encode(fields)
	if err == nil {
		t.Fatal("actual error nil, expected batch range rejection")
	}
	fields.BatchNo = 1
	fields.OrderLevel = 1024
	_, err = Encode(fields)
	if err == nil {
		t.Fatal("actual error nil, expected level range rejection")
	}
	_, err = Decode("0x00000000000000000000000000000000")
	if err == nil {
		t.Fatal("actual error nil, expected decoded identity rejection")
	}
	_, err = Decode("0X00000000000000000000000000000001")
	if err == nil {
		t.Fatal("actual error nil, expected shape rejection")
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
