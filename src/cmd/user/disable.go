package user

import (
	"os"

	"db"

	"github.com/cjey/slog"
)

func Disable(username string) {
	name, device := StdUsername(username)
	if len(name) == 0 {
		slog.Emergf("Invalid username format: %s", username)
		os.Exit(1)
	}
	if len(device) == 0 {
		disableUser(name)
	} else {
		disableDevice(name, device)
	}
}

func disableUser(name string) {
	tx, err := db.Get().Begin()
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	defer tx.Rollback()

	var flg_exists int

	// user exists check
	err = tx.QueryRow(`
        select count(1) from user
        where username=?
    `, name).Scan(&flg_exists)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	if flg_exists == 0 {
		slog.Warningf("User[%s] not exists", name)
		return
	}

	// disable user
	_, err = tx.Exec(`
        update user set expired=datetime('now')
        where username=?
    `, name)
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}

	err = tx.Commit()
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}

	// terminate all active session of openvpn
	total, effect := DisconnectDevice(name, "")
	if total > 0 {
		if effect == 0 {
			slog.Warningf("User[%s] disconnect failure, NOT/ERROR configured table: ovpn", name)
		} else {
			slog.Infof("User[%s] disconnect successfully", name)
		}
	}

	slog.Infof("User[%s] disabled successfully", name)
}

func disableDevice(name, device string) {
	tx, err := db.Get().Begin()
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	defer tx.Rollback()

	var flg_exists int

	// user exists check
	err = tx.QueryRow(`
        select count(1) from user
        where username=?
    `, name).Scan(&flg_exists)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	if flg_exists == 0 {
		slog.Warningf("User[%s] not exists", name)
		return
	}

	// device exists check
	err = tx.QueryRow(`
        select count(1) from device
        where username=? and device=?
    `, name, device).Scan(&flg_exists)
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	if flg_exists == 0 {
		slog.Warningf("Device[%s] of User[%s] not exists", device, name)
		return
	}

	// disable device
	_, err = tx.Exec(`
        update device set
            expired=datetime('now')
        where
            username=? and device=?
    `, name, device)
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}

	err = tx.Commit()
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}

	// terminate associate device active session of openvpn
	total, effect := DisconnectDevice(name, device)
	if total > 0 {
		if effect == 0 {
			slog.Warningf("Device[%s] of User[%s] disconnect failure, NOT/ERROR configured table: ovpn", device, name)
		} else {
			slog.Infof("Device[%s] of User[%s] disconnect successfully", device, name)
		}
	}

	slog.Infof("Device[%s] of User[%s] disabled successfully", device, name)
}
