package svg

import (
	"strings"
	"testing"

	"github.com/cheekybits/is"
)

const testSvg = `<?xml version="1.0" encoding="utf-8"?>
<!-- Generator: Adobe Illustrator 15.0.2, SVG Export Plug-In . SVG Version: 6.00 Build 0)  -->
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
	 width="595.201px" height="841.922px" viewBox="0 0 595.201 841.922" enable-background="new 0 0 595.201 841.922"
	 xml:space="preserve">
<rect x="207" y="53" fill="#009FE3" width="181.667" height="85.333"/>
<text transform="matrix(1 0 0 1 232.3306 107.5952)" fill="#FFFFFF" font-family="'ArialMT'" font-size="31.9752">PODIUM</text>
<g><text transform="matrix(1 0 0 1 232.3306 107.5952)" fill="#FFFFFF" font-family="'ArialMT'" font-size="31.9752">PODIUM</text></g>
</svg>`

func TestParse(t *testing.T) {
	is := is.New(t)

	svg, err := ParseSvg(testSvg, "test", 0)
	is.NoErr(err)
	is.NotNil(svg)

	svg, err = ParseSvgFromReader(strings.NewReader(testSvg), "test", 0)
	is.NoErr(err)
	is.NotNil(svg)
}

func TestTransform(t *testing.T) {
	content := `<?xml version="1.0" standalone="no"?>
	<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 20010904//EN"
	 "http://www.w3.org/TR/2001/REC-SVG-20010904/DTD/svg10.dtd">
	<svg version="1.0" xmlns="http://www.w3.org/2000/svg"
	 width="684.000000pt" height="630.000000pt" viewBox="0 0 684.000000 630.000000"
	 preserveAspectRatio="xMidYMid meet">
	<g transform="translate(0.000000,630.000000) scale(0.100000,-0.100000)"
	fill="#ff0000" stroke="none">
	<path d="M1705 6274 c-758 -115 -1377 -637 -1614 -1360 -74 -227 -86 -311 -86
	-624 0 -248 2 -286 23 -389 64 -315 210 -626 414 -880 34 -42 715 -731 1515
	-1531 l1453 -1455 1444 1445 c794 795 1473 1481 1510 1524 98 117 189 259 261
	409 233 479 267 1011 99 1516 -240 718 -861 1235 -1615 1345 -138 21 -430 21
	-564 1 -213 -32 -397 -87 -580 -174 -188 -90 -319 -177 -478 -316 l-77 -66
	-77 66 c-309 270 -657 431 -1054 489 -132 20 -447 20 -574 0z"/>
	</g>
	</svg>
	`
	s, err := ParseSvg(content, "transformed heart", 0)
	if err != nil {
		t.Fatalf("cannot parse svg %v", content)
	}
	if len(s.Groups) < 1 {
		t.Fatal("group not found")
	}
	g := s.Groups[0]
	t.Logf("original: transform=\"%v\"", g.TransformString)
	m := *g.Transform
	a, c, e := m[0][0], m[0][1], m[0][2]
	b, d, f := m[1][0], m[1][1], m[1][2]
	// see https://www.w3.org/TR/SVGTiny12 for [a b c d e f] vector notation
	t.Logf("accumulated: transform=\"matrix(%v %v %v %v %v %v)\"", a, b, c, d, e, f)
	if !(a == 0.1 && d == -0.1 && f == 630) {
		t.Error("mismatch expected transform matrix(0.1 0 0 -0.1 0 630)")
	}
}

func Test2Curves(t *testing.T) {
	content := `<?xml version="1.0" standalone="no"?>
	<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 20010904//EN"
	 "http://www.w3.org/TR/2001/REC-SVG-20010904/DTD/svg10.dtd">
	<svg version="1.0" xmlns="http://www.w3.org/2000/svg"
	viewBox="0 -450 1000 1000">
	<path fill="none" stroke="red" stroke-width="5"
	d="M100 200 C25 100 400 100 400 200 400 100 775 100 700 200 L400 450z" />
	</svg>
	`
	s, err := ParseSvg(content, "heart", 1)
	if err != nil {
		t.Fatalf("cannot parse svg %v", content)
	}
	dis, _ := s.ParseDrawingInstructions()
	strux := []*DrawingInstruction{}
	for di := range dis {
		strux = append(strux, di)
	}
	curveIdx := 2
	di := strux[curveIdx]
	if di.Kind != CurveInstruction {
		t.Fatalf("expect curve drawing instructions, got %v", di)
	}
	p := di.CurvePoints // 400 100 775 100 700 200
	if !(p.C1[0] == 400 && p.C2[0] == 775 && p.T[0] == 700) {
		t.Fatalf("expect [400 100] [775 100] [700 200], got %v %v %v", *p.C1, *p.C2, *p.T)
	}
}
