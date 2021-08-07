package install

import (
	"github.com/kuaifan/sdos/pkg/sshcmd/sshutil"
	"sync"
)

var (
	NodeIPs []string

	SSHConfig   sshutil.SSH
	ManageImage string
	ServerUrl   string
	ServerToken string

	FileMd5 sync.Map

	resultInstall   sync.Map
	resultRemove sync.Map
)
