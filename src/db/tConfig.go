package db

type tConfig struct {
}

func (t tConfig) Name() string {
	return "config"
}

func (t tConfig) Latest() (ver uint, cols []string) {
	hist := t.history()
	ver = 1
	for k, _ := range hist {
		if k > ver {
			ver = k
		}
	}
	return ver, hist[ver]
}

func (t tConfig) Version(ver uint) []string {
	return t.history()[ver]
}

func (t tConfig) history() map[uint][]string {
	hist := map[uint][]string{
		1: []string{
			"-- 全局配置",
			"-- support keys:",
			"-- allow-net, all-domain, all-city",
			"-- ipset-assign, ipset-access",
			"-- otp-sameip, otp-samecity",
			"key text not null default ''",
			"value text not null default ''",
			"primary key(key)",
		},
	}
	return hist
}

func (t tConfig) Upgrade(from uint, oldT, newT string) error {
	switch from {
	case 1:
	}
	return nil
}
