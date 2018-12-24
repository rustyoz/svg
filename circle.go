package svg

import mt "github.com/rustyoz/Mtransform"

//Circle contains the data from a circle svg element
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
