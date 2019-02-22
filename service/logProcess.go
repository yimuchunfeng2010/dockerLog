package service

import (
	"bufio"
	"dockerLog/config"
	"dockerLog/global"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// 初始化，缓存所有服务名与ID映射关键
func InitServices() {
	cmdString := "rancher ps"
	output, err := execShellCmd(cmdString)
	if err != nil {
		fmt.Printf("Exec Shell Cmd Failed! [Err: %s]", err.Error())
		return
	}

	global.GlobalVar.ServiceNameID = make(map[string]string)
	for _, value := range output {
		for _, service := range config.Config.Services {
			if strings.Contains(value, service) {
				tmpValueArray := strings.Split(value, " ")
				for _, part := range tmpValueArray {
					if strings.Contains(part, service) {
						tmpName := strings.Split(part, "/")
						global.GlobalVar.ServiceNameID[tmpName[0]] = tmpValueArray[0]
					}
				}
			}
		}
	}
}

// 日志处理主函数
func LogMain() {
	serviceName := GetServiceNameFromUser()
	serviceID, err := GetServiceID(serviceName)
	if err != nil {
		fmt.Printf("service.GetServiceID Failed[Err:%s]", err.Error())
		return
	}

	chLog := make(chan map[string]string)
	chUserInput := make(chan string)
	chCmd := make(chan string)

	// 查询日志协程
	go GetLogsByID(serviceID)
	// 日志处理协程
	go ResolveLog(chLog, chCmd)
	// 获取用户输入协程
	go GetUseInput(chUserInput)

	for {
		select {
		// 打印日志
		case logMsg, _ := <-chLog:
			PrintLog(logMsg)
			// 处理用户命令输入
		case cmd := <-chUserInput:
			switch cmd {
			case global.Command.Switch: // 切换服务

				// 停止当前日志
				chCmd <- global.InternalCmd.Stop
				StopSysProcess(serviceID)

				serviceName := GetServiceNameFromChannel(chUserInput)
				serviceID, err = GetServiceID(serviceName)
				if err != nil {
					fmt.Printf("service.GetServiceID Failed[Err:%s]", err.Error())
					return
				}

				go GetLogsByID(serviceID)
				go ResolveLog(chLog, chCmd)

			case global.Command.Quit:
				QuitLog(serviceID)
				return
			case global.Command.CleanScreen:
				CleanScreen()
			case global.Command.PausePrint:
				chCmd <- global.InternalCmd.Pause
			case global.Command.RestartPrint:
				chCmd <- global.InternalCmd.Restart
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
	return "", errors.New(fmt.Sprintf("Service Not Found[ServiceName:%s]", serviceName))
}

func GetLogsByID(serviceID string) (err error) {
	arg := fmt.Sprintf("rancher logs -f %s > %s 2>&1", serviceID, global.LogFile.TmpLogFile)
	_, err = execShellCmd(arg)
	if err != nil {
		if 0 == strings.Compare(global.ErrVar.SysErr, err.Error()) {
			return
		}
		fmt.Printf("Shell Cmd Failed![Err:%s]\n", err.Error())
		return
	}
	return
}

func ResolveLog(chLog chan map[string]string, chCmd chan string) {
	for {
		// 等待日志文件创建
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
		case stop := <-chCmd:
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

func GetServiceNameFromUser() string {
	fmt.Printf("Please Enter Service Name: ")
	var serviceName string

	for {
	loop:
		fmt.Scanf("%s", &serviceName)
		// 首先在缓存中查询
		tmpServiceList := make([]string, 0)
		for name := range global.GlobalVar.ServiceNameID {
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

func GetServiceNameFromChannel(chUserInput chan string) string {
	fmt.Printf("Please Enter Service Name: ")
	var serviceName string

	for {
	loop:
		serviceName = <-chUserInput
		// 首先在缓存中查询
		tmpServiceList := make([]string, 0)
		for name := range global.GlobalVar.ServiceNameID {
			if true == strings.Contains(name, serviceName) {
				tmpServiceList = append(tmpServiceList, name)
			}
		}
		if len(tmpServiceList) != 0 {
			return ChooseServiceFromChannel(tmpServiceList, chUserInput)
		}

		serviceList, err := GetCompleteServiceName(serviceName)
		if err != nil {
			fmt.Printf("Please Enter Right Service Name: ")
			goto loop
		}
		return ChooseServiceFromChannel(serviceList, chUserInput)

	}
	return ""
}

func ChooseServiceFromChannel(serviceList []string, chUserInput chan string) string {
	if len(serviceList) == 1 {
		return serviceList[0]
	}
	if len(serviceList) > 1 {
		fmt.Println("Which Service Do You Want: ")
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
			fmt.Println(idx, ":  ", value)
		}
		fmt.Printf("Which Service Do You Want: ")
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

		// 若出现错误或者警告级别的日志，则红色警示
		if logMsg["level"] == "error" || logMsg["level"] == "warning" {
			format = config.Config.PrintFormat.ErrWarning
		}
		msg := fmt.Sprintf("   %s: %s", strings.ToUpper(field), logMsg[field])
		fmt.Printf(" %c[%d;%d;%dm%s%s%c[0m\n", 0x1B, config.Config.PrintColor.BackgroundColor, config.Config.PrintColor.FrontColor, format, "", msg, 0x1B)
	}
	fmt.Printf("}\n")

}
