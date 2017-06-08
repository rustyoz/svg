package svg

import (
	"fmt"
	"strconv"

	mt "github.com/rustyoz/Mtransform"
	gl "github.com/rustyoz/genericlexer"
)

// Path is an SVG XML path element
type Path struct {
	ID              string `xml:"id,attr"`
	D               string `xml:"d,attr"`
	Style           string `xml:"style,attr"`
	TransformString string `xml:"transform,attr"`
	properties      map[string]string
	StrokeWidth     float64 `xml:"stroke-width,attr"`
	Fill            *string `xml:"fill,attr"`
	Stroke          string  `xml:"stroke,attr"`
	Segments        chan Segment
	instructions    chan *DrawingInstruction
	group           *Group
}

// A Segment of a path that contains a list of connected points, its
// stroke Width and if the segment forms a closed loop.  Points are
// defined in world space after any matrix transformation is applied.
type Segment struct {
	Width  float64
	Closed bool
	Points [][2]float64
}

func (p Path) newSegment(start [2]float64) *Segment {
	var s Segment
	s.Width = p.StrokeWidth * p.group.Owner.scale
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

// Parse interprets path description, transform and style atttributes to
// create a channel of segments.
func (p *Path) Parse() chan Segment {
	p.parseStyle()
	pdp := newPathDParse()
	pdp.p = p
	if p.group == nil {
		p.group = new(Group)
		temp := mt.Identity()
		p.group.Transform = &temp
	}
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

// ParseDrawingInstructions returns two channels. One is a channel of
// Segments identical to the one returned by Parse() and the other one
// is a channel of DrawingInstruction. The latter should be used to pass
// to a path drawing library (like Cairo or something comparable)
//
// Note that you have to drain both channels even if you don't need the
// results for one. Otherwise we will get a deadlock.
func (p *Path) ParseDrawingInstructions() (chan Segment, chan *DrawingInstruction) {
	p.parseStyle()
	pdp := newPathDParse()
	pdp.p = p
	if p.group == nil {
		p.group = new(Group)
		temp := mt.Identity()
		p.group.Transform = &temp
	}
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

	p.instructions = make(chan *DrawingInstruction, 100)
	p.Segments = make(chan Segment)
	l, _ := gl.Lex(fmt.Sprint(p.ID), p.D)

	pdp.lex = *l
	go func() {
		defer close(p.instructions)
		defer close(p.Segments)
		var count int
		for {
			i := pdp.lex.NextItem()
			count++
			switch {
			case i.Type == gl.ItemError:
				return
			case i.Type == gl.ItemEOS:
				if pdp.currentsegment != nil {
					p.Segments <- *pdp.currentsegment
				}

				scaledStrokeWidth := p.StrokeWidth * pdp.p.group.Owner.scale

				pdp.p.instructions <- &DrawingInstruction{
					Kind:        PaintInstruction,
					StrokeWidth: &scaledStrokeWidth,
					Stroke:      &p.Stroke,
					Fill:        p.Fill,
				}
				return
			case i.Type == gl.ItemLetter:
				err := pdp.parseCommandDrawingInstructions(l, i)
				if err != nil {
					fmt.Printf("ParseCommand error: %s\n", err)
				}

			default:
				fmt.Printf("Default invoked: %d item %v\n", count, i)
			}
		}
	}()

	return p.Segments, p.instructions
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
	case "z", "Z":
		err = pdp.parseClose()
	}

	return err
}

func (pdp *pathDescriptionParser) parseCommandDrawingInstructions(l *gl.Lexer, i gl.Item) error {
	var err error

	switch i.Value {
	case "M":
		err = pdp.parseMoveToAbsDI()
	case "m":
		err = pdp.parseMoveToRelDI()
	case "c":
		err = pdp.parseCurveToRelDI()
	case "C":
		err = pdp.parseCurveToAbsDI()
	case "L":
		err = pdp.parseLineToAbsDI()
	case "l":
		err = pdp.parseLineToRelDI()
	case "H":
		err = pdp.parseHLineToAbs()
	case "h":
		err = pdp.parseHLineToRel()
	case "z", "Z":
		err = pdp.parseCloseDI()
	}

	return err
}

func (pdp *pathDescriptionParser) parseMoveToAbsDI() error {
	var tuples []Tuple

	t, err := parseTuple(&pdp.lex)
	if err != nil {
		return fmt.Errorf("Error Passing MoveToAbs Expected Tuple\n%s", err)
	}

	pdp.x = t[0]
	pdp.y = t[1]

	if pdp.p.group.Owner == nil {
		pdp.p.group.Owner = &Svg{scale: 1}
	}
	if pdp.p.StrokeWidth == 0 {
		pdp.p.StrokeWidth = 1
	}

	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error Passing MoveToAbs\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
	}

	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.p.instructions <- &DrawingInstruction{Kind: MoveInstruction, M: &Tuple{x, y}}

	if len(tuples) > 0 {
		for _, nt := range tuples {
			pdp.x = nt[0]
			pdp.y = nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			pdp.p.instructions <- &DrawingInstruction{Kind: MoveInstruction, M: &Tuple{x, y}}
		}
	}

	return nil
}

