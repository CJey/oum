package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"

	"utils"

	"github.com/cjey/slog"
)

const (
	errpre = "    [error]"
)

func getOUM(gw string, quick bool) string {
	def := "/usr/local/bin/oum"
	p, err := exec.LookPath("oum")
	if err == nil {
		def = p
	}
	var line string
	for {
		if quick {
			line = def
		} else {
			fmt.Printf("OUM executable file path(%s): ", def)
			line = utils.Readline(def)
		}
		if line[0] != '/' {
			fmt.Printf("%s absolute path required\n", errpre)
			continue
		}
		fi, err := os.Stat(line)
		if err != nil {
			fmt.Printf("%s %s\n", errpre, err.Error())
			continue
		}
		if fi.IsDir() {
			fmt.Printf("%s oum executable file path required\n", errpre)
			continue
		}
		return fmt.Sprintf("# use setenv to control oum\n") +
			fmt.Sprintf("#setenv oum_sameip   1296000 # 15days\n") +
			fmt.Sprintf("#setenv oum_samecity 604800 # 7days\n") +
			fmt.Sprintf("#setenv oum_gateway  %s\n", gw) +
			fmt.Sprintf("#setenv oum_dns      %s # support csv\n", gw) +
			fmt.Sprintf("\n") +
			fmt.Sprintf("up '%s hook'\n", line) +
			fmt.Sprintf("auth-user-pass-verify '%s hook' via-env\n", line) +
			fmt.Sprintf("client-connect '%s hook'\n", line) +
			fmt.Sprintf("client-disconnect '%s hook'\n", line) +
			fmt.Sprintf("down '%s hook'\n", line)
	}
}

func getType(quick bool) string {
	def := "tun"
	var line string
	for {
		if quick {
			line = def
		} else {
			fmt.Printf("Choose TUN/TAP driver mode(TUN/tap): ")
			line = strings.ToLower(utils.Readline(def))
		}
		switch line {
		case "tun":
			return "dev-type tun\n\n" +
				"#dev-type tap\n"
		case "tap":
			return "#dev-type tun\n\n" +
				"dev-type tap\n"
		default:
			continue
		}
	}
}

func getDev(quick bool) string {
	def := "oum"
	var line string
	for {
		if quick {
			line = def
		} else {
			fmt.Printf("Network interface name(%s): ", def)
			line = utils.Readline(def)
		}
		if len(line) > 15 {
			fmt.Printf("%s name too long\n", errpre)
			continue
		}
		return fmt.Sprintf("dev %s # Network interface name\n", line)
	}
}

func getProto(quick bool) string {
	def := "tcp"
	var line string
	for {
		if quick {
			line = def
		} else {
			fmt.Printf("Choose transport layer protocol(TCP/udp): ")
			line = strings.ToLower(utils.Readline(def))
		}
		switch line {
		case "tcp":
			return "proto tcp-server\n" +
				"tcp-nodelay\n\n" +
				"#proto udp\n"
		case "udp":
			return "#proto tcp-server\n" +
				"#tcp-nodelay\n\n" +
				"proto udp\n"
		default:
			continue
		}
	}
}

func getPort(quick bool) string {
	def := "1194"
	var line string
	for {
		if quick {
			line = def
		} else {
			fmt.Printf("Listen port(%s): ", def)
			line = utils.Readline(def)
		}
		var port uint16
		_, err := fmt.Sscan(line, &port)
		if err != nil {
			fmt.Printf("%s %s\n", errpre, err.Error())
			continue
		}
		return fmt.Sprintf("port %s # Listen port\n", line)
	}
}

func getIfconfig(quick bool) (string, string) {
	def := "192.168.94.1/24"
	var ip net.IP
	var ipnet *net.IPNet
	var line string
	for {
		if quick {
			line = def
		} else {
			fmt.Printf("Server ifconfig(%s): ", def)
			line = utils.Readline(def)
		}
		var err error
		ip, ipnet, err = net.ParseCIDR(line)
		if err != nil {
			fmt.Printf("%s %s\n", errpre, err.Error())
			continue
		}
		if ip.String() == ipnet.IP.String() {
			fmt.Printf("%s please specify the part of the server ip\n", errpre)
			continue
		}
		break
	}
	first := fmt.Sprintf("ifconfig %s %s # Server network ifconfig\n", ip.String(), net.IP(ipnet.Mask).String())
	ones, bits := ipnet.Mask.Size()

	var pooldef string
	if ones >= 30 {
		pooldef = "192.168.94.100 192.168.94.199 255.255.255.0"
	} else {
		ipstart := net.IP(make([]byte, 4))
		ipend := net.IP(make([]byte, 4))

		zero := binary.BigEndian.Uint32(ip.To4())
		start := binary.BigEndian.Uint32(ipnet.IP.To4()) + 1
		end := (start - 1) + (1<<uint(bits-ones) - 1) - 1 - 1 // broadcast & dhcp(win convention)

		if (end - zero) >= (zero - start) {
			binary.BigEndian.PutUint32(ipstart, zero+1)
			binary.BigEndian.PutUint32(ipend, end)
		} else {
			binary.BigEndian.PutUint32(ipstart, start)
			binary.BigEndian.PutUint32(ipend, zero-1)
		}

		pooldef = fmt.Sprintf("%s %s %s", ipstart, ipend, net.IP(ipnet.Mask).String())
	}

	var second string
	help := func() {
		fmt.Printf("    Format: <pool from> <pool to> <netmask>\n")
		fmt.Printf("    Default: %s\n", pooldef)
	}
	for {
		if quick {
			line = ""
		} else {
			fmt.Printf("Client ifconfig pool(help): ")
			line = utils.Readline("")
		}
		if len(line) == 0 {
			second = fmt.Sprintf("ifconfig-pool %s # Client dhcp ip pool\n", pooldef)
			break
		}
		if line == "help" {
			help()
			continue
		}
		{
			tmp := strings.Split(line, " ")
			res := []string{}
			for _, item := range tmp {
				item = strings.TrimSpace(item)
				if len(item) > 0 {
					res = append(res, item)
				}
			}
			if len(res) != 3 {
				fmt.Printf("%s need 3 args\n", errpre)
				help()
				continue
			}
			ip := net.ParseIP(res[0]).To4()
			if ip == nil {
				fmt.Printf("%s invalid ipv4 address(%s)\n", errpre, res[0])
				help()
				continue
			}
			ip = net.ParseIP(res[1]).To4()
			if ip == nil {
				fmt.Printf("%s invalid ipv4 address(%s)\n", errpre, res[1])
				help()
				continue
			}
			ones, bits := net.IPMask(net.ParseIP(res[2]).To4()).Size()
			if ones|bits == 0 {
				fmt.Printf("%s invalid network mask(%s)\n", errpre, res[2])
				help()
				continue
			}
			second = fmt.Sprintf("ifconfig-pool %s %s %s # Client dhcp ip pool\n", res[0], res[1], res[2])
			break
		}
	}
	return ip.String(), first + second
}

