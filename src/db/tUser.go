package db

type tUser struct {
}

func (t tUser) Name() string {
	return "user"
}

func (t tUser) Latest() (ver uint, cols []string) {
	hist := t.history()
	ver = 1
	for k, _ := range hist {
		if k > ver {
			ver = k
		}
	}
	return ver, hist[ver]
}

func (t tUser) Version(ver uint) []string {
	return t.history()[ver]
}

func (t tUser) history() map[uint][]string {
	hist := map[uint][]string{
		1: []string{
			"-- 存储用户信息",
			"-- 用户身份验证支持：仅password，仅OTP，password+OTP，也就是说passowrd和OTP都为空表示禁止登录",
			"-- 登陆时支持的username格式为： [!]username[%device]，开头!代表仅连通网络，不要重定向客户端的网关/DNS信息，相同的username[%device]会被视为同一台设备，同一设备同一时间仅允许一份登录实例，后登陆的会导致前登陆的终端下线",
			"-- 省略device即表示设备号为default的设备",
			"-- 新建用户时会自动为该用户创建default设备",
			"-- 登陆时支持的password格式为：[otp code][%password]",
			"-- allow_net, allow_domain, allow_city均为空表示不做IP限制，只要其中一项非空表示开启白名单模式，三个条件任意一项match则允许登录",
			"username text not null default ''",
			"password text not null default '', -- <salt>:<sha256sum>，hmac(password, salt), 单向Hash存储密码，密码为空则表示仅使用OTP进行身份验证",
			"secret text not null default '', -- OTP secret，secret为空则表示仅使用password进行身份验证",
			"\"allow.net\" text not null default '', -- IP白名单，IP支持CIDR格式，多个net/ip用逗号分割(csv)",
			"\"allow.domain\" text not null default '', -- 域名白名单，域名会先被解析成IP, 多个域名用逗号分割(csv)",
			"\"allow.city\" text not null default '', -- 城市白名单，匹配ipip.net的地址解析结果(ref: 免费API http://www.ipip.net/api.html)，格式为(不要附加行政后缀): 国家名[/省名[/城市名]]，例: 中国，中国/江苏，中国/江苏/南京，中国/上海/上海; 多个城市用逗号分割(csv)",
			"\"ipset.assign\" text not null default '', -- ipset名称，多个用逗号分割(csv); 用户设备完成ip配置后，将分配给设备的ip地址追加到所有的set中; set如果已经存在，则应当兼容hash:ip类型，如果不存在，oum将以此名称自动创建类型为hash:ip的set",
			"\"ipset.access\" text not null default '', -- 格式和约束参见ipset.access; 用户设备完成ip配置后，将该设备的远程介入ip地址追加到所有的set中",
			"\"otp.sameip\" text not null default ''",
			"\"otp.samecity\" text not null default ''",
			"total_login integer not null default 0, -- 累计登录次数",
			"total_uptime integer not null default 0, -- 累计在线时长，单位秒",
			"total_bytes_sent integer not null default 0, -- 累计发送字节数",
			"total_bytes_received integer not null default 0, -- 累计接收字节数",
			"memo text not null default ''",
			"expired datetime not null default (datetime('now', '1000 years')), -- 格式: 2016-11-17 11:06:54 CST, 超过该时间后用户会被禁用",
			"created datetime not null default (datetime('now'))",
			"updated datetime not null default (datetime('now'))",
			"primary key(username)",
		},
	}
	return hist
}

func (t tUser) Upgrade(from uint, oldT, newT string) error {
	switch from {
	case 1:
	}
	return nil
}
