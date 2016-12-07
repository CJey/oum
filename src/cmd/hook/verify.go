package hook

import (
	"database/sql"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"cmd/user"
	"db"
	"utils"

	"github.com/cjey/slog"
)

func Verify(env map[string]string, authPath string) {
	var username, password string
	if len(authPath) == 0 {
		username = env["username"]
		password = env["password"]
	} else {
		body, err := ioutil.ReadFile(authPath)
		if err != nil {
			slog.Emergf("Invalid request, can not read pass file[%s]", authPath)
			os.Exit(1)
		}
		tmp := strings.Split(string(body), "\n")
		username = strings.TrimSpace(tmp[0])
		if len(tmp) > 1 {
			password = strings.TrimSpace(tmp[1])
		}
	}

	ipdot := env["untrusted_ip"]

	ip := net.ParseIP(ipdot).To4()
	if ip == nil {
		slog.Emergf("Invalid request, invalid untrusted_ip[%s]", ipdot)
		os.Exit(1)
	}

	name, device := user.StdUsername(username)
	if len(name) == 0 {
		slog.Emergf("Invalid username format: %s", username)
		os.Exit(1)
	}
	if len(device) == 0 {
		device = "default"
	}
	show := name + "%" + device

	tx, err := db.Get().Begin()
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	defer tx.Rollback()

	// user exists/expired check
	var pass, secret, a_net, a_domain, a_city string
	var expired time.Time
	err = tx.QueryRow(`
        select password,secret,expired,allow_net,allow_domain,allow_city from user
        where username=?
    `, name).Scan(&pass, &secret, &expired, &a_net, &a_domain, &a_city)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warningf("User[%s], Account not found", name)
		} else {
			slog.Emerg(err.Error())
		}
		os.Exit(1)
	}
	if expired.Before(time.Now()) {
		slog.Warningf("User[%s], Account disabled", name)
		os.Exit(1)
	}

	// device exists/expired check
	var last_code, last_ip string
	var last_time time.Time
	err = tx.QueryRow(`
        select expired,otp_last_code,otp_last_time,otp_last_ip from device
        where username=? and device=?
    `, name, device).Scan(&expired, &last_code, &last_time, &last_ip)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warningf("Device[%s], Device not found", show)
		} else {
			slog.Emerg(err.Error())
		}
		os.Exit(1)
	}
	if expired.Before(time.Now()) {
		slog.Warningf("Device[%s], Device disabled", show)
		os.Exit(1)
	}

	err = tx.Commit()
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}

	// access scope check
	if len(a_net)+len(a_domain)+len(a_city) != 0 {
		if inet, ok := allowNet(ip, a_net); !ok {
			if domain, ok := allowDomain(ip, a_domain); !ok {
				if city, ok := allowCity(ip, a_city); !ok {
					slog.Warningf("User[%s], Access scope denied", name)
					os.Exit(1)
				} else {
					slog.Infof("User[%s], Access allow city[%s]", name, city)
				}
			} else {
				slog.Infof("User[%s], Access allow domain[%s]", name, domain)
			}
		} else {
			slog.Infof("User[%s], Access allow net[%s]", name, inet)
		}
	}

	// password check
	if len(pass) == 0 && len(secret) == 0 {
		slog.Warningf("User[%s], Login disabled", name)
		os.Exit(1)
	}
	code, pwd := user.StdPassword(password)
	if len(pass) > 0 {
		if pass != pwd {
			slog.Warningf("User[%s], Password dismatch", name)
			os.Exit(1)
		}
	}
	if len(secret) == 0 {
		slog.Infof("Device[%s], Verify successfully by Single Password", show)
		return
	}

	// OTP check
	ok, err := utils.OTPValidate(code, secret)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	if ok {
		_, err = db.Get().Exec(`
            update device set
                otp_last_code=?,
                otp_last_time=datetime('now'),
                otp_last_ip=?
            where username=? and device=?
        `, code, ipdot, name, device)
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}

		slog.Infof("Device[%s], Verify successfully by OTP Code", show)
		return
	}

	// last_code verify
	if last_code != code {
		slog.Warningf("Device[%s], OTP Code dismatch", show)
		os.Exit(1)
	}

	sameip := time.Hour * 24 * 15  // 15days
	samecity := time.Hour * 24 * 7 // 7days
	if sec := env["oum_sameip"]; len(sec) > 0 {
		i, err := strconv.ParseUint(sec, 0, 64)
		if err == nil {
			sameip = time.Duration(i) * time.Second
		}
	}
	if sec := env["oum_samecity"]; len(sec) > 0 {
		i, err := strconv.ParseUint(sec, 0, 64)
		if err == nil {
			samecity = time.Duration(i) * time.Second
		}
	}

	now := time.Now()
	switch {
	case last_ip == ipdot:
		// same ip
		if last_time.Add(sameip).Before(now) {
			slog.Warningf("Device[%s], Last OTP Code failure, expired", show)
			os.Exit(1)
		}
		slog.Infof("Device[%s], Login with same ip[%s]", show, ipdot)
	case ReservedIPv4(ip):
		// reserved ip
		if last_time.Add(sameip).Before(now) {
			slog.Warningf("Device[%s], Last OTP Code failure, expired", show)
			os.Exit(1)
		}
		slog.Infof("Device[%s], Login with reserved ip[%s]", show, ipdot)
	default:
		ip2 := net.ParseIP(last_ip).To4()
		if ip2 == nil {
			slog.Warningf("Device[%s], Last OTP Code failure, invalid otp_last_ip[%s]", show, last_ip)
			os.Exit(1)
		}
		if ReservedIPv4(ip2) {
			slog.Warningf("Device[%s], Last OTP Code failure, cross net", show)
			os.Exit(1)
		}
		last, err := NewIP(ip2)
		if err != nil {
			slog.Warningf("Device[%s], Last OTP Code failure, parse last login ip info failure, %s", show, err.Error())
			os.Exit(1)
		}
		// same city
		iplogin, err := NewIP(ip)
		if err != nil {
			slog.Warningf("Device[%s], Last OTP Code failure, parse login ip info failure, %s", show, err.Error())
			os.Exit(1)
		}
		if last.SameCity(iplogin) {
			if last_time.Add(samecity).Before(now) {
				slog.Warningf("Device[%s], Last OTP Code failure, expired", show)
				os.Exit(1)
			}
			slog.Infof("Device[%s], Login at same city(%s), last %s, now %s", show, last.City(), last_ip, ipdot)
		} else {
			slog.Warningf("Device[%s], Last OTP Code failure, cross city", show)
			os.Exit(1)
		}
	}
	slog.Infof("Device[%s], Verify successfully by Last OTP Code", show)
}

func allowNet(ip net.IP, inet string) (string, bool) {
	ipdot := ip.String()
	inets := utils.CSVSet(inet)
	for _, inet := range inets {
		if strings.Index(inet, "/") < 0 {
			if ipdot == inet {
				return inet, true
			}
		} else {
			_, ipnet, err := net.ParseCIDR(inet)
			if err == nil && ipnet.Contains(ip) {
				return inet, true
			}
		}
	}
	return "", false
}

func allowDomain(ip net.IP, domain string) (string, bool) {
	ipdot := ip.String()
	domains := utils.CSVSet(domain)
	for _, domain := range domains {
		addrs, err := net.LookupHost(domain)
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if addr == ipdot {
				return domain, true
			}
		}
	}
	return "", false
}

func allowCity(ip net.IP, city string) (string, bool) {
	iip, err := NewIP(ip)
	if err != nil {
		return "", false
	}
	country := iip.Country()
	province := iip.Province()
	ct := iip.City()

	cities := utils.CSVSet(city)
	for _, city := range cities {
		switch city {
		case country:
			return country, true
		case province:
			return province, true
		case ct:
			return ct, true
		}
	}
	return "", false
}
