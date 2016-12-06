package user

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"time"

	"db"

	"github.com/cjey/slog"
)

func List(usernames ...string) {
	DB := db.Get()
	rows, err := DB.Query(`
        select
            username,expired,
            total_login,total_uptime,total_bytes_sent,total_bytes_received
        from user
    `)
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	var f_split bool
	err = db.RangeRows(rows, func() error {
		var name string
		var expired time.Time
		var login, uptime, bytes_sent, bytes_recv int
		err = rows.Scan(&name, &expired, &login, &uptime, &bytes_sent, &bytes_recv)
		if err != nil {
			return err
		}

		if len(usernames) > 0 {
			out := false
			for _, username := range usernames {
				if username == name {
					out = true
				}
			}
			if out == false {
				return nil
			}
		}

		if f_split {
			fmt.Printf("--------\n")
		}
		f_split = true

		pre := ""
		if expired.Before(time.Now()) {
			pre = "!"
		}
		uptime_show := (time.Duration(uptime) * time.Second).String()
		fmt.Printf("%s%s  %d  %s  TX: %s  RX: %s\n",
			pre,
			name,
			login,
			uptime_show,
			human_bytes(bytes_sent),
			human_bytes(bytes_recv),
		)

		rows2, err := DB.Query(`
            select
                device,expired,
                total_login,total_uptime,total_bytes_sent,total_bytes_received
            from device
            where username=?
        `, name)
		if err != nil {
			return err
		}
		err = db.RangeRows(rows2, func() error {
			var device string
			err = rows2.Scan(&device, &expired, &login, &uptime, &bytes_sent, &bytes_recv)
			if err != nil {
				return err
			}
			pre := ""
			if expired.Before(time.Now()) {
				pre = "!"
			}

			sub := ""
			var dev, ip, netmask, a_ip string
			var ctime time.Time
			err = DB.QueryRow(`
                select
                    ovpn_dev,ip,netmask,access_ip,connect_time
                from active
                where username=? and device=?
            `, name, device).Scan(&dev, &ip, &netmask, &a_ip, &ctime)
			if err != nil && err != sql.ErrNoRows {
				return err
			}
			if err == nil && len(pre) == 0 {
				pre = "+"
				ones, _ := net.IPMask(net.ParseIP(netmask).To4()).Size()
				uptime := (time.Now().Sub(ctime) / time.Second * time.Second).String()
				sub = fmt.Sprintf(" (%s %s %s/%d %s)", dev, uptime, ip, ones, a_ip)
			}

			uptime_show = (time.Duration(uptime) * time.Second).String()
			fmt.Printf("%s%s  %d  %s  TX: %s  RX: %s%s\n",
				pre,
				name+"%"+device,
				login,
				uptime_show,
				human_bytes(bytes_sent),
				human_bytes(bytes_recv),
				sub,
			)
			return nil
		})
		return nil
	})
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
}

func human_bytes(ibytes int) string {
	bytes := float64(ibytes)
	switch {
	case bytes/1e15 > 10:
		return fmt.Sprintf("%.2fP", bytes/1e15)
	case bytes/1e12 > 10:
		return fmt.Sprintf("%.2fT", bytes/1e12)
	case bytes/1e9 > 10:
		return fmt.Sprintf("%.2fG", bytes/1e9)
	case bytes/1e6 > 10:
		return fmt.Sprintf("%.1fM", bytes/1e6)
	case bytes/1e3 > 10:
		return fmt.Sprintf("%.1fK", bytes/1e3)
	default:
		return fmt.Sprintf("%.0f", bytes)
	}
}
