package install

import (
	"fmt"
	"strings"
)

//BuildUfw is
func BuildUfw() {
	if UFWConfig.Type == "add" {
		add()
	} else if UFWConfig.Type == "del" {
		del()
	}
}

func add() {
	content := ReadFile(UFWConfig.Path)
	if !strings.Contains(content, "*nat") {
		content = fmt.Sprintf("*nat\n:PREROUTING ACCEPT [0:0]\n:POSTROUTING ACCEPT [0:0]\nCOMMIT\n%s", content)
	}
	array := strings.Split(content, "\n")
	index := FindIndex(array, ":POSTROUTING ACCEPT [0:0]")
	value := ""
	if UFWConfig.Dip == "" {
		if strings.Contains(UFWConfig.Protocol, "/") {
			value = fmt.Sprintf("-A PREROUTING -p tcp --dport %s -j REDIRECT --to-port %s\n-A PREROUTING -p udp --dport %s -j REDIRECT --to-port %s", UFWConfig.Sport, UFWConfig.Dport, UFWConfig.Sport, UFWConfig.Dport)
		} else {
			value = fmt.Sprintf("-A PREROUTING -p %s --dport %s -j REDIRECT --to-port %s", UFWConfig.Protocol, UFWConfig.Sport, UFWConfig.Dport)
		}
	} else {
		value = fmt.Sprintf("-A PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s\n-A POSTROUTING -d %s -j MASQUERADE", UFWConfig.Protocol, UFWConfig.Sport, UFWConfig.Dip, UFWConfig.Dport, UFWConfig.Dip)
	}
	array = SliceInsert(array, index + 1, value)
	WriteFile(UFWConfig.Path, strings.Join(array, "\n"))
}

func del() {
	content := ReadFile(UFWConfig.Path)
	value := ""
	if UFWConfig.Dip == "" {
		if strings.Contains(UFWConfig.Protocol, "/") {
			value = fmt.Sprintf("-A PREROUTING -p tcp --dport %s -j REDIRECT --to-port %s\n-A PREROUTING -p udp --dport %s -j REDIRECT --to-port %s", UFWConfig.Sport, UFWConfig.Dport, UFWConfig.Sport, UFWConfig.Dport)
		} else {
			value = fmt.Sprintf("-A PREROUTING -p %s --dport %s -j REDIRECT --to-port %s", UFWConfig.Protocol, UFWConfig.Sport, UFWConfig.Dport)
		}
	} else {
		value = fmt.Sprintf("-A PREROUTING -p %s --dport %s -j DNAT --to-destination %s:%s\n-A POSTROUTING -d %s -j MASQUERADE", UFWConfig.Protocol, UFWConfig.Sport, UFWConfig.Dip, UFWConfig.Dport, UFWConfig.Dip)
	}
	WriteFile(UFWConfig.Path, strings.ReplaceAll(content, fmt.Sprintf("%s\n", value), ""))
}
