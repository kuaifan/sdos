package install

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"github.com/togettoyou/wsc"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var connectRand = ""
var wireguardTransfers = make(map[string]*Transfer)

//BuildWork is
func BuildWork() {
	if os.Getenv("NODE_MODE") == "" {
		logger.Error("System env is error")
		os.Exit(1)
	}
	_ = logger.SetLogger(`{"File":{"filename":"/tmp/sdwan.log","level":"TRAC","daily":true,"maxlines":100000,"maxsize":10,"maxdays":3,"append":true,"permit":"0660"}}`)
	//
	done := make(chan bool)
	ws := wsc.New(ServerUrl)
	// 自定义配置
	ws.SetConfig(&wsc.Config{
		WriteWait:         10 * time.Second,
		MaxMessageSize:    512 * 1024, // 512KB
		MinRecTime:        2 * time.Second,
		MaxRecTime:        30 * time.Second,
		RecFactor:         1.5,
		MessageBufferSize: 1024,
	})
	// 设置回调处理
	ws.OnConnected(func() {
		logger.Debug("OnConnected: ", ws.WebSocket.Url)
		logger.SetWebsocket(ws)
		connectRand = RandString(6)
		// 连接成功后，每50秒发送消息
		go func() {
			r := connectRand
			t := time.NewTicker(50 * time.Second)
			for {
				select {
				case <-t.C:
					if r != connectRand {
						return
					}
					err := timedTask(ws)
					if err == wsc.CloseErr {
						return
					}
				}
			}
		}()
	})
	ws.OnConnectError(func(err error) {
		logger.Debug("OnConnectError: ", err.Error())
	})
	ws.OnDisconnected(func(err error) {
		logger.Debug("OnDisconnected: ", err.Error())
	})
	ws.OnClose(func(code int, text string) {
		logger.Debug("OnClose: ", code, text)
		done <- true
	})
	ws.OnTextMessageSent(func(message string) {
		logger.Debug("OnTextMessageSent: ", message)
	})
	ws.OnBinaryMessageSent(func(data []byte) {
		logger.Debug("OnBinaryMessageSent: ", string(data))
	})
	ws.OnSentError(func(err error) {
		logger.Debug("OnSentError: ", err.Error())
	})
	ws.OnPingReceived(func(appData string) {
		logger.Debug("OnPingReceived: ", appData)
	})
	ws.OnPongReceived(func(appData string) {
		logger.Debug("OnPongReceived: ", appData)
	})
	ws.OnTextMessageReceived(func(message string) {
		logger.Debug("OnTextMessageReceived: ", message)
		handleMessageReceived(ws, message)
	})
	ws.OnBinaryMessageReceived(func(data []byte) {
		logger.Debug("OnBinaryMessageReceived: ", string(data))
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

// 定时任务
func timedTask(ws *wsc.Wsc) error {
	nodeMode := os.Getenv("NODE_MODE")
	sendMessage := ""
	if nodeMode == "manage" {
		// ping 信息
		fileName := "/usr/sdwan/work/ips"
		if !Exists(fileName) {
			logger.Debug("The ips file doesn’t exist")
			return nil
		}
		logger.Debug("start oping...")
		result, err := PingFile(fileName)
		if err != nil {
			logger.Debug("Run oping error: %s", err)
			return nil
		}
		sendMessage = fmt.Sprintf(`{"type":"node","action":"ping","data":"%s"}`, base64Encode(result))
	} else {
		// wg 流量
		result, _, err := RunCommand("-c", "wg show all transfer")
		if err != nil {
			logger.Debug("Run wg show error: %s", err)
			return nil
		}
		value := handleWireguardTransfer(result)
		if value != "" {
			sendMessage = fmt.Sprintf(`{"type":"node","action":"transfer","data":"%s"}`, base64Encode(value))
		}
	}
	if sendMessage != "" {
		return ws.SendTextMessage(sendMessage)
	}
	return nil
}

// 处理Wireguard Transfers
func handleWireguardTransfer(data string) string {
	scanner := bufio.NewScanner(strings.NewReader(data))
	var array []string
	for scanner.Scan() {
		context := strings.Fields(scanner.Text())
		t := &Transfer{}
		t.Name = context[0]
		t.Public = context[1]
		t.Received, _ = strconv.ParseInt(context[2], 10, 64)
		t.Sent, _ = strconv.ParseInt(context[3], 10, 64)
		if t.Received == 0 && t.Sent == 0 {
			continue
		}
		t.ReceivedDiff = t.Received
		t.SentDiff = t.Sent
		//
		key := StringMd5(fmt.Sprintf("%s,%s", t.Name, t.Public))
		if o, ok := wireguardTransfers[key]; ok {
			if t.Received > o.Received {
				t.ReceivedDiff = t.Received - o.Received
			} else if t.Received == o.Received {
				t.ReceivedDiff = 0
			}
			if t.Sent > o.Sent {
				t.SentDiff = t.Sent - o.Sent
			} else if t.Sent == o.Sent {
				t.SentDiff = 0
			}
		}
		wireguardTransfers[key] = t
		//
		if t.ReceivedDiff > 0 || t.SentDiff > 0 {
			val, err := json.Marshal(t)
			if err == nil {
				array = append(array, string(val))
			}
		}
	}
	//
	if len(array) == 0 {
		return ""
	}
	value, err := json.Marshal(array)
	if err != nil {
		return ""
	}
	return string(value)
}

// 处理消息
func handleMessageReceived(ws *wsc.Wsc, message string) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(message), &data); err == nil {
		if data["file"] != nil {
			file, _ := data["file"].(string)
			handleMessageFile(file)
		}
		if data["cmd"] != nil {
			cmd, _ := data["cmd"].(string)
			stdout, stderr, cmderr := handleMessageCmd(cmd)
			if data["callback"] != nil {
				errStr := ""
				if cmderr != nil {
					errStr = cmderr.Error()
				}
				_ = ws.SendTextMessage(fmt.Sprintf(`{"type":"node","action":"cmd","callback":"%s","data":{"stdout":"%s","stderr":"%s","err":"%s"}}`, data["callback"], base64Encode(stdout), base64Encode(stderr), base64Encode(errStr)))
			}
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
			logger.Warn("File empty: %s", fileName)
			continue
		}
		//
		fileKey := StringMd5(fileName)
		contentKey := StringMd5(fileContent)
		md5Value, _ := FileMd5.Load(fileKey)
		if md5Value != nil && md5Value.(string) == contentKey {
			logger.Debug("File same: %s", fileName)
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
				logger.Error("Run nic error: [%s install] %s", fileName, err)
				continue
			} else {
				logger.Info("Run nic success: [%s install]", fileName)
			}
		} else if arr[1] == "exec" {
			_, _, _ = RunCommand("-c", fmt.Sprintf("chmod +x %s", fileName))
			_, _, err = RunCommand(fileName)
			if err != nil {
				logger.Error("Run file error: [%s] %s", fileName, err)
				continue
			} else {
				logger.Info("Run file success: [%s]", fileName)
			}
		} else if arr[1] == "yml" {
			cmd := fmt.Sprintf("cd %s && docker-compose up -d --remove-orphans", fileDir)
			_, _, err = RunCommand("-c", cmd)
			if err != nil {
				logger.Error("Run yml error: [%s] %s", fileName, err)
				continue
			} else {
				logger.Info("Run yml success: [%s]", fileName)
			}
		}
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
				logger.Error("Run nic error: [%s remove] %s", file, err)
			} else {
				logger.Info("Run nic success: [%s remove]", file)
			}
		}
	}
}

// 运行自定义脚本
func handleMessageCmd(data string) (string, string, error) {
	cmd := fmt.Sprintf("cd /usr/sdwan/work && %s", data)
	stdout, stderr, err := RunCommand("-c", cmd)
	if err != nil {
		logger.Error("Run cmd error: %s", err)
	} else {
		logger.Info("Run cmd success: %s", cmd)
	}
	return stdout, stderr, err
}
