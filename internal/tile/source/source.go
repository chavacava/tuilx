package source

import (
	"context"

	"github.com/chavacava/tuilx/internal/tile"
)

// Source models the behavior of a tile retriever
type Source interface {
	GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error)
	Name() string
}
