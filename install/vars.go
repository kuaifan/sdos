package install

import (
	"github.com/kuaifan/sdos/pkg/sshcmd/sshutil"
	"sync"
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
	Ping  float64
	Issue int64
}

var (
	NodeIPs     []string
	ManageImage string
	ServerUrl   string
	ReportUrl   string
	ServerToken string
	Mtu         string
	InReset     bool

	SSHConfig sshutil.SSH

	FileMd5 sync.Map

	ResultInstall sync.Map
	ResultRemove  sync.Map

	NetInterface  string
	NetCount      uint
	NetUpdateTime float64
)
