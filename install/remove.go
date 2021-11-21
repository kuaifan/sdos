package install

import (
	"github.com/kuaifan/sdos/pkg/logger"
	"github.com/nahid/gohttp"
	"strconv"
	"sync"
	"time"
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
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/")
			_ = SSHConfig.SaveFile(node, "/root/.sdwan/base", BaseUtils(nodeName, node))
			_ = SSHConfig.CmdAsync(node, "/root/.sdwan/base remove")
			_ = SSHConfig.CmdAsync(node, "rm -rf /root/.sdwan/")
			reportRemove(node, nodeName)
		}(node)
	}
	wg.Wait()
	ResultRemove.Range(resultRemoveWalk)
}

func reportRemove(node string, nodeName string) {
	res := SSHConfig.CmdToStringNoLog(node, "cat /tmp/.sdwan_install", "")
	if res == "success" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		resp, err := gohttp.NewRequest().
			FormData(map[string]string{
				"action":    "remove",
				"name":      nodeName,
				"ip":        RemoveIpPort(node),
				"timestamp": timestamp,
			}).
			Post(ReportUrl)
		if err != nil || resp == nil {
			logger.Error("[%s] remove error %s", node, err)
		} else {
			body, errp := resp.GetBodyAsString()
			if errp != nil {
				logger.Error("[%s] remove failed %s", node, errp)
			} else {
				ResultRemove.Store(node, body)
			}
		}
	} else {
		ResultRemove.Store(node, res)
	}
}

func resultRemoveWalk(key interface{}, value interface{}) bool {
	if value.(string) == "success" {
		logger.Info("[%s] remove %s", key, value)
	} else {
		logger.Error("[%s] remove %s", key, value)
	}
	return true
}