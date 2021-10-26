package install

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/kuaifan/sdos/pkg/logger"
	"github.com/togettoyou/wsc"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	connectRand string
	wireguardTransfers = make(map[string]*Wireguard)

	monitorRand string
	monitorRecord = make(map[string]*Monitor)
)

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
		// docker-compose
		fileName := fmt.Sprintf("/usr/sdwan/work/docker-compose.yml")
		if Exists(fileName) {
			cmd := fmt.Sprintf("cd %s && docker-compose up -d --remove-orphans", filepath.Dir(fileName))
			_, _, _ = RunCommand("-c", cmd)
		}
		// 公网 ping
		sendErr := pingFileAndSend(ws, "/usr/sdwan/work/ips", "")
		if sendErr != nil {
			return sendErr
		}
		// 专线 ping
		dirPath := "/usr/sdwan/work/vpc_ip"
		if IsDir(dirPath) {
			files := GetIpsFiles(dirPath)
			if files != nil {
				for _, file := range files {
					go func(file string) {
						_ = pingFileAndSend(ws, fmt.Sprintf("%s/%s", dirPath, file), file)
					}(file)
				}
			}
		}
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

// ping 文件并发送
func pingFileAndSend(ws *wsc.Wsc, fileName string, source string) error {
	if !Exists(fileName) {
		logger.Debug("File no exist [%s]", fileName)
		return nil
	}
	logger.Debug("Start ping [%s]", fileName)
	result, err := PingFile(fileName, source)
	if err != nil {
		logger.Debug("Ping error [%s]: %s", fileName, err)
		return nil
	}
	sendMessage := fmt.Sprintf(`{"type":"node","action":"ping","data":"%s","source":"%s"}`, base64Encode(result), source)
	return ws.SendTextMessage(sendMessage)
}

