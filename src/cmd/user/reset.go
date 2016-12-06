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
func Reset(username string, password string, disableotp, randompass bool) {
	name, device := StdUsername(username)
	if len(name) == 0 {
		slog.Emergf("Invalid username format: %s", username)
		os.Exit(1)
	}
	if len(device) > 0 {
		slog.Warning("Do not support reset device")
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
	if flg_exists == 0 {
		slog.Warningf("User[%s] not exists", name)
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
		password = hashPassword(password)
	} else if randompass {
		tmp := make([]byte, 15)
		rand.Read(tmp)
		plain := base64.StdEncoding.EncodeToString(tmp)
		password = hashPassword(plain)
		fmt.Printf("Password = %s\n", plain)
		fmt.Printf("\n------------------------------------------------\n\n")
	}
	var secret, text string
	if disableotp == false {
		host, _ := os.Hostname()
		secret, text, err = utils.OTPGenerate("OUM-"+host, name)
	}

	// reset user
	_, err = tx.Exec(`
        update user set password=?, secret=?
        where username=?
    `, password, secret, name)
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}

	// reset device
	_, err = tx.Exec(`
        update device set otp_last_code='',otp_last_time='',otp_last_ip=''
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

	// terminate all active session of openvpn
	total, effect := DisconnectDevice(name, "")
	if total > 0 {
		if effect == 0 {
			slog.Warningf("User[%s] disconnect failure, NOT/ERROR configured table: ovpn", name)
		} else {
			slog.Infof("User[%s] disconnect successfully", name)
		}
	}

	slog.Infof("User[%s] password & otp secret reset successfully", name)
}
