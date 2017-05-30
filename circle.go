package svg

import mt "github.com/rustyoz/Mtransform"

// Circle is an SVG circle element
type Circle struct {
	ID        string `xml:"id,attr"`
	Transform string `xml:"transform,attr"`
	Style     string `xml:"style,attr"`
	Cx        string `xml:"cx,attr"`
	Cy        string `xml:"cy,attr"`
	Radius    string `xml:"r,attr"`

	transform mt.Transform
	group     *Group
}

// ParseDrawingInstructions implements the DrawingInstructionParser
// interface
func (c *Circle) ParseDrawingInstructions() (chan Segment, chan DrawingInstruction) {
	return make(chan Segment), make(chan DrawingInstruction)
}
