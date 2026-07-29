package hyperliquid

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// OrderRequest contains one official Hyperliquid wire Order.
type OrderRequest struct {
	Asset      int       `json:"a"`
	IsBuy      bool      `json:"b"`
	Price      string    `json:"p"`
	Size       string    `json:"s"`
	ReduceOnly bool      `json:"r"`
	Type       OrderType `json:"t"`
	CLOID      string    `json:"c"`
}

// OrderType contains exactly one official limit or trigger definition.
type OrderType struct {
	Limit   *LimitOrderType   `json:"limit,omitempty"`
	Trigger *TriggerOrderType `json:"trigger,omitempty"`
}

// LimitOrderType contains one official limit time-in-force value.
type LimitOrderType struct {
	TimeInForce string `json:"tif"`
}

// TriggerOrderType contains one official trigger definition.
type TriggerOrderType struct {
	IsMarket     bool   `json:"isMarket"`
	TriggerPrice string `json:"triggerPx"`
	TPSL         string `json:"tpsl"`
}

// PlaceOrderAction contains one official order action.
type PlaceOrderAction struct {
	Type     string         `json:"type"`
	Orders   []OrderRequest `json:"orders"`
	Grouping string         `json:"grouping"`
}

// CancelByCLOIDRequest contains one official cancellation identity.
type CancelByCLOIDRequest struct {
	Asset int    `json:"asset"`
	CLOID string `json:"cloid"`
}

// CancelByCLOIDAction contains one official cancel-by-CLOID action.
type CancelByCLOIDAction struct {
	Type    string                 `json:"type"`
	Cancels []CancelByCLOIDRequest `json:"cancels"`
}

// UpdateLeverageAction contains one official leverage update.
type UpdateLeverageAction struct {
	Type     string `json:"type"`
	Asset    int    `json:"asset"`
	IsCross  bool   `json:"isCross"`
	Leverage uint32 `json:"leverage"`
}

// SubmitResponse contains one validated official order response.
type SubmitResponse struct {
	Status   string
	Type     string
	Statuses []SubmitStatus
	Raw      string
}

// SubmitStatus contains one official ordered submission outcome.
type SubmitStatus struct {
	Kind         string
	VenueOrderID uint64
	CLOID        string
	TotalSize    string
	AveragePrice string
	Fee          *string
	Error        string
}

// CancelResponse contains one validated official cancellation response.
type CancelResponse struct {
	Status   string
	Type     string
	Statuses []CancelStatus
	Raw      string
}

// CancelStatus contains one official ordered cancellation outcome.
type CancelStatus struct {
	Success bool
	Error   string
}

// OpenOrder contains one official open-Order row.
type OpenOrder struct {
	Coin         string `json:"coin"`
	LimitPrice   string `json:"limitPx"`
	VenueOrderID uint64 `json:"oid"`
	Side         string `json:"side"`
	Size         string `json:"sz"`
	TimestampMS  uint64 `json:"timestamp"`
	CLOID        string `json:"cloid,omitempty"`
}

// OrderStatus contains one official exact Order-status response.
type OrderStatus struct {
	Status          string
	Order           *OpenOrder
	OrderStatus     string
	StatusTimestamp uint64
	Raw             string
}

// HistoricalOrder contains one official historical Order row.
type HistoricalOrder struct {
	Order           OpenOrder `json:"order"`
	Status          string    `json:"status"`
	StatusTimestamp uint64    `json:"statusTimestamp"`
}

// Fill contains one official user-Fill row.
type Fill struct {
	Coin          string  `json:"coin"`
	Price         string  `json:"px"`
	Size          string  `json:"sz"`
	Side          string  `json:"side"`
	TimestampMS   uint64  `json:"time"`
	VenueOrderID  uint64  `json:"oid"`
	VenueTID      uint64  `json:"tid"`
	CLOID         string  `json:"cloid,omitempty"`
	StartPosition string  `json:"startPosition"`
	ClosedPnL     string  `json:"closedPnl"`
	Direction     string  `json:"dir"`
	Crossed       bool    `json:"crossed"`
	Fee           *string `json:"fee,omitempty"`
	FeeToken      string  `json:"feeToken,omitempty"`
	Hash          string  `json:"hash,omitempty"`
}

