package svg

// InstructionType tells our path drawing library which function it has
// to call
type InstructionType int

// These are instruction types that we use with our path drawing library
const (
	PathInstruction InstructionType = iota
	MoveInstruction
	CircleInstruction
	CurveInstruction
	LineInstruction
	HLineInstruction
	CloseInstruction
	PaintInstruction
)

// DrawingInstruction contains enough information that a simple drawing
// library can draw the shapes contained in an SVG file.
//
// The struct contains all necessary fields but only the ones needed (as
// indicated byt the InstructionType) will be non-nil.
type DrawingInstruction struct {
	Kind        InstructionType
	M           *Tuple
	C1          *Tuple
	C2          *Tuple
	T           *Tuple
	Radius      *float64
	StrokeWidth *float64
	Fill        *string
	Stroke      *string
}
