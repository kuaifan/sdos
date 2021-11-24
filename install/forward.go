package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"os"
	"strings"
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
	return key, fmt.Sprintf("%s -m comment --comment \"%s\"", cmd, key)
}

func iptablesForwardAdd() {
	key, cmd := iptablesForwardTemplate("add")
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/forward_%s", key)
	WriteFile(cmdFile, strings.Join(os.Args, " "))
	//
	if !existNatPrerouting(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func iptablesForwardDel() {
	key, cmd := iptablesForwardTemplate("del")
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/forward_%s", key)
	_ = os.RemoveAll(cmdFile)
	//
	if existNatPrerouting(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func existNatPrerouting(key string) bool {
	result, _, _ := RunCommand("-c", fmt.Sprintf("iptables -t nat -L PREROUTING | grep '%s'", key))
	if strings.Contains(result, key) {
		return true
	}
	return false
}