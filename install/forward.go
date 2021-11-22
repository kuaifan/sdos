package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"strings"
)

var ufwPath = "/etc/ufw/before.rules"

//BuildForward is
func BuildForward() {
	if FirewallConfig.Mode == "add" {
		// 添加
		forwardAdd()
	} else if FirewallConfig.Mode == "del" {
		// 删除
		forwardDel()
	} else {
		logger.Error("Mode error")
	}
}

func forwardAdd() {
	if Exists("/usr/sbin/ufw") {
		ufwForwardAdd()
	} else if Exists("/usr/sbin/firewalld") {
		cmdForwardAdd()
	} else if Exists("/etc/init.d/iptables") {
		iptablesForwardAdd()
	}
}

func forwardDel() {
	if Exists("/usr/sbin/ufw") {
		ufwForwardDel()
	} else if Exists("/usr/sbin/firewalld") {
		cmdForwardDel()
	} else if Exists("/etc/init.d/iptables") {
		iptablesForwardDel()
	}
}

func ufwForwardAdd() {
	content := ReadFile(ufwPath)
	if !strings.Contains(content, "*nat") {
		content = fmt.Sprintf("*nat\n:PREROUTING ACCEPT [0:0]\n:POSTROUTING ACCEPT [0:0]\nCOMMIT\n%s", content)
	}
	array := strings.Split(content, "\n")
	index := FindIndex(array, ":POSTROUTING ACCEPT [0:0]")
	value := ""
	if ForwardConfig.Dip == "" {
		if strings.Contains(ForwardConfig.Protocol, "/") {
			value = fmt.Sprintf("-A PREROUTING -p tcp --dport %s -j REDIRECT --to-port %s\n-A PREROUTING -p udp --dport %s -j REDIRECT --to-port %s", ForwardConfig.Sport, ForwardConfig.Dport, ForwardConfig.Sport, ForwardConfig.Dport)
		} else {
			value = fmt.Sprintf("-A PREROUTING -p %s --dport %s -j REDIRECT --to-port %s", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Dport)
		}
	} else {
		value = fmt.Sprintf("-A PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s\n-A POSTROUTING -d %s -j MASQUERADE", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Dip, ForwardConfig.Dport, ForwardConfig.Dip)
	}
	array = SliceInsert(array, index + 1, value)
	WriteFile(ufwPath, strings.Join(array, "\n"))
}

func ufwForwardDel() {
	content := ReadFile(ufwPath)
	value := ""
	if ForwardConfig.Dip == "" {
		if strings.Contains(ForwardConfig.Protocol, "/") {
			value = fmt.Sprintf("-A PREROUTING -p tcp --dport %s -j REDIRECT --to-port %s\n-A PREROUTING -p udp --dport %s -j REDIRECT --to-port %s", ForwardConfig.Sport, ForwardConfig.Dport, ForwardConfig.Sport, ForwardConfig.Dport)
		} else {
			value = fmt.Sprintf("-A PREROUTING -p %s --dport %s -j REDIRECT --to-port %s", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Dport)
		}
	} else {
		value = fmt.Sprintf("-A PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s\n-A POSTROUTING -d %s -j MASQUERADE", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Dip, ForwardConfig.Dport, ForwardConfig.Dip)
	}
	WriteFile(ufwPath, strings.ReplaceAll(content, fmt.Sprintf("%s\n", value), ""))
}

func cmdForwardTemplate(mode string) string {
	value := ""
	if strings.Contains(ForwardConfig.Protocol, "/") {
		tcp := fmt.Sprintf("firewall-cmd --permanent --zone=public --{MODE}-forward-port=port=\"%s\":proto=tcp:toaddr=\"%s\":toport=\"%s\"", ForwardConfig.Sport, ForwardConfig.Dip, ForwardConfig.Dport)
		udp := fmt.Sprintf("firewall-cmd --permanent --zone=public --{MODE}-forward-port=port=\"%s\":proto=udp:toaddr=\"%s\":toport=\"%s\"", ForwardConfig.Sport, ForwardConfig.Dip, ForwardConfig.Dport)
		value = fmt.Sprintf("%s && %s", tcp, udp)
	} else {
		value = fmt.Sprintf("firewall-cmd --permanent --zone=public --{MODE}-forward-port=port=\"%s\":proto=\"%s\":toaddr=\"%s\":toport=\"%s\"", ForwardConfig.Sport, ForwardConfig.Protocol, ForwardConfig.Dip, ForwardConfig.Dport)
	}
	if mode == "del" {
		value = strings.ReplaceAll(value, "{MODE}", "remove")
	} else {
		value = strings.ReplaceAll(value, "{MODE}", "add")
	}
	return value
}

func cmdForwardAdd() {
	cmd := cmdForwardTemplate("add")
	_, s, err := RunCommand("-c", cmd)
	if err != nil {
		logger.Error(err, s)
	}
}

func cmdForwardDel() {
	cmd := cmdForwardTemplate("del")
	_, s, err := RunCommand("-c", cmd)
	if err != nil {
		logger.Error(err, s)
	}
}

func iptablesForwardTemplate(mode string) string {
	value := ""
	if ForwardConfig.Dip == "" {
		if strings.Contains(ForwardConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p tcp --dport %s -j REDIRECT --to-port %s", ForwardConfig.Sport, ForwardConfig.Dport)
			udp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p udp --dport %s -j REDIRECT --to-port %s", ForwardConfig.Sport, ForwardConfig.Dport)
			value = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			value = fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p %s --dport %s -j REDIRECT --to-port %s", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Dport)
		}
	} else {
		if strings.Contains(ForwardConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s", ForwardConfig.Sport, ForwardConfig.Dip, ForwardConfig.Dport)
			udp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p udp --dport %s -j DNAT --to-destination %s:%s", ForwardConfig.Sport, ForwardConfig.Dip, ForwardConfig.Dport)
			value = fmt.Sprintf("%s && %s && iptables -t nat {MODE} POSTROUTING -j MASQUERADE", tcp, udp)
		} else {
			value = fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s && iptables -t nat {MODE} POSTROUTING -j MASQUERADE", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Dip, ForwardConfig.Dport)
		}
	}
	if mode == "del" {
		value = strings.ReplaceAll(value, "{MODE}", "-D")
	} else {
		value = strings.ReplaceAll(value, "{MODE}", "-A")
	}
	return value
}

func iptablesForwardAdd() {
	cmd := iptablesForwardTemplate("add")
	_, s, err := RunCommand("-c", cmd)
	if err != nil {
		logger.Error(err, s)
	}
}

func iptablesForwardDel() {
	cmd := iptablesForwardTemplate("del")
	_, s, err := RunCommand("-c", cmd)
	if err != nil {
		logger.Error(err, s)
	}
}
