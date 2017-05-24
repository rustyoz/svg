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
			return t, fmt.Errorf("Error passing number %s", ok)
		}
		t[0] = n
	} else {
		return t, fmt.Errorf("Error passing Tuple expected Number got %v", ni)
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
	lexer, _ := gl.Lex("tlexer", tstring)
	for {
		i := lexer.NextItem()
		switch i.Type {
		case gl.ItemEOS:
			return mt.Identity(),
				fmt.Errorf("transform parse failed")
		case gl.ItemWord:
			switch i.Value {
			case "matrix":
				return parseMatrix(lexer)
				// case "scale":
				// case "rotate":
			case "translate":
				return parseTranslate(lexer)
			}
		}
	}
}

func parseMatrix(l *gl.Lexer) (mt.Transform, error) {
	nums, err := parseParenNumList(l, 6)
	if err != nil {
		return mt.Identity(),
			fmt.Errorf("Error Parsing Transform Matrix: %v", err)
	}
	var tm mt.Transform
	for i := 0; i < 6; i++ {
		tm[i%3][i/3] = nums[i]
	}
	return tm, nil
}

func parseTranslate(l *gl.Lexer) (mt.Transform, error) {
	nums, err := parseParenNumList(l, 2)
	if err != nil {
		return mt.Identity(), fmt.Errorf("Error Parsing Translate: %v", err)
	}
	tm := mt.Identity()
	tm[0][2] = nums[0]
	tm[1][2] = nums[1]
	return tm, nil
}

// Parse a parenthesized list of ncount numbers.
func parseParenNumList(l *gl.Lexer, ncount int) ([]float64, error) {
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
		if l.PeekItem().Type != gl.ItemNumber {
			return nil, fmt.Errorf("Expected Number got %v", l.PeekItem().String())
		}
		n, err := parseNumber(l.NextItem())
		if err != nil {
			return nil, err
		}
		nums = append(nums, n)
		if len(nums) >= ncount {
			i = l.PeekItem()
			if i.Type != gl.ItemParan {
				return nil, fmt.Errorf("Expected Closing Parantheses")
			}
			l.NextItem() // consume Parantheses
			return nums, nil
		}
	}
}
