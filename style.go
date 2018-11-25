package svg

import (
	"strings"
)

func splitStyle(style string) map[string]string {
	var r map[string]string
	r = make(map[string]string)
	props := strings.Split(style, ";")

	for _, keyval := range props {
		kv := strings.Split(keyval, ":")
		if len(kv) >= 2 {
			r[kv[0]] = kv[1]
		}
	}

	return r
}
