package middleware

import (
	"context"

	"github.com/chavacava/lab-tileserver/internal/tile"
	"github.com/chavacava/lab-tileserver/internal/tile/source"
)

type Switcher struct {
	method SwitchingMethod
	name   string
}

func (s Switcher) Name() string {
	return s.name
}

type SwitchingMethod func(ctx context.Context, props tile.Properties) source.Source

func NewSwitcher(name string, method SwitchingMethod) *Switcher {
	return &Switcher{method: method, name: name}
}

func (s Switcher) GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) {
	src := s.method(ctx, props)

	return src.GetTile(ctx, props)
}
