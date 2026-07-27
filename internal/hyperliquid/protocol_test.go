package hyperliquid

import "testing"

// Section 1 - Program Flow

func TestProtocolDecodesWorkingSimulatorResponses(t *testing.T) {
	var submit, err = DecodeSubmitResponse([]byte(
		`{"status":"ok","response":{"type":"order","data":{"statuses":[` +
			`{"resting":{"cloid":"0x01"}},` +
			`{"filled":{"totalSz":"1","avgPx":"100","oid":8,"fee":"0"}},` +
			`"waitingForFill"]}}}`,
	))
	if err != nil {
		t.Fatalf("decode submit response: %v", err)
	}
	if len(submit.Statuses) != 3 || submit.Statuses[0].CLOID != "0x01" ||
		submit.Statuses[1].Fee == nil || *submit.Statuses[1].Fee != "0" ||
		submit.Statuses[2].Kind != "waitingForFill" {
		t.Fatalf("unexpected submit response: %+v", submit)
	}

	var open []OpenOrder
	open, err = DecodeOpenOrders([]byte(
		`[{"coin":"BTC","limitPx":"100","oid":0,"side":"B","sz":"1",` +
			`"timestamp":10,"cloid":"0x01"}]`,
	))
	if err != nil {
		t.Fatalf("decode CLOID-only open Order: %v", err)
	}
	if len(open) != 1 || open[0].CLOID != "0x01" {
		t.Fatalf("unexpected CLOID-only open Order: %+v", open)
	}

	var fills []Fill
	fills, err = DecodeFills([]byte(
		`[{"coin":"BTC","px":"100","sz":"1","side":"B","time":10,` +
			`"oid":7,"tid":9,"startPosition":"0","closedPnl":"0",` +
			`"dir":"Open Long","crossed":true},` +
			`{"coin":"BTC","px":"101","sz":"1","side":"A","time":11,` +
			`"oid":8,"tid":10,"startPosition":"1","closedPnl":"1",` +
			`"dir":"Close Long","crossed":true,"fee":"0"}]`,
	))
	if err != nil {
		t.Fatalf("decode Fill response: %v", err)
	}
	if fills[0].Fee != nil || fills[1].Fee == nil || *fills[1].Fee != "0" {
		t.Fatalf("Fill fee presence was lost: %+v", fills)
	}
}

func TestProtocolReturnsFreshJSONAndRejectsMalformedEvidence(t *testing.T) {
	var action = PlaceOrderAction{
		Type:     "order",
		Grouping: "na",
		Orders: []OrderRequest{{
			Asset: 0, IsBuy: true, Price: "100", Size: "1",
			Type:  OrderType{Limit: &LimitOrderType{TimeInForce: "Gtc"}},
			CLOID: "0x00000000000000000000000000000001",
		}},
	}
	var first, err = Encode(action)
	if err != nil {
		t.Fatalf("encode first payload: %v", err)
	}
	first[0] = 'x'
	var second []byte
	second, err = Encode(action)
	if err != nil {
		t.Fatalf("encode second payload: %v", err)
	}
	if second[0] != '{' {
		t.Fatal("encoded payload reused caller memory")
	}
	if _, err = DecodeOpenOrders([]byte(`[{"coin":"BTC"}]`)); err == nil {
		t.Fatal("incomplete open Order was accepted")
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
