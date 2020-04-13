package middleware

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/chavacava/tuilx/internal/tile"
	"github.com/chavacava/tuilx/internal/tile/source"
)

// Delayer is a tile source that returns the tile in at least n milliseconds
// where min < n < max
// useful for testing and simulations
type Delayer struct {
	min   uint
	delta uint // in milliseconds
	name  string
	r     *rand.Rand
	src   source.Source
}

// New yields a fresh new Delayer tile retriever
func NewDelayer(min, max uint, src source.Source, name string) (*Delayer, error) {
	delta := max - min
	if delta <= 0 {
		return nil, fmt.Errorf("min must be smaller than max, got %d and %d", min, max)
	}
	return &Delayer{min: min, delta: delta, r: rand.New(rand.NewSource(time.Now().UnixNano())), src: src, name: name}, nil
}

func (d Delayer) Name() string {
	return d.name
}

func (d Delayer) GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) {
	delay := int(d.min) + rand.Intn(int(d.delta))

	start := time.Now()
	tile, err := d.src.GetTile(ctx, props)
	elapsed := time.Since(start)

	wait := int64(delay) - (elapsed.Nanoseconds() / 1000000)
	if wait > 0 {
		_ = <-time.After(time.Duration(wait) * time.Millisecond)
	}

	return tile, err
}
