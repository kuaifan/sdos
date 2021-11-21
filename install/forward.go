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
		if ForwardConfig.Type == "add" {
			ufwAdd()
		} else {
			ufwDel()
		}
	} else if Exists("/usr/sbin/iptables") {
		if ForwardConfig.Type == "add" {
			iptablesAdd()
		} else {
			iptablesDel()
		}
	}
}

func ufwAdd() {
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

func ufwDel() {
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

func iptablesTemplate(mode string) string {
	value := ""
	if ForwardConfig.Eip == "" {
		if strings.Contains(ForwardConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t nat -A PREROUTING -p tcp --dport %s -j REDIRECT --to-port %s", ForwardConfig.Sport, ForwardConfig.Eport)
			udp := fmt.Sprintf("iptables -t nat -A PREROUTING -p udp --dport %s -j REDIRECT --to-port %s", ForwardConfig.Sport, ForwardConfig.Eport)
			value = fmt.Sprintf("%s && %s", tcp, udp)
		} else {
			value = fmt.Sprintf("iptables -t nat -A PREROUTING -p %s --dport %s -j REDIRECT --to-port %s", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Eport)
		}
	} else {
		if strings.Contains(ForwardConfig.Protocol, "/") {
			tcp := fmt.Sprintf("iptables -t nat -A PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s", ForwardConfig.Sport, ForwardConfig.Eip, ForwardConfig.Eport)
			udp := fmt.Sprintf("iptables -t nat -A PREROUTING -p udp --dport %s -j DNAT --to-destination %s:%s", ForwardConfig.Sport, ForwardConfig.Eip, ForwardConfig.Eport)
			value = fmt.Sprintf("%s && %s && iptables -t nat -A POSTROUTING -j MASQUERADE", tcp, udp)
		} else {
			value = fmt.Sprintf("iptables -t nat -A PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s && iptables -t nat -A POSTROUTING -j MASQUERADE", ForwardConfig.Protocol, ForwardConfig.Sport, ForwardConfig.Eip, ForwardConfig.Eport)
		}
	}
	if mode == "del" {
		value = strings.ReplaceAll(value, "-A", "-D")
	}
	return value
}

func iptablesAdd() {
	cmd := iptablesTemplate("add")
	_, s, err := RunCommand("-c", cmd)
	if err != nil {
		logger.Error(err, s)
	}
}

func iptablesDel() {
	cmd := iptablesTemplate("del")
	_, s, err := RunCommand("-c", cmd)
	if err != nil {
		logger.Error(err, s)
	}
}
