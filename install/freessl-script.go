package install

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/nahid/gohttp"
)

func FreeSSLNode() {

	done := make(chan bool)
	go DisplayRunning("Generating", done)

	nodeName, _, err := RunCommand("-c", "hostname")
	if err == nil {
		nodeName = strings.Trim(nodeName, "\n\r ")
	}

	// 创建目录
	err = Mkdir(homeDir, 0755)
	if err != nil {
		PrintResult(done, fmt.Sprintf("Failed to create home dir: %s\n", err.Error()))
		return
	}

	// 生成base文件，并赋可执行权限
	err = ioutil.WriteFile(baseFile, []byte(BaseScriptUtils(nodeName)), 0755)
	if err != nil {
		PrintResult(done, fmt.Sprintf("Failed to create base file: %s\n", err.Error()))
		return
	}

	// 执行安装命令
	stdOut, stdErr, err := RunCommand("-c", fmt.Sprintf("%s freessl", baseFile))
	if err != nil {
		PrintResult(done, fmt.Sprintf("Failed to generate certificate: %s, %s\n", stdOut, stdErr))
		return
	}

	reportFreeSSLScript(done, nodeName)
}

func reportFreeSSLScript(done chan bool, nodeName string) {

	result, _ := ioutil.ReadFile("/tmp/.sdwan_install")
	res := strings.Trim(string(result), "\r\n")

	if res == "success" {

		keyContentB, _ := ioutil.ReadFile(fmt.Sprintf("/root/.sdwan/ssl/%s/site.key", ServerDomain))
		crtContentB, _ := ioutil.ReadFile(fmt.Sprintf("/root/.sdwan/ssl/%s/site.crt", ServerDomain))

		keyContent := string(keyContentB)
		crtContent := string(crtContentB)

		if !strings.Contains(keyContent, "PRIVATE KEY") {
			PrintResult(done, "Failed to read key file")
			return
		}
		if !strings.Contains(crtContent, "END CERTIFICATE") {
			PrintResult(done, "Failed to read crt file")
			return
		}

		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		resp, err := gohttp.NewRequest().
			FormData(map[string]string{
				"action":           "freessl",
				"name":             nodeName,
				"ip":               "",
				"domain":           ServerDomain,
				"domain_key":       keyContent,
				"domain_crt":       crtContent,
				"certificate_auto": "yes",
				"timestamp":        timestamp,
			}).
			Post(ReportUrl)
		if err != nil || resp == nil {
			PrintResult(done, fmt.Sprintf("Failed to report certificate generation: %s", err.Error()))
		} else {
			body, err := resp.GetBodyAsString()
			if err != nil {
				PrintResult(done, fmt.Sprintf("Failed to report certificate generation, occurred an error when get response body: %s", err.Error()))
			} else {
				if body != "success" {
					PrintResult(done, fmt.Sprintf("Failed to report certificate generation, got response: %s", body))
				} else {
					PrintResult(done, "Certificate Generated!")
				}
			}
		}
	} else {
		PrintResult(done, "Failed to Generate Certificate!")
	}
}
