package log

import (
	"fmt"

	"github.com/gookit/color"
)

// 简易的日志封装, 利于学生快速实践开发工具包的封装

// Error 红色[Error]开头的错误日志
func Error(args ...any) {
	text := color.FgRed.Render("[Error]")
	for _, arg := range args {
		text += " " + fmt.Sprint(arg)
	}
	fmt.Println(text)
}

// Info 蓝色[Info]开头的普通日志
func Info(args ...any) {
	text := color.FgBlue.Render("[Info]")
	for _, arg := range args {
		text += " " + fmt.Sprint(arg)
	}
	fmt.Println(text)
}

// Success 绿色[Success]开头的成功日志
func Success(args ...any) {
	text := color.FgGreen.Render("[Success]")
	for _, arg := range args {
		text += " " + fmt.Sprint(arg)
	}
	fmt.Println(text)
}

// Warn 黄色[Warn]开头的警告日志
func Warn(args ...any) {
	text := color.FgYellow.Render("[Warn]")
	for _, arg := range args {
		text += " " + fmt.Sprint(arg)
	}
	fmt.Println(text)
}
