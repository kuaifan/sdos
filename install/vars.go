package install

import (
	"github.com/kuaifan/sdos/pkg/sshcmd/sshutil"
)

var (
	NodeIPs []string

	SSHConfig   sshutil.SSH
	ManageImage string
	ServerUrl   string
	ServerToken string
)
