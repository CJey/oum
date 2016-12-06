package main

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

func ipset(interval uint, sets ...string) {
	for {
		doipset(sets...)
		if interval == 0 {
			break
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func doipset(sets ...string) {
	for _, set := range sets {
		flush := true
		tmp := strings.SplitN(set, "=", 2)
		if len(tmp) != 2 {
			tmp = strings.SplitN(set, "+", 2)
			if len(tmp) != 2 {
				fmt.Printf("[ERROR] Invalid set format, format: <set name>{=|+}domain[,domain...] ...\n")
				continue
			}
			flush = false
		}
		name := tmp[0]
		dnss := strings.Split(tmp[1], ",")
		ips := make(map[string][]string, 0)
		for _, dns := range dnss {
			addrs, err := net.LookupHost(dns)
			if err != nil {
				fmt.Printf("[ERROR] Lookup[%s] failure, %s\n", dns, err.Error())
				continue
			}
			ips[dns] = addrs
		}
		if len(ips) == 0 {
			continue
		}

		shell := "#! /bin/sh -e\n"
		if flush {
			shell += fmt.Sprintf("ipset flush %s 2>/dev/null || ipset create %s hash:net\n",
				name, name,
			)
		} else {
			shell += fmt.Sprintf("ipset create %s hash:net || true\n", name)
		}
		for _, ip := range ips {
			for _, addr := range ip {
				shell += fmt.Sprintf("ipset -exist add %s %s\n", name, addr)
			}
		}
		err := exec.Command("sh", "-c", shell).Run()
		if err != nil {
			fmt.Printf("[ERROR] Handle set[%s] failure, %s\n", name, err.Error())
			fmt.Printf("%s\n", shell)
			continue
		}
		for dns, ip := range ips {
			if flush {
				fmt.Printf("%s = %s %s\n", name, dns, ip)
			} else {
				fmt.Printf("%s + %s %s\n", name, dns, ip)
			}
		}
	}
}
