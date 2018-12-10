package global


var Command = struct{
	Quit string
	Switch string
	CleanScreen string
	PausePrint string
	RestartPrint string
}{
	Quit:"q",
	Switch:"s",
	CleanScreen:"c",
	PausePrint:"p",
	RestartPrint:"r",
}

var InternalCmd = struct {
	Stop string
	Pause string
	Restart string
}{
	Stop:"sop",
	Pause:"pause",
	Restart:"restart",
}

var GlobalVar = struct{
	PauseFlag bool
	ServiceNameID map[string]string
}{
	PauseFlag:false,
}

var ErrVar = struct {
	SysErr string
}{
	SysErr:"exit status 143",
}

var LogFile = struct {
	TmpLogFile string
}{
	TmpLogFile:"tmp.txt",
}