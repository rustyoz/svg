package svg

import (
	"fmt"
	"strconv"

	mt "github.com/rustyoz/Mtransform"
	gl "github.com/rustyoz/genericlexer"
)

type Path struct {
	ID              string `xml:"id,attr"`
	D               string `xml:"d,attr"`
	Style           string `xml:"style,attr"`
	TransformString string `xml:"transform,attr"`
	properties      map[string]string
	strokeWidth     float64
	Segments        chan Segment
	group           *Group
}

// Segment
// A segment of a path that contains a list of connected points, its stroke Width and if the segment forms a closed loop.
// Points are defined in world space after any matrix transformation is applied.
type Segment struct {
	Width  float64
	Closed bool
	Points [][2]float64
}

func (p Path) newSegment(start [2]float64) *Segment {
	var s Segment
	s.Width = p.strokeWidth * p.group.Owner.scale
	s.Points = append(s.Points, start)
	return &s
}

func (s *Segment) addPoint(p [2]float64) {
	s.Points = append(s.Points, p)
}

type pathDescriptionParser struct {
	p              *Path
	lex            gl.Lexer
	x, y           float64
	currentcommand int
	tokbuf         [4]gl.Item
	peekcount      int
	lasttuple      Tuple
	transform      mt.Transform
	svg            *Svg
	currentsegment *Segment
}

func newPathDParse() *pathDescriptionParser {
	pdp := &pathDescriptionParser{}
	pdp.transform = mt.Identity()
	return pdp
}

// Parse()
// interprets path description, transform and style atttributes to create a channel of segments.
func (p *Path) Parse() chan Segment {
	p.parseStyle()
	pdp := newPathDParse()
	pdp.p = p
	pdp.svg = p.group.Owner
	pathTransform := mt.Identity()
	if p.TransformString != "" {
		pt, err := parseTransform(p.TransformString)
		if err == nil {
			pathTransform = pt
		}
	}
	pdp.transform = mt.MultiplyTransforms(pdp.transform, *p.group.Transform)
	pdp.transform = mt.MultiplyTransforms(pdp.transform, pathTransform)
	p.Segments = make(chan Segment)
	l, _ := gl.Lex(fmt.Sprint(p.ID), p.D)
	pdp.lex = *l
	go func() {
		defer close(p.Segments)
		for {
			i := pdp.lex.NextItem()
			switch {
			case i.Type == gl.ItemError:
				return
			case i.Type == gl.ItemEOS:
				if pdp.currentsegment != nil {
					p.Segments <- *pdp.currentsegment
				}
				return
			case i.Type == gl.ItemLetter:
				pdp.parseCommand(l, i)
			default:
			}
		}
	}()
	return p.Segments
}

func (pdp *pathDescriptionParser) parseCommand(l *gl.Lexer, i gl.Item) error {
	var err error
	switch i.Value {
	case "M":
		err = pdp.parseMoveToAbs()
	case "m":
		err = pdp.parseMoveToRel()
	case "c":
		err = pdp.parseCurveToRel()
	case "C":
		err = pdp.parseCurveToAbs()
	case "L":
		err = pdp.parseLineToAbs()
	case "l":
		err = pdp.parseLineToRel()
	case "H":
		err = pdp.parseHLineToAbs()
	case "h":
		err = pdp.parseHLineToRel()
	case "Z":
	case "z":
		err = pdp.parseClose()
	}
	return err

}

func (pdp *pathDescriptionParser) parseMoveToAbs() error {
	t, err := parseTuple(&pdp.lex)
	if err != nil {
		return fmt.Errorf("Error Passing MoveToAbs Expected Tuple\n%s", err)
	}

	pdp.x = t[0]
	pdp.y = t[1]

	var tuples []Tuple
	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing MoveToAbs\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}

	if pdp.currentsegment != nil {
		pdp.p.Segments <- *pdp.currentsegment
		pdp.currentsegment = nil
	} else {

		var s Segment
		s.Width = pdp.p.strokeWidth * pdp.p.group.Owner.scale
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		s.addPoint([2]float64{x, y})
		pdp.currentsegment = &s

	}

	if len(tuples) > 0 {
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		s := pdp.p.newSegment([2]float64{x, y})
		for _, nt := range tuples {
			pdp.x = nt[0]
			pdp.y = nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			s.addPoint([2]float64{x, y})
		}
		pdp.currentsegment = s
	}
	return nil

}

func (pdp *pathDescriptionParser) parseLineToAbs() error {
	var tuples []Tuple
	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing LineToAbs\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}
	if len(tuples) > 0 {
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		pdp.currentsegment.addPoint([2]float64{x, y})
		for _, nt := range tuples {
			pdp.x = nt[0]
			pdp.y = nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			pdp.currentsegment.addPoint([2]float64{x, y})
		}
	}

	return nil

}

func (pdp *pathDescriptionParser) parseMoveToRel() error {
	pdp.lex.ConsumeWhiteSpace()
	t, err := parseTuple(&pdp.lex)
	if err != nil {
		return fmt.Errorf("Error Passing MoveToRel Expected First Tuple\n%s", err)
	}

	pdp.x += t[0]
	pdp.y += t[1]

	var tuples []Tuple
	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing MoveToRel\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}
	if pdp.currentsegment != nil {
		pdp.p.Segments <- *pdp.currentsegment
		pdp.currentsegment = nil
	} else {
		var s Segment
		s.Width = pdp.p.strokeWidth * pdp.svg.scale
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		s.addPoint([2]float64{x, y})
		pdp.currentsegment = &s
	}
	if len(tuples) > 0 {
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		pdp.currentsegment.addPoint([2]float64{x, y})
		for _, nt := range tuples {
			pdp.x += nt[0]
			pdp.y += nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			pdp.currentsegment.addPoint([2]float64{x, y})
		}
	}

	return nil
}

