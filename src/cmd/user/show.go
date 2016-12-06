package user

import (
	"database/sql"
	"fmt"
	"os"

	"db"

	"github.com/cjey/slog"
)

func Show(username string) {
	name, _ := StdUsername(username)
	if len(name) == 0 {
		slog.Emergf("Invalid username format: %s", username)
		os.Exit(1)
	}

	DB := db.Get()

	var a_net, a_domain, a_city string
	var assign, access string
	err := DB.QueryRow(`
        select allow_net,allow_domain,allow_city,ipset_assign,ipset_access from user
        where username=?
    `, name).Scan(&a_net, &a_domain, &a_city, &assign, &access)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warningf("User[%s] not exists", name)
		} else {
			slog.Emerg(err.Error())
		}
		os.Exit(1)
	}

	fmt.Printf("User: %s\n", name)
	fmt.Printf("----\n")
	fmt.Printf("allow-net: %s\n", a_net)
	fmt.Printf("allow-domain: %s\n", a_domain)
	fmt.Printf("allow-city: %s\n", a_city)
	fmt.Printf("ipset-assign: %s\n", assign)
	fmt.Printf("ipset-access: %s\n", access)
}
