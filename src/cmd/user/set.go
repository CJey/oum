package user

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	"db"
	"utils"

	"github.com/cjey/slog"
)

var SupportConfig []string = []string{
	"allow.net",
	"allow.domain",
	"allow.city",
	"ipset.assign",
	"ipset.access",
	"otp.sameip",
	"otp.samecity",
}

var RSupportConfig map[string]string

func init() {
	RSupportConfig = make(map[string]string, len(SupportConfig))
	for _, v := range SupportConfig {
		RSupportConfig[v] = ""
	}
}

func Set(def bool, username string, configs ...string) {
	if def {
		configs = append(configs, username)
	}

	re := regexp.MustCompile(`^([a-zA-Z0-9.]+)([=|\+|-].*)$`)
	valid := map[string]string{}
	for _, config := range configs {
		matches := re.FindStringSubmatch(config)
		if len(matches) == 0 {
			slog.Emergf("Invalid config format: %s\n", config)
			os.Exit(1)
		}
		k := strings.ToLower(matches[1])
		_, ok := RSupportConfig[k]
		if !ok {
			slog.Emergf("Unsupported config %s\n", k)
			os.Exit(1)
		}
		valid[k] = matches[2]
	}

	if def {
		setDefault(valid)
	} else {
		setUser(username, valid)
	}
}

func setDefault(config map[string]string) {
	if len(config) == 0 {
		slog.Infof("Suported config: %s\n", strings.Join(SupportConfig, ", "))
		return
	}

	DB := db.Get()

	do_txt := func(k, v string) {
		_, err := DB.Exec(`
            insert or replace into config
                (key, value)
            values
                (?, ?)
        `, k, v)
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
		fmt.Printf("%s = %s\n", k, v)
	}

	do_csv := func(k, v string) {
		var before string
		err := DB.QueryRow(`
            select value from config
            where key=?
        `, k).Scan(&before)
		if err != nil && err != sql.ErrNoRows {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
		after := utils.SetOp(before, v)
		do_txt(k, after)
	}

	fmt.Printf("Default\n")
	fmt.Printf("----\n")
	for k, v := range config {
		switch k {
		case "allow.net", "allow.domain", "allow.city",
			"ipset.assign", "ipset.access":
			do_csv(k, v)
		case "otp.sameip", "otp.samecity":
			if v[0] == '=' {
				do_txt(k, v[1:])
			}
		}
	}
}

func setUser(username string, config map[string]string) {
	name, _ := StdUsername(username)
	if len(name) == 0 {
		slog.Emergf("Invalid username format: %s", username)
		os.Exit(1)
	}

	DB := db.Get()

	var fexists int
	err := DB.QueryRow(`
        select count(1) from user
        where username=?
    `, name).Scan(&fexists)
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	if fexists == 0 {
		slog.Warningf("User[%s] not exists", name)
		os.Exit(1)
	}

	if len(config) == 0 {
		slog.Infof("Suported config: %s\n", strings.Join(SupportConfig, ", "))
		return
	}

	do_txt := func(k, v string) {
		_, err := DB.Exec(fmt.Sprintf(`
            update user set "%s"=?
            where username=?
        `, k), v, name)
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
		fmt.Printf("%s = %s\n", k, v)
	}

	do_csv := func(k, v string) {
		var before string
		err := DB.QueryRow(fmt.Sprintf(`
            select "%s" from user
            where username=?
        `, k), name).Scan(&before)
		if err != nil && err != sql.ErrNoRows {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
		after := utils.SetOp(before, v)
		do_txt(k, after)
	}

	fmt.Printf("User: %s\n", name)
	fmt.Printf("----\n")
	for k, v := range config {
		switch k {
		case "allow.net", "allow.domain", "allow.city",
			"ipset.assign", "ipset.access":
			do_csv(k, v)
		case "otp.sameip", "otp.samecity":
			if v[0] == '=' {
				do_txt(k, v[1:])
			}
		}
	}
}
