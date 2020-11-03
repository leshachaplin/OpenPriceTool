// Package with types and their methods

package model

import (
	"context"
)

// PositionCloser interface with close position method;
// Close should be implement
type PositionCloser interface {
	Close(ctx context.Context, username, symbol, priceUUID string) error
}
