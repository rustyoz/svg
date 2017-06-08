package svg

import mt "github.com/smallpdf/Mtransform"

// Line is an SVG XML line element
type Line struct {
	ID        string `xml:"id,attr"`
	Transform string `xml:"transform,attr"`
	Style     string `xml:"style,attr"`
	X1        string `xml:"x1,attr"`
	X2        string `xml:"x2,attr"`
	Y1        string `xml:"y1,attr"`
	Y2        string `xml:"y2,attr"`

	transform mt.Transform
	group     *Group
}
