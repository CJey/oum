package user

import (
	"database/sql"
	"os"
	"time"

	"db"
	"utils"

	"github.com/cjey/slog"
)

// username:
//     name[%device]
// password:
//     [code][%password]
func Verify(username string, password string) {
	name, device := StdUsername(username)
	if len(name) == 0 {
		slog.Emergf("Invalid username format: %s", username)
		os.Exit(1)
	}
	if len(device) == 0 {
		device = "default"
	}

	tx, err := db.Get().Begin()
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	defer tx.Rollback()

	// user exists/expired check
	var pass, secret string
	var expired time.Time
	err = tx.QueryRow(`
        select password,secret,expired from user
        where username=?
    `, name).Scan(&pass, &secret, &expired)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warningf("User[%s] not exists", name)
		} else {
			slog.Emerg(err.Error())
		}
		os.Exit(1)
	}
	if expired.Before(time.Now()) {
		slog.Warningf("User[%s] disabled", name)
		os.Exit(1)
	}

	// device exists/expired check
	err = tx.QueryRow(`
        select expired from device
        where username=? and device=?
    `, name, device).Scan(&expired)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warningf("Device[%s] of User[%s] not exists", device, name)
		} else {
			slog.Emerg(err.Error())
		}
		os.Exit(1)
	}
	if expired.Before(time.Now()) {
		slog.Warningf("Device [%s] of User[%s] disabled", device, name)
		os.Exit(1)
	}

	err = tx.Commit()
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}

	// password check
	if len(pass) == 0 && len(secret) == 0 {
		slog.Warningf("User[%s] disable login", name)
		os.Exit(1)
	}
	code, pwd := StdPassword(password)
	if len(pass) > 0 {
		if !MatchPassword(pass, pwd) {
			slog.Warningf("Password dismatch")
			os.Exit(1)
		}
	}
	if len(secret) == 0 {
		slog.Infof("Verify successfully by Single Password")
		return
	}

	// OTP check
	ok, err := utils.OTPValidate(code, secret)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	if !ok {
		slog.Warningf("OTP code dismatch")
		os.Exit(1)
	}

	slog.Infof("Verify successfully by OTP Code")
}
