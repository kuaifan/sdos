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

//BuildFreessl is
func BuildFreessl(freesslNodes []string) {
	if len(freesslNodes) > 0 {
		freesslNodesFunc(freesslNodes)
	}
}

func freesslNodesFunc(freesslNodes []string) {
	// 所有node节点
	nodes := freesslNodes
	i := &SdosInstaller{
		Nodes: nodes,
	}
	i.FreesslNodes()
	NodeIPs = append(NodeIPs, freesslNodes...)
}

//FreesslNodes is
func (s *SdosInstaller) FreesslNodes() {
	var wg sync.WaitGroup
	for _, node := range s.Nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			nodeName := GetRemoteHostName(node)
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.sdwan/")
			_ = SSHConfig.SaveFileAndChmodX(node, "/root/.sdwan/base", BaseUtils(nodeName, node))
			_ = SSHConfig.CmdAsync(node, "/root/.sdwan/base freessl")
			reportFreessl(node, nodeName)
		}(node)
	}
	wg.Wait()
	ResultFreessl.Range(resultFreesslWalk)
}

func reportFreessl(node string, nodeName string) {
	res := SSHConfig.CmdToStringNoLog(node, "cat /tmp/.sdwan_install", "")
	if res == "success" {
		keyContent := SSHConfig.CmdToStringNoLog(node, fmt.Sprintf("cat /root/.sdwan/ssl/%s/site.key", ServerDomain), "\n")
		crtContent := SSHConfig.CmdToStringNoLog(node, fmt.Sprintf("cat /root/.sdwan/ssl/%s/site.crt", ServerDomain), "\n")
		if !strings.Contains(keyContent, "PRIVATE KEY") {
			ResultFreessl.Store(node, "read key error")
			return
		}
		if !strings.Contains(crtContent, "END CERTIFICATE") {
			ResultFreessl.Store(node, "read crt error")
			return
		}
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		resp, err := gohttp.NewRequest().
			FormData(map[string]string{
				"action":           "freessl",
				"name":             nodeName,
				"ip":               RemoveIpPort(node),
				"domain":           ServerDomain,
				"domain_key":       keyContent,
				"domain_crt":       crtContent,
				"certificate_auto": "yes",
				"timestamp":        timestamp,
			}).
			Post(ReportUrl)
		if err != nil || resp == nil {
			logger.Error("[%s] freessl error %s", node, err)
		} else {
			body, errp := resp.GetBodyAsString()
			if errp != nil {
				logger.Error("[%s] freessl failed %s", node, errp)
			} else {
				ResultFreessl.Store(node, body)
			}
		}
	} else {
		ResultFreessl.Store(node, res)
	}
}

func resultFreesslWalk(key interface{}, value interface{}) bool {
	if value.(string) == "success" {
		logger.Info("[%s] freessl %s", key, value)
	} else {
		logger.Error("[%s] freessl %s", key, value)
	}
	return true
}
