package sshutil

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/kuaifan/sdos/pkg/logger"
)

func (ss *SSH) Cmd(host string, cmd string, desc ...string) []byte {
	if desc != nil {
		logger.Debug("[ssh] [%s] %s", host, strings.Join(desc, ""))
	} else {
		logger.Info("[ssh] [%s] %s", host, cmd)
	}
	session, err := ss.Connect(host)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[ssh] [%s] Error create ssh session failed,%s", host, err)
		}
	}()
	if err != nil {
		panic(1)
	}
	defer session.Close()
	b, err := session.CombinedOutput(cmd)
	logger.Debug("[ssh] [%s] command result is: %s", host, strings.TrimSpace(string(b)))
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[ssh] [%s] Error exec command failed: %s", host, err)
		}
	}()
	if err != nil {
		panic(1)
	}
	return b
}

func (ss *SSH) CmdNoLog(host string, cmd string) []byte {
	session, err := ss.Connect(host)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[ssh] [%s] Error create ssh session failed,%s", host, err)
		}
	}()
	if err != nil {
		panic(1)
	}
	defer session.Close()
	b, err := session.CombinedOutput(cmd)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[ssh] [%s] Error exec command failed: %s", host, err)
		}
	}()
	if err != nil {
		panic(1)
	}
	return b
}

func readPipe(host string, pipe io.Reader, isErr bool) {
	r := bufio.NewReader(pipe)
	for {
		line, _, err := r.ReadLine()
		if line == nil {
			return
		} else if err != nil {
			logger.Info("[ssh] [%s] %s", host, line)
			logger.Error("[ssh] [%s] %s", host, err)
			return
		} else {
			if isErr {
				logger.Error("[ssh] [%s] %s", host, line)
			} else {
				logger.Info("[ssh] [%s] %s", host, line)
			}
		}
	}
}

func (ss *SSH) CmdAsync(host string, cmd string, desc ...string) error {
	if desc != nil {
		logger.Debug("[ssh] [%s] %s", host, strings.Join(desc, ""))
	} else {
		logger.Debug("[ssh] [%s] %s", host, cmd)
	}
	session, err := ss.Connect(host)
	if err != nil {
		logger.Error("[ssh] [%s] Error create ssh session failed,%s", host, err)
		return err
	}
	defer session.Close()
	stdout, err := session.StdoutPipe()
	if err != nil {
		logger.Error("[ssh] [%s] Unable to request StdoutPipe(): %s", host, err)
		return err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		logger.Error("[ssh] [%s] Unable to request StderrPipe(): %s", host, err)
		return err
	}
	if err := session.Start(cmd); err != nil {
		logger.Error("[ssh] [%s] Unable to execute command: %s", host, err)
		return err
	}
	doneout := make(chan bool, 1)
	doneerr := make(chan bool, 1)
	go func() {
		readPipe(host, stderr, true)
		doneerr <- true
	}()
	go func() {
		readPipe(host, stdout, false)
		doneout <- true
	}()
	<-doneerr
	<-doneout
	return session.Wait()
}

func (ss *SSH) CmdToString(host, cmd, spilt string) string {
	data := ss.Cmd(host, cmd)
	if data != nil {
		str := string(data)
		str = strings.ReplaceAll(str, "\r\n", spilt)
		return str
	}
	return ""
}

func (ss *SSH) CmdToStringNoLog(host, cmd, spilt string) string {
	data := ss.CmdNoLog(host, cmd)
	if data != nil {
		str := string(data)
		str = strings.ReplaceAll(str, "\r\n", spilt)
		return str
	}
	return ""
}

func Base64Encode(data string) string {
	sEnc := base64.StdEncoding.EncodeToString([]byte(data))
	return fmt.Sprintf(sEnc)
}

func (ss *SSH) SaveFile(node string, path string, content string) error {
	cmd := fmt.Sprintf(`echo -n "%s" | base64 -d > %s && chmod +x %s`, Base64Encode(content), path, path)
	return ss.CmdAsync(node, cmd, fmt.Sprintf("Save file %s", path))
}
