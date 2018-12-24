package svg

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"

	mt "github.com/rustyoz/Mtransform"
)

//Tuple is a container for 2 float64s
type Tuple [2]float64

//SVG is a container the svg
type SVG struct {
	Title     string  `xml:"title"`
	Groups    []Group `xml:"g"`
	Name      string
	Transform *mt.Transform
	scale     float64
}

//Group contains the data from a group svg element
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
	Owner           *SVG
}

//UnmarshalXML implements encoding.xml.Unmarshaler interface
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
			err = decoder.DecodeElement(elementStruct, &tok)
			if err != nil {
				return fmt.Errorf("Error decoding element of Group\n%s", err)
			}
			g.Elements = append(g.Elements, elementStruct)

		case xml.EndElement:
			return nil
		}
	}
}

//ParseSVG parses a string into the SVG type
func ParseSVG(str string, name string, scale float64) (*SVG, error) {
	var svg SVG
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
		return nil, fmt.Errorf("parseSvg Error: %v", err)
	}
	for i := range svg.Groups {
		svg.Groups[i].SetOwner(&svg)
		if svg.Groups[i].Transform == nil {
			svg.Groups[i].Transform = mt.NewTransform()
		}
	}
	return &svg, nil
}

//ParseReader parses an io.Reader into the SVG type
func ParseReader(reader io.Reader, name string, scale float64) (*SVG, error) {
	var svg SVG
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
	dec := xml.NewDecoder(reader)
	err := dec.Decode(&svg)
	if err != nil {
		return nil, fmt.Errorf("failled to parse SVG: %v", err)
	}
	for i := range svg.Groups {
		svg.Groups[i].SetOwner(&svg)
		if svg.Groups[i].Transform == nil {
			svg.Groups[i].Transform = mt.NewTransform()
		}
	}
	return &svg, nil
}

//SetOwner sets the owner of a group
func (g *Group) SetOwner(svg *SVG) {
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
