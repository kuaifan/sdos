package install

import (
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
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/work/")
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/deploy/")
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy/docker-compose.yml", DockerCompose(nodeName, node))
			if InReset {
				_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy/utils", BaseUtils(nodeName, node))
				_ = SSHConfig.CmdAsync(node, "/root/.sdwan/deploy/utils remove")
			}
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy/utils", BaseUtils(nodeName, node))
			_ = SSHConfig.CmdAsync(node, "/root/.sdwan/deploy/utils install")
			publishInstall(node, nodeName)
		}(node)
	}
	wg.Wait()
	ResultInstall.Range(resultInstallWalk)
}

func publishInstall(node, nodeName string) {
	res := SSHConfig.CmdToStringNoLog(node, "cat /tmp/sdwan_install", "")
	if res == "success" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		resp, err := gohttp.NewRequest().
			FormData(map[string]string{
				"action":    "install",
				"name":      nodeName,
				"ip":        RemoveIpPort(node),
				"pw":        SSHConfig.GetPassword(node),
				"tk":        ServerToken,
				"timestamp": timestamp,
			}).
			Post(ServerUrl)
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
