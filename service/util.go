package service

import (
	"os"
	"strings"
	"os/exec"
)

// 判断文件是否存在
func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

// 执行shell命令
func execShellCmd(arg string) (output []string, err error) {
	cmd := exec.Command("/bin/bash", "-c", arg)
	out, err := cmd.Output()
	if err != nil {
		return
	}
	tmp := strings.Split(string(out), "\n")

	if len(tmp) < 2 {
		output = tmp
	} else {
		output = tmp[:len(tmp)-1]
	}
	return
}
