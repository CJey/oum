package db

type tLog struct {
}

func (t tLog) Name() string {
	return "log"
}

func (t tLog) Latest() (ver uint, cols []string) {
	hist := t.history()
	ver = 1
	for k, _ := range hist {
		if k > ver {
			ver = k
		}
	}
	return ver, hist[ver]
}

func (t tLog) Version(ver uint) []string {
	return t.history()[ver]
}

func (t tLog) history() map[uint][]string {
	hist := map[uint][]string{
		1: []string{
			"-- 用户登录历史记录",
			"id integer auto_increment",
			"username text not null default ''",
			"device text not null default ''",
			"cname text not null default ''",
			"ovpn_dev text not null default ''",
			"ip text not null default ''",
			"netmask text not null default ''",
			"access_ip text not null default ''",
			"access_port text not null default ''",
			"access_city text not null default ''",
			"access_isp text not null default ''",
			"connect_time datetime not null default (datetime('now'))",
			"disconnect_time datetime not null default (datetime('now', '-1000 years'))",
			"bytes_sent integer, -- 单位Byte",
			"bytes_received integer -- 单位Byte",
		},
	}
	return hist
}

func (t tLog) Upgrade(from uint, oldT, newT string) error {
	switch from {
	case 1:
	}
	return nil
}
