package install

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/kuaifan/sdos/pkg/sys"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kuaifan/sdos/pkg/logger"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	gopsnet "github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/v3/process"
)

// YmdHis 返回示例：2021-08-05 00:00:01
func YmdHis() string {
	timeObj := time.Now()
	year := timeObj.Year()
	month := timeObj.Month()
	day := timeObj.Day()
	hour := timeObj.Hour()
	minute := timeObj.Minute()
	second := timeObj.Second()
	//注意：%02d 中的 2 表示宽度，如果整数不够 2 列就补上 0
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", year, month, day, hour, minute, second)
}

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

// VersionToInt v1.15.6  => 115
func VersionToInt(version string) int {
	// v1.15.6  => 1.15.6
	version = strings.Replace(version, "v", "", -1)
	versionArr := strings.Split(version, ".")
	if len(versionArr) >= 2 {
		versionStr := versionArr[0] + versionArr[1]
		if i, err := strconv.Atoi(versionStr); err == nil {
			return i
		}
	}
	return 0
}

// VersionToIntAll v1.19.1 ==> 1191
func VersionToIntAll(version string) int {
	version = strings.Replace(version, "v", "", -1)
	arr := strings.Split(version, ".")
	if len(arr) >= 3 {
		str := arr[0] + arr[1] + arr[2]
		if i, err := strconv.Atoi(str); err == nil {
			return i
		}
	}
	return 0
}

// IpFormat is
func IpFormat(host string) string {
	ipAndPort := strings.Split(host, ":")
	if len(ipAndPort) != 2 {
		logger.Warn("invalied host fomat [%s], must like 172.0.0.2:22", host)
		os.Exit(1)
	}
	return ipAndPort[0]
}

// RandString 生成随机字符串
func RandString(len int) string {
	var r *rand.Rand
	r = rand.New(rand.NewSource(time.Now().Unix()))
	bs := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bs[i] = byte(b)
	}
	return string(bs)
}

// RandNum 生成随机数
func RandNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

// Cmp compares two IPs, returning the usual ordering:
// a < b : -1
// a == b : 0
// a > b : 1
func Cmp(a, b net.IP) int {
	aa := ipToInt(a)
	bb := ipToInt(b)

	if aa == nil || bb == nil {
		logger.Warn("ip range %s-%s is invalid", a.String(), b.String())
		os.Exit(-1)
	}
	return aa.Cmp(bb)
}

func ipToInt(ip net.IP) *big.Int {
	if v := ip.To4(); v != nil {
		return big.NewInt(0).SetBytes(v)
	}
	return big.NewInt(0).SetBytes(ip.To16())
}

func intToIP(i *big.Int) net.IP {
	return net.IP(i.Bytes())
}

func stringToIP(i string) net.IP {
	return net.ParseIP(i).To4()
}

// NextIP returns IP incremented by 1
func NextIP(ip net.IP) net.IP {
	i := ipToInt(ip)
	return intToIP(i.Add(i, big.NewInt(1)))
}

// ParseIPs 解析ip 192.168.0.2-192.168.0.6
func ParseIPs(ips []string) []string {
	return DecodeIPs(ParsePasss(ips))
}

