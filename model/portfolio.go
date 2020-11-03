// Package with types and their methods

package model

import (
	"context"
	log "github.com/sirupsen/logrus"
	"sync"
)

// Portfolio structure which define portfolio type
type Portfolio struct {
	UserID    string
	positions map[string]*PositionPortfolio
	Balance   float64
	closer    PositionCloser
	converter PriceConverter
	mu        sync.RWMutex
}

// NewPortfolio initialize Portfolio structure
func NewPortfolio(userID string, balance float64, closer PositionCloser) *Portfolio {

	return &Portfolio{
		UserID:    userID,
		Balance:   balance,
		positions: make(map[string]*PositionPortfolio, 0),
		closer:    closer,
	}
}

// TriggerMarginCall
func (p *Portfolio) TriggerMarginCall() (bool, error) {
	balance := p.Balance
	for i := range p.positions {
		pnl, err := p.positions[i].PnL()
		if err != nil {
			log.Error("error in portfolio: error in TriggerMarginCall: ", err)
			return false, err
		}
		balance += *pnl
	}
	return balance < 0, nil
}

// PriceUpdate update last price and if trigger stop loss is true
// or trigger take profit is true close position
// TODO:margin call take profit
func (p *Portfolio) PriceUpdate(ctx context.Context, price Price) error {
	log.Info("Price Update")

	position, ok := p.positions[price.Symbol]
	if ok {
		//if position.LastPrice == nil {
		//	return fmt.Errorf("error in portfolio: PriceUpdate: last price has nil expression")
		//}
		lastPr := price.GetPrice(position.Short)
		position.LastPrice = &lastPr
		log.Info("UPDATE LAST PRICE")

		trStopLoss, err := position.TriggerStopLoss()
		if err != nil {
			log.Error("error in portfolio: PriceUpdate: trigger stop loss error ", err)
			return err
		}

		if trStopLoss {
			err = p.closer.Close(ctx, p.UserID, position.Symbol, price.ID)
			if err != nil {
				return err
			}
			position.cnsl()
		}

		//trTakeProfit, err := position.TriggerTakeProfit()
		//if err != nil {
		//	log.Error("error in portfolio: PriceUpdate: trigger take profit error ", err)
		//	return err
		//}
		//
		//if trTakeProfit {
		//	p.closer.Close(p.UserID, position, price.GetPrice(position.Short), price.Date)
		//}
		//
		//trMarginCall, err := p.TriggerMarginCall()
		//if err != nil {
		//	log.Error("error in portfolio: PriceUpdate: trigger margin call error ", err)
		//	return err
		//}
		//
		//if trMarginCall {
		//	// do something
		//}
	}
	return nil
}

func (p *Portfolio) AddPosition(ctx context.Context, pos *Position, balance float64) {
	done, cnsl := context.WithCancel(context.Background())
	priceChan := make(chan Price)

	position := NewPositionPortfolio(pos, priceChan, done, cnsl)

	p.mu.Lock()
	p.positions[position.Symbol] = position
	p.Balance += balance
	p.mu.Unlock()

	p.ChannelsReader(ctx, pos.Symbol)
}

func (p *Portfolio) DeletePosition(symbol string) {
	p.mu.RLock()
	delete(p.positions, symbol)
	p.mu.RUnlock()
}

func (p *Portfolio) UpdatePrice(price Price) bool {
	p.mu.RLock()
	position, ok := p.positions[price.Symbol]
	p.mu.RUnlock()
	if ok {
		log.Info("PriceManager:SEND TO PRICE CHANNEL ")
		position.PriceChan <- price
	} else {
		return false
	}
	return true
}

// ChannelsReader start reading price channel for open
// position and update price in separate goroutine
func (p *Portfolio) ChannelsReader(ctx context.Context, symbol string) {
	p.mu.RLock()
	position := p.positions[symbol]
	p.mu.RUnlock()
	go func(ctx context.Context, pos *PositionPortfolio) {
		for {
			select {
			case <-ctx.Done():
				{
					log.Info("CHANNELS READER: CONTEXT DONE FOR SYMBOL ", pos.Symbol)
					return
				}
			case <-pos.ctx.Done():
				{
					log.Info("CHANNELS READER: POSITION CONTEXT DONE FOR SYMBOL ", pos.Symbol)
					return
				}
			default:
				{
					price := <-pos.PriceChan
					p.mu.RLock()
					err := p.PriceUpdate(ctx, price)
					p.mu.RUnlock()
					if err != nil {
						//TODO: make retry
					}
				}
			}
		}
	}(ctx, position)
}
