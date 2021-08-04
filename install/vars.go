package install

import (
	"strconv"

	"github.com/kuaifan/sdos/pkg/sshcmd/sshutil"
)

var (
	NodeIPs   []string

	SSHConfig sshutil.SSH
	ServerUrl     string

	Vlog int
)

func vlogToStr() string {
	str := strconv.Itoa(Vlog)
	return " -v " + str
}
