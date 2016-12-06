package db

type tDevice struct {
}

func (t tDevice) Name() string {
	return "device"
}

func (t tDevice) Latest() (ver uint, cols []string) {
	hist := t.history()
	ver = 1
	for k, _ := range hist {
		if k > ver {
			ver = k
		}
	}
	return ver, hist[ver]
}

func (t tDevice) Version(ver uint) []string {
	return t.history()[ver]
}

func (t tDevice) history() map[uint][]string {
	hist := map[uint][]string{
		1: []string{
			"-- 存储用户设备相关配置",
			"-- 登陆时使用相同的username和device会被视作同一个终端",
			"-- 验证otp通过后，15天内只要客户端IP地址不更换，该次otp code仍然有效",
			"-- 验证otp通过后，7天内只要客户端IP地址与通过时所用IP地址位于同一城市，该次otp code仍然有效",
			"username text not null default ''",
			"device text not null default '', -- 设备名",
			"otp_last_code text not null default '', -- 最近一次验证有效的otp时的code",
			"otp_last_time datetime not null default (datetime('now')), -- 最近一次验证有效的otp时的时间",
			"otp_last_ip text not null default '', -- 最近一次验证有效的otp时的客户端IP",
			"total_login integer not null default 0, -- 累计登录次数",
			"total_uptime integer not null default 0, -- 累计在线时长，单位秒",
			"total_bytes_sent integer not null default 0, -- 累计发送字节数",
			"total_bytes_received integer not null default 0, -- 累计接收字节数",
			"memo text not null default ''",
			"expired datetime not null default (datetime('now', '1000 years')), -- 格式: 2016-11-17 11:06:54 CST, 超过该时间后设备会被禁用",
			"created datetime not null default (datetime('now'))",
			"updated datetime not null default (datetime('now'))",
			"primary key(username, device)",
		},
	}
	return hist
}

func (t tDevice) Upgrade(from uint, oldT, newT string) error {
	switch from {
	case 1:
	}
	return nil
}
