package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"os"
	"strconv"
	"strings"
	"time"
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

func iptablesFirewallTemplate(mode string) (string, string) {
	FirewallConfig.Ports = strings.Replace(FirewallConfig.Ports, "-", ":", -1)
	cmd := ""
	if FirewallConfig.Address == "" {
		if strings.Contains(FirewallConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t mangle {MODE} PREROUTING -p tcp -m state --state NEW -m tcp --dport %s -j %s", FirewallConfig.Ports, FirewallConfig.Type)
			udp := fmt.Sprintf("iptables -t mangle {MODE} PREROUTING -p udp -m state --state NEW -m udp --dport %s -j %s", FirewallConfig.Ports, FirewallConfig.Type)
			cmd = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			cmd = fmt.Sprintf("iptables -t mangle {MODE} PREROUTING -p tcp -m state --state NEW -m %s --dport %s -j %s", FirewallConfig.Protocol, FirewallConfig.Ports, FirewallConfig.Type)
		}
	} else {
		if strings.Contains(FirewallConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t mangle {MODE} PREROUTING -s %s -p tcp --dport %s -j %s", FirewallConfig.Address, FirewallConfig.Ports, FirewallConfig.Type)
			udp := fmt.Sprintf("iptables -t mangle {MODE} PREROUTING -s %s -p udp --dport %s -j %s", FirewallConfig.Address, FirewallConfig.Ports, FirewallConfig.Type)
			cmd = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			cmd = fmt.Sprintf("iptables -t mangle {MODE} PREROUTING -s %s -p %s --dport %s -j %s", FirewallConfig.Address, FirewallConfig.Protocol, FirewallConfig.Ports, FirewallConfig.Type)
		}
	}
	key := StringMd5(cmd)
	if mode == "del" {
		cmd = strings.ReplaceAll(cmd, "{MODE}", "-D")
	} else {
		cmd = strings.ReplaceAll(cmd, "{MODE}", "-I")
	}
	return key, cmd
}

func iptablesFirewallAdd() {
	key, cmd := iptablesFirewallTemplate("add")
	file := fmt.Sprintf("/tmp/.sdwan/tmp/firewall_%s", key)
	if FirewallConfig.Force || !Exists(file) {
		WriteFile(file, strconv.FormatInt(time.Now().Unix(), 10))
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func iptablesFirewallDel() {
	key, cmd := iptablesFirewallTemplate("del")
	file := fmt.Sprintf("/tmp/.sdwan/tmp/firewall_%s", key)
	if FirewallConfig.Force || Exists(file) {
		_ = os.RemoveAll(file)
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func iptablesDefaultAccept()  {
	file := "/tmp/.sdwan/tmp/firewall_default"
	if Exists(file) {
		_ = os.RemoveAll(file)
		_, _, _ = RunCommand("-c", "iptables -t mangle -D PREROUTING -p icmp --icmp-type any -j ACCEPT")
		_, _, _ = RunCommand("-c", "iptables -t mangle -D PREROUTING -s localhost -d localhost -j ACCEPT")
		_, _, _ = RunCommand("-c", "iptables -t mangle -D PREROUTING -m state --state ESTABLISHED,RELATED -j ACCEPT")
	}
}

func iptablesDefaultDrop()  {
	file := "/tmp/.sdwan/tmp/firewall_default"
	if !Exists(file) {
		WriteFile(file, strconv.FormatInt(time.Now().Unix(), 10))
		_, _, _ = RunCommand("-c", "iptables -t mangle -A PREROUTING -p icmp --icmp-type any -j ACCEPT")
		_, _, _ = RunCommand("-c", "iptables -t mangle -A PREROUTING -s localhost -d localhost -j ACCEPT")
		_, _, _ = RunCommand("-c", "iptables -t mangle -A PREROUTING -m state --state ESTABLISHED,RELATED -j ACCEPT")
	}
}

func iptablesDefault() {
	if FirewallConfig.Type == "ACCEPT" {
		iptablesDefaultAccept()
	} else if FirewallConfig.Type == "DROP" {
		iptablesDefaultDrop()
	}
	_, s, err := RunCommand("-c", fmt.Sprintf("iptables -t mangle -P PREROUTING %s", FirewallConfig.Type))
	if err != nil {
		logger.Panic(s, err)
	}
}