// Package compress /cmd/compress/compress.go
package compress

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GoFurry/gf-file-tool/cmd"
	"github.com/GoFurry/gf-file-tool/core/compress"
	"github.com/GoFurry/gf-file-tool/utils"
	uc "github.com/GoFurry/gf-file-tool/utils/compress"
	"github.com/GoFurry/gf-file-tool/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// compressCmd 压缩命令实例
var compressCmd = &cobra.Command{
	Use:   "compress [source...]",
	Short: "压缩文件/目录",
	Long: `压缩文件/目录, 支持多格式、批量处理、分卷压缩:
  简易模式: gf-file-tool compress ./test.txt -o test.zip
  高级模式: gf-file-tool compress ./docs -f zip -s 104857600 -e -k 123456 -l 32 -r -v`,
	Args: cobra.MinimumNArgs(1), // 至少需要 1 个源文件/目录参数
	Run: func(cmd *cobra.Command, args []string) {
		// 解析命令参数
		outputPath, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")
		splitSize, _ := cmd.Flags().GetInt64("split")
		encrypt, _ := cmd.Flags().GetBool("encrypt")
		key, _ := cmd.Flags().GetString("key")
		verify, _ := cmd.Flags().GetBool("verify")
		keyLength, _ := cmd.Flags().GetInt("key-length")

		// 校验支持的格式
		format = strings.ToLower(strings.TrimSpace(format))
		supportedFormats := map[string]bool{"zip": true, "targz": true}
		if !supportedFormats[format] {
			log.Error("不支持的格式:", format, ", 仅支持 zip/targz")
			return
		}

		// 自动补全输出路径
		if outputPath == "" {
			// 目录名或文件名
			base := filepath.Base(args[0])
			if !strings.HasSuffix(base, "."+format) {
				outputPath = fmt.Sprintf("%s.%s", base, format) // 文件
			} else {
				outputPath = base // 目录
			}
		}

		// 获取所有待压缩文件
		var sourcePaths []string
		for _, src := range args {
			files, err := uc.GetFileList(src)
			if err != nil {
				log.Warn("跳过无效路径:", src, ", 错误:", err)
				continue
			}
			sourcePaths = append(sourcePaths, files...)
		}
		if len(sourcePaths) == 0 {
			log.Error("无有效待压缩文件")
			return
		}

		// 加密参数校验
		var keyBytes []byte
		var salt string
		if encrypt {
			if key == "" {
				log.Error("加密模式下必须指定密钥 (--key/-k)")
				return
			}
			// 校验密钥长度
			if keyLength == 0 {
				keyLength = uc.AES256KeyLength
			}
			if keyLength != uc.AES128KeyLength && keyLength != uc.AES192KeyLength && keyLength != uc.AES256KeyLength {
				log.Error("无效的密钥长度:", keyLength, ", 仅支持 16/24/32")
				return
			}
			// 补全密钥
			paddedKey, err := uc.PadKey("aes", key)
			if err != nil {
				log.Error("密钥处理失败:", err)
				return
			}
			if len(paddedKey) < keyLength {
				paddedKey = append(paddedKey, make([]byte, keyLength-len(paddedKey))...)
			}
			keyBytes = paddedKey[:keyLength]

			// 生成盐值
			saltBytes, err := uc.GenerateSalt(16)
			if err != nil {
				log.Error("生成盐值失败:", err)
				return
			}
			salt = saltBytes
			if utils.VerboseMode() {
				log.Success("盐值成功生成!")
			}
		}

		// 构建压缩配置
		opts := compress.CompressOptions{
			SourcePaths: sourcePaths,
			OutputPath:  outputPath,
			Format:      format,
			SplitSize:   splitSize,
			Encrypt:     encrypt,
			Key:         keyBytes,
			Verify:      verify,
			SplitSuffix: ".%03d", // 分卷后缀 .001/.002
			EncryptSalt: salt,
			KeyLength:   keyLength,
		}

		// 执行压缩
		if err := compress.RunCompress(opts); err != nil {
			log.Error("压缩失败:", err)

			// 失败清理逻辑
			log.Info("开始清理损坏的压缩文件...")
			// 清理主压缩包
			if uc.CheckPathExist(opts.OutputPath) {
				if err := os.Remove(opts.OutputPath); err != nil {
					log.Warn("清理主压缩包失败:", opts.OutputPath, ", 错误:", err)
				} else {
					log.Success("已清理:", opts.OutputPath)
				}
			}
			// 清理分卷文件
			if opts.SplitSize > 0 {
				volumeNum := 1
				for {
					splitPath := fmt.Sprintf("%s.%03d", opts.OutputPath, volumeNum)
					if !uc.CheckPathExist(splitPath) {
						break
					}
					if err := os.Remove(splitPath); err != nil {
						log.Warn("清理分卷失败:", splitPath, ", 错误:", err)
					} else {
						log.Success("已清理:", splitPath)
					}
					volumeNum++
				}
			}
			// 清理临时文件
			if opts.TempFilePath != "" && uc.CheckPathExist(opts.TempFilePath) {
				if err := os.Remove(opts.TempFilePath); err != nil {
					log.Error("清理临时文件失败:", opts.TempFilePath, ", 错误:", err)
				} else {
					log.Success("已清理:", opts.TempFilePath)
				}
			}

			return
		}

		// 成功提示
		log.Success("压缩完成, 输出路径:", outputPath)
		if encrypt {
			log.Info("加密盐值:", salt)
			log.Info("密钥长度:", keyLength)
		}
		if verify {
			log.Success("完整性校验通过")
		}
	},
}

// InitCompress 初始化命令
func InitCompress() {
	cmd.GetRootCmd().AddCommand(compressCmd)

	// 注册命令参数
	compressCmd.Flags().StringP("output", "o", "", "输出压缩包路径, 简易模式自动补全")
	compressCmd.Flags().StringP("format", "f", "zip", "压缩格式 (zip/targz)")
	compressCmd.Flags().Int64P("split", "s", 0, "分卷大小 (字节, 如 104857600 = 100MB)")
	compressCmd.Flags().BoolP("encrypt", "e", false, "启用 AES 加密 (需指定 --key)")
	compressCmd.Flags().StringP("key", "k", "", "加密密钥")
	compressCmd.Flags().BoolP("verify", "r", false, "压缩后校验完整性 (CRC32)")
	compressCmd.Flags().IntP("key-length", "l", uc.AES256KeyLength, "密钥长度 (16/24/32, 对应 AES-128/192/256)")

	// 绑定参数到 Viper
	_ = viper.BindPFlag("compress.format", compressCmd.Flags().Lookup("format"))
	_ = viper.BindPFlag("compress.split", compressCmd.Flags().Lookup("split"))
	_ = viper.BindPFlag("compress.key-length", compressCmd.Flags().Lookup("key-length"))
}
