package user

import (
	"bufio"
	"database/sql"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"db"
	"utils"

	"github.com/cjey/slog"
)

func Reconnect(username string) {
	name, device := StdUsername(username)
	if len(name) == 0 {
		slog.Emergf("Invalid username format: %s", username)
		os.Exit(1)
	}
	if len(device) == 0 {
		reconnectUser(name)
	} else {
		reconnectDevice(name, device)
	}
}

func reconnectUser(name string) {
	total, effect := DisconnectDevice(name, "")
	if total == 0 {
		slog.Infof("Inactive User[%s]", name)
		return
	}

	if effect == 0 {
		slog.Warningf("User[%s] reconnect failure, NOT/ERROR configured table: ovpn", name)
	} else {
		slog.Infof("User[%s] reconnect successfully", name)
	}
}

func reconnectDevice(name, device string) {
	total, effect := DisconnectDevice(name, device)
	if total == 0 {
		slog.Infof("Inactive Device[%s] of User[%s]", device, name)
		return
	}

	if effect == 0 {
		slog.Warningf("Device[%s] of User[%s] reconnect failure, NOT/ERROR configured table: ovpn", device, name)
	} else {
		slog.Infof("Device[%s] of User[%s] reconnect successfully", device, name)
	}
}

func DisconnectDevice(name, dev string) (total, effect int) {
	DB := db.Get()
	rows, err := DB.Query(`
        select device,cname,ovpn_dev from active
        where username=?
    `, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return
		}
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	speed := make(map[string][]string, 0)
	err = db.RangeRows(rows, func() error {
		var device, cname, ovpn string
		err := rows.Scan(&device, &cname, &ovpn)
		if err != nil {
			return err
		}
		if len(dev) == 0 || device == dev {
			total++
			cnames := speed[ovpn]
			if len(cnames) == 0 {
				speed[ovpn] = []string{cname}
			} else {
				speed[ovpn] = append(cnames, cname)
			}
		}
		return nil
	})
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}

	for ovpn, cnames := range speed {
		effect += disconnect(ovpn, cnames...)
	}

	return
}

func disconnect(dev string, cnames ...string) (affect int) {
	DB := db.Get()

	var srvpath string
	err := DB.QueryRow(`
        select conffile from ovpn
        where dev=?
    `, dev).Scan(&srvpath)
	if err != nil && err != sql.ErrNoRows {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	if err == sql.ErrNoRows {
		err = DB.QueryRow(`
            select conffile from ovpn
            where dev=''
        `).Scan(&srvpath)
		if err != nil && err != sql.ErrNoRows {
			slog.Emergf(err.Error())
			os.Exit(1)
		}
		if err == sql.ErrNoRows {
			return
		}
	}

	lines, err := utils.StripConf(srvpath)
	if err != nil {
		return
	}
	line := utils.FetchConfKey(lines, "management", "")
	fs := strings.Fields(line)
	if len(fs) < 3 {
		return
	}
	ip := fs[1]
	port := fs[2]
	var pw string
	if len(fs) > 3 && fs[3][0] != '#' {
		_pw, err := ioutil.ReadFile(fs[3])
		if err != nil {
			return
		}
		pw = strings.TrimRight(string(_pw), "\r\n")
	}

	return sendKill(dev, ip, port, pw, cnames...)
}

func sendKill(name, ip, port, pw string, cnames ...string) (affect int) {
	if ip == "tunnel" {
		return
	}
	if len(cnames) == 0 {
		return
	}

	cmds := make([]string, 0)
	for _, cname := range cnames {
		cmds = append(cmds, "kill "+cname)
	}

	var con net.Conn
	if port == "unix" {
		raddr, err := net.ResolveUnixAddr("unix", ip)
		if err != nil {
			return
		}
		con, err = net.DialUnix("unix", nil, raddr)
		if err != nil {
			return
		}
	} else {
		raddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(ip, port))
		if err != nil {
			return
		}
		con, err = net.DialTCP("tcp", nil, raddr)
		if err != nil {
			return
		}
	}
	defer con.Close()
	rdr := bufio.NewReader(con)

	if len(pw) > 0 {
		_, err := con.Write([]byte(pw + "\n"))
		if err != nil {
			return
		}
		for {
			reply, err := rdr.ReadString('\n')
			if err != nil {
				return
			}
			reply = strings.TrimSpace(reply)
			if reply[0] == '>' {
				continue
			}
			break
		}
	}
	for _, cmd := range cmds {
		_, err := con.Write([]byte(cmd + "\n"))
		if err != nil {
			return
		}
		var reply string
		for {
			var err error
			reply, err = rdr.ReadString('\n')
			if err != nil {
				return
			}
			reply = strings.TrimSpace(reply)
			if reply[0] == '>' {
				continue
			}
			break
		}
		affect++
		slog.Infof("Send command to %s: %s\n", name, cmd)
	}
	return
}
