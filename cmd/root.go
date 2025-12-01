// Package cmd /cmd/root.go
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// 在根命令的文件内我们通常需要考虑去实现一些全局的用法, 比如注册全局参数、初始化全局配置等

// 全局参数
var (
	verbose bool // 详细日志模式
	quiet   bool // 静默模式
)

// rootCmd 挂载根命令实例 非导出全局变量
var rootCmd = &cobra.Command{
	Use:   "gf-file-tool",              // 命令行关键字
	Short: "gf-file-tool 是一款多功能文件处理工具", // 简短描述
	Long: `gf-file-tool 支持跨平台的文件压缩/解压缩、加密/解密, 核心特性:
  1. 支持 zip/7z/tar.gz 多种压缩格式
  2. AES/DES 加密
  3. 批量处理、分卷压缩、完整性校验
  4. 实时进度条、简易/高级双模式使用`, // 详细描述
	Version: "gf-file-tool V1.0.0",
	// 根命令逻辑
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`  ░██████             ░██████████                                        
 ░██   ░██            ░██                                                
░██         ░███████  ░██        ░██    ░██ ░██░████ ░██░████ ░██    ░██ 
░██  █████ ░██    ░██ ░█████████ ░██    ░██ ░███     ░███     ░██    ░██ 
░██     ██ ░██    ░██ ░██        ░██    ░██ ░██      ░██      ░██    ░██ 
 ░██  ░███ ░██    ░██ ░██        ░██   ░███ ░██      ░██      ░██   ░███ 
  ░█████░█  ░███████  ░██         ░█████░██ ░██      ░██       ░█████░██ 
                                                                     ░██ 
                                                               ░███████
` + cmd.Version + `
gf-file-tool 是一款私人用途的文件加密工具.
支持zip/7z/tar.gz格式的压缩/解压缩, 还支持文件校验、加密等功能.
使用 "gf-file-tool help" 查看所有命令.
	`)
	},
}

// Execute 启动根命令
func Execute() {
	// 执行根命令
	if err := rootCmd.Execute(); err != nil {
		log.Fatal("执行命令失败: ", err)
	}
}

// InitRoot 初始化全局参数以及 Viper 配置
func InitRoot() {
	// 初始化 Viper
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GF_FILE_TOOL") // 环境变量前缀 如 GF_FILE_TOOL_VERBOSE=true

	// 注册全局参数
	// --verbose / -v 启用详细日志
	rootCmd.PersistentFlags().BoolVarP(
		&verbose,   // 绑定的变量
		"verbose",  // 参数名 --verbose
		"v",        // 短参数 -v
		false,      // 默认值
		"启用详细日志输出", // 参数描述
	)
	// --quiet / -q 静默模式
	rootCmd.PersistentFlags().BoolVarP(
		&quiet,
		"quiet",
		"q",
		false,
		"静默模式",
	)

	// 全局参数绑定到 Viper
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))

}

// GetRootCmd rootCmd 的导出方法
func GetRootCmd() *cobra.Command {
	return rootCmd
}
