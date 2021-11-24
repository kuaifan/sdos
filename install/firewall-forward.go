package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"os"
	"strings"
)

//BuildFirewallForward is
func BuildFirewallForward() {
	if FirewallForwardConfig.Mode == "add" {
		// 添加
		iptablesFirewallForwardAdd()
	} else if FirewallForwardConfig.Mode == "del" {
		// 删除
		iptablesFirewallForwardDel()
	} else {
		logger.Panic("Mode error")
	}
}

func iptablesFirewallForwardTemplate(mode string) (string, string) {
	cmd := ""
	if FirewallForwardConfig.Dip == "" {
		if strings.Contains(FirewallForwardConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p tcp --dport %s -j REDIRECT --to-port %s", FirewallForwardConfig.Sport, FirewallForwardConfig.Dport)
			udp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p udp --dport %s -j REDIRECT --to-port %s", FirewallForwardConfig.Sport, FirewallForwardConfig.Dport)
			cmd = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			cmd = fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p %s --dport %s -j REDIRECT --to-port %s", FirewallForwardConfig.Protocol, FirewallForwardConfig.Sport, FirewallForwardConfig.Dport)
		}
	} else {
		if strings.Contains(FirewallForwardConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s", FirewallForwardConfig.Sport, FirewallForwardConfig.Dip, FirewallForwardConfig.Dport)
			udp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p udp --dport %s -j DNAT --to-destination %s:%s", FirewallForwardConfig.Sport, FirewallForwardConfig.Dip, FirewallForwardConfig.Dport)
			cmd = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			cmd = fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s", FirewallForwardConfig.Protocol, FirewallForwardConfig.Sport, FirewallForwardConfig.Dip, FirewallForwardConfig.Dport)
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

func iptablesFirewallForwardAdd() {
	key, cmd := iptablesFirewallForwardTemplate("add")
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/firewall_forward_%s", key)
	WriteFile(cmdFile, strings.Join(os.Args, " "))
	//
	if !existNatPrerouting(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func iptablesFirewallForwardDel() {
	key, cmd := iptablesFirewallForwardTemplate("del")
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/firewall_forward_%s", key)
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