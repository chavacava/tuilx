package middleware

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/chavacava/lab-tileserver/internal/tile"
	"github.com/chavacava/lab-tileserver/internal/tile/source"
)

type Timer struct {
	s source.Source
}

func Time(s source.Source) *Timer {
	return &Timer{s}
}

func (t Timer) Name() string {
	return t.s.Name()
}

type timedContextKey string

const NestingLevel timedContextKey = "tckn"

func (t Timer) GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) {
	nl, _ := ctx.Value(NestingLevel).(int)
	ctx = context.WithValue(ctx, NestingLevel, nl+1)

	defer timeTrack(ctx, time.Now(), t.Name(), nl)
	return t.s.GetTile(ctx, props)
}

func timeTrack(ctx context.Context, start time.Time, name string, nl int) {
	elapsed := time.Since(start)
	rc := "├"
	if nl == 0 {
		rc = "└"
	}
	log.Printf("%s%s %s %s", rc, strings.Repeat("─", nl), name, elapsed)
}
