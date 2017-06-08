package svg

import mt "github.com/smallpdf/Mtransform"

// PolyLine is a set of connected line segments that typically form a
// closed shape
type PolyLine struct {
	ID        string `xml:"id,attr"`
	Transform string `xml:"transform,attr"`
	Style     string `xml:"style,attr"`
	Points    string `xml:"points,attr"`

	transform mt.Transform
	group     *Group
}
