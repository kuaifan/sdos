package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"strings"
)

//BuildFirewall is
func BuildFirewall() {
	if FirewallConfig.Mode == "install" {
		// 安装
		firewallInstall()
	} else if FirewallConfig.Mode == "uninstall" {
		// 卸载
		firewallUnInstall()
	} else if FirewallConfig.Mode == "check" {
		// 检测
		firewallCheckRule()
		firewallCheckForward()
	}
}

func firewallInstall() {
	key := StringMd5("firewall-default")
	cmd := strings.Join([]string{
		fmt.Sprintf("iptables -t mangle -A INPUT -p icmp --icmp-type any -j ACCEPT -m comment --comment \"%s\"", key),
		fmt.Sprintf("iptables -t mangle -A INPUT -s localhost -d localhost -j ACCEPT -m comment --comment \"%s\"", key),
		fmt.Sprintf("iptables -t mangle -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT -m comment --comment \"%s\"", key),
		"iptables -t mangle -P INPUT DROP",
	}, " && ")
	if !FirewallRuleExist(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func firewallUnInstall() {
	key := StringMd5("firewall-default")
	cmd := strings.Join([]string{
		fmt.Sprintf("iptables -t mangle -D INPUT -p icmp --icmp-type any -j ACCEPT -m comment --comment \"%s\"", key),
		fmt.Sprintf("iptables -t mangle -D INPUT -s localhost -d localhost -j ACCEPT -m comment --comment \"%s\"", key),
		fmt.Sprintf("iptables -t mangle -D INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT -m comment --comment \"%s\"", key),
		"iptables -t mangle -P INPUT ACCEPT",
	}, " && ")
	if FirewallRuleExist(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func firewallCheckRule() {
	keys := strings.Split(FirewallConfig.Keys, ",")
	condition := true
	for condition {
		condition = false
		result, _, _ := RunCommand("-c", "iptables -L INPUT -nvt mangle --line-number | grep 'sdwan-'")
		items := strings.Split(result, "\n")
		for _, item := range items {
			if item == "" {
				continue
			}
			del := true
			for _, key := range keys {
				if strings.Contains(item, key) {
					del = false
					break
				}
			}
			if del {
				// 规则不存在
				condition = true
				words := strings.Fields(item)
				_, _, _ = RunCommand("-c", fmt.Sprintf("iptables -t mangle -D INPUT %s", words[0]))
				break
			}
		}
	}
}

func firewallCheckForward() {
	keys := strings.Split(FirewallConfig.Keys, ",")
	condition := true
	for condition {
		condition = false
		result, _, _ := RunCommand("-c", "iptables -L PREROUTING -nvt nat --line-number | grep 'sdwan-'")
		items := strings.Split(result, "\n")
		for _, item := range items {
			if item == "" {
				continue
			}
			del := true
			for _, key := range keys {
				if strings.Contains(item, key) {
					del = false
					break
				}
			}
			if del {
				// 规则不存在
				condition = true
				words := strings.Fields(item)
				_, _, _ = RunCommand("-c", fmt.Sprintf("iptables -t nat -D PREROUTING %s", words[0]))
				break
			}
		}
	}
}