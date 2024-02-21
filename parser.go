package svg

import (
	"fmt"
	"strconv"

	mt "github.com/rustyoz/Mtransform"
	gl "github.com/rustyoz/genericlexer"
)

func parseNumber(i gl.Item) (float64, error) {
	var n float64
	var ok error
	if i.Type == gl.ItemNumber {
		n, ok = strconv.ParseFloat(i.Value, 64)
		if ok != nil {
			return n, fmt.Errorf("Error passing number %s", ok)
		}
	}
	return n, nil
}

func parseTuple(l *gl.Lexer) (Tuple, error) {
	t := Tuple{}

	l.ConsumeWhiteSpace()

	ni := l.NextItem()
	if ni.Type == gl.ItemNumber {
		n, ok := strconv.ParseFloat(ni.Value, 64)
		if ok != nil {
			return t, fmt.Errorf("Error parsing number %s", ok)
		}
		t[0] = n
	} else {
		return t, fmt.Errorf("Error parsing Tuple expected Number got: %s", ni.Value)
	}

	if l.PeekItem().Type == gl.ItemWSP || l.PeekItem().Type == gl.ItemComma {
		l.NextItem()
	}
	ni = l.NextItem()
	if ni.Type == gl.ItemNumber {
		n, ok := strconv.ParseFloat(ni.Value, 64)
		if ok != nil {
			return t, fmt.Errorf("Error passing Number %s", ok)
		}
		t[1] = n
	} else {
		return t, fmt.Errorf("Error passing Tuple expected Number got: %v", ni)
	}

	return t, nil
}

func parseTransform(tstring string) (mt.Transform, error) {
	var x *mt.Transform
	lexer, _ := gl.Lex("tlexer", tstring)
	for {
		i := lexer.NextItem()
		var t mt.Transform
		var err error
		switch i.Type {
		case gl.ItemEOS:
			if x == nil {
				return mt.Identity(),
					fmt.Errorf("transform parse failed")
			}
			return *x, nil
		case gl.ItemWord:
			switch i.Value {
			case "matrix":
				t, err = parseMatrix(lexer)
			case "translate":
				t, err = parseTranslate(lexer)
			case "rotate":
				t, err = parseRotate(lexer)
			case "scale":
				t, err = parseScale(lexer)
			}
		case gl.ItemWSP:
			continue
		}
		if err != nil {
			return t, err
		} else if x == nil {
			x = &t
		} else {
			x.MultiplyWith(t)
		}
	}
}

func parseMatrix(l *gl.Lexer) (mt.Transform, error) {
	nums, err := parseParenNumList(l)
	if err != nil || len(nums) != 6 {
		return mt.Identity(),
			fmt.Errorf("Error Parsing Transform Matrix: %v", err)
	}
	var tm mt.Transform
	tm[0][0] = nums[0]
	tm[0][1] = nums[2]
	tm[0][2] = nums[4]
	tm[1][0] = nums[1]
	tm[1][1] = nums[3]
	tm[1][2] = nums[5]
	tm[2][0] = 0
	tm[2][1] = 0
	tm[2][2] = 1

	return tm, nil
}

func parseTranslate(l *gl.Lexer) (mt.Transform, error) {
	nums, err := parseParenNumList(l)
	if err != nil || len(nums) != 2 {
		return mt.Identity(), fmt.Errorf("Error Parsing Translate: %v", err)
	}
	tm := mt.Identity()
	tm[0][2] = nums[0]
	tm[1][2] = nums[1]
	return tm, nil
}

func parseRotate(l *gl.Lexer) (mt.Transform, error) {
	nums, err := parseParenNumList(l)
	if err != nil || (len(nums) != 1 && len(nums) != 3) {
		return mt.Identity(), fmt.Errorf("Error Parsing Rotate: %v", err)
	}
	a, px, py := nums[0], 0.0, 0.0
	if len(nums) == 3 {
		px, py = nums[1], nums[2]
	}

	tm := mt.Identity()
	(&tm).RotatePoint(a, px, py)
	return tm, nil
}

func parseScale(l *gl.Lexer) (mt.Transform, error) {
	nums, err := parseParenNumList(l)
	if err != nil || (len(nums) != 1 && len(nums) != 2) {
		return mt.Identity(), fmt.Errorf("Error Parsing Scale: %v", err)
	}
	x, y := nums[0], nums[0]
	if len(nums) == 2 {
		y = nums[1]
	}
	tm := mt.Identity()
	(&tm).Scale(x, y)
	return tm, nil
}

// Parse a parenthesized list of ncount numbers.
func parseParenNumList(l *gl.Lexer) ([]float64, error) {
	i := l.NextItem()
	if i.Type != gl.ItemParan {
		return nil, fmt.Errorf("Expected Opening Parantheses")
	}
	var nums []float64
	for {
		if len(nums) > 0 {
			for l.PeekItem().Type == gl.ItemComma || l.PeekItem().Type == gl.ItemWSP {
				l.NextItem()
			}
		}
		if l.PeekItem().Type == gl.ItemParan {
			l.NextItem() // consume Parantheses
			break
		}
		if l.PeekItem().Type != gl.ItemNumber {
			return nil, fmt.Errorf("Expected Number got %v", l.PeekItem().String())
		}
		n, err := parseNumber(l.NextItem())
		if err != nil {
			return nil, err
		}
		nums = append(nums, n)

	}
	return nums, nil
}
