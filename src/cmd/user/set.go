package user

import (
	"database/sql"
	"os"

	"db"
	"utils"

	"github.com/cjey/slog"
)

func Set(username string, config map[string]string) {
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

	a_net = utils.SetOp(a_net, config["allow-net"])
	a_domain = utils.SetOp(a_domain, config["allow-domain"])
	a_city = utils.SetOp(a_city, config["allow-city"])
	assign = utils.SetOp(assign, config["ipset-assign"])
	access = utils.SetOp(access, config["ipset-access"])

	_, err = DB.Exec(`
        update user set
            allow_net=?,allow_domain=?,allow_city=?,
            ipset_assign=?,ipset_access=?
        where username=?
    `, a_net, a_domain, a_city, assign, access, name)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
}
