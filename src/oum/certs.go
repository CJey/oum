package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"utils"
)

func genCAKey(ecdsa bool) (key string, err error) {
	cmd := exec.Command("openssl", "genrsa", "2048")
	if ecdsa {
		cmd = exec.Command("openssl", "ecparam", "-name", "prime256v1", "-genkey")
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return
	}
	key = out.String()
	return
}

func genCACert(key, subj string) (crt string, err error) {
	pkey := utils.Tmpfile(key)
	defer utils.Rmtmp(pkey)
	cmd := exec.Command(
		"openssl", "req", "-x509", "-new", "-nodes", "-sha256", "-days", "36500",
		"-key", pkey,
		"-subj", "/"+subj,
	)
	cmd.Stdin = strings.NewReader(key)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return
	}
	crt = out.String()
	return
}

func genSrvKey(ecdsa bool) (key string, err error) {
	cmd := exec.Command("openssl", "genrsa", "2048")
	if ecdsa {
		cmd = exec.Command("openssl", "ecparam", "-name", "prime256v1", "-genkey")
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return
	}
	key = out.String()
	return
}

func genSrvCert(key, cacrt, cakey, subj string) (crt string, err error) {
	pkey := utils.Tmpfile(key)
	defer utils.Rmtmp(pkey)
	cmd := exec.Command(
		"openssl", "req", "-new", "-days", "36500", "-nodes",
		"-key", pkey,
		"-subj", "/"+subj,
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return
	}
	csr := out.String()

	pcacrt := utils.Tmpfile(cacrt)
	defer utils.Rmtmp(pcacrt)
	pcakey := utils.Tmpfile(cakey)
	defer utils.Rmtmp(pcakey)
	pserial := utils.Tmpfile("01")
	defer utils.Rmtmp(pserial)
	cmd = exec.Command(
		"openssl", "x509", "-req", "-sha256",
		"-CAserial", pserial,
		"-CA", pcacrt,
		"-CAkey", pcakey,
	)
	cmd.Stdin = strings.NewReader(csr)
	out.Reset()
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("stderr = %s\n", stderr.String())
		return
	}
	crt = out.String()
	return
}

func genDH(quick bool) (dh string, err error) {
	if quick {
		// pre 2048 bits
		dh = `-----BEGIN DH PARAMETERS-----
MIIBCAKCAQEA2fQGI69/j3DiaYP+BAhdIUt477QV88hQW6CxLzWDrEOtmV3XHKS7
E8jFnv2hYD03vl89cLs977Ir05sI424DBcbbxgi7V7YMkaJbdhZd3df2AqAG/QPq
keWVPZ7usgnxokUI73o1PJkL0PGYv8Mzrb1+xQuy50vUf95VP8PsMn7mxh77FaMT
AISdpnfPn5jvEBD32KW8jKnM8/iHGCFOYUf+9cQMayT0CLVYAroK4aVLkevcrNzW
DxkHOi75IoiNhtulgOdRsgO3XPozHP5izuPZ7QV3gPlU4NvwHjA5v+x0DZ3F+09I
dBM1iy9qTKBDFRvJpH3iWq5yBFo9Hx6muwIBAg==
-----END DH PARAMETERS-----
`
	} else {
		cmd := exec.Command("openssl", "dhparam", "1024")
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			return
		}
		dh = out.String()
	}
	return
}
