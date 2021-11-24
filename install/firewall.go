package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"os"
	"strings"
)

//BuildFirewall is
func BuildFirewall() {
	if FirewallConfig.Mode == "add" {
		// 添加
		iptablesFirewallAdd()
	} else if FirewallConfig.Mode == "del" {
		// 删除
		iptablesFirewallDel()
	} else if FirewallConfig.Mode == "install" {
		// 安装
		iptablesInstall()
	} else if FirewallConfig.Mode == "uninstall" {
		// 卸载
		iptablesUnInstall()
	} else {
		logger.Panic("Mode error")
	}
}

func iptablesFirewallTemplate(mode string) (string, string) {
	FirewallConfig.Ports = strings.Replace(FirewallConfig.Ports, "-", ":", -1)
	cmd := ""
	if FirewallConfig.Address == "" {
		if strings.Contains(FirewallConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t mangle {MODE} INPUT -p tcp -m state --state NEW -m tcp --dport %s -j %s", FirewallConfig.Ports, FirewallConfig.Type)
			udp := fmt.Sprintf("iptables -t mangle {MODE} INPUT -p udp -m state --state NEW -m udp --dport %s -j %s", FirewallConfig.Ports, FirewallConfig.Type)
			cmd = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			cmd = fmt.Sprintf("iptables -t mangle {MODE} INPUT -p tcp -m state --state NEW -m %s --dport %s -j %s", FirewallConfig.Protocol, FirewallConfig.Ports, FirewallConfig.Type)
		}
	} else {
		if strings.Contains(FirewallConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t mangle {MODE} INPUT -s %s -p tcp --dport %s -j %s", FirewallConfig.Address, FirewallConfig.Ports, FirewallConfig.Type)
			udp := fmt.Sprintf("iptables -t mangle {MODE} INPUT -s %s -p udp --dport %s -j %s", FirewallConfig.Address, FirewallConfig.Ports, FirewallConfig.Type)
			cmd = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			cmd = fmt.Sprintf("iptables -t mangle {MODE} INPUT -s %s -p %s --dport %s -j %s", FirewallConfig.Address, FirewallConfig.Protocol, FirewallConfig.Ports, FirewallConfig.Type)
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

func iptablesFirewallAdd() {
	key, cmd := iptablesFirewallTemplate("add")
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/firewall_%s", key)
	WriteFile(cmdFile, strings.Join(os.Args, " "))
	//
	if !existMangleInput(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func iptablesFirewallDel() {
	key, cmd := iptablesFirewallTemplate("del")
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/firewall_%s", key)
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
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/firewall_%s", key)
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
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/firewall_%s", key)
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