package hyperliquid

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// AccountState contains one translated perpetual clearinghouse response.
type AccountState struct {
	TimeMS            uint64          `json:"time_ms"`
	Margin            Margin          `json:"margin"`
	CrossMargin       Margin          `json:"cross_margin"`
	MaintenanceMargin decimal.Decimal `json:"maintenance_margin"`
	Withdrawable      decimal.Decimal `json:"withdrawable"`
	Positions         []Position      `json:"positions"`
}

// Margin contains one readable Hyperliquid margin summary.
type Margin struct {
	Equity     decimal.Decimal `json:"equity"`
	MarginUsed decimal.Decimal `json:"margin_used"`
	Notional   decimal.Decimal `json:"notional"`
	RawUSD     decimal.Decimal `json:"raw_usd"`
}

// Position contains one readable perpetual position.
type Position struct {
	Symbol           string           `json:"symbol"`
	Mode             string           `json:"mode"`
	SignedSize       decimal.Decimal  `json:"signed_size"`
	EntryPrice       *decimal.Decimal `json:"entry_price"`
	LiquidationPrice *decimal.Decimal `json:"liquidation_price"`
	MarginUsed       decimal.Decimal  `json:"margin_used"`
	Notional         decimal.Decimal  `json:"notional"`
	ReturnOnEquity   decimal.Decimal  `json:"return_on_equity"`
	UnrealizedPnL    decimal.Decimal  `json:"unrealized_pnl"`
	MaxLeverage      uint32           `json:"max_leverage"`
	Leverage         Leverage         `json:"leverage"`
	Funding          Funding          `json:"funding"`
}

// Leverage contains one position's leverage settings.
type Leverage struct {
	Type   string           `json:"type"`
	Value  uint32           `json:"value"`
	RawUSD *decimal.Decimal `json:"raw_usd"`
}

// Funding contains cumulative funding values.
type Funding struct {
	AllTime     decimal.Decimal `json:"all_time"`
	SinceChange decimal.Decimal `json:"since_change"`
	SinceOpen   decimal.Decimal `json:"since_open"`
}

type clearinghouseStateResponse struct {
	AssetPositions             []assetPositionResponse `json:"assetPositions"`
	CrossMaintenanceMarginUsed string                  `json:"crossMaintenanceMarginUsed"`
	CrossMarginSummary         marginSummaryResponse   `json:"crossMarginSummary"`
	MarginSummary              marginSummaryResponse   `json:"marginSummary"`
	Time                       uint64                  `json:"time"`
	Withdrawable               string                  `json:"withdrawable"`
}

type marginSummaryResponse struct {
	AccountValue    string `json:"accountValue"`
	TotalMarginUsed string `json:"totalMarginUsed"`
	TotalNtlPos     string `json:"totalNtlPos"`
	TotalRawUSD     string `json:"totalRawUsd"`
}

type assetPositionResponse struct {
	Position positionResponse `json:"position"`
	Type     string           `json:"type"`
}

type positionResponse struct {
	Coin           string           `json:"coin"`
	CumFunding     fundingResponse  `json:"cumFunding"`
	EntryPrice     *string          `json:"entryPx"`
	Leverage       leverageResponse `json:"leverage"`
	LiquidationPx  *string          `json:"liquidationPx"`
	MarginUsed     string           `json:"marginUsed"`
	MaxLeverage    uint32           `json:"maxLeverage"`
	PositionValue  string           `json:"positionValue"`
	ReturnOnEquity string           `json:"returnOnEquity"`
	Size           string           `json:"szi"`
	UnrealizedPnL  string           `json:"unrealizedPnl"`
}

type leverageResponse struct {
	RawUSD *string `json:"rawUsd"`
	Type   string  `json:"type"`
	Value  uint32  `json:"value"`
}

type fundingResponse struct {
	AllTime     string `json:"allTime"`
	SinceChange string `json:"sinceChange"`
	SinceOpen   string `json:"sinceOpen"`
}

// Section 1 - Program Flow

// Section 2 - Domain Helpers

