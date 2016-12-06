package serv

import (
	"fmt"
	"os"

	"db"

	"github.com/cjey/slog"
)

func List() {
	DB := db.Get()
	rows, err := DB.Query(`
        select dev,name,remote,port,conffile from ovpn
    `)
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
	var i int
	err = db.RangeRows(rows, func() error {
		var dev, name, remote, port, conffile string
		err = rows.Scan(&dev, &name, &remote, &port, &conffile)
		if err != nil {
			return err
		}
		i++
		fmt.Printf("%d. %s(%s) %s:%s %s\n", i, dev, name, remote, port, conffile)
		return nil
	})
	if err != nil {
		slog.Emergf(err.Error())
		os.Exit(1)
	}
}
