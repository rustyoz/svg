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
)

// DrawingInstruction contains enough information that a simple drawing
// library can draw the shapes contained in an SVG file.
type DrawingInstruction struct {
	Kind InstructionType
	M    *Tuple
	C1   *Tuple
	C2   *Tuple
	T    *Tuple
}
