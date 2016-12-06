package api

import (
	"fmt"
)

func sliceU64tostring(b []uint64) (str string) {
	if len(b) != 0 {
		for i, v := range b {
			if i == len(b)-1 {
				str = str + fmt.Sprintf("%d", v)
			} else {
				str = str + fmt.Sprintf("%d", v) + ","
			}
		}
	}
	return
}
