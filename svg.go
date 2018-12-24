package svg

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	mt "github.com/smallpdf/Mtransform"
)

// DrawingInstructionParser allow getting segments and drawing
// instructions from them. All SVG elements should implement this
// interface.
type DrawingInstructionParser interface {
	ParseDrawingInstructions() (chan *DrawingInstruction, chan error)
}

// Tuple is an X,Y coordinate
type Tuple [2]float64

// Svg represents an SVG file containing at least a top level group or a
// number of Paths
type Svg struct {
	Title        string  `xml:"title"`
	Groups       []Group `xml:"g"`
	ViewBox      string  `xml:"viewBox,attr"`
	Elements     []DrawingInstructionParser
	Name         string
	Transform    *mt.Transform
	scale        float64
	instructions chan *DrawingInstruction
	errors       chan error
	segments     chan Segment
}

// Group represents an SVG group (usually located in a 'g' XML element)
type Group struct {
	ID              string
	Stroke          string
	StrokeWidth     int32
	Fill            string
	FillRule        string
	Elements        []DrawingInstructionParser
	TransformString string
	Transform       *mt.Transform // row, column
	Parent          *Group
	Owner           *Svg
	instructions    chan *DrawingInstruction
	errors          chan error
	segments        chan Segment
}

// ParseDrawingInstructions implements the DrawingInstructionParser interface
//
// This method makes it easier to get all the drawing instructions.
func (g *Group) ParseDrawingInstructions() (chan *DrawingInstruction, chan error) {
	g.instructions = make(chan *DrawingInstruction, 100)
	g.errors = make(chan error, 100)

	errWg := &sync.WaitGroup{}

	go func() {
		defer close(g.instructions)
		defer func() { errWg.Wait(); close(g.errors) }()
		for _, e := range g.Elements {
			instrs, errs := e.ParseDrawingInstructions()
			errWg.Add(1)
			go func() {
				for er := range errs {
					g.errors <- er
				}
				errWg.Done()
			}()
			for is := range instrs {
				g.instructions <- is
			}
		}
	}()

	return g.instructions, g.errors
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
			var elementStruct DrawingInstructionParser

			switch tok.Name.Local {
			case "g":
				elementStruct = &Group{Parent: g, Owner: g.Owner, Transform: mt.NewTransform()}
			case "rect":
				elementStruct = &Rect{group: g}
			case "circle":
				elementStruct = &Circle{group: g}
			case "path":
				elementStruct = &Path{group: g, StrokeWidth: float64(g.StrokeWidth), Stroke: &g.Stroke, Fill: &g.Fill}
			default:
				continue
			}
			if err = decoder.DecodeElement(elementStruct, &tok); err != nil {
				return fmt.Errorf("error decoding element of Group: %s", err)
			}
			g.Elements = append(g.Elements, elementStruct)
		case xml.EndElement:
			if tok.Name.Local == "g" {
				return nil
			}
		}
	}
}

// ParseDrawingInstructions implements the DrawingInstructionParser interface
//
// This method makes it easier to get all the drawing instructions.
func (s *Svg) ParseDrawingInstructions() (chan *DrawingInstruction, chan error) {
	s.instructions = make(chan *DrawingInstruction, 100)
	s.errors = make(chan error, 100)

	go func() {
		errWg := &sync.WaitGroup{}
		var elecount int
		defer close(s.instructions)
		defer func() { errWg.Wait(); close(s.errors) }()
		for _, e := range s.Elements {
			elecount++
			instrs, errs := e.ParseDrawingInstructions()
			errWg.Add(1)
			go func(count int) {
				for er := range errs {
					s.errors <- fmt.Errorf("error when parsing element nr. %d: %s", count, er)
				}
				errWg.Done()
			}(elecount)

			for is := range instrs {
				s.instructions <- is
			}
		}

		for _, g := range s.Groups {
			instrs, errs := g.ParseDrawingInstructions()
			errWg.Add(1)
			go func() {
				for er := range errs {
					s.errors <- er
				}
				errWg.Done()
			}()
			for is := range instrs {
				s.instructions <- is
			}
		}
	}()

	return s.instructions, s.errors
}

// UnmarshalXML implements the encoding.xml.Unmarshaler interface
func (s *Svg) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	for {
		for _, attr := range start.Attr {
			if attr.Name.Local == "viewBox" {
				s.ViewBox = attr.Value
				break
			}
		}

		token, err := decoder.Token()
		if err != nil {
			return err
		}

		switch tok := token.(type) {
		case xml.StartElement:
			var dip DrawingInstructionParser

			switch tok.Name.Local {
			case "g":
				g := &Group{Owner: s, Transform: mt.NewTransform()}
				if err = decoder.DecodeElement(g, &tok); err != nil {
					return fmt.Errorf("error decoding group element within SVG struct: %s", err)
				}
				s.Groups = append(s.Groups, *g)
				continue
			case "rect":
				dip = &Rect{}
			case "circle":
				dip = &Circle{}
			case "path":
				dip = &Path{}

			default:
				continue
			}

			if err = decoder.DecodeElement(dip, &tok); err != nil {
				return fmt.Errorf("error decoding element of SVG struct: %s", err)
			}

			s.Elements = append(s.Elements, dip)

		case xml.EndElement:
			if tok.Name.Local == "svg" {
				return nil
			}
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

	for i := range svg.Groups {
		svg.Groups[i].SetOwner(&svg)
		if svg.Groups[i].Transform == nil {
			svg.Groups[i].Transform = mt.NewTransform()
		}
	}
	return &svg, nil
}

// ViewBoxValues returns all the numerical values in the viewBox
// attribute.
func (s *Svg) ViewBoxValues() ([]float64, error) {
	var vals []float64

	if s.ViewBox == "" {
		return vals, errors.New("viewBox attribute is empty")
	}

	split := strings.Split(s.ViewBox, " ")

	for _, val := range split {
		ival, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return vals, err
		}
		vals = append(vals, ival)
	}

	return vals, nil
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