func getCerts(quick, ecdsa bool) string {
	cakey, err := genCAKey(ecdsa)
	if err != nil {
		fmt.Printf("%s %s\n", errpre, err.Error())
		os.Exit(1)
	}
	cacrt, err := genCACert(cakey, "CN=CA")
	if err != nil {
		fmt.Printf("%s %s\n", errpre, err.Error())
		os.Exit(1)
	}
	key, err := genSrvKey(ecdsa)
	if err != nil {
		fmt.Printf("%s %s\n", errpre, err.Error())
		os.Exit(1)
	}
	crt, err := genSrvCert(key, cacrt, cakey, "CN=Server")
	if err != nil {
		fmt.Printf("%s %s\n", errpre, err.Error())
		os.Exit(1)
	}
	if !quick {
		fmt.Printf("Generating dhparams, it will take some time...\n")
	}
	dh, err := genDH(quick)
	if err != nil {
		fmt.Printf("%s %s\n", errpre, err.Error())
		os.Exit(1)
	}
	return fmt.Sprintf("<ca>\n%s</ca>\n\n"+
		"<cert>\n%s</cert>\n\n"+
		"<key>\n%s</key>\n\n"+
		"<dh>\n%s</dh>\n",
		cacrt, crt, key, dh)
}

func showPattern(outpath string, ecdsa, quick bool) {
	devname := getDev(quick)
	devtype := getType(quick)
	proto := getProto(quick)
	port := getPort(quick)
	gw, ifconfig := getIfconfig(quick)
	hooks := getOUM(gw, quick)
	certs := getCerts(quick, ecdsa)

	var buf bytes.Buffer

	mgr := make([]byte, 16)
	_, _ = rand.Read(mgr)
	mgrc := fmt.Sprintf("management /var/run/oum-%x unix\n", mgr)

	buf.WriteString("# openvpn conf - server\n")
	buf.WriteString("# auto generated by oum\n\n")
	buf.WriteString(devname)
	buf.WriteString(port)
	buf.WriteString(ifconfig)
	buf.WriteString("# ----------------\n\n")
	buf.WriteString("# you can push local associated networks route to each clients\n")
	buf.WriteString("# push '<network> <netmask> <ip in the ifconfig>'\n\n")
	buf.WriteString(fmt.Sprintf("# push '192.168.0.0 255.255.0.0 %s'\n", gw))
	buf.WriteString(fmt.Sprintf("# push '172.16.0.0  255.240.0.0 %s'\n", gw))
	buf.WriteString(fmt.Sprintf("# push '10.0.0.0    255.0.0.0   %s'\n", gw))
	buf.WriteString("# ----------------\n\n")
	buf.WriteString(proto)
	buf.WriteString("# ----------------\n\n")
	buf.WriteString(devtype)
	buf.WriteString("# ----------------\n\n")
	buf.WriteString("keepalive 30 100\n")
	buf.WriteString("cipher AES-128-CBC\n")
	buf.WriteString("auth SHA1\n")
	buf.WriteString("# ----------------\n\n")
	buf.WriteString(hooks)
	buf.WriteString(mgrc)
	buf.WriteString("# ----------------\n\n")
	buf.WriteString("verb 3\n")
	buf.WriteString("reneg-sec 0\n")
	buf.WriteString("topology subnet\n")
	buf.WriteString("push 'topology subnet'\n")
	buf.WriteString("route-metric 1000\n")
	buf.WriteString("username-as-common-name\n")
	buf.WriteString("client-cert-not-required\n")
	buf.WriteString("script-security 3\n")
	buf.WriteString("mode server\n")
	buf.WriteString("tls-server\n")
	buf.WriteString("tun-mtu 1400\n")
	buf.WriteString("persist-key\n")
	buf.WriteString("persist-tun\n")
	buf.WriteString("# ----------------\n\n")
	buf.WriteString(certs)

	if outpath == "-" {
		fmt.Printf("\n>>>>>>>> copy all blow to file with suffix .conf <<<<<<<<\n\n")
		fmt.Printf("%s", buf.String())
	} else {
		err := ioutil.WriteFile(outpath, buf.Bytes(), 0600)
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
	}
}
