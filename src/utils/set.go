package utils

import (
	"strings"
)

func CSVSet(set string) []string {
	tmp := strings.Split(set, ",")
	val_map := make(map[string]struct{})
	for _, v := range tmp {
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}
		_, ok := val_map[v]
		if !ok {
			val_map[v] = struct{}{}
		}
	}
	ret := make([]string, 0, len(val_map))
	for k, _ := range val_map {
		ret = append(ret, k)
	}
	return ret
}

func SetOp(origin string, op string) string {
	if len(op) == 0 {
		return origin
	}

	mode := op[0]
	op = op[1:]

	a := CSVSet(origin)
	b := CSVSet(op)

	r_a := map[string]struct{}{}
	for _, i := range a {
		r_a[i] = struct{}{}
	}

	switch mode {
	case '+':
		for _, i := range b {
			r_a[i] = struct{}{}
		}
	case '-':
		for _, i := range b {
			delete(r_a, i)
		}
	case '=':
		return strings.Join(b, ",")
	default:
		return strings.Join(a, ",")
	}

	res := make([]string, 0, len(r_a))
	for i, _ := range r_a {
		res = append(res, i)
	}
	return strings.Join(res, ",")
}
