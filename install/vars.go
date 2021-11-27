package install

import (
	"github.com/kuaifan/sdos/pkg/sshcmd/sshutil"
	"sync"
	"time"
)

type Wireguard struct {
	Name         string
	Public       string
	Received     int64
	ReceivedDiff int64
	Sent         int64
	SentDiff     int64
}

type Monitor struct {
	State string
	Sn    string
	Ping  float64
	Unix  int64
}

type NetIoNic struct {
	T    time.Time
	Up   uint64
	Down uint64
	Sent uint64
	Recv uint64
}

type State struct {
	T   time.Time
	Cpu float64
	Mem struct {
		Current uint64
		Total   uint64
	}
	Swap struct {
		Current uint64
		Total   uint64
	}
	Disk struct {
		Current uint64
		Total   uint64
	}
	Uptime   uint64
	Loads    []float64
	TcpCount int
	UdpCount int
	NetIO    struct {
		Up   uint64
		Down uint64
	}
	NetTraffic struct {
		Sent uint64
		Recv uint64
	}
}

type Firewall struct {
	Mode     string
	Keys     string
}

type FirewallRule struct {
	Mode     string
	Ports    string
	Type     string
	Address  string
	Protocol string
	Key      string
}

type FirewallForward struct {
	Mode     string
	Sport    string
	Dip      string
	Dport    string
	Protocol string
	Key      string
}

var (
	NodeIPs      []string
	ManageImage  string
	ServerDomain string
	ServerKey    string
	ServerCrt    string
	ServerUrl    string
	ReportUrl    string
	SwapFile     string
	ServerToken  string
	Mtu          string
	InReset      bool

	SSHConfig sshutil.SSH

	FileMd5 sync.Map

	ResultInstall sync.Map
	ResultRemove  sync.Map

	FirewallConfig Firewall
	FirewallRuleConfig FirewallRule
	FirewallForwardConfig FirewallForward
)