// RunCommand 执行命令
func RunCommand(arg ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("/bin/sh", arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// RemoveIpPort 去除IP中的端口
func RemoveIpPort(ip string) string {
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	return ip
}

// GetIpAndPort 返回ip、端口
func GetIpAndPort(ip string) (string, string) {
	if strings.Contains(ip, ":") {
		arr := strings.Split(ip, ":")
		return arr[0], arr[1]
	}
	return ip, "22"
}

// ParsePasss 将ip中的密码保存到全局变量
func ParsePasss(ips []string) []string {
	if SSHConfig.UserPass == nil {
		SSHConfig.UserPass = map[string]string{}
	}
	var hosts []string
	for _, host := range ips {
		if strings.Contains(host, "@") {
			arr := strings.Split(host, "@")
			ip := arr[0]
			pass := host[len(ip)+1:]
			if strings.Contains(ip, ":") {
				SSHConfig.UserPass[ip] = pass
			} else {
				SSHConfig.UserPass[ip+":22"] = pass
			}
			hosts = append(hosts, ip)
		} else {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

func DecodeIPs(ips []string) []string {
	var res []string
	var port string
	for _, ip := range ips {
		port = "22"
		if ipport := strings.Split(ip, ":"); len(ipport) == 2 {
			ip = ipport[0]
			port = ipport[1]
		}
		if iprange := strings.Split(ip, "-"); len(iprange) == 2 {
			for Cmp(stringToIP(iprange[0]), stringToIP(iprange[1])) <= 0 {
				res = append(res, fmt.Sprintf("%s:%s", iprange[0], port))
				iprange[0] = NextIP(stringToIP(iprange[0])).String()
			}
		} else {
			if stringToIP(ip) == nil {
				logger.Warn("ip [%s] is invalid", ip)
				os.Exit(1)
			}
			res = append(res, fmt.Sprintf("%s:%s", ip, port))
		}
	}
	return res
}

func GetRemoteHostName(hostIP string) string {
	hostName := SSHConfig.CmdToStringNoLog(hostIP, "hostname", "")
	return strings.ToLower(hostName)
}

func Base64Encode(data string) string {
	sEnc := base64.StdEncoding.EncodeToString([]byte(data))
	return fmt.Sprintf(sEnc)
}

func Base64Decode(data string) string {
	uDec, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		logger.Warn("Error decoding string: %s ", err.Error())
		return ""
	}
	return string(uDec)
}

func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}

func StringMd5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func InArray(item string, items []string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

func PingFile(path string, source string) (string, error) {
	result, err := PingFileMap(path, source, 2000, 5)
	if err != nil {
		return "", err
	}
	value, errJson := json.Marshal(result)
	return string(value), errJson
}

// PingFileMap 遍历ping文件内ip，并返回ping键值（最小）
func PingFileMap(path string, source string, timeout int, count int) (map[string]float64, error) {
	cmd := fmt.Sprintf("fping -A -u -q -4 -t %d -c %d -f %s", timeout, count, path)
	if source != "" {
		cmd = fmt.Sprintf("fping -A -u -q -4 -S %s -t %d -c %d -f %s", source, timeout, count, path)
	}
	_, result, err := RunCommand("-c", cmd)
	if result == "" && err != nil {
		return nil, err
	}
	result = strings.Replace(result, " ", "", -1)
	spaceRe, errRe := regexp.Compile(`[/:=]`)
	if errRe != nil {
		return nil, err
	}
	var pingMap = make(map[string]float64)
	scanner := bufio.NewScanner(strings.NewReader(result))
	for scanner.Scan() {
		s := spaceRe.Split(scanner.Text(), -1)
		if len(s) > 9 {
			float, _ := strconv.ParseFloat(s[9], 64)
			pingMap[s[0]] = float
		} else {
			pingMap[s[0]] = 0
		}
	}
	return pingMap, nil
}

func ReadLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	var ret []string
	r := bufio.NewReader(f)
	for {
		line, readErr := r.ReadString('\n')
		if readErr != nil {
			break
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}
	return ret, nil
}

func ComputePing(var1, var2 float64) bool {
	diff := math.Abs(var1 - var2)
	if diff < 5 {
		return false
	}
	if diff >= 50 {
		return true
	}
	var multiple float64
	if var1 > var2 {
		multiple = var1 / var2
	} else {
		multiple = var2 / var1
	}
	if multiple < 1.1 {
		return false
	}
	return true
}

// GetIpsFiles 获取指定目录下的所有ips文件
func GetIpsFiles(dirPath string) []string {
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil
	}
	var files []string
	for _, fi := range dir {
		if !fi.IsDir() {
			ok := strings.HasSuffix(fi.Name(), ".ips")
			if ok {
				ip := strings.TrimSuffix(fi.Name(), ".ips")
				if stringToIP(ip) != nil {
					files = append(files, ip)
				}
			}
		}
	}
	return files
}

// StringsContains 数组是否包含
func StringsContains(array []string, val string) (index int) {
	index = -1
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			index = i
			return
		}
	}
	return
}

// KillProcess 根据名字杀死进程
func KillProcess(name string) error {
	processes, err := process.Processes()
	if err != nil {
		return err
	}
	for _, p := range processes {
		n, err := p.Name()
		if err != nil {
			return err
		}
		if n == name {
			return p.Kill()
		}
	}
	return fmt.Errorf("process not found")
}

// GetNetIoNic 获取指定网卡的网速
func GetNetIoNic(nicName string, lastNetIoNic *NetIoNic) *NetIoNic {
	ioStats, err := gopsnet.IOCounters(true)
	if err != nil {
		logger.Warn("get io counters failed:", err)
	} else if len(ioStats) > 0 {
		var netIoNic *NetIoNic
		for _, stat := range ioStats {
			if stat.Name == nicName {
				now := time.Now()
				netIoNic = &NetIoNic{
					T: now,
				}
				netIoNic.Sent = stat.BytesSent
				netIoNic.Recv = stat.BytesRecv

				if lastNetIoNic != nil {
					duration := now.Sub(lastNetIoNic.T)
					seconds := float64(duration) / float64(time.Second)
					up := uint64(float64(netIoNic.Sent-lastNetIoNic.Sent) / seconds)
					down := uint64(float64(netIoNic.Recv-lastNetIoNic.Recv) / seconds)
					netIoNic.Up = up
					netIoNic.Down = down
				}
				break
			}
		}
		return netIoNic
	} else {
		logger.Warn("can not find io counters")
	}
	return nil
}

// GetNetIoInNic 获取入口容器的网速
func GetNetIoInNic(lastNetIoNic *NetIoNic) *NetIoNic {
	ioStats, err := gopsnet.IOCounters(true)
	if err != nil {
		logger.Warn("get io counters failed:", err)
	} else if len(ioStats) > 0 {
		stat := gopsnet.IOCountersStat{
			Name: "all",
		}
		for _, nic := range ioStats {
			if nic.Name == "wgcenter" || strings.HasSuffix(nic.Name, "wg2s_") {
				stat.BytesRecv += nic.BytesRecv
				stat.PacketsRecv += nic.PacketsRecv
				stat.Errin += nic.Errin
				stat.Dropin += nic.Dropin
				stat.BytesSent += nic.BytesSent
				stat.PacketsSent += nic.PacketsSent
				stat.Errout += nic.Errout
				stat.Dropout += nic.Dropout
			}
		}
		now := time.Now()
		netIoNic := &NetIoNic{
			T: now,
			Sent: stat.BytesSent,
			Recv: stat.BytesRecv,
		}
		if lastNetIoNic != nil {
			duration := now.Sub(lastNetIoNic.T)
			seconds := float64(duration) / float64(time.Second)
			up := uint64(float64(netIoNic.Sent-lastNetIoNic.Sent) / seconds)
			down := uint64(float64(netIoNic.Recv-lastNetIoNic.Recv) / seconds)
			netIoNic.Up = up
			netIoNic.Down = down
		}
		return netIoNic
	} else {
		logger.Warn("can not find io counters")
	}
	return nil
}

// GetManageState 获取主容器的状态
func GetManageState(lastState *State) *State {
	now := time.Now()
	state := &State{
		T: now,
	}

	percents, err := cpu.Percent(0, false)
	if err != nil {
		logger.Warn("get cpu percent failed:", err)
	} else {
		state.Cpu = percents[0]
	}

	upTime, err := host.Uptime()
	if err != nil {
		logger.Warn("get uptime failed:", err)
	} else {
		state.Uptime = upTime
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		logger.Warn("get virtual memory failed:", err)
	} else {
		state.Mem.Current = memInfo.Used
		state.Mem.Total = memInfo.Total
	}

	swapInfo, err := mem.SwapMemory()
	if err != nil {
		logger.Warn("get swap memory failed:", err)
	} else {
		state.Swap.Current = swapInfo.Used
		state.Swap.Total = swapInfo.Total
	}

	distInfo, err := disk.Usage("/")
	if err != nil {
		logger.Warn("get dist usage failed:", err)
	} else {
		state.Disk.Current = distInfo.Used
		state.Disk.Total = distInfo.Total
	}

	avgState, err := load.Avg()
	if err != nil {
		logger.Warn("get load avg failed:", err)
	} else {
		state.Loads = []float64{avgState.Load1, avgState.Load5, avgState.Load15}
	}

	ioStats, err := gopsnet.IOCounters(false)
	if err != nil {
		logger.Warn("get io counters failed:", err)
	} else if len(ioStats) > 0 {
		ioStat := ioStats[0]
		state.NetTraffic.Sent = ioStat.BytesSent
		state.NetTraffic.Recv = ioStat.BytesRecv

		if lastState != nil {
			duration := now.Sub(lastState.T)
			seconds := float64(duration) / float64(time.Second)
			up := uint64(float64(state.NetTraffic.Sent-lastState.NetTraffic.Sent) / seconds)
			down := uint64(float64(state.NetTraffic.Recv-lastState.NetTraffic.Recv) / seconds)
			state.NetIO.Up = up
			state.NetIO.Down = down
		}
	} else {
		logger.Warn("can not find io counters")
	}

	state.TcpCount, err = sys.GetTCPCount()
	if err != nil {
		logger.Warn("get tcp connections failed:", err)
	}

	state.UdpCount, err = sys.GetUDPCount()
	if err != nil {
		logger.Warn("get udp connections failed:", err)
	}

	return state
}