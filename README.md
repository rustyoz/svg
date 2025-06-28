# SVG Parser for Go 
[![Go](https://github.com/rustyoz/svg/actions/workflows/go.yml/badge.svg)](https://github.com/rustyoz/svg/actions/workflows/go.yml)

A comprehensive Go library for parsing SVG (Scalable Vector Graphics) files with support for path parsing, shape elements, transformations, and Bezier curve rasterization.

<span style="color: red;">**Warning: Readme is AI generated**</span>

## Features

- **SVG Parsing**: Parse SVG files from strings or readers
- **Shape Support**: Full support for all major SVG shapes:
  - `<path>` with complete path command support (M, L, H, V, C, S, Q, T, A, Z)
  - `<rect>`, `<circle>`, `<ellipse>`, `<line>`
  - `<polygon>`, `<polyline>`
- **Transformations**: Complete SVG transform support (translate, rotate, scale, matrix)
- **Groups**: Nested group (`<g>`) support with inheritance
- **Styling**: Stroke, fill, opacity, stroke-width, and other styling attributes
- **Bezier Curves**: Advanced Bezier curve rasterization with recursive interpolation
- **Drawing Instructions**: Generate drawing instructions suitable for graphics libraries
- **ViewBox Support**: Proper viewport and coordinate system handling

## Installation

```bash
go get github.com/rustyoz/svg
```

## Quick Start

### Basic SVG Parsing

```go
package main

import (
    "fmt"
    "log"
    "github.com/rustyoz/svg"
)

func main() {
    svgContent := `<svg width="100" height="100" viewBox="0 0 100 100">
        <rect x="10" y="10" width="30" height="30" fill="red"/>
        <circle cx="70" cy="25" r="15" fill="blue"/>
        <path d="M10 70 L50 70 L30 90 Z" fill="green"/>
    </svg>`
    
    // Parse SVG from string
    parsed, err := svg.ParseSvg(svgContent, "example", 1.0)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("SVG dimensions: %s x %s\n", parsed.Width, parsed.Height)
    fmt.Printf("ViewBox: %s\n", parsed.ViewBox)
}
```

### Working with Drawing Instructions

Drawing instructions provide a higher-level interface suitable for graphics libraries:

```go
// Parse drawing instructions from SVG
instructions, errors := parsed.ParseDrawingInstructions()

// Process instructions
go func() {
    for err := range errors {
        log.Printf("Error: %v", err)
    }
}()

for instruction := range instructions {
    switch instruction.Kind {
    case svg.MoveInstruction:
        // Move to instruction.M
    case svg.LineInstruction:
        // Line to instruction.L
    case svg.CurveInstruction:
        // Bezier curve with control points
        fmt.Printf("Curve: %v -> %v -> %v\n", 
            instruction.CurvePoints.C1,
            instruction.CurvePoints.C2, 
            instruction.CurvePoints.T)
    case svg.PaintInstruction:
        // Apply styling (stroke, fill, etc.)
    }
}
```

### Working with Path Segments

For lower-level access, you can work directly with path segments:

```go
// Assuming you have a Path element
segments := path.Parse()

for segment := range segments {
    fmt.Printf("Segment with %d points, width: %.2f, closed: %t\n", 
        len(segment.Points), segment.Width, segment.Closed)
    
    for _, point := range segment.Points {
        fmt.Printf("  Point: (%.2f, %.2f)\n", point[0], point[1])
    }
}
```

## Advanced Usage

### Custom Transformations

The library properly handles SVG transformations:

```go
svgWithTransforms := `<svg viewBox="0 0 100 100">
    <g transform="translate(50,50) rotate(45) scale(2,1)">
        <rect x="-10" y="-10" width="20" height="20"/>
    </g>
</svg>`

parsed, _ := svg.ParseSvg(svgWithTransforms, "transformed", 1.0)
// Transformations are automatically applied to drawing instructions
```

### Reading from File

```go
import (
    "os"
    "github.com/rustyoz/svg"
)

file, err := os.Open("drawing.svg")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

parsed, err := svg.ParseSvgFromReader(file, "drawing", 1.0)
if err != nil {
    log.Fatal(err)
}
```

### Bezier Curve Rasterization

The library includes sophisticated Bezier curve rasterization:

```go
// Bezier curves in paths are automatically rasterized
// with configurable precision based on angle thresholds
svgWithCurves := `<svg viewBox="0 0 200 200">
    <path d="M50 50 C 50 10, 150 10, 150 50 S 250 90, 150 90"/>
</svg>`

parsed, _ := svg.ParseSvg(svgWithCurves, "curves", 1.0)
instructions, _ := parsed.ParseDrawingInstructions()

for instruction := range instructions {
    if instruction.Kind == svg.CurveInstruction {
        // Access rasterized curve points
        points := instruction.CurvePoints
        // Use points for rendering
    }
}
```

## Supported SVG Elements

| Element | Support | Notes |
|---------|---------|-------|
| `<svg>` | ✅ | Root element with viewBox, width, height |
| `<g>` | ✅ | Groups with transformations and styling |
| `<path>` | ✅ | Complete path command support |
| `<rect>` | ✅ | Rectangles with rounded corners |
| `<circle>` | ✅ | Circles |
| `<ellipse>` | ✅ | Ellipses |
| `<line>` | ✅ | Lines |
| `<polygon>` | ✅ | Closed polygons |
| `<polyline>` | ✅ | Open polygons |
| `<text>` | ⚠️ | Parsed but not rasterized |

## Path Commands

All SVG path commands are supported:

- **M/m**: Move to (absolute/relative)
- **L/l**: Line to (absolute/relative)  
- **H/h**: Horizontal line to (absolute/relative)
- **V/v**: Vertical line to (absolute/relative)
- **C/c**: Cubic Bezier curve (absolute/relative)
- **S/s**: Smooth cubic Bezier curve (absolute/relative)
- **Q/q**: Quadratic Bezier curve (absolute/relative)
- **T/t**: Smooth quadratic Bezier curve (absolute/relative)
- **A/a**: Elliptical arc (absolute/relative)
- **Z/z**: Close path

## API Reference

### Main Functions

```go
// Parse SVG from string
func ParseSvg(svgString, name string, scale float64) (*Svg, error)

// Parse SVG from io.Reader
func ParseSvgFromReader(r io.Reader, name string, scale float64) (*Svg, error)
```

### Core Types

```go
type Svg struct {
    Title    string
    Groups   []Group
    Width    string
    Height   string
    ViewBox  string
    Elements []DrawingInstructionParser
    // ...
}

type DrawingInstruction struct {
    Kind           InstructionType
    M              *[2]float64     // Move to point
    L              *[2]float64     // Line to point
    CurvePoints    *CurvePoints    // Bezier curve points
    StrokeWidth    *float64
    Stroke         *string
    Fill           *string
    Opacity        *float64
    // ...
}
```

## Dependencies

- `github.com/rustyoz/Mtransform` - Matrix transformations
- `github.com/rustyoz/genericlexer` - Lexical analysis for path parsing
- Standard Go XML parsing

## Applications

This library is actively used in:
- [svg2kicad](http://github.com/rustyoz/svg2kicad) - Convert SVG to KiCad PCB format

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

See [LICENCE](LICENCE) file for details.
