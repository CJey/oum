package version

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

var (
	name      string
	version   string
	gitNumber string
	gitHash   string
	gitBranch string
	goVersion string
	buildTime string
	codeRoot  string
)

func Show() string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "Version\t\t%s\n", Version())
	fmt.Fprintf(buf, "BuildInfo\t%s@%s\n", GoVersion(), BuildTime().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(buf, "GitInfo\t\t%s:%d.%s\n", GitBranch(), GitNumber(), GitFullHash())
	return buf.String()
}

func Name() string {
	return name
}

func Version() string {
	return version
}

func GitNumber() uint64 {
	n, err := strconv.ParseUint(gitNumber, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func GitShortHash() string {
	if len(gitHash) >= 7 {
		return gitHash[0:7]
	}
	return ""
}

func GitFullHash() string {
	return gitHash
}

func GitBranch() string {
	return gitBranch
}

func GoVersion() string {
	return goVersion
}

func BuildTime() time.Time {
	t, err := strconv.ParseInt(buildTime, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(t, 0)
}

func CodeRoot() string {
	return codeRoot
}
