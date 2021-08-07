package install

import (
	"github.com/kuaifan/sdos/pkg/logger"
	"github.com/nahid/gohttp"
	"strconv"
	"sync"
	"time"
)

//BuildJoin is
func BuildJoin(joinNodes []string) {
	if len(joinNodes) > 0 {
		joinNodesFunc(joinNodes)
	}
}

func joinNodesFunc(joinNodes []string) {
	nodes := joinNodes
	i := &SdosInstaller{
		Nodes: nodes,
	}
	i.JoinNodes()
	NodeIPs = append(NodeIPs, joinNodes...)
}

func (s *SdosInstaller) JoinNodes() {
	var wg sync.WaitGroup
	for _, node := range s.Nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			nodeName := GetRemoteHostName(node)
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/work/")
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/deploy/")
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy/docker-compose.yml", DockerCompose(nodeName, node))
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy/install", BaseUtils(nodeName, node))
			_ = SSHConfig.CmdAsync(node, "/root/.sdwan/deploy/install join")
			publishJoin(node, nodeName)
		}(node)
	}
	wg.Wait()
	resultJoin.Range(resultJoinWalk)
}

func publishJoin(node, nodeName string) {
	res := SSHConfig.CmdToStringNoLog(node, "cat /tmp/sdwan_install", "")
	if res == "success" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		resp, err := gohttp.NewRequest().
			FormData(map[string]string{
				"action":    "join",
				"name":      nodeName,
				"ip":        RemoveIpPort(node),
				"pw":        SSHConfig.GetPassword(node),
				"tk":        ServerToken,
				"timestamp": timestamp,
			}).
			Post(ServerUrl)
		if err != nil || resp == nil {
			logger.Error("[%s] join error %s", node, err)
		} else {
			body, errp := resp.GetBodyAsString()
			if errp != nil {
				logger.Error("[%s] join failed %s", node, errp)
			} else {
				resultJoin.Store(node, body)
			}
		}
	} else {
		resultJoin.Store(node, res)
	}
}

func resultJoinWalk(key interface{}, value interface{}) bool {
	if value.(string) == "success" {
		logger.Info("[%s] join %s", key, value)
	} else {
		logger.Error("[%s] join %s", key, value)
	}
	return true
}
