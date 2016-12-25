package main

import (
	"os"
	"strings"

	"cmd/hook"
	"cmd/serv"
	"cmd/user"
	"conf"
	"db"
	"github.com/cjey/slog"
)

func main() {
	switch conf.FullCommand {
	case conf.Guide.FullCommand:
		guide()
		return
	case conf.Table.FullCommand:
		db.ShowCreateTable(conf.Table.Names...)
		return
	case conf.Iptables.FullCommand:
		iptables(conf.Iptables.Out)
		return
	case conf.Ipset.FullCommand:
		ipset(conf.Ipset.Interval, conf.Ipset.Sets...)
		return
	case conf.Pattern.FullCommand:
		switch {
		case len(conf.Pattern.Dev) > 0:
			err := db.Init(conf.DBFilePath)
			if err != nil {
				slog.Emergf("Open sqlite db file failure, %s", err.Error())
				os.Exit(1)
			}
			showPatternDev(conf.Pattern.Dev, conf.Pattern.Out)
		case len(conf.Pattern.From) > 0:
			showPatternFrom(conf.Pattern.From, conf.Pattern.Out, conf.Pattern.Quick)
		default:
			showPattern(conf.Pattern.Out, conf.Pattern.ECDSA, conf.Pattern.Quick)
		}
		return
	}

	err := db.Init(conf.DBFilePath)
	if err != nil {
		slog.Emergf("Open sqlite db file failure, %s", err.Error())
		os.Exit(1)
	}

	switch conf.FullCommand {
	case conf.Hook.FullCommand:
		subHook()
	case conf.List.FullCommand:
		user.List(conf.List.Usernames...)
	case conf.Add.FullCommand:
		user.Add(
			conf.Add.Username,
			conf.Add.Password,
			conf.Add.DisableOTP,
			conf.Add.RandomPass,
		)
	case conf.Reset.FullCommand:
		user.Reset(
			conf.Reset.Username,
			conf.Reset.Password,
			conf.Reset.DisableOTP,
			conf.Reset.RandomPass,
		)
	case conf.Set.FullCommand:
		user.Set(conf.Set.Def, conf.Set.Username, conf.Set.Config...)
	case conf.Ifconfig.FullCommand:
		user.Ifconfig(conf.Ifconfig.Username, conf.Ifconfig.Dev, conf.Ifconfig.Config...)
	case conf.Show.FullCommand:
		user.Show(conf.Show.Username)
	case conf.Enable.FullCommand:
		user.Enable(conf.Enable.Username)
	case conf.Disable.FullCommand:
		user.Disable(conf.Disable.Username)
	case conf.Reconnect.FullCommand:
		user.Reconnect(conf.Reconnect.Username)
	case conf.Delete.FullCommand:
		user.Delete(conf.Delete.Username)
	case conf.Verify.FullCommand:
		user.Verify(conf.Verify.Username, conf.Verify.Password)
	case conf.Web.FullCommand:
		web()
	case conf.ServAdd.FullCommand:
		serv.Add(conf.ServAdd.Conffile)
	case conf.ServDelete.FullCommand:
		serv.Delete(conf.ServDelete.Dev)
	case conf.ServList.FullCommand:
		serv.List()
	}
}

func subHook() {
	env := map[string]string{}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		env[pair[0]] = pair[1]
	}
	switch env["script_type"] {
	case "client-connect":
		if len(conf.Hook.Args) == 0 {
			slog.Emerg("should be called by openvpn hook directive")
			os.Exit(1)
		}
		dynpath := conf.Hook.Args[len(conf.Hook.Args)-1]
		hook.Connect(env, dynpath)
	case "client-disconnect":
		hook.Disconnect(env)
	case "up":
		hook.Up(env)
	case "down":
		hook.Down(env)
	case "user-pass-verify":
		if len(conf.Hook.Args) == 0 {
			hook.Verify(env, "")
		} else {
			hook.Verify(env, conf.Hook.Args[0])
		}
	default:
		slog.Emerg("should be called by openvpn hook directive")
		os.Exit(1)
	}
}
