package serv

import (
	"fmt"
	"os"

	"db"

	"github.com/cjey/slog"
)

func Delete(dev string) {
	DB := db.Get()
	var fexists int

	err := DB.QueryRow(`
        select count(1) from ovpn
        where dev=?
    `, dev).Scan(&fexists)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	if fexists == 0 {
		fmt.Printf("[ERROR] Interface[%s] not found\n", dev)
		os.Exit(1)
	}
	_, err = DB.Exec(`
        delete from ovpn
        where dev=?
    `, dev)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	fmt.Printf("Delete interface[%s] successfully\n\n", dev)
	List()
}
