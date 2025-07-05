package main

import (
	"fmt"

	"github.com/rustyoz/svg"
)

func main() {
	svgContent := `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" height="639"
    width="505" viewBox="(0, 0, 505, 639)">
    <defs />
    <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" height="639"
        width="505" viewBox="0 -639 505 639">
        <defs />
    </svg>
</svg>`

	result, err := svg.ParseSvg(svgContent, "test", 1.0)
	if err != nil {
		fmt.Printf("Error parsing SVG: %v\n", err)
		return
	}

	fmt.Printf("Successfully parsed SVG: %s\n", result.Name)
	fmt.Printf("Width: %s, Height: %s, ViewBox: %s\n", result.Width, result.Height, result.ViewBox)
	fmt.Printf("Number of groups: %d\n", len(result.Groups))
	fmt.Printf("Number of elements: %d\n", len(result.Elements))

	// print the svg content
	fmt.Println(result.Elements[0])
}
