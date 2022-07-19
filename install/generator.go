package install

import (
	"bytes"
	"github.com/kuaifan/sdos/pkg/logger"
	"strings"
	"text/template"
)

func DockerCompose(nodeName string, node string) string {
	var sb strings.Builder
	sb.Write([]byte(dockerCompose))
	var envMap = make(map[string]interface{})
	envMap["SERVER_URL"] = ServerUrl
	envMap["NODE_NAME"] = nodeName
	envMap["NODE_IP"] = RemoveIpPort(node)
	envMap["NODE_TOKEN"] = ServerToken
	envMap["MANAGE_IMAGE"] = ManageImage
	return FromTemplateContent(sb.String(), envMap)
}

func BaseUtils(nodeName string, node string) string {
	var sb strings.Builder
	sb.Write([]byte(baseUtils))
	var envMap = make(map[string]interface{})
	nodeIp, nodePort := GetIpAndPort(node)
	envMap["SERVER_URL"] = ServerUrl
	envMap["SERVER_DOMAIN"] = ServerDomain
	if ServerKey == "" {
		envMap["CERTIFICATE_AUTO"] = "yes"
	} else {
		envMap["CERTIFICATE_AUTO"] = "no"
	}
	envMap["NODE_NAME"] = nodeName
	envMap["NODE_IP"] = nodeIp
	envMap["NODE_PORT"] = nodePort
	envMap["NODE_TOKEN"] = ServerToken
	envMap["NODE_PASSWORD"] = SSHConfig.GetPassword(node)
	envMap["SWAP_FILE"] = SwapFile
	return FromTemplateContent(sb.String(), envMap)
}

func BaseRemoteUtils(node string) string {
	var sb strings.Builder
	sb.Write([]byte(baseRemoteUtils))
	var envMap = make(map[string]interface{})
	nodeIp, nodePort := GetIpAndPort(node)
	envMap["SERVER_URL"] = ServerUrl
	envMap["NODE_IP"] = nodeIp
	envMap["NODE_PORT"] = nodePort
	envMap["NODE_PASSWORD"] = SSHConfig.GetPassword(node)
	return FromTemplateContent(sb.String(), envMap)
}

func BaseHookUtils(node string) string {
	var sb strings.Builder
	sb.Write([]byte(baseHookUtils))
	var envMap = make(map[string]interface{})
	nodeIp, _ := GetIpAndPort(node)
	envMap["NODE_IP"] = nodeIp
	envMap["EXEC_CMD"] = Base64Decode(ExecConfig.Cmd)
	return FromTemplateContent(sb.String(), envMap)
}

func FromTemplateContent(templateContent string, envMap map[string]interface{}) string {
	tmpl, err := template.New("text").Parse(templateContent)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Template parse failed:", err)
		}
	}()
	if err != nil {
		panic(1)
	}
	var buffer bytes.Buffer
	_ = tmpl.Execute(&buffer, envMap)
	return string(buffer.Bytes())
}

func LocalDockerCompose(nodeName string) string {
	var sb strings.Builder
	sb.Write([]byte(dockerCompose))
	var envMap = make(map[string]interface{})
	envMap["SERVER_URL"] = ServerUrl
	envMap["NODE_NAME"] = nodeName
	envMap["NODE_TOKEN"] = ServerToken
	envMap["MANAGE_IMAGE"] = ManageImage
	return FromTemplateContent(sb.String(), envMap)
}

func BaseScriptUtils(nodeName string) string {
	var sb strings.Builder
	sb.Write([]byte(baseUtils))
	var envMap = make(map[string]interface{})
	envMap["SERVER_URL"] = ServerUrl
	envMap["SERVER_DOMAIN"] = ServerDomain
	if ServerKey != "" && ServerCrt != "" {
		envMap["CERTIFICATE_AUTO"] = "no"
	} else {
		envMap["CERTIFICATE_AUTO"] = "yes"
	}
	envMap["NODE_NAME"] = nodeName
	envMap["NODE_TOKEN"] = ServerToken
	envMap["SWAP_FILE"] = SwapFile
	return FromTemplateContent(sb.String(), envMap)
}
