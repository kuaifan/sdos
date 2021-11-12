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
				_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy", BaseUtils(nodeName, node))
				_ = SSHConfig.CmdAsync(node, "/root/.sdwan/deploy remove")
				_ = SSHConfig.CmdAsync(node, "rm -rf /root/.sdwan/")
			}
			if ServerDomain != "" {
				_ = SSHConfig.CmdAsync(node, fmt.Sprintf("mkdir -p /root/.sdwan/ssl/%s/", ServerDomain))
				if ServerKey != "" {
					keyCmd := fmt.Sprintf("wget -N --no-check-certificate %s -O /root/.sdwan/ssl/%s/site.key", ServerKey, ServerDomain)
					_ = SSHConfig.CmdAsync(node, keyCmd)
				}
				if ServerCrt != "" {
					crtCmd := fmt.Sprintf("wget -N --no-check-certificate %s -O /root/.sdwan/ssl/%s/site.crt", ServerCrt, ServerDomain)
					_ = SSHConfig.CmdAsync(node, crtCmd)
				}
			}
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/")
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/docker-compose.yml", DockerCompose(nodeName, node))
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy", BaseUtils(nodeName, node))
			_ = SSHConfig.CmdAsync(node, "/root/.sdwan/deploy install")
			done(node, nodeName)
		}(node)
	}
	wg.Wait()
	ResultInstall.Range(walk)
}

func done(node, nodeName string) {
	res := SSHConfig.CmdToStringNoLog(node, "cat /tmp/sdwan_install", "")
	if res == "success" {
		if Mtu == "" {
			Mtu = "1360"
		}
		var (
			keyContent = ""
			crtContent = ""
		)
		if ServerDomain != "" && ServerKey == "" {
			keyContent = SSHConfig.CmdToStringNoLog(node, fmt.Sprintf("cat /root/.sdwan/ssl/%s/site.key", ServerDomain), "\n")
			crtContent = SSHConfig.CmdToStringNoLog(node, fmt.Sprintf("cat /root/.sdwan/ssl/%s/site.crt", ServerDomain), "\n")
			if !strings.Contains(keyContent, "END RSA PRIVATE KEY") {
				logger.Error("[%s] [%s] key error %s", node, ServerDomain)
				keyContent = ""
			}
			if !strings.Contains(crtContent, "END CERTIFICATE") {
				logger.Error("[%s] [%s] crt error %s", node, ServerDomain)
				crtContent = ""
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
			logger.Error("[%s] install error %s", node, err)
		} else {
			body, errp := resp.GetBodyAsString()
			if errp != nil {
				logger.Error("[%s] install failed %s", node, errp)
			} else {
				ResultInstall.Store(node, body)
			}
		}
	} else {
		ResultInstall.Store(node, res)
	}
}

func walk(key interface{}, value interface{}) bool {
	if value.(string) == "success" {
		logger.Info("[%s] install %s", key, value)
	} else {
		if value == "" {
			value = "error"
		}
		Error(fmt.Sprintf("[%s] install %s", key, value))
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
