// Package crc32 /cmd/function/crc32/crc32.go
package crc32

import (
	"github.com/GoFurry/gf-file-tool/cmd"
	"github.com/GoFurry/gf-file-tool/utils/compress"
	"github.com/GoFurry/gf-file-tool/utils/log"
	"github.com/spf13/cobra"
)

// crc32Cmd crc32校验命令实例
var crc32Cmd = &cobra.Command{
	Use:   "crc32 [file]",
	Short: "计算文件 CRC32 值",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		crc, err := compress.CalculateCRC32(args[0])
		if err != nil {
			log.Error("计算失败:", err)
			return
		}
		log.Success(args[0], "的 CRC32:", crc)
	},
}

// InitCRC32 初始化命令
func InitCRC32() {
	cmd.GetRootCmd().AddCommand(crc32Cmd)
}
