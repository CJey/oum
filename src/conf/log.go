package conf

import (
	"log/syslog"

	"version"

	"github.com/cjey/slog"
)

func initLog() {
	slog.SetPriorityString(LogLevel)
	slog.SetCodeRoot(version.CodeRoot())
	slog.SetLineOff(!LogLineOn)
	slog.SetLevelOff(LogLevelOff)
	slog.SetTimeOff(LogTimeOff)

	if UseSyslog {
		w, err := syslog.New(syslog.LOG_DAEMON, version.Name())
		if err == nil {
			slog.SetWriter(w)
			slog.SetTimeOff(true)
		}
	}
}
