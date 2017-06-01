package svg

import mt "github.com/rustyoz/Mtransform"

// Circle is an SVG circle element
type Circle struct {
	ID        string  `xml:"id,attr"`
	Transform string  `xml:"transform,attr"`
	Style     string  `xml:"style,attr"`
	Cx        float64 `xml:"cx,attr"`
	Cy        float64 `xml:"cy,attr"`
	Radius    float64 `xml:"r,attr"`
	Fill      string  `xml:"fill,attr"`

	transform mt.Transform
	group     *Group
}

// ParseDrawingInstructions implements the DrawingInstructionParser
// interface
func (c *Circle) ParseDrawingInstructions() (chan Segment, chan *DrawingInstruction) {
	seg, draw := make(chan Segment), make(chan *DrawingInstruction)

	go func() {
		defer close(seg)
		defer close(draw)

		draw <- &DrawingInstruction{
			Kind:   CircleInstruction,
			M:      &Tuple{c.Cx, c.Cy},
			Radius: &c.Radius,
		}

		draw <- &DrawingInstruction{Kind: PaintInstruction, Fill: &c.Fill}
	}()

	return seg, draw
}
