package install

import (
	"github.com/wonderivan/logger"
	"sync"
)

//BuildRemove is
func BuildRemove(removeNodes []string) {
	if len(removeNodes) > 0 {
		removeNodesFunc(removeNodes)
	}
}

func removeNodesFunc(removeNodes []string) {
	// 所有node节点
	nodes := removeNodes
	i := &SdosInstaller{
		Nodes: nodes,
	}
	i.RemoveNodes()
	NodeIPs = append(NodeIPs, removeNodes...)
}

//RemoveNodes is
func (s *SdosInstaller) RemoveNodes() {
	var wg sync.WaitGroup
	var result sync.Map
	for _, node := range s.Nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			nodeName := GetRemoteHostName(node)
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/deploy/")
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy/install", BaseUtils(nodeName, node))
			_ = SSHConfig.CmdAsync(node, "/root/.sdwan/deploy/install remove")
			_ = SSHConfig.CmdAsync(node, "rm -rf /root/.sdwan/")
			result.Store(node, SSHConfig.CmdToStringNoLog(node, "cat /tmp/sdwan_install", ""))
		}(node)
	}
	wg.Wait()
	result.Range(resultRemoveWalk)
}

func resultRemoveWalk(key interface{}, value interface{}) bool {
	if value.(string) == "success" {
		logger.Info("[%s] %s", key, value)
	} else {
		logger.Error("[%s] %s", key, value)
	}
	return true
}