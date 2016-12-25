package hook

import (
	"os"
	"os/exec"

	"db"
	"fmt"
	"utils"

	"github.com/cjey/slog"
)

func userIPset(name string) (sets_as, sets_ac []string) {
	var assign, access string
	err := db.Get().QueryRow(`
        select "ipset.assign","ipset.access" from user
        where username=?
    `, name).Scan(&assign, &access)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	sets_as = utils.CSVSet(assign)
	sets_ac = utils.CSVSet(access)

	var shell string
	for _, set := range sets_as {
		shell += fmt.Sprintf("ipset create %s hash:ip\n", set)
	}
	for _, set := range sets_ac {
		shell += fmt.Sprintf("ipset create %s hash:ip\n", set)
	}
	if len(shell) > 0 {
		exec.Command("/bin/sh", "-c", shell).Run()
	}
	return
}
