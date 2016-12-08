package user

import (
	"database/sql"
	"fmt"
	"os"

	"db"

	"github.com/cjey/slog"
)

func Show(username string) {
	if len(username) > 0 {
		showUser(username)
	} else {
		showDefault()
	}
}

func showDefault() {
	conf := GetDefaultConfig()
	fmt.Printf("Default\n")
	fmt.Printf("----\n")
	for _, v := range SupportConfig {
		fmt.Printf("%s = %s\n", v, conf[v])
	}
}

func showUser(username string) {
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

	def := GetDefaultConfig()
	conf := GetUserConfig(name)

	fmt.Printf("User: %s\n", name)
	fmt.Printf("----\n")
	for _, k := range SupportConfig {
		if len(conf[k]) == 0 && len(def[k]) > 0 {
			fmt.Printf("*%s = %s\n", k, def[k])
		} else {
			fmt.Printf("%s = %s\n", k, conf[k])
		}
	}
}

func GetUserConfig(name string) map[string]string {
	DB := db.Get()
	res := make(map[string]string, len(SupportConfig))
	for _, k := range SupportConfig {
		var v string
		err := DB.QueryRow(fmt.Sprintf(`
            select "%s" from user
            where username=?
        `, k), name).Scan(&v)
		if err != nil && err != sql.ErrNoRows {
			slog.Emergf(err.Error())
			os.Exit(1)
		}
		res[k] = v
	}
	return res
}

func GetDefaultConfig() map[string]string {
	DB := db.Get()
	rows, err := DB.Query(`
        select key, value from config
    `)
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	res := make(map[string]string, len(SupportConfig))
	err = db.RangeRows(rows, func() error {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return err
		}
		_, ok := RSupportConfig[key]
		if ok {
			res[key] = value
		}
		return nil
	})
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	return res
}

func GetFinalConfig(name string) map[string]string {
	def := GetDefaultConfig()
	user := GetUserConfig(name)
	for k, _ := range def {
		if len(user[k]) > 0 {
			def[k] = user[k]
		}
	}
	return def
}
