package middleware

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"sync"

	"github.com/chavacava/tuilx/internal/tile"
	"github.com/chavacava/tuilx/internal/tile/source"
	"go.opencensus.io/trace"
)

type Masker struct {
	primary     source.Source
	mask        source.Source
	replacement source.Source
}

func NewMasker(primary, mask, replacement source.Source) *Masker {
	return &Masker{primary, mask, replacement}
}

func (Masker) Name() string {
	return "masked"
}

func (f Masker) GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) {
	tile, err := f.getMaskerTile(ctx, props)
	if err != nil {
		return nil, fmt.Errorf("[%s] error while masking: %v", f.Name(), err)
	}

	return tile, nil
}

func (f Masker) getMaskerTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) { // TODO parallelize

	ptc := make(chan *tile.Tile)
	mtc := make(chan *tile.Tile)
	rtc := make(chan *tile.Tile)
	errc := make(chan error)

	go func() {
		t, err := f.primary.GetTile(ctx, props)
		if err != nil {
			errc <- fmt.Errorf("[%s] error getting primary tile: %v", f.Name(), err)
		}
		ptc <- t
	}()

	go func() {
		t, err := f.mask.GetTile(ctx, props)
		if err != nil {
			errc <- fmt.Errorf("[%s] error getting mask tile: %v", f.Name(), err)
		}
		mtc <- t
	}()

	go func() {
		t, err := f.replacement.GetTile(ctx, props)
		if err != nil {
			errc <- fmt.Errorf("[%s] error getting replacement tile: %v", f.Name(), err)
		}
		rtc <- t
	}()

	var primaryTile, replacementTile, maskTile *tile.Tile
	for i := 0; i < 3; i++ {
		select {
		case primaryTile = <-ptc:
		case replacementTile = <-rtc:
		case maskTile = <-mtc:
		case err := <-errc:
			return nil, err
		}
	}

	maskedImg, err := mask(ctx, primaryTile, replacementTile, maskTile)
	if err != nil {
		return nil, err
	}

	props.Encoding = tile.JPEG
	result := &tile.Tile{Props: props, Data: maskedImg}
	return result, nil
}

var transparent = color.NRGBA{0, 0, 0, 0}

func mask(ctx context.Context, src, repl, msk *tile.Tile) (image.Image, error) {
	_, span := trace.StartSpan(ctx, "mask")
	defer span.End()

	source, replacement, mask := src.Data, repl.Data, msk.Data
	if !sameSize(source, replacement, mask) {
		return nil, fmt.Errorf("can not mask image, sizes do not match")
	}
	img := image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{255, 255}})

	var wg sync.WaitGroup
	mb := mask.Bounds()
	for y := mb.Min.Y; y < mb.Max.Y; y++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()

			for x := mb.Min.X; x < mb.Max.X; x++ {
				if c := mask.At(x, y); c != transparent {
					img.Set(x, y, replacement.At(x, y))
				} else {
					img.Set(x, y, source.At(x, y))
				}
			}
		}(y)
	}

	wg.Wait()

	return img, nil
}

func sameSize(source, replacement, mask image.Image) bool {
	return (source.Bounds().Dx() == replacement.Bounds().Dx() && replacement.Bounds().Dx() == mask.Bounds().Dx()) &&
		(source.Bounds().Dy() == replacement.Bounds().Dy() && replacement.Bounds().Dy() == mask.Bounds().Dy())
}
