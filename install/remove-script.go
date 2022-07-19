package install

import (
	"fmt"
	"github.com/nahid/gohttp"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

func RemoveScript() {

	done := make(chan bool)
	go DisplayRunning("Removing", done)

	nodeName, _, err := RunCommand("-c", "hostname")
	if err != nil {
		PrintResult(done, fmt.Sprintf("Failed to get hostname: %s\n", err.Error()))
	}
	nodeName = strings.Trim(nodeName, "\n\r ")

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

	// 执行删除命令
	stdOut, stdErr, err := RunCommand("-c", fmt.Sprintf("%s remove", baseFile))
	if err != nil {
		PrintResult(done, fmt.Sprintf("Failed to execute removal command: %s, %s\n", stdOut, stdErr))
		return
	}

	// 删除.sdwan目录
	err = os.RemoveAll(homeDir)
	if err != nil {
		PrintResult(done, fmt.Sprintf("Failed to delete base dir: %s, %s\n", stdOut, stdErr))
		return
	}

	reportRemoveScript(done, nodeName)
}

func reportRemoveScript(done chan bool, nodeName string) {

	result, _ := ioutil.ReadFile("/tmp/.sdwan_install")
	res := strings.Trim(string(result), "\r\n")

	if res == "success" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		resp, err := gohttp.NewRequest().
			FormData(map[string]string{
				"action":    "remove",
				"name":      nodeName,
				"ip":        "",
				"timestamp": timestamp,
				"tk":        ServerToken,
			}).
			Post(ReportUrl)
		if err != nil || resp == nil {
			PrintResult(done, fmt.Sprintf("Faild to report node removal: %s", err))
		} else {
			body, err := resp.GetBodyAsString()
			if err != nil {
				PrintResult(done, fmt.Sprintf("Failed to report node removal, occurred an error when get body: %s", err))
			} else {
				if body != "success" {
					PrintResult(done, fmt.Sprintf("Failed to report node removal, got response: %s", body))
				} else {
					done <- true
					time.Sleep(500 * time.Microsecond)
					PrintSuccess("Remove success!")
				}
			}
		}
	} else {
		PrintResult(done, "Remove failed!")
	}
}