// DecodeClearinghouseState decodes and translates one raw clearinghouse payload.
func DecodeClearinghouseState(payload []byte) (AccountState, error) {
	// decode response payload
	var response clearinghouseStateResponse
	var err = json.Unmarshal(payload, &response)
	if err != nil {
		return AccountState{}, fmt.Errorf("decode clearinghouse payload: %v", err)
	}
	var validationErr = validateClearinghouseResponse(response)
	if validationErr != nil {
		return AccountState{}, validationErr
	}
	// translate account state
	var state AccountState
	state, err = translateAccountState(response)
	if err != nil {
		return AccountState{}, err
	}
	return state, nil
}

func validateClearinghouseResponse(response clearinghouseStateResponse) error {
	if response.Time == 0 {
		return fmt.Errorf("validate clearinghouse payload: time must be positive")
	}
	if response.AssetPositions == nil {
		return fmt.Errorf("validate clearinghouse payload: assetPositions must be an array")
	}
	var index int
	var asset assetPositionResponse
	for index, asset = range response.AssetPositions {
		var name = fmt.Sprintf("assetPositions[%d]", index)
		if asset.Type != "oneWay" {
			return fmt.Errorf("validate clearinghouse payload: %s.type %q", name, asset.Type)
		}
		if asset.Position.Coin == "" {
			return fmt.Errorf("validate clearinghouse payload: %s.position.coin is empty", name)
		}
		if asset.Position.Leverage.Type != "cross" &&
			asset.Position.Leverage.Type != "isolated" {
			return fmt.Errorf(
				"validate clearinghouse payload: %s.position.leverage.type %q",
				name,
				asset.Position.Leverage.Type,
			)
		}
		if asset.Position.Leverage.Value == 0 {
			return fmt.Errorf(
				"validate clearinghouse payload: %s.position.leverage.value must be positive",
				name,
			)
		}
		if asset.Position.MaxLeverage == 0 {
			return fmt.Errorf(
				"validate clearinghouse payload: %s.position.maxLeverage must be positive",
				name,
			)
		}
	}
	return nil
}

func translateAccountState(response clearinghouseStateResponse) (AccountState, error) {
	var margin, err = translateMargin("marginSummary", response.MarginSummary)
	if err != nil {
		return AccountState{}, err
	}
	var crossMargin Margin
	crossMargin, err = translateMargin("crossMarginSummary", response.CrossMarginSummary)
	if err != nil {
		return AccountState{}, err
	}
	var maintenanceMargin decimal.Decimal
	maintenanceMargin, err = parseDecimal(
		"crossMaintenanceMarginUsed",
		response.CrossMaintenanceMarginUsed,
	)
	if err != nil {
		return AccountState{}, err
	}
	var withdrawable decimal.Decimal
	withdrawable, err = parseDecimal("withdrawable", response.Withdrawable)
	if err != nil {
		return AccountState{}, err
	}
	var positions = make([]Position, 0, len(response.AssetPositions))
	var index int
	var assetPosition assetPositionResponse
	for index, assetPosition = range response.AssetPositions {
		var position Position
		position, err = translatePosition(index, assetPosition)
		if err != nil {
			return AccountState{}, err
		}
		positions = append(positions, position)
	}
	return AccountState{
		TimeMS:            response.Time,
		Margin:            margin,
		CrossMargin:       crossMargin,
		MaintenanceMargin: maintenanceMargin,
		Withdrawable:      withdrawable,
		Positions:         positions,
	}, nil
}

func translateMargin(name string, response marginSummaryResponse) (Margin, error) {
	var equity, err = parseDecimal(name+".accountValue", response.AccountValue)
	if err != nil {
		return Margin{}, err
	}
	var marginUsed decimal.Decimal
	marginUsed, err = parseDecimal(name+".totalMarginUsed", response.TotalMarginUsed)
	if err != nil {
		return Margin{}, err
	}
	var notional decimal.Decimal
	notional, err = parseDecimal(name+".totalNtlPos", response.TotalNtlPos)
	if err != nil {
		return Margin{}, err
	}
	var rawUSD decimal.Decimal
	rawUSD, err = parseDecimal(name+".totalRawUsd", response.TotalRawUSD)
	if err != nil {
		return Margin{}, err
	}
	return Margin{
		Equity:     equity,
		MarginUsed: marginUsed,
		Notional:   notional,
		RawUSD:     rawUSD,
	}, nil
}

