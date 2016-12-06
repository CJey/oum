package utils

import (
	"io/ioutil"
	"strings"
)

func StripConf(path string) (lines []string, err error) {
	tmp, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	tmp_lines := strings.Split(string(tmp), "\n")
	lines = make([]string, 0)
	for _, line := range tmp_lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		lines = append(lines, line)
	}
	return
}

func FetchConfKey(lines []string, target, def string) string {
	find := def
	for _, line := range lines {
		if strings.Fields(line)[0] == target {
			find = line
		}
	}
	return find
}

func FetchConfBlock(lines []string, begin, end string) []string {
	var hit bool
	var res []string
	for _, line := range lines {
		if hit {
			if strings.Fields(line)[0] == end {
				hit = false
			}
			res = append(res, line)
		} else {
			if strings.Fields(line)[0] == begin {
				hit = true
				res = []string{line}
			}
		}
	}
	return res
}
