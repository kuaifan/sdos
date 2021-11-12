package install

import (
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"os"
)

//BuildTest is
func BuildTest() {
	nodeMode := os.Getenv("NODE_MODE")
	logger.Debug("NODE_MODE: %s", nodeMode)

	test1()
}

func test2()  {
	fileName := "/tmp/.sdwan/work/nic/wgcenter"
	stdout, stderr, err := RunCommand("-e", fileName, "install")
	logger.Info("stdout: %s", stdout)
	logger.Info("stderr: %s", stderr)
	logger.Info("err: %s", err)
}

func test1() {
	nodeMode := os.Getenv("NODE_MODE")
	sendMessage := ""
	if nodeMode == "manage" {
		// ping 信息
		fileName := "/tmp/.sdwan/work/ips"
		if !Exists(fileName) {
			return
		}
		result, _, err := RunCommand("-c", fmt.Sprintf("oping -w 2 -c 5 $(cat %s) | sed '/from/d' | sed '/PING/d' | sed '/^$/d'", fileName))
		if err != nil {
			logger.Debug("Run oping error: %s", err)
			return
		}
		sendMessage = fmt.Sprintf(`{"type":"node","action":"ping","data":"%s"}`, Base64Encode(result));
	} else {
		// wg 流量
		result, _, err := RunCommand("-c", "wg show all transfer")
		if err != nil {
			logger.Debug("Run wg show error: %s", err)
			return
		}
		value := handleWireguardTransfer(result)
		if value != "" {
			sendMessage = fmt.Sprintf(`{"type":"node","action":"transfer","data":"%s"}`, Base64Encode(value));
		}
	}

	fmt.Println(sendMessage)
}