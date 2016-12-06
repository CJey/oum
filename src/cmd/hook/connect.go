package hook

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"

	"cmd/user"
	"db"
	"utils"

	"github.com/cjey/slog"
)

func Connect(env map[string]string, dconfPath string) {
	dev := env["dev"]
	username := env["username"]
	if len(username) == 0 {
		username = env["common_name"]
	}
	ifconfig_local := env["ifconfig_local"]
	ifconfig_pool_remote_ip := env["ifconfig_pool_remote_ip"]
	ifconfig_netmask := env["ifconfig_netmask"]
	ifconfig_pool_netmask := env["ifconfig_pool_netmask"]
	ipdot := env["untrusted_ip"]

	f, err := os.OpenFile(dconfPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	defer f.Close()

	name, device := user.StdUsername(username)
	if len(name) == 0 {
		slog.Emergf("Invalid username format: %s", username)
		os.Exit(1)
	}
	if len(device) == 0 {
		device = "default"
	}
	show := name + "%" + device

	// fetch ifconfig
	DB := db.Get()
	var static, netmask, gateway, route, dns string
	err = DB.QueryRow(`
        select ip,netmask,gateway,routes,dns from ifconfig
        where username=? and device=? and ovpn_dev=?
    `, name, device, dev).Scan(&static, &netmask, &gateway, &route, &dns)
	if err != nil && err != sql.ErrNoRows {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	if err == sql.ErrNoRows {
		err = DB.QueryRow(`
            select ip,netmask,gateway,routes,dns from ifconfig
            where username=? and device='' and ovpn_dev=?
        `, name, dev).Scan(&static, &netmask, &gateway, &route, &dns)
		if err != nil && err != sql.ErrNoRows {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
	}

	remote_ip := ifconfig_pool_remote_ip
	if len(static) > 0 {
		remote_ip = static
	}
	if len(remote_ip) == 0 {
		netmask = ""
	} else if len(netmask) == 0 {
		if len(ifconfig_pool_netmask) > 0 {
			netmask = ifconfig_pool_netmask
		} else if len(ifconfig_netmask) > 0 {
			netmask = ifconfig_netmask
		} else {
			netmask = "255.255.255.255"
		}
	}
	if len(gateway) == 0 {
		if defgw := env["oum_gateway"]; len(defgw) > 0 {
			gateway = defgw
		} else {
			gateway = ifconfig_local
		}
	}
	var dnss []string
	if len(dns) > 0 {
		dnss = utils.CSVSet(dns)
	} else {
		if defdns := env["oum_dns"]; len(defdns) > 0 {
			dnss = utils.CSVSet(defdns)
		} else {
			dnss = []string{ifconfig_local}
		}
	}
	var routes []string
	if len(route) > 0 {
		tmp := utils.CSVSet(route)
		for _, r := range tmp {
			if strings.Index(r, "/") < 0 {
				r += "/32"
			}
			_, rr, err := net.ParseCIDR(r)
			if err != nil || rr == nil {
				continue
			}
			routes = append(routes, rr.IP.String()+" "+net.IP(rr.Mask).String())
		}
	}

	// push ip
	if len(static) > 0 {
		f.WriteString(fmt.Sprintf("ifconfig-push %s %s\n", static, netmask))
	}

	// push route
	for _, route := range routes {
		f.WriteString(fmt.Sprintf("push 'route %s %s'\n", route, gateway))
	}

	// redirect gateway
	if username[0] != '!' {
		f.WriteString(fmt.Sprintf("push 'route-gateway %s'\n", gateway))
		for _, dns := range dnss {
			f.WriteString(fmt.Sprintf("push 'dhcp-option DNS %s'\n", dns))
		}
		f.WriteString(fmt.Sprintf("push 'register-dns'\n"))
		f.WriteString("push 'redirect-gateway'\n")
	}

	var city, isp string
	ip := net.ParseIP(ipdot).To4()
	if ReservedIPv4(ip) == false {
		iip, err := NewIP(ip)
		if err == nil {
			city = iip.City()
			isp = iip.ISP()
		}
	}
	tx, err := DB.Begin()
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
        insert or replace into active
            (username,device,cname,ovpn_dev,ip,netmask,access_ip,access_city,access_isp,connect_time)
        values
            (?,?,?,?,?,?,?,?,?,datetime('now'))
    `, name, device, env["common_name"], dev, remote_ip, netmask, ipdot, city, isp)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	_, err = tx.Exec(`
        update user set total_login=total_login+1
        where username=?
    `, name)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	_, err = tx.Exec(`
        update device set total_login=total_login+1
        where username=? and device=?
    `, name, device)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	err = tx.Commit()
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}

	slog.Infof("Device[%s], Connected", show)
}
