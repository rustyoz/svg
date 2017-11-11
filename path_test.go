package svg

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

type PathTest struct {
	Description string
	Svg     string
	Kinds   []InstructionType
	XCoords []float64
	YCoords []float64
}

var tests = []PathTest{
	{
		"absolute lines",
		`<svg viewBox="0 0 100 100"><path d="M0.000 0.000 L100.000 0.000 100.000 100.000 L0.000 100.000 Z" fill="#000000" stroke="#000000" stroke-width="2"/></svg>`,
		[]InstructionType{MoveInstruction, LineInstruction, LineInstruction, LineInstruction, CloseInstruction, PaintInstruction},
		[]float64{0, 100, 100, 0,   0},
		[]float64{0, 0,   100, 100, 0},
	},
	{
		"relative lines",
		`<svg viewBox="0 0 100 100"><path d="M0.000 0.000 l100.000 0.000 100.000 100.000 l0.000 100.000 Z" fill="#000000" stroke="#000000" stroke-width="2"/></svg>`,
		[]InstructionType{MoveInstruction, LineInstruction, LineInstruction, LineInstruction, CloseInstruction, PaintInstruction},
		[]float64{0, 100, 200, 200, 0},
		[]float64{0, 0,   100, 200, 0},
	},
	{
		"relative h-line test",
		`<svg viewBox="0 0 100 100"><path d="M0.000 0.000 h100.000 50.000" fill="#000000" stroke="#000000" stroke-width="2"/></svg>`,
		[]InstructionType{MoveInstruction, LineInstruction, LineInstruction, PaintInstruction},
		[]float64{0, 100, 150, 0},
		[]float64{0, 0, 0, 0},
	},
	{
		"absolute h-line test",
		`<svg viewBox="0 0 100 100"><path d="M0.000 0.000 H100.000 50.000" fill="#000000" stroke="#000000" stroke-width="2"/></svg>`,
		[]InstructionType{MoveInstruction, LineInstruction, LineInstruction, PaintInstruction},
		[]float64{0, 100, 50, 0},
		[]float64{0, 0, 0, 0},
	},
	{
		"relative v-line test",
		`<svg viewBox="0 0 100 100"><path d="M0.000 0.000 v100.000 50.000" fill="#000000" stroke="#000000" stroke-width="2"/></svg>`,
		[]InstructionType{MoveInstruction, LineInstruction, LineInstruction, PaintInstruction},
		[]float64{0, 0, 0, 0},
		[]float64{0, 100, 150, 0},
	},
	{
		"absolute v-line test",
		`<svg viewBox="0 0 100 100"><path d="M0.000 0.000 V100.000 50.000" fill="#000000" stroke="#000000" stroke-width="2"/></svg>`,
		[]InstructionType{MoveInstruction, LineInstruction, LineInstruction, PaintInstruction},
		[]float64{0, 0, 0, 0},
		[]float64{0, 100, 50, 0},
	},
}

func TestParsePathList(t *testing.T) {
	for _, test := range tests {
		svg, err := ParseSvg(test.Svg, "test", 0)
		require.NoError(t, err)

		dis, _ := svg.ParseDrawingInstructions()
		strux := []*DrawingInstruction{}
		for di := range dis {
			strux = append(strux, di)
			log.Printf("di: %+v, di.M: %+v", di, di.M)
		}

		if len(strux) != len(test.Kinds) {
			t.Fatalf("expected %d instructions for test %s, but received %d", len(test.Kinds), test.Description, len(strux))
		}

		for i, kind := range test.Kinds {
			if strux[i].Kind != kind {
				t.Fatalf("expected instruction %d for test %s to be %d, but was %d", i, test.Description, kind, strux[i].Kind)
			}
		}

		for i, x := range test.XCoords {
			if strux[i].M == nil {
				continue
			}

			if strux[i].M[0] != x {
				t.Fatalf("expected X coordinate %d for test %s to be %f, but was %f", i, test.Description, x, strux[i].M[0])
			}
		}

		for i, y := range test.YCoords {
			if strux[i].M == nil {
				continue
			}

			if strux[i].M[1] != y {
				t.Fatalf("expected Y coordinate %d for test %s to be %f, but was %f", i, test.Description, y, strux[i].M[1])
			}
		}
	}
}
