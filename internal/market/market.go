package market

import (
	"fmt"
	"math"
)

type BBO struct {
	TimestampMS uint64
	Price       float64
}

// Section 1 - Program Flow

// CreateBBO validates and creates one BBO.
func CreateBBO(timestampMS uint64, price float64) (BBO, error) {
	// validate bbo
	if timestampMS == 0 || math.IsNaN(price) || math.IsInf(price, 0) || price <= 0 {
		return BBO{}, fmt.Errorf("invalid BBO timestamp=%d price=%g", timestampMS, price)
	}

	// create bbo
	return BBO{TimestampMS: timestampMS, Price: price}, nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
