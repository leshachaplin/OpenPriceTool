// Package with types and their methods

package model

import (
	"encoding/json"
	"time"
)

// Price structure which define price type
type Price struct {
	ID       string
	Bid      float64
	Ask      float64
	Date     time.Time
	Symbol   string
	Currency string
}

// UnmarshalBinary unmarshaler for Price
func (p *Price) UnmarshalBinary(data []byte) (*Price, error) {
	price := &Price{}
	err := json.Unmarshal(data, price)
	if err != nil {
		return nil, err
	}

	return price, nil
}

// MarshalBinary marshaller for Price
func (p *Price) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}

// GetPrice take a price corresponding to a short or long position
func (p *Price) GetPrice(short bool) float64 {
	if short {
		return p.Ask
	}
	return p.Bid
}
