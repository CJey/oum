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
	a := CSVSet(origin)

	mode := 0
	if len(op) > 0 {
		switch op[0] {
		case '+':
			mode = 1
			op = op[1:]
		case '-':
			mode = -1
			op = op[1:]
		}
	} else {
		return strings.Join(a, ",")
	}

	b := CSVSet(op)
	if mode == 0 {
		return strings.Join(b, ",")
	}

	r_a := map[string]struct{}{}
	for _, i := range a {
		r_a[i] = struct{}{}
	}

	if mode > 0 {
		for _, i := range b {
			r_a[i] = struct{}{}
		}
	} else {
		for _, i := range b {
			delete(r_a, i)
		}
	}

	res := make([]string, 0, len(r_a))
	for i, _ := range r_a {
		res = append(res, i)
	}
	return strings.Join(res, ",")
}
