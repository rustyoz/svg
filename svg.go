package svg

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"

	mt "github.com/rustyoz/Mtransform"
)

// Tuple is an X,Y coordinate
type Tuple [2]float64

// Svg represents an SVG file containing at least a top level group or a
// number of Paths
type Svg struct {
	Title     string  `xml:"title"`
	Groups    []Group `xml:"g"`
	Paths     []Path  `xml:"path"`
	Name      string
	Transform *mt.Transform
	scale     float64
}

// Group represents an SVG group (usually located in a 'g' XML element)
type Group struct {
	ID              string
	Stroke          string
	StrokeWidth     int32
	Fill            string
	FillRule        string
	Elements        []interface{}
	TransformString string
	Transform       *mt.Transform // row, column
	Parent          *Group
	Owner           *Svg
}

// UnmarshalXML implements the encoding.xml.Unmarshaler interface
func (g *Group) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "id":
			g.ID = attr.Value
		case "stroke":
			g.Stroke = attr.Value
		case "stroke-width":
			intValue, err := strconv.ParseInt(attr.Value, 10, 32)
			if err != nil {
				return err
			}
			g.StrokeWidth = int32(intValue)
		case "fill":
			g.Fill = attr.Value
		case "fill-rule":
			g.FillRule = attr.Value
		case "transform":
			g.TransformString = attr.Value
			t, err := parseTransform(g.TransformString)
			if err != nil {
				fmt.Println(err)
			}
			g.Transform = &t
		}
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}

		switch tok := token.(type) {
		case xml.StartElement:
			var elementStruct interface{}

			switch tok.Name.Local {
			case "g":
				elementStruct = &Group{Parent: g, Owner: g.Owner, Transform: mt.NewTransform()}
			case "rect":
				elementStruct = &Rect{group: g}
			case "path":
				elementStruct = &Path{group: g, strokeWidth: 1}

			}

			if err = decoder.DecodeElement(elementStruct, &tok); err != nil {
				return fmt.Errorf("Error decoding element of Group\n%s", err)
			}
			g.Elements = append(g.Elements, elementStruct)

		case xml.EndElement:
			return nil
		}
	}
}

// ParseSvg parses an SVG string into an SVG struct
func ParseSvg(str string, name string, scale float64) (*Svg, error) {
	var svg Svg
	svg.Name = name
	svg.Transform = mt.NewTransform()
	if scale > 0 {
		svg.Transform.Scale(scale, scale)
		svg.scale = scale
	}
	if scale < 0 {
		svg.Transform.Scale(1.0/-scale, 1.0/-scale)
		svg.scale = 1.0 / -scale
	}

	err := xml.Unmarshal([]byte(str), &svg)
	if err != nil {
		return nil, fmt.Errorf("ParseSvg Error: %v", err)
	}
	fmt.Println(len(svg.Groups))
	for i := range svg.Groups {
		svg.Groups[i].SetOwner(&svg)
		if svg.Groups[i].Transform == nil {
			svg.Groups[i].Transform = mt.NewTransform()
		}
	}
	return &svg, nil
}

// ParseSvgFromReader parses an SVG struct from an io.Reader
func ParseSvgFromReader(r io.Reader, name string, scale float64) (*Svg, error) {
	var svg Svg
	svg.Name = name
	svg.Transform = mt.NewTransform()
	if scale > 0 {
		svg.Transform.Scale(scale, scale)
		svg.scale = scale
	}
	if scale < 0 {
		svg.Transform.Scale(1.0/-scale, 1.0/-scale)
		svg.scale = 1.0 / -scale
	}

	if err := xml.NewDecoder(r).Decode(&svg); err != nil {
		return nil, fmt.Errorf("ParseSvg Error: %v", err)
	}

	fmt.Println(len(svg.Groups))

	for i := range svg.Groups {
		svg.Groups[i].SetOwner(&svg)
		if svg.Groups[i].Transform == nil {
			svg.Groups[i].Transform = mt.NewTransform()
		}
	}
	return &svg, nil
}

// SetOwner sets the owner of a SVG Group
func (g *Group) SetOwner(svg *Svg) {
	g.Owner = svg
	for _, gn := range g.Elements {
		switch gn.(type) {
		case *Group:
			gn.(*Group).Owner = g.Owner
			gn.(*Group).SetOwner(svg)
		case *Path:
			gn.(*Path).group = g
		}
	}
}
