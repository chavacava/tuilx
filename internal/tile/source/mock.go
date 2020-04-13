package source

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"os"

	"github.com/chavacava/tuilx/internal/tile"
)

// MockedSource implements tile source by always returning the same tile
// useful for testing and simulations
type Mocked struct {
	data     image.Image
	encoding tile.EncodingType
	name     string
}

func NewMockedFromFile(path string, encoding tile.EncodingType, name string) (*Mocked, error) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Errorf("unable to open file '%s', got: %v", path, err)
	}

	return NewMocked(f, encoding, name)
}

func NewMocked(data io.Reader, encoding tile.EncodingType, name string) (*Mocked, error) {
	d, err := ioutil.ReadAll(data)
	if err != nil {
		return nil, err
	}

	var img image.Image
	switch encoding {
	case tile.PNG:
		img, err = png.Decode(bytes.NewBuffer(d))
	case tile.JPEG:
		img, err = jpeg.Decode(bytes.NewBuffer(d))
	default:
		return nil, fmt.Errorf("unhandled encoding '%v'", encoding)
	}
	if err != nil {
		return nil, err
	}

	return &Mocked{data: img, encoding: encoding, name: name}, nil
}

func (m Mocked) Name() string {
	return m.name
}

func (m Mocked) GetTile(_ context.Context, props tile.Properties) (*tile.Tile, error) {
	props.Encoding = m.encoding

	return &tile.Tile{Props: props, Data: m.data}, nil
}
