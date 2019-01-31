package service

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"DockerLog/config"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"encoding/json"
	"DockerLog/global"
)

func LogMain() {
	serviceName := GetServiceNameFromUser()
	serviceID, err := GetServiceID(serviceName)
	if err != nil {
		fmt.Printf("service.GetServiceID Failed[Err:%s]", err.Error())
		return
	}
	chLog := make(chan map[string]string)
	chUserInput := make(chan string)
	chInternal := make(chan string)

	go GetLogsByID(serviceID)
	go GetUseInput(chUserInput)
	go ProcLog(chLog, chInternal)

	for {
		select {
		case logMsg, _ := <-chLog:
			PrintLog(logMsg)
		case cmd := <-chUserInput:
			switch cmd {
			case global.Command.Quit:
				QuitLog(serviceID)
				return
			case global.Command.Switch: // 切换分支
				chInternal <- global.InternalCmd.Stop
				StopSysProcess(serviceID)
				serviceName := GetServiceNameFromUserNew(chUserInput)
				serviceID, err = GetServiceID(serviceName)
				if err != nil {
					fmt.Printf("service.GetServiceID Failed[Err:%s]", err.Error())
					return
				}
				go GetLogsByID(serviceID)
				go ProcLog(chLog, chInternal)
			case global.Command.CleanScreen:
				CleanScreen()
			case global.Command.PausePrint:
				chInternal <- global.InternalCmd.Pause
			case global.Command.RestartPrint:
				chInternal <- global.InternalCmd.Restart
			default:
				CmdHelp()

			}
		}
	}

}

func GetServiceID(serviceName string) (serviceID string, err error) {
	serviceID, err = GetServiceIDFromCache(serviceName)
	if err == nil {
		return
	}

	arg := fmt.Sprintf("rancher ps | grep %s", serviceName)
	output, err := execShellCmd(arg)
	if err != nil {
		fmt.Printf("Exec ShellCmd Failed[Arg:%s, Err:%s]\n", arg, err.Error())
		return
	}

	if len(output) == 0 {
		errMsg := fmt.Sprintf("Service Not Found[ServiceName:%s]", serviceName)
		fmt.Println(errMsg)
		return "", errors.New(errMsg)
	}

	for _, value := range output {
		tmpArray := strings.Split(value, " ")
		newArray := make([]string, 0)
		for _, value := range tmpArray {
			if value != "" {
				newArray = append(newArray, value)
			}
		}
		tmpService := strings.Split(newArray[2], "/")[0]
		if 0 == strings.Compare(tmpService, serviceName) {
			serviceID = tmpArray[0]
			return
		}
	}

	return "", errors.New(fmt.Sprintf("Service Not Found[ServiceName:%s]", serviceName))
}

