package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"strings"
)

var ufwPath = "/etc/ufw/before.rules"

//BuildForward is
func BuildForward() {
	if Exists("/usr/sbin/ufw") {
		if ForwardConfig.Mode == "add" {
			ufwForwardAdd()
		} else {
			ufwForwardDel()
		}
	} else if Exists("/usr/sbin/iptables") {
		if ForwardConfig.Mode == "add" {
			iptablesForwardAdd()
		} else {
			iptablesForwardDel()
		}
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
	if ForwardConfig.Eip == "" {
		if strings.Contains(ForwardConfig.Protocol, "/") {
			value = fmt.Sprintf("-A PREROUTING -p tcp --dport %s -j REDIRECT --to-port %s\n-A PREROUTING -p udp --dport %s -j REDIRECT --to-port %s", ForwardConfig.Sport, ForwardConfig.Eport, ForwardConfig.Sport, ForwardConfig.Eport)
		} else {
			value = fmt.Sprintf("-A PREROUTING -p %s --dport %s -j REDIRECT --to-port %s", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Eport)
		}
	} else {
		value = fmt.Sprintf("-A PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s\n-A POSTROUTING -d %s -j MASQUERADE", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Eip, ForwardConfig.Eport, ForwardConfig.Eip)
	}
	array = SliceInsert(array, index + 1, value)
	WriteFile(ufwPath, strings.Join(array, "\n"))
}

func ufwForwardDel() {
	content := ReadFile(ufwPath)
	value := ""
	if ForwardConfig.Eip == "" {
		if strings.Contains(ForwardConfig.Protocol, "/") {
			value = fmt.Sprintf("-A PREROUTING -p tcp --dport %s -j REDIRECT --to-port %s\n-A PREROUTING -p udp --dport %s -j REDIRECT --to-port %s", ForwardConfig.Sport, ForwardConfig.Eport, ForwardConfig.Sport, ForwardConfig.Eport)
		} else {
			value = fmt.Sprintf("-A PREROUTING -p %s --dport %s -j REDIRECT --to-port %s", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Eport)
		}
	} else {
		value = fmt.Sprintf("-A PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s\n-A POSTROUTING -d %s -j MASQUERADE", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Eip, ForwardConfig.Eport, ForwardConfig.Eip)
	}
	WriteFile(ufwPath, strings.ReplaceAll(content, fmt.Sprintf("%s\n", value), ""))
}

func iptablesForwardTemplate(mode string) string {
	value := ""
	if ForwardConfig.Eip == "" {
		if strings.Contains(ForwardConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p tcp --dport %s -j REDIRECT --to-port %s", ForwardConfig.Sport, ForwardConfig.Eport)
			udp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p udp --dport %s -j REDIRECT --to-port %s", ForwardConfig.Sport, ForwardConfig.Eport)
			value = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			value = fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p %s --dport %s -j REDIRECT --to-port %s", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Eport)
		}
	} else {
		if strings.Contains(ForwardConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s", ForwardConfig.Sport, ForwardConfig.Eip, ForwardConfig.Eport)
			udp := fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p udp --dport %s -j DNAT --to-destination %s:%s", ForwardConfig.Sport, ForwardConfig.Eip, ForwardConfig.Eport)
			value = fmt.Sprintf("%s && %s && iptables -t nat {MODE} POSTROUTING -j MASQUERADE", tcp, udp)
		} else {
			value = fmt.Sprintf("iptables -t nat {MODE} PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s && iptables -t nat {MODE} POSTROUTING -j MASQUERADE", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Eip, ForwardConfig.Eport)
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
