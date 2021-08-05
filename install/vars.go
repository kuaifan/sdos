package install

import (
	"github.com/kuaifan/sdos/pkg/sshcmd/sshutil"
)

var (
	NodeIPs []string

	SSHConfig sshutil.SSH
	Image     string
	ServerUrl string
)