func GetServiceIDFromCache(serviceName string) (serviceID string, err error) {
	for name, id := range global.GlobalVar.ServiceNameID {
		if 0 == strings.Compare(name, serviceName) {
			return id, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Service Not Found"))
}

func GetLogsByID(serviceID string) (err error) {
	arg := fmt.Sprintf("rancher logs -f %s > %s 2>&1", serviceID, global.LogFile.TmpLogFile)
	_, err = execShellCmd(arg)
	if err != nil {
		if 0 == strings.Compare(global.ErrVar.SysErr, err.Error()) {
			return
		}
		fmt.Printf("Exec ShellCmd Failed![Err:%s]\n", err.Error())
		return
	}
	return
}

func ProcLog(chLog chan map[string]string, chInternal chan string) {
	for {
		if true == IsExist(global.LogFile.TmpLogFile) {
			break
		}
		time.Sleep(time.Second)
	}

	fi, err := os.Open(global.LogFile.TmpLogFile)
	if err != nil {
		fmt.Printf("Open File Failed[Err:%s]", err.Error())
		return
	}
	defer fi.Close()
	br := bufio.NewReader(fi)

	for {
	CmdLoop:
		select {
		case stop := <-chInternal:
			switch stop {
			case global.InternalCmd.Stop:
				return
			case global.InternalCmd.Pause:
				global.GlobalVar.PauseFlag = true
			case global.InternalCmd.Restart:
				global.GlobalVar.PauseFlag = false

			}
		default:
			break
		}

		if true == global.GlobalVar.PauseFlag {
			time.Sleep(time.Second)
			goto CmdLoop
		}

		a, _, c := br.ReadLine()
		if c == io.EOF {
			time.Sleep(time.Second)
			continue
		}

		m := make(map[string]string)
		json.Unmarshal(a, &m)
		chLog <- m

	}
}
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

// 判断文件是否存在
func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func GetServiceNameFromUser() string {
	fmt.Printf("Please Enter Service Name: ")
	var serviceName string

	for {
	loop:
		fmt.Scanf("%s", &serviceName)
		// 首先在缓存中查询
		tmpServiceList := make([]string, 0)
		for name, _ := range global.GlobalVar.ServiceNameID {
			if true == strings.Contains(name, serviceName) {
				tmpServiceList = append(tmpServiceList, name)
			}
		}
		if len(tmpServiceList) != 0 {
			return ChooseService(tmpServiceList)
		}

		serviceList, err := GetCompleteServiceName(serviceName)
		if err != nil {
			fmt.Printf("Please Enter Right Service Name: ")
			goto loop
		}
		return ChooseService(serviceList)

	}
	return ""
}

func GetServiceNameFromUserNew(chUserInput chan string) string {
	fmt.Printf("Please Enter Service Name: ")
	var serviceName string

	for {
	loop:
		serviceName = <-chUserInput
		// 首先在缓存中查询
		tmpServiceList := make([]string, 0)
		for name, _ := range global.GlobalVar.ServiceNameID {
			if true == strings.Contains(name, serviceName) {
				tmpServiceList = append(tmpServiceList, name)
			}
		}
		if len(tmpServiceList) != 0 {
			return ChooseServiceNew(tmpServiceList, chUserInput)
		}

		serviceList, err := GetCompleteServiceName(serviceName)
		if err != nil {
			fmt.Printf("Please Enter Right Service Name: ")
			goto loop
		}
		return ChooseServiceNew(serviceList, chUserInput)

	}
	return ""
}

func ChooseServiceNew(serviceList []string, chUserInput chan string) string {
	if len(serviceList) == 1 {
		return serviceList[0]
	}
	if len(serviceList) > 1 {
		fmt.Println("Which Service Do You Choose:")
	idxLoop:
		for idx, value := range serviceList {
			fmt.Println(idx, value)
		}
		tmp := <-chUserInput
		serviceIdx, _ := strconv.Atoi(tmp)
		if serviceIdx >= len(serviceList) {
			fmt.Println("Please Enter Right Idx")
			goto idxLoop
		}

		return serviceList[serviceIdx]
	}
	return ""
}

func ChooseService(serviceList []string) string {
	if len(serviceList) == 1 {
		return serviceList[0]
	}
	if len(serviceList) > 1 {

	idxLoop:
		for idx, value := range serviceList {
			fmt.Println(idx, "  ", value)
		}
		fmt.Printf("Which Service Do You Choose: ")
		var serviceIdx int
		fmt.Scanf("%d", &serviceIdx)
		if serviceIdx >= len(serviceList) {
			fmt.Println("Please Enter Right Idx")
			goto idxLoop
		}

		return serviceList[serviceIdx]
	}
	return ""
}

// 获取完整的服务名，用户可能输入的服务不完整
func GetCompleteServiceName(partServiceName string) (serviceList []string, err error) {
	arg := fmt.Sprintf("rancher ps | grep %s", partServiceName)
	output, err := execShellCmd(arg)
	if err != nil {
		fmt.Printf("Service %s Not Found\n", partServiceName)
		return
	}

	for _, value := range output {
		tmpArray := strings.Split(value, " ")
		newArray := make([]string, 0)
		for _, value := range tmpArray {
			if value != "" {
				newArray = append(newArray, value)
			}
		}
		tmpService := strings.Split(newArray[2], "/")[0]
		serviceList = append(serviceList, tmpService)
	}
	if len(serviceList) == 0 {
		msg := fmt.Sprintf("Service: %s Not Found", partServiceName)
		return serviceList, errors.New(msg)
	}
	return
}

func GetUseInput(chUserInput chan string) {
	for {
		var cmd string
		fmt.Scanf("%s", &cmd)
		chUserInput <- cmd
	}
}

func PrintLog(logMsg map[string]string) {
	// 先判断是否有需要打印的日志
	isNeedPrint := false
	for _, field := range config.Config.Fields {
		if logMsg[field] != "" {
			isNeedPrint = true
			break
		}
	}
	if false == isNeedPrint {
		return
	}

	fmt.Printf("{\n")
	for _, field := range config.Config.Fields {
		if logMsg[field] == "" {
			continue
		}

		// 显示颜色
		var format int
		switch field {
		case config.Config.Fields[0]:
			format = config.Config.PrintFormat.TimeFormat
		case config.Config.Fields[1]:
			format = config.Config.PrintFormat.MsgFormat
		case config.Config.Fields[2]:
			format = config.Config.PrintFormat.FileFormat
		}

		if logMsg["level"] == "error" || logMsg["level"] == "warning" {
			format = config.Config.PrintFormat.ErrWarning
		}
		msg := fmt.Sprintf("   %s: %s", strings.ToUpper(field), logMsg[field])
		fmt.Printf(" %c[%d;%d;%dm%s%s%c[0m\n", 0x1B, config.Config.PrintColor.BackgroudColor, config.Config.PrintColor.FrontColor, format, "", msg, 0x1B)
	}
	fmt.Printf("}\n")

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

func GetAllService() {
	cmdString := "rancher ps"
	output, err := execShellCmd(cmdString)
	if err != nil {
		fmt.Printf("Exec Shell Cmd Failed! [Err: %s]", err.Error())
		return
	}
	global.GlobalVar.ServiceNameID = make(map[string]string)
	for _, value := range output {

		for _, curServiceName := range config.Config.Sevices {
			if strings.Contains(value, curServiceName) {
				tmpValueArray := strings.Split(value, " ")
				for _, part := range tmpValueArray {
					if strings.Contains(part, curServiceName) {
						tmpName := strings.Split(part, "/")
						global.GlobalVar.ServiceNameID[tmpName[0]] = tmpValueArray[0]
					}
				}
			}
		}
	}
}

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
