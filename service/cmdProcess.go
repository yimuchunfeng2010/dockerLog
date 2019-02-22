package service

import (
	"strings"
	"strconv"
	"os"
	"fmt"
	"dockerLog/global"
)

// 获取用户输入
func GetUseInput(chUserInput chan string) {
	for {
		var userCmd string
		fmt.Scanf("%s", &userCmd)
		chUserInput <- userCmd
	}
}

func StopSysProcess(ServiceID string) (err error) {

	cmdString := fmt.Sprintf("ps -ef|grep %s", ServiceID)
	output, err := execShellCmd(cmdString)
	if err != nil {
		return
	}

	var processID int
	for _, value := range output {
		if true == strings.Contains(value, "rancher logs -f") && -1 == strings.Index(value, "2>&1") {
			tmpArray := strings.Split(value, " ")
			for _, vv := range tmpArray {
				if id, newErr := strconv.Atoi(vv); newErr == nil {
					processID = id
					goto loop
				}
			}
		}
	}

loop:
	if 0 == processID {
		return
	}
	// 终止进程
	cmdString = fmt.Sprintf("kill %d", processID)
	execShellCmd(cmdString)
	return
}

func QuitLog(serviceID string) {
	// 终止读取日志进程
	StopSysProcess(serviceID)
	// 删除临时文件
	os.Remove(global.LogFile.TmpLogFile)
}

// 清屏
func CleanScreen() {
	msg := ""
	for i := 0; i < 100; i++ {
		msg += "\n"
	}
	fmt.Printf(msg)
}

func CmdHelp() {
	fmt.Println("Please Enter Right Command")
	fmt.Println("q   Quit the program")
	fmt.Println("s   Switch to other service")
	fmt.Println("c   clean screen")
	fmt.Println("p   Pause to pring log")
	fmt.Println("r   Restart to print log")
}