func translatePosition(index int, response assetPositionResponse) (Position, error) {
	var name = fmt.Sprintf("assetPositions[%d].position", index)
	var signedSize, err = parseDecimal(name+".szi", response.Position.Size)
	if err != nil {
		return Position{}, err
	}
	var entryPrice *decimal.Decimal
	entryPrice, err = parseOptionalDecimal(name+".entryPx", response.Position.EntryPrice)
	if err != nil {
		return Position{}, err
	}
	var liquidationPrice *decimal.Decimal
	liquidationPrice, err = parseOptionalDecimal(
		name+".liquidationPx",
		response.Position.LiquidationPx,
	)
	if err != nil {
		return Position{}, err
	}
	var marginUsed decimal.Decimal
	marginUsed, err = parseDecimal(name+".marginUsed", response.Position.MarginUsed)
	if err != nil {
		return Position{}, err
	}
	var notional decimal.Decimal
	notional, err = parseDecimal(name+".positionValue", response.Position.PositionValue)
	if err != nil {
		return Position{}, err
	}
	var returnOnEquity decimal.Decimal
	returnOnEquity, err = parseDecimal(
		name+".returnOnEquity",
		response.Position.ReturnOnEquity,
	)
	if err != nil {
		return Position{}, err
	}
	var unrealizedPnL decimal.Decimal
	unrealizedPnL, err = parseDecimal(name+".unrealizedPnl", response.Position.UnrealizedPnL)
	if err != nil {
		return Position{}, err
	}
	var leverage Leverage
	leverage, err = translateLeverage(name+".leverage", response.Position.Leverage)
	if err != nil {
		return Position{}, err
	}
	var funding Funding
	funding, err = translateFunding(name+".cumFunding", response.Position.CumFunding)
	if err != nil {
		return Position{}, err
	}
	return Position{
		Symbol:           response.Position.Coin,
		Mode:             response.Type,
		SignedSize:       signedSize,
		EntryPrice:       entryPrice,
		LiquidationPrice: liquidationPrice,
		MarginUsed:       marginUsed,
		Notional:         notional,
		ReturnOnEquity:   returnOnEquity,
		UnrealizedPnL:    unrealizedPnL,
		MaxLeverage:      response.Position.MaxLeverage,
		Leverage:         leverage,
		Funding:          funding,
	}, nil
}

func translateLeverage(name string, response leverageResponse) (Leverage, error) {
	var rawUSD, err = parseOptionalDecimal(name+".rawUsd", response.RawUSD)
	if err != nil {
		return Leverage{}, err
	}
	return Leverage{
		Type:   response.Type,
		Value:  response.Value,
		RawUSD: rawUSD,
	}, nil
}

func translateFunding(name string, response fundingResponse) (Funding, error) {
	var allTime, err = parseDecimal(name+".allTime", response.AllTime)
	if err != nil {
		return Funding{}, err
	}
	var sinceChange decimal.Decimal
	sinceChange, err = parseDecimal(name+".sinceChange", response.SinceChange)
	if err != nil {
		return Funding{}, err
	}
	var sinceOpen decimal.Decimal
	sinceOpen, err = parseDecimal(name+".sinceOpen", response.SinceOpen)
	if err != nil {
		return Funding{}, err
	}
	return Funding{
		AllTime:     allTime,
		SinceChange: sinceChange,
		SinceOpen:   sinceOpen,
	}, nil
}

// Section 3 - Generic Helpers

func parseDecimal(name, value string) (decimal.Decimal, error) {
	var parsed, err = decimal.NewFromString(value)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("translate %s: invalid decimal %q", name, value)
	}
	return parsed, nil
}

func parseOptionalDecimal(name string, value *string) (*decimal.Decimal, error) {
	if value == nil {
		return nil, nil
	}
	var parsed, err = parseDecimal(name, *value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
