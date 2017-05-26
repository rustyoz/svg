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
