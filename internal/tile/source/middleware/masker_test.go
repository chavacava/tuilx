package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/Nr90/imgsim"
	"github.com/chavacava/lab-tileserver/internal/tile"
	"github.com/chavacava/lab-tileserver/internal/tile/source"
)

func BenchmarkMask(b *testing.B) {
	mk, err := source.NewMockedFromFile("../../../testdata/mask.png", tile.PNG, "mask")
	if err != nil {
		b.Fatalf("error while creating mask mock: %v", err)
	}
	black, err := source.NewMockedFromFile("../../../testdata/black.png", tile.PNG, "black")
	if err != nil {
		b.Fatalf("error while creating black mock: %v", err)
	}
	blue, err := source.NewMockedFromFile("../../../testdata/blue.jpeg", tile.JPEG, "blue")
	if err != nil {
		b.Fatalf("error while creating blue mock: %v", err)
	}

	var ctx = context.Background()
	var s, _ = blue.GetTile(ctx, tile.Properties{Size: 256})
	var r, _ = black.GetTile(ctx, tile.Properties{Size: 256})
	var ma, _ = mk.GetTile(ctx, tile.Properties{Size: 256})

	for n := 0; n < b.N; n++ {
		mask(s, r, ma)
	}
}

func TestMasker(t *testing.T) {

	tt := map[string]struct {
		src       string
		srcCoding tile.EncodingType
		rep       string
		repCoding tile.EncodingType
		msk       string
		mskCoding tile.EncodingType
		exp       string
		expCoding tile.EncodingType
		err       string
	}{
		"nominal": {
			"../../../testdata/blue.jpeg", tile.JPEG,
			"../../../testdata/black.png", tile.PNG,
			"../../../testdata/mask.png", tile.PNG,
			"../../../testdata/blueMaskedByBlack.jpeg", tile.JPEG,
			"",
		},
		"not same size": {
			"../../../testdata/gopher.png", tile.PNG,
			"../../../testdata/black.png", tile.PNG,
			"../../../testdata/mask.png", tile.PNG,
			"../../../testdata/blueMaskedByBlack.jpeg", tile.JPEG,
			"[masked] error while masking: can not mask image, sizes do not match",
		},
		"error getting source tile": {
			"error", tile.PNG,
			"../../../testdata/black.png", tile.PNG,
			"../../../testdata/mask.png", tile.PNG,
			"../../../testdata/blueMaskedByBlack.jpeg", tile.JPEG,
			"[masked] error while masking: [masked] error getting primary tile: you asked for an error",
		},
		"error getting replacement tile": {
			"../../../testdata/black.png", tile.PNG,
			"error", tile.PNG,
			"../../../testdata/mask.png", tile.PNG,
			"../../../testdata/blueMaskedByBlack.jpeg", tile.JPEG,
			"[masked] error while masking: [masked] error getting replacement tile: you asked for an error",
		},
		"error getting mask tile": {
			"../../../testdata/black.png", tile.PNG,
			"../../../testdata/black.png", tile.PNG,
			"error", tile.PNG,
			"../../../testdata/blueMaskedByBlack.jpeg", tile.JPEG,
			"[masked] error while masking: [masked] error getting mask tile: you asked for an error",
		},
	}

	for _, tc := range tt {
		mk := buildSource(tc.msk, tc.mskCoding, "mask", t)

		black := buildSource(tc.rep, tc.repCoding, "black", t)

		blue := buildSource(tc.src, tc.srcCoding, "blue", t)

		var ctx = context.Background()
		masker := NewMasker(blue, mk, black)
		result, err := masker.GetTile(ctx, tile.Properties{})
		if err != nil {
			if err.Error() != tc.err {
				t.Fatalf("expected error '%v', got '%v'", tc.err, err)
			}
			continue
		}

		masked, err := source.NewMockedFromFile(tc.exp, tc.expCoding, "masked")
		if err != nil {
			t.Fatalf("error while creating masker reference: %v", err)
		}
		ref, err := masked.GetTile(ctx, tile.Properties{})
		if err != nil {
			t.Fatalf("unexpected masked.GetTile error: %v", err)
		}
		h1 := imgsim.AverageHash(ref.Data)
		h2 := imgsim.AverageHash(result.Data)
		if h1 != h2 {
			t.Fatal("resulting tile data is not the expected one")
		}
	}
}

func buildSource(file string, enc tile.EncodingType, name string, t *testing.T) source.Source {
	if file == "error" {
		return &errorSource{}
	}

	r, err := source.NewMockedFromFile(file, enc, name)
	if err != nil {
		t.Fatalf("error while creating mask %s: %v", name, err)
	}

	return r
}

type errorSource struct{}

func (e errorSource) GetTile(ctx context.Context, props tile.Properties) (*tile.Tile, error) {
	return nil, errors.New("you asked for an error")
}

func (errorSource) Name() string { return "error" }
