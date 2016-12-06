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
