package main

import (
	"fmt"
)

func guide() {
	fmt.Printf("[Install]\n")
	fmt.Printf("1. copy oum to dir /usr/local/bin/\n")
	fmt.Printf("2. install dependency: openvpn ipset iptables openssl\n")
	fmt.Printf("# optional, hack sqlite3 db of oum\n")
	fmt.Printf("3. install sqlite3\n")
	fmt.Printf("\n")

	fmt.Printf("[Add Server]\n")
	fmt.Printf("1. oum pattern --out server.conf # follow the tips to generate openvpn server configuration and put it into the file server.conf\n")
	fmt.Printf("2. copy server.conf to /etc/openvpn/\n")
	fmt.Printf("3. start openvpn service with server.conf # service openvpn restart\n")
	fmt.Printf("4. oum serv add /etc/openvpn/server.conf # assume interface name is 'oum'\n")
	fmt.Printf("\n")

	fmt.Printf("[Add User]\n")
	fmt.Printf("1. oum user add cjey # add a user named cjey\n")
	fmt.Printf("2. pass the output url link or the qrcode image to cjey\n")
	fmt.Printf("3. cjey must install mobile app freeotp\n")
	fmt.Printf("4. cjey use freeotp register the qrcode from step 2\n")
	fmt.Printf("\n")

	fmt.Printf("[Add Device, optional]\n")
	fmt.Printf("1. oum user add cjey%%phone # add a device named phone belong to user cjey\n")
	fmt.Printf("\n")

	fmt.Printf("[Be Router, optional]\n")
	fmt.Printf("1. oum iptables --out startup\n")
	fmt.Printf("2. modify the script 'startup', to match your network\n")
	fmt.Printf("3. add the script 'startup' to execute at server boot\n")
	fmt.Printf("\n")

	fmt.Printf("[Add Client - by cli]\n")
	fmt.Printf("# generate client conffile pair to interface 'oum', ref command: oum serv\n")
	fmt.Printf("1. oum pattern --dev oum --out client.conf\n")
	fmt.Printf("2. pass the client.conf to user cjey\n")
	fmt.Printf("3. cjey use openvpn client import the client.conf\n")
	fmt.Printf("4. cjey login, when ask for username/password, get the otp code, then input username <cjey> and the password <otp code>\n")
	fmt.Printf("\n")

	fmt.Printf("[Add Client - by gui]\n")
	fmt.Printf("1. oum web\n")
	fmt.Printf("2. let user cjey access the web served by oum\n")
	fmt.Printf("3. download configuration file from the web, and import the file\n")
	fmt.Printf("4. cjey login, when ask for username/password, get the otp code, then input username <cjey> and the password <otp code>\n")
}
