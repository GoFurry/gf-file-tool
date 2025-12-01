package decompress

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/GoFurry/gf-file-tool/cmd"
	"github.com/GoFurry/gf-file-tool/core/compress"
	uc "github.com/GoFurry/gf-file-tool/utils/compress"
	"github.com/GoFurry/gf-file-tool/utils/log"
	"github.com/klauspost/compress/zip"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// decompressCmd 解压缩主命令
var decompressCmd = &cobra.Command{
	Use:   "decompress [source]",
	Short: "解压缩文件/压缩包",
	Long: `解压缩文件/压缩包，支持多格式、批量处理、加密解密、分卷合并:
  简易模式:gf-file-tool decompress test.zip
  加密解密:gf-file-tool decompress test.zip -e -k 123456 -l 32
  分卷合并:gf-file-tool decompress split_big.zip.001 -o ./output
  完整性校验:gf-file-tool decompress test.zip -r --crc32 a18d2fb9`,
	Args: cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		// 解析参数
		outputDir, _ := c.Flags().GetString("output")
		format, _ := c.Flags().GetString("format")
		encrypt, _ := c.Flags().GetBool("encrypt")
		key, _ := c.Flags().GetString("key")
		keyLength, _ := c.Flags().GetInt("key-length")
		verify, _ := c.Flags().GetBool("verify")
		expectedCRC, _ := c.Flags().GetString("crc32")
		salt, _ := c.Flags().GetString("salt")

		// 自动补全输出目录
		if outputDir == "" {
			base := filepath.Base(args[0])
			ext := filepath.Ext(base)
			// 分卷文件去掉后缀
			if len(ext) == 4 && ext[0] == '.' {
				base = base[:len(base)-4]
				ext = filepath.Ext(base)
			}
			outputDir = base[:len(base)-len(ext)] + "_unzip"
		}

		// 构建配置
		opts := compress.DecompressOptions{
			SourcePath:  args[0],
			OutputDir:   outputDir,
			Format:      format,
			Encrypt:     encrypt,
			Verify:      verify,
			EncryptSalt: salt,
			ExpectedCRC: expectedCRC,
		}

		// 自动读取 Zip 注释中的盐值和密钥长度
		if encrypt && opts.Format == "zip" {
			// 读取 Zip 注释
			r, err := zip.OpenReader(opts.SourcePath)
			if err == nil {
				defer r.Close()
				comment := r.Comment
				if comment != "" && strings.HasPrefix(comment, "gf-encrypt:") {
					// 解析注释
					commentParts := strings.Split(strings.TrimPrefix(comment, "gf-encrypt:"), ";")
					for _, part := range commentParts {
						kv := strings.Split(part, "=")
						if len(kv) != 2 {
							continue
						}
						switch kv[0] {
						case "salt":
							opts.EncryptSalt = kv[1]
							log.Info("从压缩包注释读取盐值:", opts.EncryptSalt)
						case "key-length":
							kl, _ := strconv.Atoi(kv[1])
							keyLength = kl
							log.Info("从压缩包注释读取密钥长度:", keyLength)
						}
					}
				}
			}
		}

		// 解密参数处理
		if encrypt {
			if key == "" {
				log.Error("解密模式下必须指定密钥 (--key/-k)")
				return
			}
			// 默认 AES256
			if keyLength == 0 {
				keyLength = uc.AES256KeyLength
			}
			// 填充密钥
			paddedKey, err := uc.PadKey("aes", key)
			if err != nil {
				log.Error("密钥处理失败:", err)
				return
			}
			if len(paddedKey) < keyLength {
				paddedKey = append(paddedKey, make([]byte, keyLength-len(paddedKey))...)
			}
			opts.Key = paddedKey[:keyLength]
		}

		// 自动识别格式
		if opts.Format == "" {
			ext := filepath.Ext(args[0])
			switch ext {
			case ".zip":
				opts.Format = "zip"
			case ".tar.gz", ".tgz":
				opts.Format = "targz"
			default:
				log.Warn("自动识别格式失败, 默认使用 zip")
				opts.Format = "zip"
			}
		}

		// 执行解压缩
		if err := compress.RunDecompress(opts); err != nil {
			log.Error("解压缩失败:", err)

			// 清理损坏的解压文件
			log.Info("开始清理损坏的解压文件...")
			if uc.CheckPathExist(opts.OutputDir) {
				if err := os.RemoveAll(opts.OutputDir); err != nil {
					log.Warn("清理解压目录失败:", opts.OutputDir, ", 错误:", err)
				} else {
					log.Success("已清理解压目录:", opts.OutputDir)
				}
			}
			// 清理分卷合并的临时文件
			mergedPath := opts.SourcePath + ".merged"
			if uc.CheckPathExist(mergedPath) {
				if err := os.Remove(mergedPath); err != nil {
					log.Warn("清理分卷合并临时文件失败:", mergedPath, ", 错误:", err)
				} else {
					log.Success("已清理:", mergedPath)
				}
			}
			return
		}
	},
}

// InitDecompress 初始化命令
func InitDecompress() {
	cmd.GetRootCmd().AddCommand(decompressCmd)

	// 注册参数
	decompressCmd.Flags().StringP("output", "o", "", "输出目录（简易模式自动补全为 压缩包名_unzip）")
	decompressCmd.Flags().StringP("format", "f", "", "压缩格式（自动识别：zip/targz）")
	decompressCmd.Flags().BoolP("encrypt", "e", false, "启用解密（需指定 --key）")
	decompressCmd.Flags().StringP("key", "k", "", "解密密钥")
	decompressCmd.Flags().IntP("key-length", "l", 32, "密钥长度（AES：16/24/32）")
	decompressCmd.Flags().StringP("salt", "s", "", "解密盐值（与压缩时一致）")
	decompressCmd.Flags().BoolP("verify", "r", false, "解压缩后校验完整性")
	decompressCmd.Flags().StringP("crc32", "c", "", "预期 CRC32 值（用于校验）")

	// 绑定 Viper
	_ = viper.BindPFlag("decompress.format", decompressCmd.Flags().Lookup("format"))
	_ = viper.BindPFlag("decompress.key-length", decompressCmd.Flags().Lookup("key-length"))
}
