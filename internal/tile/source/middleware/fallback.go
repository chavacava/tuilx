package middleware

import (
	"context"
	"log"

	"github.com/chavacava/tuilx/internal/tile"
	"github.com/chavacava/tuilx/internal/tile/source"
)

type Fallback struct {
	primary  source.Source
	fallback source.Source
}

func NewFallback(primary, fallback source.Source) *Fallback {
	return &Fallback{primary: primary, fallback: fallback}
}

func (Fallback) Name() string {
	return "fallback"
}

func (f Fallback) GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) {
	t, err := f.primary.GetTile(ctx, props)
	if err == nil {
		return t, err
	}

	log.Printf("Returning tile from the fallback because: %v", err)
	return f.fallback.GetTile(ctx, props)
}
