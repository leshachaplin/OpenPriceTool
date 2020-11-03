// Package with types and their methods

package model

// PriceConverter interface with convert price method
// Convert should be implemented
type PriceConverter interface {
	Convert(from string, to string, value float64) float64
}