type exchangeResponse struct {
	Status   string `json:"status"`
	Response struct {
		Type string `json:"type"`
		Data struct {
			Statuses []json.RawMessage `json:"statuses"`
		} `json:"data"`
	} `json:"response"`
}

type restingStatus struct {
	Resting *struct {
		OID   uint64 `json:"oid"`
		CLOID string `json:"cloid,omitempty"`
	} `json:"resting"`
}

type filledStatus struct {
	Filled *struct {
		TotalSize    string  `json:"totalSz"`
		AveragePrice string  `json:"avgPx"`
		OID          uint64  `json:"oid"`
		CLOID        string  `json:"cloid,omitempty"`
		Fee          *string `json:"fee,omitempty"`
	} `json:"filled"`
}

type errorStatus struct {
	Error string `json:"error"`
}

type orderStatusResponse struct {
	Status string `json:"status"`
	Order  *struct {
		Order           OpenOrder `json:"order"`
		Status          string    `json:"status"`
		StatusTimestamp uint64    `json:"statusTimestamp"`
	} `json:"order"`
}

// Section 1 - Program Flow

// Section 2 - Domain Helpers

// Encode returns fresh JSON for one official request or response value.
func Encode(value any) ([]byte, error) {
	var payload, err = json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode Hyperliquid payload: %w", err)
	}
	return payload, nil
}

// DecodeSubmitResponse validates one official order response.
func DecodeSubmitResponse(payload []byte) (SubmitResponse, error) {
	var envelope exchangeResponse
	var err = decodeOne(payload, &envelope)
	if err != nil {
		return SubmitResponse{}, fmt.Errorf("decode Hyperliquid order response: %w", err)
	}
	if envelope.Status != "ok" || envelope.Response.Type != "order" {
		return SubmitResponse{}, fmt.Errorf("decode Hyperliquid order response: invalid envelope")
	}
	var result = SubmitResponse{
		Status: envelope.Status,
		Type:   envelope.Response.Type,
		Raw:    string(payload),
	}
	for _, raw := range envelope.Response.Data.Statuses {
		var status SubmitStatus
		status, err = decodeSubmitStatus(raw)
		if err != nil {
			return SubmitResponse{}, err
		}
		result.Statuses = append(result.Statuses, status)
	}
	return result, nil
}

// DecodeCancelResponse validates one official cancellation response.
func DecodeCancelResponse(payload []byte) (CancelResponse, error) {
	var envelope exchangeResponse
	var err = decodeOne(payload, &envelope)
	if err != nil {
		return CancelResponse{}, fmt.Errorf("decode Hyperliquid cancel response: %w", err)
	}
	if envelope.Status != "ok" || envelope.Response.Type != "cancel" {
		return CancelResponse{}, fmt.Errorf("decode Hyperliquid cancel response: invalid envelope")
	}
	var result = CancelResponse{
		Status: envelope.Status,
		Type:   envelope.Response.Type,
		Raw:    string(payload),
	}
	for _, raw := range envelope.Response.Data.Statuses {
		var value string
		if json.Unmarshal(raw, &value) == nil && value == "success" {
			result.Statuses = append(result.Statuses, CancelStatus{Success: true})
			continue
		}
		var rejected errorStatus
		if json.Unmarshal(raw, &rejected) != nil || rejected.Error == "" {
			return CancelResponse{}, fmt.Errorf("decode Hyperliquid cancel response: invalid status")
		}
		result.Statuses = append(result.Statuses, CancelStatus{Error: rejected.Error})
	}
	return result, nil
}

// DecodeOpenOrders validates one official open-Orders response.
func DecodeOpenOrders(payload []byte) ([]OpenOrder, error) {
	var rows []OpenOrder
	var err = decodeOne(payload, &rows)
	if err != nil {
		return nil, fmt.Errorf("decode Hyperliquid open Orders: %w", err)
	}
	for _, row := range rows {
		if row.Coin == "" || (row.CLOID == "" && row.VenueOrderID == 0) || row.Side == "" ||
			row.Size == "" || row.TimestampMS == 0 {
			return nil, fmt.Errorf("decode Hyperliquid open Orders: incomplete row")
		}
	}
	return rows, nil
}

