package user

import (
	"database/sql"
	"os"
	"time"

	"db"

	"github.com/cjey/slog"
)

func Enable(username string) {
	name, device := StdUsername(username)
	if len(name) == 0 {
		slog.Emergf("Invalid username format: %s", username)
		os.Exit(1)
	}
	if len(device) == 0 {
		enableUser(name)
	} else {
		enableDevice(name, device)
	}
}

func enableUser(name string) {
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

	// enable user
	_, err = tx.Exec(`
        update user set expired=datetime('now', '1000 years')
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

	slog.Infof("User[%s] enabled successfully", name)
}

func enableDevice(name, device string) {
	tx, err := db.Get().Begin()
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	defer tx.Rollback()

	var expired time.Time

	// user exists/expired check
	err = tx.QueryRow(`
        select expired from user
        where username=?
    `, name).Scan(&expired)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warningf("User[%s] not exists", name)
			return
		}
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	if expired.Before(time.Now()) {
		slog.Warningf("User[%s] already disabled, please enable User[%s] first", name, name)
		return
	}

	var flg_exists int

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

	// enable device
	_, err = tx.Exec(`
        update device set expired=datetime('now', '1000 years')
        where username=? and device=?
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

	slog.Infof("Device[%s] of User[%s] enabled successfully", device, name)
}
