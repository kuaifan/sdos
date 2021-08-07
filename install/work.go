package install

import (
	"encoding/json"
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"github.com/togettoyou/wsc"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//BuildWork is
func BuildWork() {
	if os.Getenv("NODE_MODE") == "" {
		logger.Error("System env is error")
		os.Exit(1)
	}
	done := make(chan bool)
	ws := wsc.New(ServerUrl)
	logger.SetWebsocket(ws)
	// 自定义配置
	ws.SetConfig(&wsc.Config{
		WriteWait:         10 * time.Second,
		MaxMessageSize:    20480,
		MinRecTime:        2 * time.Second,
		MaxRecTime:        30 * time.Second,
		RecFactor:         1.5,
		MessageBufferSize: 1024,
	})
	// 设置回调处理
	ws.OnConnected(func() {
		logger.Info("OnConnected: ", ws.WebSocket.Url)
		// 连接成功后，每60秒发送消息
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
		//logger.Info("OnTextMessageSent: ", message)
	})
	ws.OnBinaryMessageSent(func(data []byte) {
		//logger.Info("OnBinaryMessageSent: ", string(data))
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
		//logger.Info("OnTextMessageReceived: ", message)
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

// 发送ping
func checkPingip(ws *wsc.Wsc) error {
	fileName := "/usr/sdwan/work/ips"
	if !Exists(fileName) {
		return nil
	}
	cmd := fmt.Sprintf("oping -w 2 -c 5 $(cat %s) | sed '/from/d' | sed '/PING/d' | sed '/^$/d'", fileName)
	result, _, err := RunCommand("-c", cmd)
	if err != nil {
		logger.Error("Run oping error: %s", err)
		return nil
	}
	return ws.SendTextMessage(fmt.Sprintf(`{"type":"nodeping","data":"%s"}`, base64Encode(result)))
}

// 处理消息
func handleMessageReceived(message string) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(message), &data); err == nil {
		if data["file"] != nil {
			file, _ := data["file"].(string)
			handleMessageFile(file)
		}
		if data["cmd"] != nil {
			cmd, _ := data["cmd"].(string)
			handleMessageCmd(cmd)
		}
		if data["type"] == "nodenic" && data["nicDir"] != nil && data["nicName"] != nil {
			nicDir, _ := data["nicDir"].(string)
			nicName, _ := data["nicName"].(string)
			handleMessageNic(nicDir, nicName)
		}
	}
}

// 保存文件或运行文件
func handleMessageFile(data string) {
	var err error
	files := strings.Split(data, ",")
	for _, file := range files {
		arr := strings.Split(file, ":")
		if arr[0] == "" {
			continue
		}
		//
		fileContent := ""
		fileName := fmt.Sprintf("/usr/sdwan/work/%s", arr[0])
		fileDir := filepath.Dir(fileName)
		if !Exists(fileDir) {
			err = os.MkdirAll(fileDir, os.ModePerm)
			if err != nil {
				logger.Error("Mkdir error: [%s] %s", fileDir, err)
				continue
			}
		}
		if len(arr) > 2 {
			fileContent = base64Decode(arr[2])
		} else {
			fileContent = base64Decode(arr[1])
		}
		if fileContent == "" {
			logger.Error("File empty: %s", fileName)
			continue
		}
		//
		fileKey := StringMd5(fileName)
		contentKey := StringMd5(fileContent)
		md5Value, _ := FileMd5.Load(fileKey)
		if md5Value != nil && md5Value.(string) == contentKey {
			logger.Warn("File same: %s", fileName)
			continue
		}
		FileMd5.Store(fileKey, contentKey)
		//
		var fileByte = []byte(fileContent)
		err = ioutil.WriteFile(fileName, fileByte, 0666)
		if err != nil {
			logger.Error("WriteFile error: [%s] %s", fileName, err)
			continue
		}
		if arr[1] == "nic" {
			_, _, _ = RunCommand("-c", fmt.Sprintf("chmod +x %s", fileName))
			_, _, err = RunCommand(fileName, "install")
			if err != nil {
				logger.Error("Run file error: [%s install] %s", fileName, err)
				continue
			}
		} else if arr[1] == "exec" {
			_, _, _ = RunCommand("-c", fmt.Sprintf("chmod +x %s", fileName))
			_, _, err = RunCommand(fileName)
			if err != nil {
				logger.Error("Run file error: [%s] %s", fileName, err)
				continue
			}
		} else if arr[1] == "yml" {
			cmd := fmt.Sprintf("cd %s && docker-compose up -d --remove-orphans", fileDir)
			_, _, err = RunCommand("-c", cmd)
			if err != nil {
				logger.Error("Run yml error: [%s] %s", fileName, err)
				continue
			}
		}
	}
}

// 运行自定义脚本
func handleMessageCmd(data string) {
	cmd := fmt.Sprintf("cd /usr/sdwan/work && %s", data)
	_, _, err := RunCommand("-c", cmd)
	if err != nil {
		logger.Error("Run cmd error: %s", err)
	}
}

// 删除没用的网卡
func handleMessageNic(nicDir string, nicName string) {
	path := fmt.Sprintf("/usr/sdwan/work/%s", nicDir)
	nics := strings.Split(nicName, ",")

	files, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		logger.Error(err)
	}
	for i := range files {
		file := files[i]
		name := filepath.Base(file)
		if !InArray(name, nics) {
			_, _, _ = RunCommand("-c", fmt.Sprintf("chmod +x %s", file))
			_, _, err = RunCommand(file, "remove")
			if err != nil {
				logger.Error("Run file error: [%s remove] %s", file, err)
			}
		}
	}
}
