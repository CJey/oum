package db

type tIPCache struct {
}

func (t tIPCache) Name() string {
	return "ipcache"
}

func (t tIPCache) Latest() (ver uint, cols []string) {
	hist := t.history()
	ver = 1
	for k, _ := range hist {
		if k > ver {
			ver = k
		}
	}
	return ver, hist[ver]
}

func (t tIPCache) Version(ver uint) []string {
	return t.history()[ver]
}

func (t tIPCache) history() map[uint][]string {
	hist := map[uint][]string{
		1: []string{
			"-- ip地址归属地信息的缓存记录",
			"-- 涉及到ip归属地信息的匹配时",
			"--     1. 先查询本表，如果不存在，同步ipip(5秒超时)，否则3",
			"--     2. 如果同步成功，则使用最新数据，否则算作未知IP",
			"--     3. 如果updated时间在一天内，直接使用缓存记录，否则4",
			"--     4. 同步ipip(5秒超时)，如果同步成功，则使用最新数据，否则继续使用缓存记录",
			"ip text not null default ''",
			"country text not null default ''",
			"province text not null default ''",
			"city text not null default ''",
			"isp text not null default ''",
			"created datetime not null default (datetime('now'))",
			"updated datetime not null default (datetime('now'))",
			"primary key(ip)",
		},
	}
	return hist
}

func (t tIPCache) Upgrade(from uint, oldT, newT string) error {
	switch from {
	case 1:
	}
	return nil
}
