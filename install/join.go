package install

import (
	"github.com/wonderivan/logger"
	"sync"
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
	var result sync.Map
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
			result.Store(node, SSHConfig.CmdToStringNoLog(node, "cat /tmp/sdwan_install", ""))
		}(node)
	}
	wg.Wait()
	result.Range(resultJoinWalk)
}

func resultJoinWalk(key interface{}, value interface{}) bool {
	if value.(string) == "success" {
		logger.Info("[%s] %s", key, value)
	} else {
		logger.Error("[%s] %s", key, value)
	}
	return true
}
