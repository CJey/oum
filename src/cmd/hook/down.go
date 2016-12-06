package hook

import (
	"os"

	"db"

	"github.com/cjey/slog"
)

func Down(env map[string]string) {
	dev := env["dev"]
	if len(dev) == 0 {
		slog.Emergf("Invalid request, dev not found")
		os.Exit(1)
	}
	_, err := db.Get().Exec(`
        delete from active
        where ovpn_dev=?
    `, dev)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	slog.Infof("OpenVPN[%s] shutdown", dev)
}
