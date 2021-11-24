package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"os"
	"strconv"
	"strings"
	"time"
)

//BuildForward is
func BuildForward() {
	if ForwardConfig.Mode == "add" {
		// 添加
		iptablesForwardAdd()
	} else if ForwardConfig.Mode == "del" {
		// 删除
		iptablesForwardDel()
	} else {
		logger.Panic("Mode error")
	}
}

func iptablesForwardTemplate(mode string) (string, string) {
	cmd := ""
	if ForwardConfig.Dip == "" {
		if strings.Contains(ForwardConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p tcp --dport %s -j REDIRECT --to-port %s", ForwardConfig.Sport, ForwardConfig.Dport)
			udp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p udp --dport %s -j REDIRECT --to-port %s", ForwardConfig.Sport, ForwardConfig.Dport)
			cmd = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			cmd = fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p %s --dport %s -j REDIRECT --to-port %s", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Dport)
		}
	} else {
		if strings.Contains(ForwardConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s", ForwardConfig.Sport, ForwardConfig.Dip, ForwardConfig.Dport)
			udp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p udp --dport %s -j DNAT --to-destination %s:%s", ForwardConfig.Sport, ForwardConfig.Dip, ForwardConfig.Dport)
			cmd = fmt.Sprintf("%s && %s && iptables -t nat {MODE} POSTROUTING -j MASQUERADE", tcp, udp)
		} else {
			cmd = fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s && iptables -t nat {MODE} POSTROUTING -j MASQUERADE", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Dip, ForwardConfig.Dport)
		}
	}
	key := StringMd5(cmd)
	if mode == "del" {
		cmd = strings.ReplaceAll(cmd, "{MODE}", "-D")
	} else {
		cmd = strings.ReplaceAll(cmd, "{MODE}", "-A")
	}
	return key, cmd
}

func iptablesForwardAdd() {
	key, cmd := iptablesForwardTemplate("add")
	file := fmt.Sprintf("/tmp/.sdwan/tmp/forward_%s", key)
	if ForwardConfig.Force || !Exists(file) {
		WriteFile(file, strconv.FormatInt(time.Now().Unix(), 10))
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func iptablesForwardDel() {
	key, cmd := iptablesForwardTemplate("del")
	file := fmt.Sprintf("/tmp/.sdwan/tmp/forward_%s", key)
	if ForwardConfig.Force || Exists(file) {
		_ = os.RemoveAll(file)
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}
