// Package init /init/cmdInit.go
package init

import (
	"github.com/GoFurry/gf-file-tool/cmd"
	"github.com/GoFurry/gf-file-tool/cmd/compress"
	"github.com/GoFurry/gf-file-tool/cmd/decompress"
	"github.com/GoFurry/gf-file-tool/cmd/decrypt"
	"github.com/GoFurry/gf-file-tool/cmd/encrypt"
	"github.com/GoFurry/gf-file-tool/cmd/function/crc32"
	"github.com/GoFurry/gf-file-tool/cmd/function/merge"
)

// PerformInitOnStart 开始前的初始化函数, 在 Web 项目中常用于初始化数据库以及各种中间件服务.
// 而在命令行工具中通常只需要挂载对应的子命令.
func PerformInitOnStart() {
	// 必须先初始化根命令
	// 初始化根命令
	cmd.InitRoot()

	// 初始化子命令
	compress.InitCompress()     // 压缩
	decompress.InitDecompress() // 解压缩
	encrypt.InitEncrypt()       // 加密
	decrypt.InitDecrypt()       // 解密
	crc32.InitCRC32()           // CRC32 校验
	merge.InitMerge()           // 合并文件
}
