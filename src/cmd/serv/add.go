package serv

import (
	"fmt"
	"os"
	"strings"

	"db"
	"utils"

	"github.com/cjey/slog"
)

func getAlias(dev string) string {
	var alias string
	for {
		if len(dev) > 0 {
			fmt.Printf("Alias name(%s): ", dev)
		} else {
			fmt.Printf("Alias name: ")
		}
		alias = utils.Readline(dev)
		if len(alias) > 0 {
			break
		}
	}
	return alias
}

func getRemote() string {
	remote := utils.PrimaryIP()
	fmt.Printf("Remote Server Host(%s): ", remote)
	return utils.Readline(remote)
}

func getPort(lines []string) string {
	port := "1194"
	find := utils.FetchConfKey(lines, "port", "")
	fs := strings.Fields(find)
	if len(fs) >= 2 {
		var p uint16
		_, err := fmt.Sscan(fs[1], &p)
		if err == nil {
			port = fs[1]
		}
	}

	for {
		fmt.Printf("Remote Server Port(%s): ", port)
		line := utils.Readline(port)
		var p uint16
		_, err := fmt.Sscan(line, &p)
		if err != nil {
			fmt.Printf("%s %s\n", errpre, err.Error())
			continue
		}
		port = line
		break
	}
	return port
}

func getProto(lines []string) string {
	def := "proto tcp-server"
	find := utils.FetchConfKey(lines, "proto", def)
	fs := strings.Fields(find)
	if len(fs) < 2 || strings.ToLower(fs[1]) != "udp" {
		return "tcp-client"
	}
	return "udp"
}

func getDev(lines []string) string {
	dev := "oum"
	find := utils.FetchConfKey(lines, "dev", "")
	fs := strings.Fields(find)
	if len(fs) >= 2 {
		dev = fs[1]
	}

	fmt.Printf("Interface name(%s): ", dev)
	return utils.Readline(dev)
}

func Add(conffile string) {
	lines, err := utils.StripConf(conffile)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	if conffile[0] != '/' {
		wd, err := os.Getwd()
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
		conffile = wd + "/" + conffile
	}

	DB := db.Get()
	var fexists int

	dev := getDev(lines)
	for {
		err = DB.QueryRow(`
            select count(1) from ovpn
            where dev=?
        `, dev).Scan(&fexists)
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
		if fexists == 0 {
			break
		}
		fmt.Printf("%s Duplicate Interface name[%s]\n", errpre, dev)
		dev = getDev(lines)
	}

	alias := getAlias(dev)
	for {
		err = DB.QueryRow(`
            select count(1) from ovpn
            where name=?
        `, alias).Scan(&fexists)
		if err != nil {
			slog.Emerg(err.Error())
			os.Exit(1)
		}
		if fexists == 0 {
			break
		}
		fmt.Printf("%s Duplicate Alias[%s]\n", errpre, alias)
		alias = getAlias(dev)
	}

	remote := getRemote()
	port := getPort(lines)
	proto := getProto(lines)

	var memo string
	if proto == "udp" {
		memo = "性能体验好，主流ISP网络下推荐使用"
	} else {
		memo = "兼容性好，非主流ISP网络下推荐使用"
	}

	_, err = DB.Exec(`
        insert into ovpn
            (dev,name,remote,port,conffile,memo)
        values
            (?,?,?,?,?,?)
    `, dev, alias, remote, port, conffile, memo)
	if err != nil {
		slog.Emerg(err.Error())
		os.Exit(1)
	}
	fmt.Printf("\n")
	fmt.Printf("Add interface[%s] successfully\n\n", dev)
	List()
}
