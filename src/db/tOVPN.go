package db

type tOVPN struct {
}

func (t tOVPN) Name() string {
	return "ovpn"
}

func (t tOVPN) Latest() (ver uint, cols []string) {
	hist := t.history()
	ver = 1
	for k, _ := range hist {
		if k > ver {
			ver = k
		}
	}
	return ver, hist[ver]
}

func (t tOVPN) Version(ver uint) []string {
	return t.history()[ver]
}

func (t tOVPN) history() map[uint][]string {
	hist := map[uint][]string{
		1: []string{
			"-- OpenVPN 相关配置",
			"-- 缺少本配置表信息将导致OUM web子系统不能正常的分发客户端配置",
			"-- 缺少本配置表信息将导致OUM reconnect子命令，以及用户/设备disable/reset操作不能正确的断开正在连接的设备",
			"-- OUM将从这里记载的conffile文件内容中parse management配置，并用于disconnect设备",
			"-- OUM的web子系统将基于这里的信息分发客户端配置",
			"dev text not null default '', -- ovpn_dev",
			"name text not null default '', -- 配置名称",
			"remote text not null default '', -- 客户端连接时使用的remote参数",
			"port text not null default '', -- 客户端连接时使用的port参数, 默认从配置文件中parse",
			"conffile text not null default '', -- 服务端配置文件路径",
			"memo text not null default ''",
			"primary key(dev)",
			"unique(name)",
		},
	}
	return hist
}

func (t tOVPN) Upgrade(from uint, oldT, newT string) error {
	switch from {
	case 1:
	}
	return nil
}
