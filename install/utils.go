package install

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
		logger.Error("invalied host fomat [%s], must like 172.0.0.2:22", host)
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

// Cmp compares two IPs, returning the usual ordering:
// a < b : -1
// a == b : 0
// a > b : 1
func Cmp(a, b net.IP) int {
	aa := ipToInt(a)
	bb := ipToInt(b)

	if aa == nil || bb == nil {
		logger.Error("ip range %s-%s is invalid", a.String(), b.String())
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
				logger.Error("ip [%s] is invalid", ip)
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

func base64Encode(data string) string {
	sEnc := base64.StdEncoding.EncodeToString([]byte(data))
	return fmt.Sprintf(sEnc)
}

func base64Decode(data string) string {
	uDec, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		logger.Error("Error decoding string: %s ", err.Error())
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

func PingFile(path string) (string, error) {
	cmd := fmt.Sprintf("fping -A -u -q -4 -t 2000 -c 5 -f %s", path)
	_, result, err := RunCommand("-c", cmd)
	if err != nil {
		return "", err
	}
	spaceRe, errRe := regexp.Compile(`[/:]`)
	if errRe != nil {
		return "", err
	}
	var pingMap = make(map[string]float64)
	scanner := bufio.NewScanner(strings.NewReader(result))
	for scanner.Scan() {
		s := spaceRe.Split(scanner.Text(), -1)
		if len(s) > 9 {
			float, _ := strconv.ParseFloat(s[9], 64)
			pingMap[strings.TrimSpace(s[0])] = float
		} else {
			pingMap[strings.TrimSpace(s[0])] = 0
		}
	}
	value, errJson := json.Marshal(pingMap)
	return string(value), errJson
}
