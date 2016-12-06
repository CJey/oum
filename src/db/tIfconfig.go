package db

type tIfconfig struct {
}

func (t tIfconfig) Name() string {
	return "ifconfig"
}

func (t tIfconfig) Latest() (ver uint, cols []string) {
	hist := t.history()
	ver = 1
	for k, _ := range hist {
		if k > ver {
			ver = k
		}
	}
	return ver, hist[ver]
}

func (t tIfconfig) Version(ver uint) []string {
	return t.history()[ver]
}

func (t tIfconfig) history() map[uint][]string {
	hist := map[uint][]string{
		1: []string{
			"-- 如果需要管理多个openvpn实例",
			"--     1. 检测是否存在与当前openvpn网卡名称相同的配置项，如果有，选中该配置，否则下一步",
			"--     2. 检测是否存在ovpn_dev为空的配置项，如果有，选中该配置，否则下一步",
			"--     3. 使用默认ifconfig-pool配置",
			"username text not null default ''",
			"device text not null default '', -- 设备名",
			"ovpn_dev text not null default '', -- Device name of openvpn",
			"ip text not null default '', -- 静态IP配置，留空表示使用dhcp pool，即跟随openvpn的ifconfig-pool配置",
			"netmask text not null default '', -- 指定网络，留空表示使用dhcp pool，即跟随openvpn的ifconfig-pool配置",
			"gateway text not null default '', -- 指定网关，留空表示使用默认网关，即openvpn local peer的IP",
			"routes text not null default '', -- 推送到客户端的路由网络，CIDR格式，多个用逗号分割(csv): push <cidr> vpn_gateway",
			"dns text not null default '', -- 指定dns，多个dns用逗号分割，留空表示使用默认dns，即openvpn local peer的IP",
			"primary key(username,device,ovpn_dev)",
		},
	}
	return hist
}

func (t tIfconfig) Upgrade(from uint, oldT, newT string) error {
	switch from {
	case 1:
	}
	return nil
}
