package install

import (
	"github.com/kuaifan/sdos/pkg/sshcmd/sshutil"
	"sync"
)

type Transfer struct {
	Name           string
	Public         string
	Received       int64
	ReceivedDiff   int64
	Sent           int64
	SentDiff       int64
}

var (
	NodeIPs []string

	SSHConfig   sshutil.SSH
	ManageImage string
	ServerUrl   string
	ServerToken string
	InReset     bool

	FileMd5 sync.Map

	ResultInstall sync.Map
	ResultRemove  sync.Map
)
