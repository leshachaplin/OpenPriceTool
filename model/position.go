// Package with types and their methods

package model

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

// PositionError define the position package errors
type PositionError int

const (
	Null PositionError = iota
	LastPriceNull
	StopLossNull
	TakeProfitNull
	SetShortStopLoss
	SetLongStopLoss
	SetShortTakeProfit
	SetLongTakeProfit
)

// Error return string representation of error
func (p PositionError) Error() string {
	switch p {
	case Null:
		return "Null"
	case LastPriceNull:
		return "last price has nil expression"
	case StopLossNull:
		return "stop loss has nil expression"
	case TakeProfitNull:
		return "take profit has nil expression"
	case SetShortStopLoss:
		return "short position, price has to be higher that stop loss price"
	case SetLongStopLoss:
		return "short position, price has to be lower than current stop loss price"
	case SetShortTakeProfit:
		return "short position, price has to be lower that current take profit value"
	case SetLongTakeProfit:
		return "long position, price has to be higher that current take profit value"
	}

	return fmt.Sprintf("Pill(%d)", p)
}

// Position structure which define position type
type Position struct {
	Symbol     string
	Short      bool
	OpenPrice  float64
	LastPrice  *float64
	StopLoss   *float64
	TakeProfit *float64
	Amount     float64
}

// PositionPortfolio structure which contains price channel, context
// and cancel func and this structure is a field of Portfolio
type PositionPortfolio struct {
	*Position
	PriceChan chan Price
	ctx       context.Context
	cnsl      context.CancelFunc
}

// PositionDB structure which define position type and
// have time fields to use with databases
type PositionDB struct {
	*Position
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
}

// NewPositionPortfolio initialize PositionPortfolio
func NewPositionPortfolio(pos *Position, ch chan Price,
	ctx context.Context, cancelFunc context.CancelFunc) *PositionPortfolio {
	return &PositionPortfolio{
		Position:  pos,
		PriceChan: ch,
		ctx:       ctx,
		cnsl:      cancelFunc,
	}
}

// SetTakeProfit check is value valid and set TakeProfit
func (p *Position) SetTakeProfit(value float64) error {
	if p.LastPrice == nil {
		return TakeProfitNull
	}
	////TODO?????????????????
	//if value == 0 {
	//	p.TakeProfit = nil
	//	return nil
	//}

	if p.Short {
		if *p.LastPrice >= value {
			return SetShortTakeProfit
		}
		p.TakeProfit = &value
		return nil
	}
	if *p.LastPrice <= value {
		return SetLongTakeProfit
	}
	p.TakeProfit = &value
	return nil
}

// SetStopLoss check is value valid and set StopLoss
func (p *Position) SetStopLoss(value float64) error {
	if p.LastPrice == nil {
		return LastPriceNull
	}

	if value == 0 {
		p.StopLoss = &value
		return nil
	}

	if p.Short {
		if *p.LastPrice <= value {
			return SetShortStopLoss
		}
		p.StopLoss = &value
		return nil
	}
	if *p.LastPrice >= value {
		return SetLongStopLoss
	}
	p.StopLoss = &value
	return nil
}

// TriggerStopLoss notifies when to stop
func (p *Position) TriggerStopLoss() (bool, error) {
	log.Info("trigger stop loss start")
	//if p.LastPrice == nil {
	//	return false, LastPriceNull
	//}
	//?????????????????????????????????????
	//if p.StopLoss == nil {
	//	return false, StopLossNull
	//}

	if *p.StopLoss == 0 {
		return false, nil
	}

	if p.Short {
		return *p.LastPrice >= *p.StopLoss, nil
	}
	return *p.LastPrice <= *p.StopLoss, nil
}

// TriggerTakeProfit notifies when to take profit
func (p *Position) TriggerTakeProfit() (bool, error) {
	if p.LastPrice == nil {
		return false, LastPriceNull
	}

	if p.TakeProfit == nil {
		return false, TakeProfitNull
	}

	if p.Short {
		return *p.LastPrice <= *p.TakeProfit, nil
	}
	return *p.LastPrice >= *p.TakeProfit, nil
}

// PnL net profit from the transaction
func (p *Position) PnL() (*float64, error) {
	if p.LastPrice == nil {
		return nil, LastPriceNull
	}

	if p.Short {
		pnl := (p.OpenPrice - *p.LastPrice) * p.Amount
		return &pnl, nil
	}
	pnl := (*p.LastPrice - p.OpenPrice) * p.Amount
	return &pnl, nil
}
