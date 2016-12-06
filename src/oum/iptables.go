package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cjey/slog"
)

func iptables(outpath string) {
	if outpath == "-" {
		fmt.Print(iptables_script)
	} else {
		err := ioutil.WriteFile(outpath, []byte(iptables_script), 0775)
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
	}
}

const iptables_script = `#! /bin/sh
set -e

_ROOT="$(pwd)" && cd "$(dirname "$0")" && ROOT="$(pwd)"

MARK_NAT=0x80000000

ipset_clean() {
    ipset list -n | xargs -I{} ipset destroy {}
}

iptables_clean() {
    sysctl -w net.ipv4.ip_forward=0 >/dev/null

    iptables-restore <<EOF
*filter
:INPUT ACCEPT [0:0]
:FORWARD ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
COMMIT
*nat
:PREROUTING ACCEPT [0:0]
:INPUT ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
:POSTROUTING ACCEPT [0:0]
COMMIT
*mangle
:PREROUTING ACCEPT [0:0]
:INPUT ACCEPT [0:0]
:FORWARD ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
:POSTROUTING ACCEPT [0:0]
COMMIT
*raw
:PREROUTING ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
COMMIT
*security
:INPUT ACCEPT [0:0]
:FORWARD ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
COMMIT
EOF
}

define_ipset() {
    name="$1"
    shift
    ipset create $name hash:net
    for cidr in $*; do
        ipset add $name $cidr
    done
}

define_ipset_from_file() {
    name="$1"
    iplist="$2"
    ipset create $name hash:net
    grep -vP '^\s*(#|$)' $iplist | sed -nr "s/^(.+)\$/add $name \1/p" | ipset -exist restore
}

traffic_from_ipset() {
    name="$1"
    callback="$2"
    match="-m set --match-set $name src"
    [ "$name" = "unspec" ] && match=
    iptables -t filter -N input_from_$name
    iptables -t filter -N forward_from_$name
    iptables -t nat    -N dnat_from_$name
    iptables -t filter -A INPUT      -g input_from_$name   $match
    iptables -t filter -A FORWARD    -g forward_from_$name $match
    iptables -t nat    -A PREROUTING -g dnat_from_$name    $match

    iptables -t filter -A forward_from_$name -m conntrack --ctstate DNAT -j ACCEPT_NAT

    $callback input_from_$name forward_from_$name dnat_from_$name
}

iptables_init() {
    sysctl -w net.ipv4.ip_forward=1 >/dev/null

    # helper chain: ACCEPT_NAT, use sampe:
    # means: accept traffic from 192.168.1.0/24 to anywhere, *AND* *AUTO* MASQUERADE the connection
    # iptables -t filter -A FORWARD -s 192.168.1.0/24 -j ACCEPT_NAT
    iptables -t filter -N ACCEPT_NAT
    iptables -t filter -A ACCEPT_NAT -j MARK --set-mark $MARK_NAT/$MARK_NAT
    iptables -t filter -A ACCEPT_NAT -j ACCEPT

    iptables -t filter -P INPUT DROP
    iptables -t filter -A INPUT -j ACCEPT -i lo
    iptables -t filter -A INPUT -j ACCEPT -m conntrack --ctstate   ESTABLISHED,RELATED
    # dhcp
    iptables -t filter -A INPUT -j ACCEPT \
        -s 0.0.0.0 -d 255.255.255.255 -p udp --sport 68 --dport 67

    iptables -t filter -P FORWARD DROP
    iptables -t filter -A FORWARD -j ACCEPT -m conntrack --ctstate ESTABLISHED,RELATED

    iptables -t nat -A POSTROUTING -j RETURN     -m addrtype --src-type LOCAL
    iptables -t nat -A POSTROUTING -j TCPMSS     --clamp-mss-to-pmtu -p tcp --tcp-flags SYN SYN
    iptables -t nat -A POSTROUTING -j MASQUERADE --random -m mark --mark $MARK_NAT/$MARK_NAT
}

#--------------------------------------------------

__traffic_from_sample() {
    INPUT=$1
    FORWARD=$2
    PREROUTING=$3

    # INPUT: control the traffic which from sample to me
    # allow <sample> ping me
    iptables -t filter -A $INPUT -j ACCEPT -p icmp --icmp-type echo-request
    # allow <sample> access my ssh
    iptables -t filter -A $INPUT -j ACCEPT -p tcp  --dport 22
    # allow <sample> access my dns
    iptables -t filter -A $INPUT -j ACCEPT -p udp  --dport 53
    # allow <sample> access openvpn tcp mode
    iptables -t filter -A $INPUT -j ACCEPT -p tcp  --dport 1194
    # allow <sample> access openvpn udp mode
    iptables -t filter -A $INPUT -j ACCEPT -p udp  --dport 1194

    # FORWARD: control the traffic which from sample to other
    # allow <sample> forward traffic to 192.168.1.0/24
    iptables -t filter -A $FORWARD -j ACCEPT -d 192.168.1.0/24
    # allow <sample> forward traffic to 8.8.8.8, and masquerade the traffic
    # if you do not understand what the different between forward and nat, always use nat
    iptables -t filter -A $FORWARD -j ACCEPT_NAT -d 8.8.8.8

    # DNAT: change the traffic from sample to placeB, which from sample to placeA originally
    # allow <sample> access my tcp:12345, and dnat the traffic to 1.2.3.4:54321
    iptables -t nat -A $PREROUTING -m addrtype --dst-type LOCAL -p tcp --dport 12345 -j DNAT --to 1.2.3.4:54321
}

#--------------------------------------------------
# clean iptables first, then clean ipset
iptables_clean
ipset_clean

# install base iptables skeleton
iptables_init
#--------------------------------------------------

#@@@@@@@@@@@@@@@@@@ <Customize> @@@@@@@@@@@@@@@@@@@
# Define sets first
#define_ipset <set name> cidr ...
#define_ipset_from_file <set name> <cidrs list file path, one cidr one line>

define_ipset vpn      192.168.94.0/24
define_ipset localnet 192.168.1.0/24
#--------------------------------------------------
# Rules
#traffic_from_ipset <set name> <callback, func or executable>

traffic_from_vpn() {
    INPUT=$1
    FORWARD=$2
    PREROUTING=$3

    # INPUT: control the traffic which from vpn to me
    iptables -t filter -A $INPUT -j ACCEPT -p udp  --dport 53 # dns
    iptables -t filter -A $INPUT -j ACCEPT -p tcp  --dport 22 # ssh
    iptables -t filter -A $INPUT -j ACCEPT -p icmp --icmp-type echo-request # ping

    # FORWARD: control the traffic which from vpn to other
    # allow vpn clients access each other
    iptables -t filter -A $FORWARD -j ACCEPT     -m set --match-set vpn dst
    # allow vpn access any other(localnet & Internet), and do NAT
    iptables -t filter -A $FORWARD -j ACCEPT_NAT
}

traffic_from_ipset vpn traffic_from_vpn

#----

traffic_from_localnet() {
    INPUT=$1
    FORWARD=$2
    PREROUTING=$3

    # INPUT: control the traffic which from localnet to me
    iptables -t filter -A $INPUT -j ACCEPT -p udp  --dport 53 # dns
    iptables -t filter -A $INPUT -j ACCEPT -p tcp  --dport 22 # ssh
    iptables -t filter -A $INPUT -j ACCEPT -p icmp --icmp-type echo-request # ping
    iptables -t filter -A $INPUT -j ACCEPT -p tcp  --dport 1194 # openvpn tcp mode
    iptables -t filter -A $INPUT -j ACCEPT -p udp  --dport 1194 # openvpn udp mode
    iptables -t filter -A $INPUT -j ACCEPT -p tcp  --dport 1195 # oum web

    # FORWARD: control the traffic which from localnet to other
    # allow localnet access each other
    iptables -t filter -A $FORWARD -j ACCEPT -m set --match-set localnet dst
    # allow localnet access any other(vpn & Internet), and do NAT
    iptables -t filter -A $FORWARD -j ACCEPT_NAT
}

traffic_from_ipset localnet traffic_from_localnet

#----

# "unspec" is a spcial ipset
# means any other networks not matched in the given Rules order
# generally, it means Internet traffic

# WARNING: Always let unspec at the end

traffic_from_unspec() {
    INPUT=$1
    FORWARD=$2
    PREROUTING=$3

    # INPUT: control the traffic which from unsepcified network(Internet) to me
    iptables -t filter -A $INPUT -j ACCEPT -p tcp  --dport 1194 # openvpn tcp mode
    iptables -t filter -A $INPUT -j ACCEPT -p udp  --dport 1194 # openvpn udp mode
    iptables -t filter -A $INPUT -j ACCEPT -p tcp  --dport 1195 # oum web
    iptables -t filter -A $INPUT -j ACCEPT -p icmp --icmp-type echo-request # ping
}

traffic_from_ipset unspec traffic_from_unspec

#@@@@@@@@@@@@@@@@@ </Customize> @@@@@@@@@@@@@@@@@@@
`
