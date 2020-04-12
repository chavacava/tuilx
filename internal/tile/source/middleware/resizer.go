package middleware

import (
	"context"
	"strconv"

	"github.com/anthonynsimon/bild/transform"
	"github.com/chavacava/lab-tileserver/internal/tile"
	"github.com/chavacava/lab-tileserver/internal/tile/source"
)

type SamplingMethod int

const (
	NearestNeighbor SamplingMethod = iota + 1
	Box
	Linear
	Gaussian
	MitchellNetravali
	CatmullRom
	Lanczos
)

var smMapping = map[SamplingMethod]transform.ResampleFilter{
	NearestNeighbor:   transform.NearestNeighbor,
	Box:               transform.Box,
	Linear:            transform.Linear,
	MitchellNetravali: transform.MitchellNetravali,
	CatmullRom:        transform.CatmullRom,
	Lanczos:           transform.Lanczos,
}

type Resizer struct {
	source     source.Source
	method     transform.ResampleFilter
	targetSize uint16
}

func NewResizer(source source.Source, sm SamplingMethod, targetSize uint16) *Resizer {
	return &Resizer{source, smMapping[sm], targetSize}
}

func (r Resizer) Name() string {
	return "resize-to-" + strconv.Itoa(int(r.targetSize))
}

func (r Resizer) GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) {
	t, err := r.source.GetTile(ctx, props)
	if err != nil {
		return nil, err
	}

	img := t.Data
	/*
		img, _, err := image.Decode(bytes.NewReader(t.Data))
		if err != nil {
			return nil, errors.Wrap(err, "unable to decode tile data")
		}
	*/
	resized := transform.Resize(img, int(r.targetSize), int(r.targetSize), r.method)
	/*
		buf := new(bytes.Buffer)
		if err := jpeg.Encode(buf, resized, &jpeg.Options{Quality: 100}); err != nil {
			return nil, errors.Wrapf(err, "unable to encode tile data (of type %T) as jpeg", resized)
		}
	*/
	props.Size = r.targetSize
	tile := &tile.Tile{Props: props, Data: resized}

	return tile, nil
}
