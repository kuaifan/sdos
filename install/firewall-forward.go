package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"strings"
)

//BuildFirewallForward is
func BuildFirewallForward() {
	if FirewallForwardConfig.Mode == "add" {
		// 添加
		firewallForwardAdd()
	} else if FirewallForwardConfig.Mode == "del" {
		// 删除
		firewallForwardDel()
	}
}

func firewallForwardTemplate(mode string) (string, string) {
	cmd := ""
	if FirewallForwardConfig.Dip == "" {
		if strings.Contains(FirewallForwardConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p tcp --dport %s -j REDIRECT --to-port %s -m comment --comment \"{COMMENT}\"", FirewallForwardConfig.Sport, FirewallForwardConfig.Dport)
			udp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p udp --dport %s -j REDIRECT --to-port %s -m comment --comment \"{COMMENT}\"", FirewallForwardConfig.Sport, FirewallForwardConfig.Dport)
			cmd = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			cmd = fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p %s --dport %s -j REDIRECT --to-port %s -m comment --comment \"{COMMENT}\"", FirewallForwardConfig.Protocol, FirewallForwardConfig.Sport, FirewallForwardConfig.Dport)
		}
	} else {
		if strings.Contains(FirewallForwardConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s -m comment --comment \"{COMMENT}\"", FirewallForwardConfig.Sport, FirewallForwardConfig.Dip, FirewallForwardConfig.Dport)
			udp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p udp --dport %s -j DNAT --to-destination %s:%s -m comment --comment \"{COMMENT}\"", FirewallForwardConfig.Sport, FirewallForwardConfig.Dip, FirewallForwardConfig.Dport)
			cmd = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			cmd = fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s -m comment --comment \"{COMMENT}\"", FirewallForwardConfig.Protocol, FirewallForwardConfig.Sport, FirewallForwardConfig.Dip, FirewallForwardConfig.Dport)
		}
	}
	if mode == "del" {
		cmd = strings.ReplaceAll(cmd, "{MODE}", "-D")
	} else {
		cmd = strings.ReplaceAll(cmd, "{MODE}", "-A")
	}
	key := fmt.Sprintf("sdwan-forward-%s", FirewallForwardConfig.Key)
	return key, strings.ReplaceAll(cmd, "{COMMENT}", key)
}

func firewallForwardAdd() {
	key, cmd := firewallForwardTemplate("add")
	if !FirewallForwardExist(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func firewallForwardDel() {
	key, cmd := firewallForwardTemplate("del")
	if FirewallForwardExist(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}