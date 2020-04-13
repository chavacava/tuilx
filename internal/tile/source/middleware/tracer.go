package middleware

import (
	"context"

	"github.com/chavacava/tuilx/internal/tile"
	"github.com/chavacava/tuilx/internal/tile/source"
	"go.opencensus.io/trace"
)

type Tracer struct {
	s source.Source
}

func Trace(s source.Source) *Tracer {
	return &Tracer{s}
}

func (t Tracer) Name() string {
	return t.s.Name()
}

func (t Tracer) GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) {
	ctx, span := trace.StartSpan(ctx, t.Name())
	defer span.End()

	return t.s.GetTile(ctx, props)
}
