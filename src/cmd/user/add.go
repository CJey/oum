package user

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"db"
	"utils"

	"github.com/cjey/slog"
)

// username:
//     name[%device]
// password:
//     password
func Add(username string, password string, disableotp, randompass bool) {
	name, device := StdUsername(username)
	if len(name) == 0 {
		slog.Emergf("Invalid username format: %s", username)
		os.Exit(1)
	}
	if len(device) > 0 {
		addDevice(name, device)
		return
	}

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
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	if flg_exists > 0 {
		slog.Warningf("User[%s] already exists", name)
		return
	}

	// otp check
	if len(password) > 0 {
		if strings.Contains(password, "%") {
			slog.Emergf("Invalid password format, should not contain %")
			os.Exit(1)
		}
		if len(password) < 6 {
			slog.Emergf("Invalid password format, too short")
			os.Exit(1)
		}
		password = HashPassword(password)
	} else if randompass {
		tmp := make([]byte, 15)
		rand.Read(tmp)
		plain := base64.StdEncoding.EncodeToString(tmp)
		password = HashPassword(plain)
		fmt.Printf("Password = %s\n", plain)
		fmt.Printf("\n------------------------------------------------\n\n")
	}
	var secret, text string
	if disableotp == false {
		host, _ := os.Hostname()
		secret, text, err = utils.OTPGenerate("OUM-"+host, name)
	}

	// insert user
	_, err = tx.Exec(`
        insert into user (username, password, secret)
        values (?, ?, ?)
    `, name, password, secret)
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}

	err = tx.Commit()
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}

	// generate otp secret url link
	if disableotp == false {
		linkgoogle := utils.QRLinkGoogle(text)
		linkcaoliao := utils.QRLinkCaoliao(text)
		fmt.Printf("[Google QR Code HTTP Link] %s\n", linkgoogle)
		fmt.Print("\n------------------------------------------------\n\n")
		fmt.Printf("[Caoliao QR Code HTTP Link] %s\n", linkcaoliao)
		fmt.Print("\n------------------------------------------------\n\n")
		fmt.Print("You can change comment,user,issuer by hack the link!\n")
		fmt.Print("otpauth://totp/{comment}:{user}?issuer={issuer}&...\n")
		fmt.Print("url decode helper: %26 &, %2F /, %3A :, %3D =, %3F ?\n\n")
	}

	slog.Infof("User[%s] created successfully", name)

	addDevice(name, "default")
}

func addDevice(name string, device string) {
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
		slog.Emergf(err.Error())
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
	if flg_exists > 0 {
		slog.Warningf("Device[%s] of user[%s] already exists", device, name)
		return
	}

	// insert device
	_, err = tx.Exec(`
        insert into device (username, device, otp_last_time)
        values (?, ?, '')
    `, name, device)
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}

	// insert ifconfig
	_, err = tx.Exec(`
        insert into ifconfig (username, device, ovpn_dev)
        values (?, ?, '')
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

	slog.Infof("Device[%s] of User[%s] created successfully", device, name)
}
