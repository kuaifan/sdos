package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"github.com/nahid/gohttp"
	"strconv"
	"strings"
	"sync"
	"time"
)

//BuildInstall is
func BuildInstall(installNodes []string) {
	if len(installNodes) > 0 {
		installNodesFunc(installNodes)
	}
}

func installNodesFunc(installNodes []string) {
	nodes := installNodes
	i := &SdosInstaller{
		Nodes: nodes,
	}
	i.InstallNodes()
	NodeIPs = append(NodeIPs, installNodes...)
}

func (s *SdosInstaller) InstallNodes() {
	var wg sync.WaitGroup
	for _, node := range s.Nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			nodeName := GetRemoteHostName(node)
			if InReset {
				_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/")
				_ = SSHConfig.SaveFile(node, "/root/.sdwan/base", BaseUtils(nodeName, node))
				_ = SSHConfig.CmdAsync(node, "/root/.sdwan/base remove")
				_ = SSHConfig.CmdAsync(node, "rm -rf /root/.sdwan/")
			}
			if ServerKey != "" {
				_ = SSHConfig.CmdAsync(node, fmt.Sprintf("mkdir -p /root/.sdwan/ssl/%s/", ServerDomain))
				_ = SSHConfig.CmdAsync(node, fmt.Sprintf("wget -N --no-check-certificate '%s' -O /root/.sdwan/ssl/%s/site.key", ServerKey, ServerDomain))
				_ = SSHConfig.CmdAsync(node, fmt.Sprintf("wget -N --no-check-certificate '%s' -O /root/.sdwan/ssl/%s/site.crt", ServerCrt, ServerDomain))
			}
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/")
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/docker-compose.yml", DockerCompose(nodeName, node))
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/base", BaseUtils(nodeName, node))
			_ = SSHConfig.CmdAsync(node, "/root/.sdwan/base install")
			installDone(node, nodeName)
		}(node)
	}
	wg.Wait()
	ResultInstall.Range(installWalk)
}

func installDone(node string, nodeName string) {
	res := SSHConfig.CmdToStringNoLog(node, "cat /tmp/.sdwan_install", "")
	if res == "success" {
		if Mtu == "" {
			Mtu = "1360"
		}
		var (
			keyContent = ""
			crtContent = ""
		)
		if ServerDomain != "" {
			keyContent = SSHConfig.CmdToStringNoLog(node, fmt.Sprintf("cat /root/.sdwan/ssl/%s/site.key", ServerDomain), "\n")
			crtContent = SSHConfig.CmdToStringNoLog(node, fmt.Sprintf("cat /root/.sdwan/ssl/%s/site.crt", ServerDomain), "\n")
			if !strings.Contains(keyContent, "PRIVATE KEY") {
				ResultInstall.Store(node, "read key error")
				return
			}
			if !strings.Contains(crtContent, "END CERTIFICATE") {
				ResultInstall.Store(node, "read crt error")
				return
			}
		}
		nodeIp, nodePort := GetIpAndPort(node)
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		resp, err := gohttp.NewRequest().
			FormData(map[string]string{
				"action":     "install",
				"ip":         nodeIp,
				"name":       nodeName,
				"mtu":        Mtu,
				"port":       nodePort,
				"user":       SSHConfig.User,
				"pw":         SSHConfig.GetPassword(node),
				"tk":         ServerToken,
				"domain":     ServerDomain,
				"domain_key": keyContent,
				"domain_crt": crtContent,
				"timestamp":  timestamp,
			}).
			Post(ReportUrl)
		if err != nil || resp == nil {
			ResultInstall.Store(node, err.Error())
			return
		}
		body, errp := resp.GetBodyAsString()
		if errp != nil {
			ResultInstall.Store(node, errp.Error())
			return
		}
		ResultInstall.Store(node, body)
		return
	}
	ResultInstall.Store(node, res)
}

func installWalk(key interface{}, value interface{}) bool {
	if value.(string) == "success" {
		logger.Info("[%s] install %s", key, value)
	} else {
		if value != "" && value != "error" {
			value = fmt.Sprintf(": %s", value)
		}
		Error(fmt.Sprintf("[%s] install error%s", key, value))
	}
	return true
}

func Error(error string) {
	logger.Error(error)
	nodes := ParseIPs(NodeIPs)
	for _, node := range nodes {
		nodeIp, _ := GetIpAndPort(node)
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		_, _ = gohttp.NewRequest().
			FormData(map[string]string{
				"action":    "error",
				"ip":        nodeIp,
				"error":     error,
				"timestamp": timestamp,
			}).
			Post(ReportUrl)
	}
}
