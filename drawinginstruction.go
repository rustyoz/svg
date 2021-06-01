package svg

import "fmt"

// InstructionType tells our path drawing library which function it has
// to call
type InstructionType int

// These are instruction types that we use with our path drawing library
const (
	MoveInstruction InstructionType = iota
	CircleInstruction
	CurveInstruction
	LineInstruction
	CloseInstruction
	PaintInstruction
)

// CurvePoints are the points needed by a bezier curve.
type CurvePoints struct {
	C1 *Tuple
	C2 *Tuple
	T  *Tuple
}

// DrawingInstruction contains enough information that a simple drawing
// library can draw the shapes contained in an SVG file.
//
// The struct contains all necessary fields but only the ones needed (as
// indicated byt the InstructionType) will be non-nil.
type DrawingInstruction struct {
	Kind           InstructionType
	M              *Tuple
	CurvePoints    *CurvePoints
	Radius         *float64
	StrokeWidth    *float64
	Opacity		   *float64
	Fill           *string
	Stroke         *string
	StrokeLineCap  *string
	StrokeLineJoin *string
}

func (di *DrawingInstruction) String() string {
	switch di.Kind {
	case MoveInstruction:
		return fmt.Sprintf("M%v %v", di.M[0], di.M[1])
	case CircleInstruction:
		return fmt.Sprintf("circle R=%v", *di.Radius)
	case CurveInstruction:
		c := di.CurvePoints
		return fmt.Sprintf("C%v %v %v %v %v %v",
			c.C1[0], c.C1[1], c.C2[0], c.C2[1], c.T[0], c.T[1])
	case LineInstruction:
		return fmt.Sprintf("L%v %v", di.M[0], di.M[1])
	case CloseInstruction:
		return "Z"
	case PaintInstruction:
		pt := ""
		if di.Fill != nil {
			pt = fmt.Sprintf("%vfill=\"%v\" ", pt, *di.Fill)
		}
		if di.Stroke != nil {
			pt = fmt.Sprintf("%vstroke=\"%v\" ", pt, *di.Stroke)
		}
		if di.StrokeWidth != nil {
			pt = fmt.Sprintf("%vstroke-width=\"%v\" ", pt, *di.StrokeWidth)
		}
		if di.StrokeLineCap != nil {
			pt = fmt.Sprintf("%vstroke-linecap=\"%v\" ", pt, *di.StrokeLineCap)
		}
		if di.StrokeLineJoin != nil {
			pt = fmt.Sprintf("%vstroke-linejoin=\"%v\" ", pt, *di.StrokeLineJoin)
		}
		return pt
	}
	return ""
}

// PathStringFromDrawingInstructions converts drawing instructions obtained
// from svg <path/> element back into <path/> form
func PathStringFromDrawingInstructions(dis []*DrawingInstruction) string {
	data := " "
	sep := ""
	var paint *DrawingInstruction
	for _, di := range dis {
		if di.Kind == PaintInstruction {
			paint = di
		} else {
			data += sep + di.String()
			sep = " "
		}
	}
	pt := ""
	if paint != nil {
		pt = paint.String()
	}
	return `<path d="` + data + `" ` + pt + `/>`
}
