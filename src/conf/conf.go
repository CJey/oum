package conf

import (
	"os"

	"version"

	"github.com/alecthomas/kingpin"
)

var showVersion bool

var (
	UseSyslog   bool
	LogLevel    string
	LogLineOn   bool
	LogLevelOff bool
	LogTimeOff  bool
)

var (
	DBFilePath  string
	FullCommand string
)

var Hook struct {
	FullCommand string

	Args []string
}

var Add struct {
	FullCommand string

	Username string
	Password string

	DisableOTP bool
	RandomPass bool
}

var List struct {
	FullCommand string

	Usernames []string
}

var Reset struct {
	FullCommand string

	Username string
	Password string

	DisableOTP bool
	RandomPass bool
}

var Enable struct {
	FullCommand string

	Username string
}

var Disable struct {
	FullCommand string

	Username string
}

var Delete struct {
	FullCommand string

	Username string
}

var Reconnect struct {
	FullCommand string

	Username string
}

var Verify struct {
	FullCommand string

	Username string
	Password string
}

var Guide struct {
	FullCommand string
}

var Table struct {
	FullCommand string

	Names []string
}

var Pattern struct {
	FullCommand string

	Quick bool
	ECDSA bool
	From  string
	Dev   string
	Out   string
}

var Web struct {
	FullCommand string

	Cert    string
	CertKey string
	Port    uint16
	HTTPS   bool
	Root    string
	Restore string
}

var Iptables struct {
	FullCommand string

	Out string
}

var Ipset struct {
	FullCommand string

	Interval uint
	Sets     []string
}

var ServAdd struct {
	FullCommand string

	Conffile string
}

var ServDelete struct {
	FullCommand string

	Dev string
}

var ServList struct {
	FullCommand string
}

func init() {
	Hook.Args = []string{}
	List.Usernames = []string{}
	Table.Names = []string{}
	Ipset.Sets = []string{}
}

