package hook

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"time"

	"cmd/user"
	"db"
	"utils"

	"github.com/cjey/slog"
)

func Disconnect(env map[string]string) {
	now := time.Now().UTC()
	dev := env["dev"]
	username := env["username"]
	if len(username) == 0 {
		username = env["common_name"]
	}
	bytes_sent := env["bytes_sent"]
	bytes_received := env["bytes_received"]

	name, device := user.StdUsername(username)
	if len(name) == 0 {
		slog.Emergf("Invalid username format: %s", username)
		os.Exit(1)
	}
	if len(device) == 0 {
		device = "default"
	}
	show := name + "%" + device

	DB := db.Get()
	tx, err := DB.Begin()
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	defer tx.Rollback()

	var remote_ip, netmask, cname, a_ip, a_port, a_city, a_isp string
	var ctime time.Time
	err = tx.QueryRow(`
        select ip,netmask,cname,access_ip,access_port,access_city,access_isp,connect_time from active
        where username=? and device=? and ovpn_dev=?
    `, name, device, dev).Scan(&remote_ip, &netmask, &cname, &a_ip, &a_port, &a_city, &a_isp, &ctime)
	if err != nil && err != sql.ErrNoRows {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	if err == sql.ErrNoRows {
		slog.Warningf("No Active connection found")
		os.Exit(1)
	}

	_, err = tx.Exec(`
        insert into log(
            username,device,cname,ovpn_dev,
            ip,netmask,
            access_ip,access_port,access_city,access_isp,
            connect_time,disconnect_time,
            bytes_sent,bytes_received
        ) values (
            ?, ?, ?, ?,
            ?, ?,
            ?, ?, ?, ?,
            ?, ?,
            ?, ?
        )
    `, name, device, cname, dev,
		remote_ip, netmask,
		a_ip, a_port, a_city, a_isp,
		ctime.Format(utils.TIMEFORMAT), now.Format(utils.TIMEFORMAT),
		bytes_sent, bytes_received,
	)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}

	_, err = tx.Exec(`
        delete from active
        where username=? and device=? and ovpn_dev=?
    `, name, device, dev)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}

	uptime := int(now.Sub(ctime).Seconds())

	_, err = tx.Exec(`
        update user set
            total_uptime=total_uptime+?,
            total_bytes_sent=total_bytes_sent+?,
            total_bytes_received=total_bytes_received+?
        where username=?
    `, uptime, bytes_sent, bytes_received, name)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}

	_, err = tx.Exec(`
        update device set
            total_uptime=total_uptime+?,
            total_bytes_sent=total_bytes_sent+?,
            total_bytes_received=total_bytes_received+?
        where username=? and device=?
    `, uptime, bytes_sent, bytes_received, name, device)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}

	err = tx.Commit()
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}

	removeFromIPset(name, remote_ip, a_ip)

	slog.Infof("Device[%s], Disconnected", show)
}

func removeFromIPset(name, as_item, ac_item string) {
	sets_as, sets_ac := userIPset(name)

	var shell string
	for _, set := range sets_as {
		shell += fmt.Sprintf("ipset del %s %s\n", set, as_item)
	}
	for _, set := range sets_ac {
		shell += fmt.Sprintf("ipset del %s %s\n", set, ac_item)
	}
	if len(shell) > 0 {
		exec.Command("/bin/sh", "-c", shell).Run()
	}
}
