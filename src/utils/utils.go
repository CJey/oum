package utils

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/cjey/slog"
)

func PrimaryIP() string {
	target := "8.8.8.8"
	addrs, err := net.LookupHost("freeapi.ipip.net")
	if err != nil && len(addrs) > 0 {
		target = addrs[0]
	}
	cmd := exec.Command("ip", "route", "get", target)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	fs := strings.Fields(out.String())
	for i, f := range fs {
		if f == "src" {
			return fs[i+1]
		}
	}
	return ""
}

func Readline(def string) string {
	buf := bufio.NewReader(os.Stdin)
	line, err := buf.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			println()
			os.Exit(0)
		}
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	line = strings.TrimSpace(line)
	if len(line) == 0 && len(def) > 0 {
		return def
	}
	return line
}

func Tmpfile(body string) string {
	name := make([]byte, 16)
	_, _ = rand.Read(name)
	path := path.Join(os.TempDir(), fmt.Sprintf("oum-%x", name))
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	_, err = f.WriteString(body)
	if err != nil {
		os.Remove(path)
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	return path
}

func Rmtmp(path string) {
	defer os.Remove(path)
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	or00 := bytes.Repeat([]byte{0x00}, int(fi.Size()))
	orff := bytes.Repeat([]byte{0xff}, int(fi.Size()))

	for i := 0; i < 3; i++ {
		f.Seek(0, os.SEEK_SET)
		f.Write(or00)
		f.Seek(0, os.SEEK_SET)
		f.Write(orff)
	}
}