func parseFlags() {
	kingpin.Command("version", "Show Version Info")
	kingpin.Flag("syslog", "Output redirect to syslog").
		Default("false").BoolVar(&UseSyslog)
	kingpin.Flag("log-level", "log level").
		Default("info").
		EnumVar(&LogLevel, "emerg", "alert", "crit", "err", "warning", "notice", "info", "debug")
	kingpin.Flag("log-lineon", "Hide the code line of log").
		Default("false").BoolVar(&LogLineOn)
	kingpin.Flag("log-leveloff", "Hide the log level hint string").
		Default("false").BoolVar(&LogLevelOff)
	kingpin.Flag("log-timeoff", "Hide the time of log").
		Default("false").BoolVar(&LogTimeOff)
	kingpin.Flag("db", "sqlite3 db file path").
		Default("/var/lib/oum/oum.db").StringVar(&DBFilePath)

	// openvpn hook
	cmdHook := kingpin.Command("hook", "Working as openvpn hook scripts")
	Hook.FullCommand = cmdHook.FullCommand()
	cmdHook.Arg("args", "Args from openvpn").
		StringsVar(&Hook.Args)

	// list
	cmdList := kingpin.Command("list", "List users/devices info")
	List.FullCommand = cmdList.FullCommand()
	cmdList.Arg("usernames", "which users do you want to show, default show all").
		StringsVar(&List.Usernames)

		// add
	cmdAdd := kingpin.Command("add", "Add user/device")
	Add.FullCommand = cmdAdd.FullCommand()
	cmdAdd.Flag("disable-otp", "do not generate otp secret").BoolVar(&Add.DisableOTP)
	cmdAdd.Flag("random-pass", "generate randm password").BoolVar(&Add.RandomPass)
	cmdAdd.Arg("username", "which username do you want to add, if username exists, then support username%device, it means add new device of username").Required().
		StringVar(&Add.Username)
	cmdAdd.Arg("password", "which password do you want to set, when add new device, password will be ignored").
		StringVar(&Add.Password)

		// reset
	cmdReset := kingpin.Command("reset", "Reset user password")
	Reset.FullCommand = cmdReset.FullCommand()
	cmdReset.Flag("disable-otp", "do not generate otp secret").BoolVar(&Reset.DisableOTP)
	cmdReset.Flag("random-pass", "generate randm password").BoolVar(&Reset.RandomPass)
	cmdReset.Arg("username", "which user do you want to reset").Required().
		StringVar(&Reset.Username)
	cmdReset.Arg("password", "which password do you want to reset").
		StringVar(&Reset.Password)

		// enable
	cmdEnable := kingpin.Command("enable", "Enable user/device")
	Enable.FullCommand = cmdEnable.FullCommand()
	cmdEnable.Arg("username", "which user/device do you want to enable").Required().
		StringVar(&Enable.Username)

		// disable
	cmdDisable := kingpin.Command("disable", "Disable user/device")
	Disable.FullCommand = cmdDisable.FullCommand()
	cmdDisable.Arg("username", "which user/device do you want to disable").Required().
		StringVar(&Disable.Username)

		// delete
	cmdDelete := kingpin.Command("del", "Delete user/device")
	Delete.FullCommand = cmdDelete.FullCommand()
	cmdDelete.Arg("username", "which user/device do you want to delete").Required().
		StringVar(&Delete.Username)

		// reconnect
	cmdReconnect := kingpin.Command("reconnect", "Reconnect user/device")
	Reconnect.FullCommand = cmdReconnect.FullCommand()
	cmdReconnect.Arg("username", "which user/device do you want to reconnect").Required().
		StringVar(&Reconnect.Username)

		// verify
	cmdVerify := kingpin.Command("verify", "Verify user")
	Verify.FullCommand = cmdVerify.FullCommand()
	cmdVerify.Arg("username", "which user/device do you want to verify").Required().
		StringVar(&Verify.Username)
	cmdVerify.Arg("password", "which password do you want to verify").Required().
		StringVar(&Verify.Password)

		// guide
	cmdGuide := kingpin.Command("guide", "Helper: Show general operation guide")
	Guide.FullCommand = cmdGuide.FullCommand()

	// table
	cmdTable := kingpin.Command("table", "Helper: Show sqlite3 table definition, you can hack db with utils sqlite3 by your hand")
	Table.FullCommand = cmdTable.FullCommand()
	cmdTable.Arg("names", "which tables do you want to see").StringsVar(&Table.Names)

	// pattern
	cmdPattern := kingpin.Command("pattern", "Helper, depends openssl: Show pattern openvpn server configuration file")
	Pattern.FullCommand = cmdPattern.FullCommand()
	cmdPattern.Flag("quick", "use all default").
		BoolVar(&Pattern.Quick)
	cmdPattern.Flag("ecdsa", "generate ecdsa key paire, instead of rsa").
		BoolVar(&Pattern.ECDSA)
	cmdPattern.Flag("from", "generate client conf extract from this server conffile").
		StringVar(&Pattern.From)
	cmdPattern.Flag("dev", "generate client conf extract from this dev in the table: ovpn").
		StringVar(&Pattern.Dev)
	cmdPattern.Flag("out", "output pattern conf to the file, use - print to stdout").
		Required().
		StringVar(&Pattern.Out)

	// iptables
	cmdIptables := kingpin.Command("iptables", "Helper: Output iptables pattern")
	Iptables.FullCommand = cmdIptables.FullCommand()
	cmdIptables.Flag("out", "output generated shell script to the file, use - print to stdout").
		Required().
		StringVar(&Iptables.Out)

	// ipset
	cmdIpset := kingpin.Command("ipset", "Helper: resolve dns to set ipset")
	Ipset.FullCommand = cmdIpset.FullCommand()
	cmdIpset.Flag("interval", "default 0, run just one time, unit: second").
		UintVar(&Ipset.Interval)
	cmdIpset.Arg("sets", "format: <set name>{=|+}domain[,domain...] ...").
		Required().
		StringsVar(&Ipset.Sets)

		// web
	cmdWeb := kingpin.Command("web", "Helper: Run http service to serve user to download config file, please config table: ovpn first")
	Web.FullCommand = cmdWeb.FullCommand()
	cmdWeb.Flag("port", "http bind port").
		Default("1195").Uint16Var(&Web.Port)
	cmdWeb.Flag("https", "serve as https").
		Default("false").BoolVar(&Web.HTTPS)
	cmdWeb.Flag("cert", "tls cert file path, default oum will auto generate, and put it to /var/lib/oum/web.crt").
		StringVar(&Web.Cert)
	cmdWeb.Flag("certkey", "tls cert key file path, default oum will auto generate, and put it to /var/lib/oum/web.key").
		StringVar(&Web.CertKey)
	cmdWeb.Flag("root", "http server root path, use this flag just for debug").
		StringVar(&Web.Root)
	cmdWeb.Flag("restore", "restore bundled static resources into the dir").
		StringVar(&Web.Restore)

		// serv
	cmdServ := kingpin.Command("serv", "Config which openvpn server file that oum will serve")

	// serv add
	cmdServAdd := cmdServ.Command("add", "Add new server configuration file to serve")
	ServAdd.FullCommand = cmdServAdd.FullCommand()
	cmdServAdd.Arg("conffile", "server configuration file path").
		Required().
		StringVar(&ServAdd.Conffile)

	// serv delete
	cmdServDelete := cmdServ.Command("del", "Delete a specified server configuration file to serve")
	ServDelete.FullCommand = cmdServDelete.FullCommand()
	cmdServDelete.Arg("dev", "interface name").
		Required().
		StringVar(&ServDelete.Dev)

	// serv list
	cmdServList := cmdServ.Command("list", "List all server configuration file that oum serving now").Default()
	ServList.FullCommand = cmdServList.FullCommand()

	FullCommand = kingpin.Parse()
}

func prepareFlags() {
	initLog()
}

func init() {
	parseFlags()

	if FullCommand == "version" {
		print(version.Show())
		os.Exit(0)
	}

	prepareFlags()
}
