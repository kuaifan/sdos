package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"strings"
)

//BuildForward is
func BuildForward() {
	if ForwardConfig.Mode == "add" {
		// 添加
		iptablesForwardAdd()
	} else if ForwardConfig.Mode == "del" {
		// 删除
		iptablesForwardDel()
	} else {
		logger.Panic("Mode error")
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
		value = fmt.Sprintf("%s &> /dev/null", value)
	} else {
		value = strings.ReplaceAll(value, "{MODE}", "-A")
	}
	return value
}

func iptablesForwardAdd() {
	if ForwardConfig.Force {
		_, _, _ = RunCommand("-c", iptablesForwardTemplate("del"))
	}
	_, s, err := RunCommand("-c", iptablesForwardTemplate("add"))
	if err != nil {
		logger.Panic(s, err)
	}
}

func iptablesForwardDel() {
	_, s, err := RunCommand("-c", iptablesForwardTemplate("del"))
	if err != nil {
		logger.Panic(s, err)
	}
}
