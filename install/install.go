package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"github.com/nahid/gohttp"
	"strconv"
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
				_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/deploy/")
				_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy/utils", BaseUtils(nodeName, node))
				_ = SSHConfig.CmdAsync(node, "/root/.sdwan/deploy/utils remove")
				_ = SSHConfig.CmdAsync(node, "rm -rf /root/.sdwan/")
			}
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/work/")
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/deploy/")
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy/docker-compose.yml", DockerCompose(nodeName, node))
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy/utils", BaseUtils(nodeName, node))
			_ = SSHConfig.CmdAsync(node, "/root/.sdwan/deploy/utils install")
			reportInstall(node, nodeName)
		}(node)
	}
	wg.Wait()
	ResultInstall.Range(resultInstallWalk)
}

func reportInstall(node, nodeName string) {
	res := SSHConfig.CmdToStringNoLog(node, "cat /tmp/sdwan_install", "")
	if res == "success" {
		if Mtu == "" {
			Mtu = "1360"
		}
		var (
			keyContent string
			crtContent string
		)
		if ServerDomain != "" && ServerKey == "" {
			keyContent = SSHConfig.CmdToStringNoLog(node, fmt.Sprintf("cat /root/.sdwan/ssl/%s/site.key", ServerDomain), "")
			crtContent = SSHConfig.CmdToStringNoLog(node, fmt.Sprintf("cat /root/.sdwan/ssl/%s/site.crt", ServerDomain), "")
		}
		nodeIp, nodePort := GetIpAndPort(node)
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		resp, err := gohttp.NewRequest().
			FormData(map[string]string{
				"action":    "install",
				"ip":        nodeIp,
				"name":      nodeName,
				"mtu":       Mtu,
				"port":      nodePort,
				"user":      SSHConfig.User,
				"pw":        SSHConfig.GetPassword(node),
				"tk":        ServerToken,
				"key":       keyContent,
				"crt":       crtContent,
				"timestamp": timestamp,
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

func resultInstallWalk(key interface{}, value interface{}) bool {
	if value.(string) == "success" {
		logger.Info("[%s] install %s", key, value)
	} else {
		logger.Error("[%s] install %s", key, value)
	}
	return true
}
