package hook

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
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
	a_ip := env["untrusted_ip"]
	a_port := env["untrusted_port"]

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

	rows, err := DB.Query(`
        select ovpn_dev,device,ip,netmask,gateway,routes,dns,iroutes from ifconfig
        where username=?
    `, name)
	if err != nil && err != sql.ErrNoRows {
		slog.Emerg(err.Error())
		os.Exit(1)
	}

	var static, netmask, gateway, route, dns, iroute string
	if err == nil {
		var hit, a, b, c, d []string
		hit = []string{static, netmask, gateway, route, dns, iroute}
		err = db.RangeRows(rows, func() error {
			var cdev, dvc string
			err := rows.Scan(&cdev, &dvc, &static, &netmask, &gateway, &route, &dns, &iroute)
			if err != nil {
				return err
			}
			if cdev == dev && dvc == device {
				a = []string{static, netmask, gateway, route, dns, iroute}
			} else if cdev == dev && dvc == "" {
				b = []string{static, netmask, gateway, route, dns, iroute}
			} else if cdev == "" && dvc == device {
				c = []string{static, netmask, gateway, route, dns, iroute}
			} else if cdev == "" && dvc == "" {
				d = []string{static, netmask, gateway, route, dns, iroute}
			}
			return nil
		})
		if err != nil {
			slog.Emergf(err.Error())
			os.Exit(1)
		}
		switch {
		case len(a) > 0:
			hit = a
		case len(b) > 0:
			hit = b
		case len(c) > 0:
			hit = c
		case len(d) > 0:
			hit = d
		}
		static, netmask, gateway, route, dns, iroute = hit[0], hit[1], hit[2], hit[3], hit[4], hit[5]
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
	var iroutes []string
	if len(iroute) > 0 {
		tmp := utils.CSVSet(iroute)
		for _, r := range tmp {
			if strings.Index(r, "/") < 0 {
				r += "/32"
			}
			_, rr, err := net.ParseCIDR(r)
			if err != nil || rr == nil {
				continue
			}
			iroutes = append(iroutes, rr.IP.String()+" "+net.IP(rr.Mask).String())
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

	// config iroute
	for _, iroute := range iroutes {
		f.WriteString(fmt.Sprintf("iroute %s\n", iroute))
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

	var a_city, isp string
	ip := net.ParseIP(a_ip).To4()
	if ReservedIPv4(ip) == false {
		iip, err := NewIP(ip)
		if err == nil {
			a_city = iip.City()
			isp = iip.ISP()
		}
	}

	user.DisconnectDevice(name, device)

	tx, err := DB.Begin()
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
        insert or replace into active
            (username,device,cname,ovpn_dev,ip,netmask,access_ip,access_port,access_city,access_isp,connect_time)
        values
            (?,?,?,?,?,?,?,?,?,?,datetime('now'))
    `, name, device, env["common_name"], dev, remote_ip, netmask, a_ip, a_port, a_city, isp)
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

	addToIPset(name, remote_ip, a_ip)

	slog.Infof("Device[%s], Connected", show)
}

func addToIPset(name, as_item, ac_item string) {
	sets_as, sets_ac := userIPset(name)

	var shell string
	for _, set := range sets_as {
		shell += fmt.Sprintf("ipset add %s %s\n", set, as_item)
	}
	for _, set := range sets_ac {
		shell += fmt.Sprintf("ipset add %s %s\n", set, ac_item)
	}
	if len(shell) > 0 {
		exec.Command("/bin/sh", "-c", shell).Run()
	}
}