// DecodeOrderHistory validates one official historical-Orders response.
func DecodeOrderHistory(payload []byte) ([]HistoricalOrder, error) {
	var rows []HistoricalOrder
	var err = decodeOne(payload, &rows)
	if err != nil {
		return nil, fmt.Errorf("decode Hyperliquid Order history: %w", err)
	}
	for _, row := range rows {
		if row.Order.Coin == "" ||
			(row.Order.CLOID == "" && row.Order.VenueOrderID == 0) ||
			row.Order.Side == "" || row.Order.TimestampMS == 0 ||
			row.Status == "" || row.StatusTimestamp == 0 {
			return nil, fmt.Errorf("decode Hyperliquid Order history: incomplete row")
		}
	}
	return rows, nil
}

// DecodeOrderStatus validates one official exact Order-status response.
func DecodeOrderStatus(payload []byte) (OrderStatus, error) {
	var response orderStatusResponse
	var err = decodeOne(payload, &response)
	if err != nil {
		return OrderStatus{}, fmt.Errorf("decode Hyperliquid Order status: %w", err)
	}
	if response.Status == "unknownOid" {
		return OrderStatus{Status: response.Status, Raw: string(payload)}, nil
	}
	if response.Status != "order" || response.Order == nil ||
		(response.Order.Order.CLOID == "" &&
			response.Order.Order.VenueOrderID == 0) ||
		response.Order.Status == "" {
		return OrderStatus{}, fmt.Errorf("decode Hyperliquid Order status: invalid response")
	}
	var row = response.Order.Order
	return OrderStatus{
		Status:          response.Status,
		Order:           &row,
		OrderStatus:     response.Order.Status,
		StatusTimestamp: response.Order.StatusTimestamp,
		Raw:             string(payload),
	}, nil
}

// DecodeFills validates one official user-Fills response.
func DecodeFills(payload []byte) ([]Fill, error) {
	var rows []Fill
	var err = decodeOne(payload, &rows)
	if err != nil {
		return nil, fmt.Errorf("decode Hyperliquid Fills: %w", err)
	}
	for _, row := range rows {
		if row.Coin == "" || row.Price == "" || row.Size == "" ||
			row.Side == "" || row.TimestampMS == 0 ||
			row.VenueOrderID == 0 || row.VenueTID == 0 {
			return nil, fmt.Errorf("decode Hyperliquid Fills: incomplete row")
		}
	}
	return rows, nil
}

// Section 3 - Generic Helpers

func decodeSubmitStatus(raw json.RawMessage) (SubmitStatus, error) {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		if text == "waitingForFill" || text == "waitingForTrigger" {
			return SubmitStatus{Kind: text}, nil
		}
		return SubmitStatus{}, fmt.Errorf("decode Hyperliquid order response: invalid status")
	}
	var resting restingStatus
	if json.Unmarshal(raw, &resting) == nil && resting.Resting != nil &&
		(resting.Resting.OID != 0 || resting.Resting.CLOID != "") {
		return SubmitStatus{
			Kind:         "resting",
			VenueOrderID: resting.Resting.OID,
			CLOID:        resting.Resting.CLOID,
		}, nil
	}
	var filled filledStatus
	if json.Unmarshal(raw, &filled) == nil && filled.Filled != nil &&
		(filled.Filled.OID != 0 || filled.Filled.CLOID != "") &&
		filled.Filled.TotalSize != "" &&
		filled.Filled.AveragePrice != "" {
		return SubmitStatus{
			Kind:         "filled",
			VenueOrderID: filled.Filled.OID,
			CLOID:        filled.Filled.CLOID,
			TotalSize:    filled.Filled.TotalSize,
			AveragePrice: filled.Filled.AveragePrice,
			Fee:          filled.Filled.Fee,
		}, nil
	}
	var rejected errorStatus
	if json.Unmarshal(raw, &rejected) == nil && rejected.Error != "" {
		return SubmitStatus{Kind: "error", Error: rejected.Error}, nil
	}
	return SubmitStatus{}, fmt.Errorf("decode Hyperliquid order response: invalid status")
}

func decodeOne(payload []byte, value any) error {
	var decoder = json.NewDecoder(bytes.NewReader(payload))
	var err = decoder.Decode(value)
	if err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) == nil {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}