func (pdp *pathDescriptionParser) parseLineToRel() error {

	var tuples []Tuple
	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing LineToRel\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}
	if len(tuples) > 0 {
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		pdp.currentsegment.addPoint([2]float64{x, y})
		for _, nt := range tuples {
			pdp.x += nt[0]
			pdp.y += nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			pdp.currentsegment.addPoint([2]float64{x, y})
		}
	}

	return nil
}

func (pdp *pathDescriptionParser) parseHLineToAbs() error {
	pdp.lex.ConsumeWhiteSpace()
	var n float64
	var err error
	if pdp.lex.PeekItem().Type != gl.ItemNumber {
		n, err = parseNumber(pdp.lex.NextItem())
		if err != nil {
			return fmt.Errorf("Error Passing HLineToAbs\n%s", err)
		}
	}

	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.addPoint([2]float64{x, y})
	pdp.x = n
	x, y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.addPoint([2]float64{x, y})

	return nil
}

func (pdp *pathDescriptionParser) parseHLineToRel() error {
	pdp.lex.ConsumeWhiteSpace()
	var n float64
	var err error
	if pdp.lex.PeekItem().Type != gl.ItemNumber {
		n, err = parseNumber(pdp.lex.NextItem())
		if err != nil {
			return fmt.Errorf("Error Passing HLineToRel\n%s", err)
		}
	}

	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.addPoint([2]float64{x, y})
	pdp.x += n
	x, y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.addPoint([2]float64{x, y})

	return nil

}

func (pdp *pathDescriptionParser) parseVLineToAbs() error {
	pdp.lex.ConsumeWhiteSpace()
	var n float64
	var err error
	if pdp.lex.PeekItem().Type != gl.ItemNumber {
		n, err = parseNumber(pdp.lex.NextItem())
		if err != nil {
			return fmt.Errorf("Error Passing VLineToAbs\n%s", err)
		}
	}

	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.addPoint([2]float64{x, y})
	pdp.y = n
	x, y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.addPoint([2]float64{x, y})

	return nil
}

func (pdp *pathDescriptionParser) parseClose() error {
	pdp.lex.ConsumeWhiteSpace()
	if pdp.currentsegment != nil {
		pdp.currentsegment.addPoint(pdp.currentsegment.Points[0])
		pdp.currentsegment.Closed = true
		pdp.p.Segments <- *pdp.currentsegment
		pdp.currentsegment = nil
		return nil
	}
	return fmt.Errorf("Error Parsing closepath command, no previous path")

}

func (pdp *pathDescriptionParser) parseVLineToRel() error {
	pdp.lex.ConsumeWhiteSpace()
	var n float64
	var err error
	if pdp.lex.PeekItem().Type != gl.ItemNumber {
		n, err = parseNumber(pdp.lex.NextItem())
		if err != nil {
			return fmt.Errorf("Error Passing VLineToRel\n%s", err)
		}
	}

	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.addPoint([2]float64{x, y})
	pdp.y += n
	x, y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.addPoint([2]float64{x, y})

	return nil

}

func (pdp *pathDescriptionParser) parseCurveToRel() error {
	var tuples []Tuple
	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing CurveToRel\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}
	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.addPoint([2]float64{x, y})

	for j := 0; j < len(tuples)/3; j++ {
		var cb cubicBezier
		cb.controlpoints[0][0] = pdp.x
		cb.controlpoints[0][1] = pdp.y

		cb.controlpoints[1][0] = pdp.x + tuples[j*3][0]
		cb.controlpoints[1][1] = pdp.y + tuples[j*3][1]

		cb.controlpoints[2][0] = pdp.x + tuples[j*3+1][0]
		cb.controlpoints[2][1] = pdp.y + tuples[j*3+1][1]

		pdp.x += tuples[j*3+2][0]
		pdp.y += tuples[j*3+2][1]

		cb.controlpoints[3][0] = pdp.x
		cb.controlpoints[3][1] = pdp.y

		vertices := cb.recursiveInterpolate(10, 0)
		for _, v := range vertices {
			x, y = pdp.transform.Apply(v[0], v[1])
			pdp.currentsegment.addPoint([2]float64{x, y})
		}
	}

	return nil
}

func (pdp *pathDescriptionParser) parseCurveToAbs() error {
	var tuples []Tuple
	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing CurveToRel\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}

	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.addPoint([2]float64{x, y})

	for j := 0; j < len(tuples)/3; j++ {
		var cb cubicBezier
		cb.controlpoints[0][0] = pdp.x
		cb.controlpoints[0][1] = pdp.y
		for i, nt := range tuples[j*3 : (j+1)*3] {
			pdp.x = nt[0]
			pdp.y = nt[1]
			cb.controlpoints[i+1][0] = pdp.x
			cb.controlpoints[i+1][1] = pdp.y
		}
		vertices := cb.recursiveInterpolate(10, 0)
		for _, v := range vertices {
			x, y = pdp.transform.Apply(v[0], v[1])
			pdp.currentsegment.addPoint([2]float64{x, y})
		}
	}

	return nil
}

func (p *Path) parseStyle() {
	p.properties = splitStyle(p.Style)
	for key, val := range p.properties {
		switch key {
		case "stroke-width":
			sw, ok := strconv.ParseFloat(val, 64)
			if ok == nil {
				p.strokeWidth = sw
			}

		}
	}
}
