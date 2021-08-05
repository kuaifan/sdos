package install

import (
	"fmt"
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
	for _, node := range s.Nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			nodeName := GetRemoteHostName(node)
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/deploy/")
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy/baseUtils", BaseUtils(nodeName, node))
			_ = SSHConfig.CmdAsync(node, "/root/.sdwan/deploy/baseUtils remove")
			_ = SSHConfig.CmdAsync(node, "rm -rf /root/.sdwan/")
			logger.Info(fmt.Sprintf("[%s] Done", node))
		}(node)
	}
	wg.Wait()
}