// 处理Wireguard Transfers
func handleWireguardTransfer(data string) string {
	scanner := bufio.NewScanner(strings.NewReader(data))
	var array []string
	for scanner.Scan() {
		context := strings.Fields(scanner.Text())
		t := &Wireguard{}
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
	if ok := json.Unmarshal([]byte(message), &data); ok == nil {
		content, _ := data["content"].(string)
		if data["type"] == "nodework:file" {
			// 保存文件
			handleMessageFile(content)
		} else if data["type"] == "nodework:nic" {
			// 保存文件
			handleMessageFile(content)
			// 删除没用的网卡
			if data["dir"] != nil && data["names"] != nil {
				dir, _ := data["dir"].(string)
				names, _ := data["names"].(string)
				handleDeleteUnusedNic(dir, names)
			}
		} else if data["type"] == "nodework:cmd" {
			// 执行命令
			stdout, stderr, err := handleMessageCmd(content, data["log"] != "no")
			if data["callback"] != nil {
				cmderr := ""
				if err != nil {
					cmderr = err.Error()
				}
				err = ws.SendTextMessage(fmt.Sprintf(`{"type":"node","action":"cmd","callback":"%s","data":{"stdout":"%s","stderr":"%s","err":"%s"}}`, data["callback"], base64Encode(stdout), base64Encode(stderr), base64Encode(cmderr)))
				if err != nil {
					logger.Debug("Send cmd callback error: %s", err)
				}
			}
		} else if data["type"] == "nodework:monitorip" {
			// 监听ip状态
			monitorRand = RandString(6)
			go handleMessageMonitorIp(ws, monitorRand, content)
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
			logger.Info("Run nic start: [%s] [%s install]", contentKey, fileName)
			_, _, _ = RunCommand("-c", fmt.Sprintf("chmod +x %s", fileName))
			_, _, err = RunCommand(fileName, "install")
			if err != nil {
				logger.Error("Run nic error: [%s] [%s install] %s", contentKey, fileName, err)
				continue
			} else {
				logger.Info("Run nic success: [%s] [%s install]", contentKey, fileName)
			}
		} else if arr[1] == "exec" {
			logger.Info("Run file start: [%s] [%s]", contentKey, fileName)
			_, _, _ = RunCommand("-c", fmt.Sprintf("chmod +x %s", fileName))
			_, _, err = RunCommand(fileName)
			if err != nil {
				logger.Error("Run file error: [%s] [%s] %s", contentKey, fileName, err)
				continue
			} else {
				logger.Info("Run file success: [%s] [%s]", contentKey, fileName)
			}
		} else if arr[1] == "yml" {
			logger.Info("Run yml start: [%s] [%s]", contentKey, fileName)
			cmd := fmt.Sprintf("cd %s && docker-compose up -d --remove-orphans", fileDir)
			_, _, err = RunCommand("-c", cmd)
			if err != nil {
				logger.Error("Run yml error: [%s] [%s] %s", contentKey, fileName, err)
				continue
			} else {
				logger.Info("Run yml success: [%s] [%s]", contentKey, fileName)
			}
		}
	}
}

// 删除没用的网卡
func handleDeleteUnusedNic(nicDir string, nicName string) {
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
func handleMessageCmd(data string, addLog bool) (string, string, error) {
	cmd := fmt.Sprintf("cd /usr/sdwan/work && %s", data)
	stdout, stderr, err := RunCommand("-c", cmd)
	if addLog {
		if err != nil {
			logger.Error("Run cmd error: [%s] %s", data, err)
		} else {
			logger.Info("Run cmd success: [%s]", data)
		}
	}
	return stdout, stderr, err
}

// 监听ip通或不通上报（ping值变化超过5也上报）
func handleMessageMonitorIp(ws *wsc.Wsc, rand string, content string) {
	var fileText []string
	array := strings.Split(content, ",")
	for _, value := range array {
		arr := strings.Split(value, ":")
		address := net.ParseIP(arr[0])
		if address == nil {
			continue
		}
		ip := address.String()
		if len(arr) >= 4 {
			state := arr[1]
			ping, _ := strconv.ParseFloat(arr[2], 64)
			unix, _ := strconv.ParseInt(arr[3], 10, 64)
			monitorRecord[ip] = &Monitor{State: state, Ping: ping, Unix: unix}
		}
		fileText = append(fileText, ip)
	}
	fileName := fmt.Sprintf("/tmp/monitorip_%s.txt", rand)
	var fileByte = []byte(strings.Join(fileText, "\n"))
	err := ioutil.WriteFile(fileName, fileByte, 0666)
	if err != nil {
		logger.Error("[MonitorIp] [%s] WriteFile error: [%s] %s", rand, fileName, err)
		return
	}
	//
	for {
		if rand != monitorRand {
			_ = os.Remove(fileName)
			logger.Debug("[MonitorIp] [%s] Jump thread", rand)
			return
		}
		result, pingErr := PingFileMap(fileName, "", 2000, 4)
		if pingErr != nil {
			logger.Debug("[MonitorIp] [%s] Ping error: %s", rand, pingErr)
			time.Sleep(2 * time.Second)
			continue
		}
		var state string
		var record *Monitor
		var report = make(map[string]*Monitor)
		var unix = time.Now().Unix()
		for ip, ping := range result {
			state = "reject"
			if ping > 0 {
				state = "accept"	// ping值大于0表示线路通
			}
			record = monitorRecord[ip]
			/**
			1、记录没有
			2、状态改变（通 不通 发生改变
			3、大于10分钟
			4、大于10秒钟且（与上次ping值相差大于等于50或与上次相差1.1倍）
			 */
			if record == nil || record.State != state || unix - record.Unix >= 600 || (unix - record.Unix >= 10 && ComputePing(record.Ping, ping)) {
				report[ip] = &Monitor{State: state, Ping: ping, Unix: unix}
				monitorRecord[ip] = report[ip]
			}
		}
		if len(report) > 0 {
			reportValue, jsonErr := json.Marshal(report)
			if jsonErr != nil {
				logger.Debug("[MonitorIp] [%s] Marshal error: %s", rand, jsonErr)
				for ip := range report {
					delete(monitorRecord, ip)
				}
			}
			if ws != nil {
				sendErr := ws.SendTextMessage(fmt.Sprintf(`{"type":"node","action":"monitorip","data":"%s"}`, base64Encode(string(reportValue))))
				if sendErr != nil {
					logger.Debug("[MonitorIp] [%s] Send error: %s", rand, sendErr)
					for ip := range report {
						delete(monitorRecord, ip)
					}
				}
			} else {
				logger.Debug("[MonitorIp] record: %s", string(reportValue))
			}
		}
	}
}
