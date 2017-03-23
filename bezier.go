package svg

import "math"

// cubicBezier
//
type cubicBezier struct {
	controlpoints [4][2]float64
	vertices      [][2]float64
	level         int
}

func (c *cubicBezier) interpolate(n int) [][2]float64 {
	var vertices [][2]float64
	for i := 0; i < n+1; i++ {
		var t float64
		t = float64(i) / float64(n)

		var v [2]float64
		v[0] = (1-t)*(1-t)*(1-t)*c.controlpoints[0][0] + 3*(1-t)*(1-t)*t*c.controlpoints[1][0] + 3*(1-t)*t*t*c.controlpoints[2][0] + t*t*t*c.controlpoints[3][0]
		v[1] = (1-t)*(1-t)*(1-t)*c.controlpoints[0][1] + 3*(1-t)*(1-t)*t*c.controlpoints[1][1] + 3*(1-t)*t*t*c.controlpoints[2][1] + t*t*t*c.controlpoints[3][1]
		vertices = append(vertices, v)
	}
	return vertices
}

func (c *cubicBezier) recursiveInterpolate(limit int, level int) [][2]float64 {
	c.level = level
	//	fmt.Println(level)
	var m12 [2]float64
	var m23 [2]float64
	var m34 [2]float64
	var m123 [2]float64
	var m234 [2]float64
	var m1234 [2]float64
	m12[0] = (c.controlpoints[0][0] + c.controlpoints[1][0]) / 2.0
	m12[1] = (c.controlpoints[0][1] + c.controlpoints[1][1]) / 2.0

	m23[0] = (c.controlpoints[1][0] + c.controlpoints[2][0]) / 2.0
	m23[1] = (c.controlpoints[1][1] + c.controlpoints[2][1]) / 2.0

	m34[0] = (c.controlpoints[2][0] + c.controlpoints[3][0]) / 2.0
	m34[1] = (c.controlpoints[2][1] + c.controlpoints[3][1]) / 2.0

	m123[0] = (m12[0] + m23[0]) / 2.0
	m123[1] = (m12[1] + m23[1]) / 2.0

	m234[0] = (m23[0] + m34[0]) / 2.0
	m234[1] = (m23[1] + m34[1]) / 2.0

	m1234[0] = (m123[0] + m234[0]) / 2.0
	m1234[1] = (m123[1] + m234[1]) / 2.0

	if limit == 0 {
		var vertices [][2]float64
		vertices = append(vertices, c.controlpoints[0])
		vertices = append(vertices, m1234)
		vertices = append(vertices, c.controlpoints[3])
		return vertices
	}

	a12 := math.Atan2(c.controlpoints[0][1]-c.controlpoints[1][1], c.controlpoints[1][0]-c.controlpoints[0][0])
	a34 := math.Atan2(c.controlpoints[2][1]-c.controlpoints[3][1], c.controlpoints[3][0]-c.controlpoints[2][0])
	a1234 := math.Atan2(m123[1]-m1234[1], m1234[0]-m123[0])

	//	fmt.Println("a12", a12*180.0/math.Pi)
	//	fmt.Println("a34", a34*180.0/math.Pi)
	//	fmt.Println("a1234", a1234*180.0/math.Pi)

	d1 := math.Abs(a1234 - a12)
	d2 := math.Abs(a34 - a1234)
	if d1 >= math.Pi {
		d1 = 2*math.Pi - d1
	}
	if d2 >= math.Pi {
		d2 = 2*math.Pi - d2
	}

	//	fmt.Println("angle d1", d1*180.0/math.Pi)
	//fmt.Println("angle d2", d2*180.0/math.Pi)

	if d1+d2 > 5.0*math.Pi/180.0 {
		var vertices [][2]float64
		var c1 cubicBezier
		c1.controlpoints[0] = c.controlpoints[0]
		c1.controlpoints[1] = m12
		c1.controlpoints[2] = m123
		c1.controlpoints[3] = m1234

		vertices = append(vertices, c1.recursiveInterpolate(limit-1, c.level+1)...)

		var c2 cubicBezier
		c2.controlpoints[0] = m1234
		c2.controlpoints[1] = m234
		c2.controlpoints[2] = m34
		c2.controlpoints[3] = c.controlpoints[3]
		vertices = append(vertices, c2.recursiveInterpolate(limit-1, c.level+1)...)
		return vertices
	} else {
		var vertices [][2]float64
		vertices = append(vertices, c.controlpoints[0])
		vertices = append(vertices, m1234)
		vertices = append(vertices, c.controlpoints[3])
		return vertices
	}
}
