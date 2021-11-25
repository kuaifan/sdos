package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"strings"
)

//BuildFirewallRule is
func BuildFirewallRule() {
	if FirewallRuleConfig.Mode == "add" {
		// 添加
		firewallRuleAdd()
	} else if FirewallRuleConfig.Mode == "del" {
		// 删除
		firewallRuleDel()
	}
}

func firewallRuleTemplate(mode string) (string, string) {
	FirewallRuleConfig.Ports = strings.Replace(FirewallRuleConfig.Ports, "-", ":", -1)
	cmd := ""
	if FirewallRuleConfig.Address == "" {
		if strings.Contains(FirewallRuleConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t mangle {MODE} INPUT -p tcp -m state --state NEW -m tcp --dport %s -j %s", FirewallRuleConfig.Ports, FirewallRuleConfig.Type)
			udp := fmt.Sprintf("iptables -t mangle {MODE} INPUT -p udp -m state --state NEW -m udp --dport %s -j %s", FirewallRuleConfig.Ports, FirewallRuleConfig.Type)
			cmd = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			cmd = fmt.Sprintf("iptables -t mangle {MODE} INPUT -p tcp -m state --state NEW -m %s --dport %s -j %s", FirewallRuleConfig.Protocol, FirewallRuleConfig.Ports, FirewallRuleConfig.Type)
		}
	} else {
		if strings.Contains(FirewallRuleConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t mangle {MODE} INPUT -s %s -p tcp --dport %s -j %s", FirewallRuleConfig.Address, FirewallRuleConfig.Ports, FirewallRuleConfig.Type)
			udp := fmt.Sprintf("iptables -t mangle {MODE} INPUT -s %s -p udp --dport %s -j %s", FirewallRuleConfig.Address, FirewallRuleConfig.Ports, FirewallRuleConfig.Type)
			cmd = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			cmd = fmt.Sprintf("iptables -t mangle {MODE} INPUT -s %s -p %s --dport %s -j %s", FirewallRuleConfig.Address, FirewallRuleConfig.Protocol, FirewallRuleConfig.Ports, FirewallRuleConfig.Type)
		}
	}
	key := fmt.Sprintf("sdwan-rule-%s", FirewallRuleConfig.Key)
	if mode == "del" {
		cmd = strings.ReplaceAll(cmd, "{MODE}", "-D")
	} else {
		cmd = strings.ReplaceAll(cmd, "{MODE}", "-I")
	}
	return key, fmt.Sprintf("%s -m comment --comment \"%s\"", cmd, key)
}

func firewallRuleAdd() {
	key, cmd := firewallRuleTemplate("add")
	if !FirewallRuleExist(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func firewallRuleDel() {
	key, cmd := firewallRuleTemplate("del")
	if FirewallRuleExist(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}