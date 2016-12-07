package user

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"db"

	"github.com/cjey/slog"
)

func Ifconfig(username, dev string, configs ...string) {
	name, device := StdUsername(username)
	if len(name) == 0 {
		slog.Emergf("Invalid username format: %s", username)
		os.Exit(1)
	}
	var alldev bool
	if len(device) == 0 {
		device = "default"
		alldev = true
	}

	DB := db.Get()

	var fexists int
	if len(dev) > 0 {
		err := DB.QueryRow(`
            select count(1) from ovpn
            where dev=?
        `, dev).Scan(&fexists)
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
		if fexists == 0 {
			slog.Emergf("Dev[%s] not served\n", dev)
			os.Exit(1)
		}
	}

	err := DB.QueryRow(`
        select count(1) from device
        where username=? and device=?
    `, name, device).Scan(&fexists)
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	if fexists == 0 {
		slog.Warningf("Device[%s] of user[%s] not exists", device, name)
		return
	}

	valid := map[string]string{}
	for _, config := range configs {
		tmp := strings.SplitN(config, "=", 2)
		key := strings.ToLower(tmp[0])
		switch key {
		case "ip", "dns":
		default:
			slog.Emergf("Unsupported config field[%s]", key)
			os.Exit(1)
		}
		if len(tmp) > 1 {
			valid[key] = tmp[1]
		} else {
			valid[key] = ""
		}
	}

	if len(valid) > 0 {
		var ip, dns string
		err := DB.QueryRow(`
            select ip,dns from ifconfig
            where username=? and device=? and ovpn_dev=?
        `, name, device, dev).Scan(&ip, &dns)
		if err != nil && err != sql.ErrNoRows {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
		if v, ok := valid["ip"]; ok {
			ip = v
		}
		if v, ok := valid["dns"]; ok {
			dns = v
		}

		_, err = DB.Exec(`
            insert or replace into ifconfig
                (username,device,ovpn_dev,ip,dns)
            values
                (?,?,?,?,?)
        `, name, device, dev, ip, dns)
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
	}

	if alldev {
		ifconfigShow(name, "")
	} else {
		ifconfigShow(name, device)
	}
}

func ifconfigShow(name, device string) {
	DB := db.Get()
	rows, err := DB.Query(`
        select ovpn_dev,device,ip,dns from ifconfig
        where username=?
    `, name)
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}

	var fmore bool
	err = db.RangeRows(rows, func() error {
		var dev, dvc, ip, dns string
		err := rows.Scan(&dev, &dvc, &ip, &dns)
		if err != nil {
			return err
		}
		if len(device) > 0 && device != dvc {
			return nil
		}
		if fmore {
			fmt.Printf("\n")
		}
		fmt.Printf("Device: %s%%%s\n", name, dvc)
		fmt.Printf("Interface: %s\n", dev)
		if len(ip) == 0 {
			ip = "dhcp"
		}
		fmt.Printf("IP: %s\n", ip)
		if len(dns) == 0 {
			dns = "dhcp"
		}
		fmt.Printf("DNS: %s\n", dns)
		fmore = true
		return nil
	})
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
}