func (pdp *pathDescriptionParser) parseMoveToAbs() error {
	var tuples []Tuple

	t, err := parseTuple(&pdp.lex)
	if err != nil {
		return fmt.Errorf("Error Passing MoveToAbs Expected Tuple\n%s", err)
	}

	pdp.x = t[0]
	pdp.y = t[1]

	if pdp.p.group.Owner == nil {
		pdp.p.group.Owner = &Svg{scale: 1}
	}
	if pdp.p.StrokeWidth == 0 {
		pdp.p.StrokeWidth = 1
	}

	scaledStroke := pdp.p.StrokeWidth * pdp.p.group.Owner.scale

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
		s.Width = scaledStroke
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		s.addPoint([2]float64{x, y})
		pdp.currentsegment = &s
		//fmt.Printf("orig x %f y %f, applied x %f y %f\n", pdp.x, pdp.y, x, y)
		pdp.p.instructions <- &DrawingInstruction{Kind: MoveInstruction, M: &Tuple{x, y}}
	}

	if len(tuples) > 0 {
		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		s := pdp.p.newSegment([2]float64{x, y})
		pdp.p.instructions <- &DrawingInstruction{Kind: MoveInstruction, M: &Tuple{x, y}}
		for _, nt := range tuples {
			pdp.x = nt[0]
			pdp.y = nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			s.addPoint([2]float64{x, y})
			pdp.p.instructions <- &DrawingInstruction{Kind: MoveInstruction, M: &Tuple{x, y}}
		}
		pdp.currentsegment = s
	}
	return nil

}

func (pdp *pathDescriptionParser) parseLineToAbsDI() error {
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

		for _, nt := range tuples {
			pdp.x = nt[0]
			pdp.y = nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			pdp.p.instructions <- &DrawingInstruction{Kind: LineInstruction, M: &Tuple{x, y}}
		}
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
			pdp.p.instructions <- &DrawingInstruction{Kind: LineInstruction, M: &Tuple{x, y}}
		}
	}

	return nil

}

