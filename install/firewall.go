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
	} else if FirewallConfig.Mode == "default" {
		// 修改默认
		iptablesDefault()
	} else {
		logger.Panic("Mode error")
	}
}

func iptablesFirewallTemplate(mode string) string {
	FirewallConfig.Ports = strings.Replace(FirewallConfig.Ports, "-", ":", -1)
	value := ""
	if FirewallConfig.Address == "" {
		if strings.Contains(FirewallConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t mangle {MODE} PREROUTING -p tcp -m state --state NEW -m tcp --dport %s -j %s", FirewallConfig.Ports, FirewallConfig.Type)
			udp := fmt.Sprintf("iptables -t mangle {MODE} PREROUTING -p udp -m state --state NEW -m udp --dport %s -j %s", FirewallConfig.Ports, FirewallConfig.Type)
			value = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			value = fmt.Sprintf("iptables -t mangle {MODE} PREROUTING -p tcp -m state --state NEW -m %s --dport %s -j %s", FirewallConfig.Protocol, FirewallConfig.Ports, FirewallConfig.Type)
		}
	} else {
		if strings.Contains(FirewallConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t mangle {MODE} PREROUTING -s %s -p tcp --dport %s -j %s", FirewallConfig.Address, FirewallConfig.Ports, FirewallConfig.Type)
			udp := fmt.Sprintf("iptables -t mangle {MODE} PREROUTING -s %s -p udp --dport %s -j %s", FirewallConfig.Address, FirewallConfig.Ports, FirewallConfig.Type)
			value = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			value = fmt.Sprintf("iptables -t mangle {MODE} PREROUTING -s %s -p %s --dport %s -j %s", FirewallConfig.Address, FirewallConfig.Protocol, FirewallConfig.Ports, FirewallConfig.Type)
		}
	}
	if mode == "del" {
		value = strings.ReplaceAll(value, "{MODE}", "-D")
		value = fmt.Sprintf("%s &> /dev/null", value)
	} else {
		value = strings.ReplaceAll(value, "{MODE}", "-I")
	}
	return value
}

func iptablesFirewallAdd() {
	if FirewallConfig.Force {
		_, _, _ = RunCommand("-c", iptablesFirewallTemplate("del"))
	}
	_, s, err := RunCommand("-c", iptablesFirewallTemplate("add"))
	if err != nil {
		logger.Panic(s, err)
	}
}

func iptablesFirewallDel() {
	_, s, err := RunCommand("-c", iptablesFirewallTemplate("del"))
	if err != nil {
		logger.Panic(s, err)
	}
}

func iptablesDefaultAccept()  {
	_, _, _ = RunCommand("-c", "iptables -t mangle -D PREROUTING -p icmp --icmp-type any -j ACCEPT &> /dev/null")
	_, _, _ = RunCommand("-c", "iptables -t mangle -D PREROUTING -s localhost -d localhost -j ACCEPT &> /dev/null")
	_, _, _ = RunCommand("-c", "iptables -t mangle -D PREROUTING -m state --state ESTABLISHED,RELATED -j ACCEPT &> /dev/null")
}

func iptablesDefaultDrop()  {
	_, _, _ = RunCommand("-c", "iptables -t mangle -A PREROUTING -p icmp --icmp-type any -j ACCEPT")
	_, _, _ = RunCommand("-c", "iptables -t mangle -A PREROUTING -s localhost -d localhost -j ACCEPT")
	_, _, _ = RunCommand("-c", "iptables -t mangle -A PREROUTING -m state --state ESTABLISHED,RELATED -j ACCEPT")
}

func iptablesDefault() {
	if FirewallConfig.Type == "ACCEPT" || FirewallConfig.Force {
		iptablesDefaultAccept()
	}
	if FirewallConfig.Type == "DROP" {
		iptablesDefaultDrop()
	}
	_, s, err := RunCommand("-c", fmt.Sprintf("iptables -t mangle -P PREROUTING %s", FirewallConfig.Type))
	if err != nil {
		logger.Panic(s, err)
	}
}