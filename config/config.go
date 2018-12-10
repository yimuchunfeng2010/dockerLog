package config

var Config = struct {
	Fields       []string
	Sevices      []string
	PrintColor PrintColor
	PrintFormat PrintFormat

}{
	Fields: []string{
		"time",
		"file",
		"msg",
	},
	Sevices:[]string{
		"service-etrip-approval",
		"service-etrip-bill",
		"service-finance-bill",
		"service-fin-mgr",
		"service-etrip-app-gateway",
		"service-glp-gateway",
		"service-mybank-gateway",
		"service-wacai-gateway",
		"service-finance-timing-task",
		"service-finance-kafka-task",
		"service-freight-task",
		"service-freight-approval",
	},
	PrintColor:PrintColor{
		FrontColor:40,
		BackgroudColor:1,
	},
	PrintFormat:PrintFormat{
		TimeFormat:32,
		FileFormat:35,
		MsgFormat:33,
		ErrWarning:31,
	},
}

/*
 // 前景 背景 颜色
    // ---------------------------------------
    // 30  40  黑色
    // 31  41  红色
    // 32  42  绿色
    // 33  43  黄色
    // 34  44  蓝色
    // 35  45  紫红色
    // 36  46  青蓝色
    // 37  47  白色
    //
    // 代码 意义
    // -------------------------
    //  0  终端默认设置
    //  1  高亮显示
    //  4  使用下划线
    //  5  闪烁
    //  7  反白显示
    //  8  不可见
*/

type PrintColor struct{
	FrontColor int
	BackgroudColor int
}

type PrintFormat struct {
	TimeFormat int
	FileFormat int
	MsgFormat int
	ErrWarning int
}
