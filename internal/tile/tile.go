package tile

import (
	"fmt"
	"image"
)

type Properties struct {
	Col, Row, Zoom uint64
	Size           uint16 // rectangular tile
	Encoding       EncodingType
}

func (p Properties) String() string {
	return fmt.Sprintf("%d/%d/%d/%d.%s", p.Size, p.Zoom, p.Col, p.Row, p.Encoding)
}

type Tile struct {
	Props Properties
	Data  image.Image
}

func (t Tile) String() string {
	return fmt.Sprintf("%s (%d bytes)", t.Props.String(), 666)
}

type EncodingType string

const JPEG EncodingType = "image/jpeg"
const PNG EncodingType = "image/png"
