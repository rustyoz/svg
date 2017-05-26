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
