package install

import (
	"encoding/json"
	"fmt"
	"github.com/togettoyou/wsc"
	"github.com/wonderivan/logger"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//BuildWork is
func BuildWork() {
	done := make(chan bool)
	ws := wsc.New(ServerUrl)
	// 自定义配置
	ws.SetConfig(&wsc.Config{
		WriteWait: 10 * time.Second,
		MaxMessageSize: 20480,
		MinRecTime: 2 * time.Second,
		MaxRecTime: 30 * time.Second,
		RecFactor: 1.5,
		MessageBufferSize: 1024,
	})
	// 设置回调处理
	ws.OnConnected(func() {
		logger.Info("OnConnected: ", ws.WebSocket.Url)
		// 连接成功后，测试每60秒发送消息
		go func() {
			t := time.NewTicker(60 * time.Second)
			for {
				select {
				case <-t.C:
					err := checkPingip(ws)
					if err == wsc.CloseErr {
						return
					}
				}
			}
		}()
	})
	ws.OnConnectError(func(err error) {
		logger.Info("OnConnectError: ", err.Error())
	})
	ws.OnDisconnected(func(err error) {
		logger.Info("OnDisconnected: ", err.Error())
	})
	ws.OnClose(func(code int, text string) {
		logger.Info("OnClose: ", code, text)
		done <- true
	})
	ws.OnTextMessageSent(func(message string) {
		logger.Info("OnTextMessageSent: ", message)
	})
	ws.OnBinaryMessageSent(func(data []byte) {
		logger.Info("OnBinaryMessageSent: ", string(data))
	})
	ws.OnSentError(func(err error) {
		logger.Info("OnSentError: ", err.Error())
	})
	ws.OnPingReceived(func(appData string) {
		logger.Info("OnPingReceived: ", appData)
	})
	ws.OnPongReceived(func(appData string) {
		logger.Info("OnPongReceived: ", appData)
	})
	ws.OnTextMessageReceived(func(message string) {
		logger.Info("OnTextMessageReceived: ", message)
		handleMessageReceived(message)
	})
	ws.OnBinaryMessageReceived(func(data []byte) {
		logger.Info("OnBinaryMessageReceived: ", string(data))
	})
	// 开始连接
	go ws.Connect()
	for {
		select {
		case <-done:
			return
		}
	}
}

func checkPingip(ws *wsc.Wsc) error {
	fileName := "/usr/sdwan/assets/ips"
	if !Exists(fileName) {
		return nil
	}
	cmd := fmt.Sprintf("oping -w 2 -c 5 $(cat %s) | sed '/from/d' | sed '/PING/d' | sed '/^$/d'", fileName)
	result, _, err := RunShellInSystem(cmd)
	if err != nil {
		logger.Error("Run oping error: %s", err)
		return nil
	}
	return ws.SendTextMessage(fmt.Sprintf("{\"type\":\"nodeping\",\"data\":\"%s\"}", base64Encode(result)))
}

func handleMessageReceived(message string) {
	//json str 转map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(message), &data); err == nil {
		if data["file"] != nil {
			files := strings.Split(data["file"].(string), ",")
			for _, file := range files {
				arr := strings.Split(file, ":")
				if arr[0] == "" {
					return
				}
				fileContent := ""
				fileName := fmt.Sprintf("/usr/sdwan/assets/%s", arr[0])
				fileDir := filepath.Dir(fileName)
				if !Exists(fileDir) {
					err = os.MkdirAll(fileDir, os.ModePerm)
					if err != nil {
						logger.Error("Mkdir error: [%s] %s", fileDir, err)
						return
					}
				}
				if len(arr) > 2 {
					fileContent = base64Decode(arr[2])
				} else {
					fileContent = base64Decode(arr[1])
				}
				if fileContent == "" {
					return
				}
				var fileData = []byte(fileContent)
				err = ioutil.WriteFile(fileName, fileData, 0666)
				if err != nil {
					logger.Error("WriteFile error: [%s] %s", fileName, err)
					return
				}
				if arr[1] == "exec" {
					_, _, _ = RunShellInSystem(fmt.Sprintf("chmod +x %s", fileName))
					_, _, err = RunShellInFile(fileName)
					if err != nil {
						logger.Error("Run file error: [%s] %s", fileName, err)
						return
					}
				} else if arr[1] == "yml" {
					cmd := fmt.Sprintf("cd %s && docker-compose up -d", fileDir)
					_, _, err = RunShellInSystem(cmd)
					if err != nil {
						logger.Error("Run yml error: [%s] %s", fileName, err)
						return
					}
				}
			}
		}
		if data["cmd"] != nil {
			cmd := fmt.Sprintf("cd /usr/sdwan/assets && %s", data["cmd"])
			_, _, err = RunShellInSystem(cmd)
			if err != nil {
				logger.Error("Run cmd error: %s", err)
				return
			}
		}
	}
}
