package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
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
	} else if FirewallConfig.Mode == "accept" {
		// 默认接收
		iptablesDefaultAccept()
	} else if FirewallConfig.Mode == "drop" {
		// 默认丢弃
		iptablesDefaultDrop()
	} else {
		logger.Error("Mode error")
	}
}

func iptablesFirewallTemplate(mode string) string {
	FirewallConfig.Ports = strings.Replace(FirewallConfig.Ports, "-", ":", -1)
	value := ""
	if FirewallConfig.Address == "" {
		if strings.Contains(FirewallConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables {MODE} INPUT -p tcp -m state --state NEW -m tcp --dport %s -j %s", FirewallConfig.Ports, FirewallConfig.Type)
			udp := fmt.Sprintf("iptables {MODE} INPUT -p udp -m state --state NEW -m udp --dport %s -j %s", FirewallConfig.Ports, FirewallConfig.Type)
			value = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			value = fmt.Sprintf("iptables {MODE} INPUT -p tcp -m state --state NEW -m %s --dport %s -j %s", FirewallConfig.Protocol, FirewallConfig.Ports, FirewallConfig.Type)
		}
	} else {
		if strings.Contains(FirewallConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables {MODE} INPUT -s %s -p tcp --dport %s -j %s", FirewallConfig.Address, FirewallConfig.Ports, FirewallConfig.Type)
			udp := fmt.Sprintf("iptables {MODE} INPUT -s %s -p udp --dport %s -j %s", FirewallConfig.Address, FirewallConfig.Ports, FirewallConfig.Type)
			value = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			value = fmt.Sprintf("iptables {MODE} INPUT -s %s -p %s --dport %s -j %s", FirewallConfig.Address, FirewallConfig.Protocol, FirewallConfig.Ports, FirewallConfig.Type)
		}
	}
	if mode == "del" {
		value = strings.ReplaceAll(value, "{MODE}", "-D")
	} else {
		value = strings.ReplaceAll(value, "{MODE}", "-I")
	}
	return value
}

func iptablesFirewallAdd() {
	// 先删除（防止重复添加）
	_, _, _ = RunCommand("-c", iptablesFirewallTemplate("del"))
	// 后添加
	_, s, err := RunCommand("-c", iptablesFirewallTemplate("add"))
	if err != nil {
		logger.Error(err, s)
	}
}

func iptablesFirewallDel() {
	_, s, err := RunCommand("-c", iptablesFirewallTemplate("del"))
	if err != nil {
		logger.Error(err, s)
	}
}

func iptablesDefaultAccept() {
	_, _, _ = RunCommand("-c", "iptables -D INPUT -p icmp --icmp-type any -j ACCEPT")
	_, _, _ = RunCommand("-c", "iptables -D INPUT -s localhost -d localhost -j ACCEPT")
	_, _, _ = RunCommand("-c", "iptables -D INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT")
	_, _, _ = RunCommand("-c", "iptables -P INPUT ACCEPT")
}

func iptablesDefaultDrop() {
	_, _, _ = RunCommand("-c", "iptables -A INPUT -p icmp --icmp-type any -j ACCEPT")
	_, _, _ = RunCommand("-c", "iptables -A INPUT -s localhost -d localhost -j ACCEPT")
	_, _, _ = RunCommand("-c", "iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT")
	_, _, _ = RunCommand("-c", "iptables -P INPUT DROP")
}
