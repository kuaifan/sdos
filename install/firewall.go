package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"os"
	"strings"
)

//BuildFirewall is
func BuildFirewall() {
	if FirewallConfig.Mode == "install" {
		// 安装
		FirewallInstall()
	} else if FirewallConfig.Mode == "uninstall" {
		// 卸载
		FirewallUnInstall()
	} else {
		logger.Panic("Mode error")
	}
}

func FirewallInstall() {
	key := StringMd5("sdwan-default")
	cmd := strings.Join([]string{
		fmt.Sprintf("iptables -t mangle -A INPUT -p icmp --icmp-type any -j ACCEPT -m comment --comment \"%s\"", key),
		"iptables -t mangle -A INPUT -s localhost -d localhost -j ACCEPT",
		"iptables -t mangle -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT",
		"iptables -t mangle -P INPUT DROP",
	}, " && ")
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/firewall_rule_%s", key)
	WriteFile(cmdFile, strings.Join(os.Args, " "))
	//
	if !ExistMangleInput(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func FirewallUnInstall() {
	key := StringMd5("sdwan-default")
	cmd := strings.Join([]string{
		fmt.Sprintf("iptables -t mangle -D INPUT -p icmp --icmp-type any -j ACCEPT -m comment --comment \"%s\"", key),
		"iptables -t mangle -D INPUT -s localhost -d localhost -j ACCEPT",
		"iptables -t mangle -D INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT",
		"iptables -t mangle -P INPUT ACCEPT",
	}, " && ")
	cmdFile := fmt.Sprintf("/usr/.sdwan/startcmd/firewall_rule_%s", key)
	_ = os.RemoveAll(cmdFile)
	//
	if ExistMangleInput(key) {
		_, s, err := RunCommand("-c", cmd)
		if err != nil {
			logger.Panic(s, err)
		}
	}
}

func ExistMangleInput(key string) bool {
	result, _, _ := RunCommand("-c", fmt.Sprintf("iptables -L INPUT -nvt mangle | grep '%s'", key))
	if strings.Contains(result, key) {
		return true
	}
	return false
}

func ExistNatPrerouting(key string) bool {
	result, _, _ := RunCommand("-c", fmt.Sprintf("iptables -L PREROUTING -nvt nat | grep '%s'", key))
	if strings.Contains(result, key) {
		return true
	}
	return false
}