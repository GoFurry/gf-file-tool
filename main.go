package main

import (
	"github.com/GoFurry/gf-file-tool/cmd"
	cmdInit "github.com/GoFurry/gf-file-tool/init"
)

/**
 *    ┏┓  ┏┓
 *    ┃┓┏┓┣ ┓┏┏┓┏┓┓┏
 *    ┗┛┗┛┻ ┗┻┛ ┛ ┗┫
 *                 ┛
 * @title gf-file-tool
 * @version v1.0.0
 * @description 一个简易的私人压缩/解压缩/加密/解密命令行工具
 *   A training program from Chengdu University of Information Technology (CUIT).
 *   The aim is to cultivate Golang engineers.
 *   Go语言CLI工具开发教学案例, 通过这个案例你可以学习到 Cobra 框架的基本用法, Viper 配置管理库的基本用法,
 *   大量的 IO 读写训练, 文件/协议头部的解析, 密码学的一些基础知识. 你可以尝试修复该工具中一些显而易见的错误或是优
 *   化和新增更多的相关命令, 适合在学习完 Go 基础后配套使用, 祝你早日成为一名合格的 Golang 软件工程师.
 * @author 🐺福狼
 */

// 在命令行工具的入口我们需要先初始化 Cobra 框架的 Root 命令才可以执行, 而根命令就是整个工具的入口,
// 其余子命令都挂载在根命令上
func main() {
	cmdInit.PerformInitOnStart() // 初始化命令
	cmd.Execute()                // 执行根命令
}
