package install

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nahid/gohttp"
)

var (
	homeDir           = "/root/.sdwan/"
	baseFile          = "/root/.sdwan/base"
	dockerComposeFile = "/root/.sdwan/docker-compose.yml"
)

func ScriptInstallNode() {

	done := make(chan bool)
	go DisplayRunning("Installing", done)

	nodeName, _, err := RunCommand("-c", "hostname")
	if err == nil {
		nodeName = strings.Trim(nodeName, "\n\r ")
	}

	if InReset {
		// 创建目录
		err := Mkdir(homeDir, 0755)
		if err != nil {
			PrintResult(done, fmt.Sprintf("Failed to create home dir: %s\n", err.Error()))
			return
		}
		// 生成base文件，并赋可执行权限
		WriteFile(baseFile, BaseScriptUtils(nodeName))
		err = os.Chmod(baseFile, 0755)
		if err != nil {
			PrintResult(done, fmt.Sprintf("Failed to create base file: %s\n", err.Error()))
			return
		}
		// 执行删除命令
		stdOut, stdErr, err := RunCommand("-c", fmt.Sprintf("%s remove", baseFile))
		if err != nil {
			PrintResult(done, fmt.Sprintf("Failed to execute remove command: %s %s\n", stdOut, stdErr))
			return
		}
		// 删除目录
		_ = os.RemoveAll(homeDir)
	}
	// 创建目录
	err = Mkdir(homeDir, 0755)
	if err != nil {
		PrintResult(done, fmt.Sprintf("Failed to create home dir: %s\n", err.Error()))
		return
	}
	// 保存证书和秘钥
	if ServerDomain != "" && ServerKey != "" && ServerCrt != "" {
		sslDir := fmt.Sprintf("/root/.sdwan/ssl/%s/", ServerDomain)
		err := Mkdir(sslDir, 0755)
		if err != nil {
			PrintResult(done, fmt.Sprintf("Failed to create ssl dir: %v\n", err))
			return
		}
		keyFile := fmt.Sprintf("/root/.sdwan/ssl/%s/site.key", ServerDomain)
		certFile := fmt.Sprintf("/root/.sdwan/ssl/%s/site.crt", ServerDomain)
		WriteFile(keyFile, ServerKey)
		WriteFile(certFile, ServerCrt)
	}

	// 生成docker-compose.yml
	WriteFile(dockerComposeFile, LocalDockerCompose(nodeName))

	// 生成base文件，并赋可执行权限
	err = ioutil.WriteFile(baseFile, []byte(BaseScriptUtils(nodeName)), 0755)
	if err != nil {
		PrintResult(done, fmt.Sprintf("Failed to create base file: %s\n", err.Error()))
		return
	}

	// 执行安装命令
	stdOut, stdErr, err := RunCommand("-c", fmt.Sprintf("%s install", baseFile))
	if err != nil {
		PrintResult(done, fmt.Sprintf("Failed to execute installation command: %s, %s\n", stdOut, stdErr))
		return
	}

	scriptInstallDone(done, nodeName)

}

func scriptInstallDone(done chan bool, nodeName string) {

	result, _ := ioutil.ReadFile("/tmp/.sdwan_install")
	res := strings.Trim(string(result), "\r\n")

	if res == "success" {
		if Mtu == "" {
			Mtu = "1360"
		}
		var (
			keyContent      = ""
			crtContent      = ""
			certificateAuto = ""
		)
		if ServerDomain != "" {

			keyFile := fmt.Sprintf("/root/.sdwan/ssl/%s/site.key", ServerDomain)
			keyContentB, _ := ioutil.ReadFile(keyFile)
			keyContent := strings.ReplaceAll(string(keyContentB), "\r\n", "\n")

			crtFile := fmt.Sprintf("/root/.sdwan/ssl/%s/site.crt", ServerDomain)
			crtContentB, _ := ioutil.ReadFile(crtFile)
			crtContent := strings.ReplaceAll(string(crtContentB), "\r\n", "\n")

			// 证书秘钥内容错误
			if !strings.Contains(keyContent, "PRIVATE KEY") {
				PrintResult(done, "Key content format error")
				return
			}
			// 证书内容错误
			if !strings.Contains(crtContent, "END CERTIFICATE") {
				PrintResult(done, "Crt content format error")
				return
			}
		}
		if ServerKey != "" && ServerCrt != "" {
			certificateAuto = "no"
		} else {
			certificateAuto = "yes"
		}

		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		resp, err := gohttp.NewRequest().
			FormData(map[string]string{
				"action":           "install",
				"name":             nodeName,
				"mtu":              Mtu,
				"ip":               "",
				"port":             "",
				"user":             "",
				"pw":               "",
				"tk":               ServerToken,
				"domain":           ServerDomain,
				"domain_key":       keyContent,
				"domain_crt":       crtContent,
				"certificate_auto": certificateAuto,
				"timestamp":        timestamp,
			}).
			Post(ReportUrl)
		if err != nil || resp == nil {
			PrintResult(done, fmt.Sprintf("Failed to report node installation: %s\n", err.Error()))
			return
		}

		body, err := resp.GetBodyAsString()
		if err != nil {
			PrintResult(done, fmt.Sprintf("Failed to report node installation, occurred an error when get response body: %s\n", err.Error()))
			return
		} else {
			if body != "success" {
				PrintResult(done, fmt.Sprintf("Failed to report node installation, got response: %s\n", body))
			} else {
				done <- true
				time.Sleep(500 * time.Microsecond)
				PrintSuccess("Install success!")
			}
		}
	} else {
		PrintResult(done, res)
	}
}

func PrintResult(done chan bool, error string) {
	done <- true
	time.Sleep(500 * time.Microsecond)
	PrintError(error)
}

func GetIp() (ip string) {
	resp, _ := gohttp.NewRequest().Headers(map[string]string{
		"User-Agent": "curl/7.79.1",
	}).Get("http://ip.sb")

	body, _ := resp.GetBodyAsString()
	ip = strings.Trim(body, "\n\r ")
	return
}
