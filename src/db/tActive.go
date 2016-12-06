package db

type tActive struct {
}

func (t tActive) Name() string {
	return "active"
}

func (t tActive) Latest() (ver uint, cols []string) {
	hist := t.history()
	ver = 1
	for k, _ := range hist {
		if k > ver {
			ver = k
		}
	}
	return ver, hist[ver]
}

func (t tActive) Version(ver uint) []string {
	return t.history()[ver]
}

func (t tActive) history() map[uint][]string {
	hist := map[uint][]string{
		1: []string{
			"-- 当前登录的活跃设备",
			"username text not null default ''",
			"device text not null default ''",
			"cname text not null default ''",
			"ovpn_dev text not null default ''",
			"ip text not null default ''",
			"netmask text not null default ''",
			"access_ip text not null default ''",
			"access_city text not null default ''",
			"access_isp text not null default ''",
			"connect_time datetime not null default (datetime('now'))",
			"primary key(username, device)",
		},
	}
	return hist
}

func (t tActive) Upgrade(from uint, oldT, newT string) error {
	switch from {
	case 1:
	}
	return nil
}
