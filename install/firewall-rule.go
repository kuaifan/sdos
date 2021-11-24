package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"os"
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
	} else if FirewallRuleConfig.Mode == "install" {
		// 安装
		iptablesInstall()
	} else if FirewallRuleConfig.Mode == "uninstall" {
		// 卸载
		iptablesUnInstall()
	} else {
		logger.Panic("Mode error")
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
	key := StringMd5(cmd)
	if mode == "del" {
		cmd = strings.ReplaceAll(cmd, "{MODE}", "-D")
	} else {
		cmd = strings.ReplaceAll(cmd, "{MODE}", "-I")
	}
	return key, fmt.Sprintf("%s -m comment --comment \"%s\"", cmd, key)
}

func firewallRuleAdd() {
	key, cmd := firewallRuleTemplate("add")
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/firewall_rule_%s", key)
	WriteFile(cmdFile, strings.Join(os.Args, " "))
	//
	if !existMangleInput(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func firewallRuleDel() {
	key, cmd := firewallRuleTemplate("del")
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/firewall_rule_%s", key)
	_ = os.RemoveAll(cmdFile)
	//
	if existMangleInput(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func iptablesInstall() {
	key := StringMd5("sdwan-default")
	cmd := strings.Join([]string{
		fmt.Sprintf("iptables -t mangle -A INPUT -p icmp --icmp-type any -j ACCEPT -m comment --comment \"%s\"", key),
		"iptables -t mangle -A INPUT -s localhost -d localhost -j ACCEPT",
		"iptables -t mangle -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT",
		"iptables -t mangle -P INPUT DROP",
		"iptables -t nat -A POSTROUTING -j MASQUERADE",
	}, " && ")
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/firewall_rule_%s", key)
	WriteFile(cmdFile, strings.Join(os.Args, " "))
	//
	if !existMangleInput(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func iptablesUnInstall() {
	key := StringMd5("sdwan-default")
	cmd := strings.Join([]string{
		fmt.Sprintf("iptables -t mangle -D INPUT -p icmp --icmp-type any -j ACCEPT -m comment --comment \"%s\"", key),
		"iptables -t mangle -D INPUT -s localhost -d localhost -j ACCEPT",
		"iptables -t mangle -D INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT",
		"iptables -t mangle -P INPUT ACCEPT",
		"iptables -t nat -D POSTROUTING -j MASQUERADE",
	}, " && ")
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/firewall_rule_%s", key)
	_ = os.RemoveAll(cmdFile)
	//
	if existMangleInput(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func existMangleInput(key string) bool {
	result, _, _ := RunCommand("-c", fmt.Sprintf("iptables -t mangle -L INPUT | grep '%s'", key))
	if strings.Contains(result, key) {
		return true
	}
	return false
}