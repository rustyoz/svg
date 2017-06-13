package svg

import mt "github.com/smallpdf/Mtransform"

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
func (c *Circle) ParseDrawingInstructions() (chan *DrawingInstruction, chan error) {
	draw := make(chan *DrawingInstruction)
	errs := make(chan error)

	go func() {
		defer close(draw)
		defer close(errs)

		draw <- &DrawingInstruction{
			Kind:   CircleInstruction,
			M:      &Tuple{c.Cx, c.Cy},
			Radius: &c.Radius,
		}

		draw <- &DrawingInstruction{Kind: PaintInstruction, Fill: &c.Fill}
	}()

	return draw, errs
}
