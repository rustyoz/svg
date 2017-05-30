package svg

import mt "github.com/rustyoz/Mtransform"

// Rect is an SVG XML rect element
type Rect struct {
	ID        string `xml:"id,attr"`
	Width     string `xml:"width,attr"`
	Height    string `xml:"height,attr"`
	Transform string `xml:"transform,attr"`
	Style     string `xml:"style,attr"`
	Rx        string `xml:"rx,attr"`
	Ry        string `xml:"ry,attr"`

	transform mt.Transform
	group     *Group
}

// ParseDrawingInstructions implements the DrawingInstructionParser
// interface
func (r *Rect) ParseDrawingInstructions() (chan Segment, chan DrawingInstruction) {
	seg, draw := make(chan Segment), make(chan DrawingInstruction)

	defer close(seg)
	defer close(draw)

	return seg, draw
}
