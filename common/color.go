package common

import (
	"fmt"
)

// ANSI 颜色码常量
const (
	codeReset  = "\033[0m"
	codeRed    = "\033[31m"
	codeGreen  = "\033[32m"
	codeYellow = "\033[33m"
	codeBlue   = "\033[34m"
	codePurple = "\033[35m"
	codeCyan   = "\033[36m"
	codeGray   = "\033[37m"
)

// Console 是我们的工具结构体
type Console struct {
	EnableColor bool // 全局开关，如果设置为 false，则打印纯文本
}

// New 创建一个新的实例，默认开启颜色
func New() *Console {
	return &Console{EnableColor: true}
}

// printColor 内部核心打印逻辑
func (c *Console) printColor(colorCode string, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if c.EnableColor {
		fmt.Print(colorCode + msg + codeReset + "\n")
	} else {
		fmt.Println(msg)
	}
}

// --- 基础颜色方法 ---

func (c *Console) Red(format string, a ...interface{}) {
	c.printColor(codeRed, format, a...)
}

func (c *Console) Green(format string, a ...interface{}) {
	c.printColor(codeGreen, format, a...)
}

func (c *Console) Yellow(format string, a ...interface{}) {
	c.printColor(codeYellow, format, a...)
}

func (c *Console) Blue(format string, a ...interface{}) {
	c.printColor(codeBlue, format, a...)
}

func (c *Console) Cyan(format string, a ...interface{}) {
	c.printColor(codeCyan, format, a...)
}

func (c *Console) Purple(format string, a ...interface{}) {
	c.printColor(codePurple, format, a...)
}

// --- 语义化快捷方法 (业务常用) ---

// Success 绿色：用于成功提示
func (c *Console) Success(format string, a ...interface{}) {
	c.printColor(codeGreen, "[SUCCESS] "+format, a...)
}

// Error 红色：用于错误提示
func (c *Console) Error(format string, a ...interface{}) {
	c.printColor(codeRed, "[ERROR] "+format, a...)
}

// Warning 黄色：用于警告
func (c *Console) Warning(format string, a ...interface{}) {
	c.printColor(codeYellow, "[WARN] "+format, a...)
}

// Info 蓝色：用于一般信息
func (c *Console) Info(format string, a ...interface{}) {
	c.printColor(codeBlue, "[INFO] "+format, a...)
}

// --- 默认全局单例 (方便直接调用) ---
var Default = New()