func (pdp *pathDescriptionParser) parseMoveToRelDI() error {
	pdp.lex.ConsumeWhiteSpace()
	t, err := parseTuple(&pdp.lex)
	if err != nil {
		return fmt.Errorf("Error Passing MoveToRel Expected First Tuple %s", err)
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

	x, y := pdp.transform.Apply(pdp.x, pdp.y)
	pdp.p.instructions <- &DrawingInstruction{Kind: MoveInstruction, M: &Tuple{x, y}}

	if len(tuples) > 0 {
		for _, nt := range tuples {
			pdp.x += nt[0]
			pdp.y += nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			pdp.p.instructions <- &DrawingInstruction{Kind: MoveInstruction, M: &Tuple{x, y}}
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
		scaledStroke := pdp.p.StrokeWidth * pdp.svg.scale
		s.Width = scaledStroke

		x, y := pdp.transform.Apply(pdp.x, pdp.y)
		s.addPoint([2]float64{x, y})
		pdp.p.instructions <- &DrawingInstruction{Kind: MoveInstruction, M: &Tuple{x, y}}
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
			pdp.p.instructions <- &DrawingInstruction{Kind: MoveInstruction, M: &Tuple{x, y}}
		}
	}

	return nil
}

func (pdp *pathDescriptionParser) parseLineToRelDI() error {
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

		for _, nt := range tuples {
			pdp.x += nt[0]
			pdp.y += nt[1]
			x, y = pdp.transform.Apply(pdp.x, pdp.y)
			pdp.p.instructions <- &DrawingInstruction{Kind: LineInstruction, M: &Tuple{x, y}}
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
			pdp.p.instructions <- &DrawingInstruction{Kind: LineInstruction, M: &Tuple{x, y}}
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
	pdp.p.instructions <- &DrawingInstruction{Kind: MoveInstruction, M: &Tuple{x, y}}
	pdp.x = n
	x, y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.currentsegment.addPoint([2]float64{x, y})
	pdp.p.instructions <- &DrawingInstruction{Kind: LineInstruction, M: &Tuple{x, y}}

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
	pdp.p.instructions <- &DrawingInstruction{Kind: MoveInstruction, M: &Tuple{x, y}}
	pdp.x += n
	x, y = pdp.transform.Apply(pdp.x, pdp.y)
	pdp.p.instructions <- &DrawingInstruction{Kind: LineInstruction, M: &Tuple{x, y}}
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

func (pdp *pathDescriptionParser) parseCloseDI() error {
	pdp.lex.ConsumeWhiteSpace()

	pdp.p.instructions <- &DrawingInstruction{Kind: CloseInstruction}

	return nil
}

func (pdp *pathDescriptionParser) parseClose() error {
	pdp.lex.ConsumeWhiteSpace()

	if pdp.currentsegment != nil {
		pdp.currentsegment.addPoint(pdp.currentsegment.Points[0])
		pdp.currentsegment.Closed = true
		pdp.p.Segments <- *pdp.currentsegment
		pdp.currentsegment = nil
	}

	pdp.p.instructions <- &DrawingInstruction{Kind: CloseInstruction}
	return nil
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

func (pdp *pathDescriptionParser) parseCurveToRelDI() error {
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

	for j := 0; j < len(tuples)/3; j++ {
		pdp.x += tuples[j*3+2][0]
		pdp.y += tuples[j*3+2][1]

		c1x, c1y := pdp.transform.Apply(x+tuples[j*3][0], y+tuples[j*3][1])
		c2x, c2y := pdp.transform.Apply(x+tuples[j*3+1][0], y+tuples[j*3+1][1])
		tx, ty := pdp.transform.Apply(x+tuples[j*3+2][0], y+tuples[j*3+2][1])

		pdp.p.instructions <- &DrawingInstruction{
			Kind: CurveInstruction,
			C1:   &Tuple{c1x, c1y},
			C2:   &Tuple{c2x, c2y},
			T:    &Tuple{tx, ty},
		}
	}

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

		c1x, c1y := pdp.transform.Apply(x+tuples[j*3][0], y+tuples[j*3][1])
		c2x, c2y := pdp.transform.Apply(x+tuples[j*3+1][0], y+tuples[j*3+1][1])
		tx, ty := pdp.transform.Apply(x+tuples[j*3+2][0], y+tuples[j*3+2][1])

		pdp.p.instructions <- &DrawingInstruction{
			Kind: CurveInstruction,
			C1:   &Tuple{c1x, c1y},
			C2:   &Tuple{c2x, c2y},
			T:    &Tuple{tx, ty},
		}

		vertices := cb.recursiveInterpolate(10, 0)
		for _, v := range vertices {
			x, y = pdp.transform.Apply(v[0], v[1])
			pdp.currentsegment.addPoint([2]float64{x, y})
		}
	}

	return nil
}

func (pdp *pathDescriptionParser) parseCurveToAbsDI() error {
	var (
		tuples      []Tuple
		instrTuples []Tuple
	)

	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error parsing CurveToRel\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
		pdp.lex.ConsumeComma()
	}

	for j := 0; j < len(tuples)/3; j++ {
		for _, nt := range tuples[j*3 : (j+1)*3] {
			pdp.x = nt[0]
			pdp.y = nt[1]

			tx, ty := pdp.transform.Apply(pdp.x, pdp.y)
			instrTuples = append(instrTuples, Tuple{tx, ty})
		}

		pdp.p.instructions <- &DrawingInstruction{
			Kind: CurveInstruction,
			C1:   &instrTuples[0],
			C2:   &instrTuples[1],
			T:    &instrTuples[2],
		}
	}

	return nil
}

func (pdp *pathDescriptionParser) parseCurveToAbs() error {
	var (
		tuples      []Tuple
		instrTuples []Tuple
	)

	pdp.lex.ConsumeWhiteSpace()
	for pdp.lex.PeekItem().Type == gl.ItemNumber {
		t, err := parseTuple(&pdp.lex)
		if err != nil {
			return fmt.Errorf("Error parsing CurveToRel\n%s", err)
		}
		tuples = append(tuples, t)
		pdp.lex.ConsumeWhiteSpace()
		pdp.lex.ConsumeComma()
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

			tx, ty := pdp.transform.Apply(pdp.x, pdp.y)
			instrTuples = append(instrTuples, Tuple{tx, ty})
		}

		pdp.p.instructions <- &DrawingInstruction{
			Kind: CurveInstruction,
			C1:   &instrTuples[0],
			C2:   &instrTuples[1],
			T:    &instrTuples[2],
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
				p.StrokeWidth = sw
			}

		}
	}
}
