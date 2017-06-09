package svg

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePath(t *testing.T) {
	const svgAbsoluteLine = `<svg viewBox="0 0 100 100"><path d="M0.000 0.000 L100.000 0.000 L100.000 100.000 L0.000 100.000 Z" fill="#000000" stroke="#000000" stroke-width="2"/></svg>`

	svg, err := ParseSvg(svgAbsoluteLine, "test", 0)
	require.NoError(t, err)

	dis := svg.ParseDrawingInstructions()
	for di := range dis {
		log.Printf("di: %+v, di.M: %+v", di, di.M)
	}
}
