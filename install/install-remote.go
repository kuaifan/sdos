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

//BuildInstallRemote is
func BuildInstallRemote(installRemoteNodes []string) {
	if len(installRemoteNodes) > 0 {
		installRemoteNodesFunc(installRemoteNodes)
	}
}

func installRemoteNodesFunc(installRemoteNodes []string) {
	nodes := installRemoteNodes
	i := &SdosInstaller{
		Nodes: nodes,
	}
	i.InstallRemoteNodes()
	NodeIPs = append(NodeIPs, installRemoteNodes...)
}

func (s *SdosInstaller) InstallRemoteNodes() {
	var wg sync.WaitGroup
	for _, node := range s.Nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			_ = SSHConfig.CmdAsync(node, "mkdir -p /root/.remote/")
			_ = SSHConfig.SaveFileX(node, "/root/.remote/base", BaseRemoteUtils(node))
			_ = SSHConfig.CmdAsync(node, "/root/.remote/base install")
			installRemoteDone(node)
		}(node)
	}
	wg.Wait()
	ResultInstall.Range(installRemoteWalk)
}

func installRemoteDone(node string) {
	res := SSHConfig.CmdToStringNoLog(node, "cat /tmp/.remote_install", "")
	if res == "success" {
		caContent := SSHConfig.CmdToStringNoLog(node, "cat /etc/docker/certs/ca.pem", "\n")
		certContent := SSHConfig.CmdToStringNoLog(node, "cat /etc/docker/certs/cert.pem", "\n")
		keyContent := SSHConfig.CmdToStringNoLog(node, "cat /etc/docker/certs/key.pem", "\n")
		if !strings.Contains(caContent, "END CERTIFICATE") {
			ResultInstall.Store(node, "read ca error")
			return
		}
		if !strings.Contains(certContent, "END CERTIFICATE") {
			ResultInstall.Store(node, "read cert error")
			return
		}
		if !strings.Contains(keyContent, "END RSA PRIVATE KEY") {
			ResultInstall.Store(node, "read key error")
			return
		}
		nodeIp, nodePort := GetIpAndPort(node)
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		resp, err := gohttp.NewRequest().
			FormData(map[string]string{
				"action":       "install",
				"ip":           nodeIp,
				"port":         nodePort,
				"user":         SSHConfig.User,
				"pw":           SSHConfig.GetPassword(node),
				"ca_content":   caContent,
				"cert_content": certContent,
				"key_content":  keyContent,
				"timestamp":    timestamp,
			}).
			Post(ReportUrl)
		if err != nil || resp == nil {
			ResultInstall.Store(node, err.Error())
			return
		}
		body, errp := resp.GetBodyAsString()
		if errp != nil {
			ResultInstall.Store(node, errp.Error())
			return
		}
		ResultInstall.Store(node, body)
		return
	}
	ResultInstall.Store(node, res)
}

func installRemoteWalk(key interface{}, value interface{}) bool {
	if value.(string) == "success" {
		logger.Info("[%s] install %s", key, value)
	} else {
		if value == "error" {
			value = ""
		}
		if value != "" {
			value = fmt.Sprintf(": %s", value)
		}
		RemoteError(fmt.Sprintf("[%s] install error%s", key, value))
	}
	return true
}

func RemoteError(error string) {
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
