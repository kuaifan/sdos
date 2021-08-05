package install

import (
	"sync"
)

//BuildJoin is
func BuildJoin(joinNodes []string) {
	if len(joinNodes) > 0 {
		joinNodesFunc(joinNodes)
	}
}

func joinNodesFunc(joinNodes []string) {
	// 所有node节点
	nodes := joinNodes
	i := &SdosInstaller{
		Nodes: nodes,
	}
	i.JoinNodes()
	NodeIPs = append(NodeIPs, joinNodes...)
}

//JoinNodes is
func (s *SdosInstaller) JoinNodes() {
	var wg sync.WaitGroup
	for _, node := range s.Nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			nodeName := GetRemoteHostName(node)

			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/work/")
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/deploy/")
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy/docker-compose.yml", DockerCompose(nodeName))
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/deploy/baseUtils", BaseUtils(nodeName))
			_ = SSHConfig.CmdAsync(node, "/root/.sdwan/deploy/baseUtils init")
			_ = SSHConfig.CmdAsync(node, "rm -f /root/.sdwan/deploy/baseUtils")
		}(node)
	}
	wg.Wait()
}
