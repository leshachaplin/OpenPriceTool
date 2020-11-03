package test

import (
	"github.com/leshachaplin/OpenPriceTool/model"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestPnlLong calculates the net profit for a long position
func TestPnlLong(t *testing.T) {
	lastPriceValue := float64(5)

	p := &model.Position{
		Amount:    10,
		OpenPrice: 1,
		LastPrice: &lastPriceValue,
		Short:     false,
	}
	pnl, err := p.PnL()
	if err != nil {
		t.Error(err)
	}

	require.Equal(t, *pnl, float64(40))
}

// TestPnlLongNegative calculates the loss for a long position
func TestPnlLongNegative(t *testing.T) {
	lastPriceValue := float64(1)

	p := &model.Position{
		Amount:    10,
		OpenPrice: 5,
		LastPrice: &lastPriceValue,
		Short:     false,
	}

	pnl, err := p.PnL()
	if err != nil {
		t.Error(err)
	}

	require.Equal(t, *pnl, float64(-40))
}

// TestPnlShort calculates the net profit for a short position
func TestPnlShort(t *testing.T) {
	lastPriceValue := float64(1)

	p := &model.Position{
		Amount:    10,
		OpenPrice: 5,
		LastPrice: &lastPriceValue,
		Short:     true,
	}

	pnl, err := p.PnL()
	if err != nil {
		t.Error(err)
	}

	require.Equal(t, *pnl, float64(40))
}

// TestPnlShortNegative calculates the loss for a short position
func TestPnlShortNegative(t *testing.T) {
	lastPriceValue := float64(5)

	p := &model.Position{
		Amount:    10,
		OpenPrice: 1,
		LastPrice: &lastPriceValue,
		Short:     true,
	}

	pnl, err := p.PnL()
	if err != nil {
		t.Error(err)
	}

	require.Equal(t, *pnl, float64(-40))
}

// TestLongTakeProfit check is TakeProfit value is valid
// and trigger to take profit if all conditions is true
func TestLongTakeProfit(t *testing.T) {
	lastPriceValue := float64(8)

	p := &model.Position{
		Amount:    10,
		OpenPrice: 5,
		LastPrice: &lastPriceValue,
		Short:     false,
	}

	err := p.SetTakeProfit(6)
	if err != nil {
		t.Error(err)
	}

	tr, err := p.TriggerTakeProfit()
	if err != nil {
		t.Error(err)
	}

	require.True(t, tr)
}

// TestLongTakeProfitFalse check is TakeProfit value is valid, value should be invalid in this test
func TestLongTakeProfitFalse(t *testing.T) {
	lastPriceValue := float64(8)

	p := &model.Position{
		Amount:    10,
		OpenPrice: 5,
		LastPrice: &lastPriceValue,
		Short:     false,
	}

	err := p.SetTakeProfit(9)

	require.EqualError(t, err, "long position, price has to be higher that current value")
}

// TestLongStopLossFalse check is StopLoss value is valid, value should be invalid in this test
func TestLongStopLossFalse(t *testing.T) {
	lastPriceValue := float64(3)

	p := &model.Position{
		Amount:    10,
		OpenPrice: 5,
		LastPrice: &lastPriceValue,
		Short:     false,
	}

	err := p.SetStopLoss(2)

	require.EqualError(t, err, model.SetLongStopLoss.Error())
}

// TestLongStopLoss check is StopLoss value is valid
// and trigger to stop loss if all conditions is true
func TestLongStopLoss(t *testing.T) {
	lastPriceValue := float64(1)

	p := &model.Position{
		Amount:    10,
		OpenPrice: 5,
		LastPrice: &lastPriceValue,
		Short:     false,
	}

	err := p.SetStopLoss(4)
	if err != nil {
		t.Error(err)
	}

	tr, err := p.TriggerStopLoss()
	if err != nil {
		t.Error(err)
	}

	require.True(t, tr)
}

// TestShortTakeProfit check is TakeProfit value is valid
// and trigger to take profit if all conditions is true
func TestShortTakeProfit(t *testing.T) {
	lastPriceValue := float64(2)

	p := &model.Position{
		Amount:    10,
		OpenPrice: 5,
		LastPrice: &lastPriceValue,
		Short:     true,
	}

	err := p.SetTakeProfit(4)
	if err != nil {
		t.Error(err)
	}

	tr, err := p.TriggerTakeProfit()
	if err != nil {
		t.Error(err)
	}

	require.True(t, tr)
}

// TestShortTakeProfitFalse check is TakeProfit value is valid, value should be invalid in this test
func TestShortTakeProfitFalse(t *testing.T) {
	lastPriceValue := float64(3)

	p := &model.Position{
		Amount:    10,
		OpenPrice: 5,
		LastPrice: &lastPriceValue,
		Short:     true,
	}

	err := p.SetTakeProfit(2)

	require.EqualError(t, err, model.SetShortTakeProfit.Error())
}

// TestShortStopLossFalse check is StopLoss value is valid, value should be invalid in this test
func TestShortStopLossFalse(t *testing.T) {
	lastPriceValue := float64(7)

	p := &model.Position{
		Amount:    10,
		OpenPrice: 5,
		LastPrice: &lastPriceValue,
		Short:     true,
	}

	err := p.SetStopLoss(9)

	require.EqualError(t, err, model.SetShortStopLoss.Error())
}

// TestShortStopLoss check is StopLoss value is valid
// and trigger to stop loss if all conditions is true
func TestShortStopLoss(t *testing.T) {
	lastPriceValue := float64(10)

	p := &model.Position{
		Amount:    10,
		OpenPrice: 5,
		LastPrice: &lastPriceValue,
		Short:     true,
	}

	err := p.SetStopLoss(6)
	if err != nil {
		t.Error(err)
	}

	tr, err := p.TriggerStopLoss()
	if err != nil {
		t.Error(err)
	}

	require.True(t, tr)
}
