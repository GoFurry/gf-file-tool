// Package encrypt /cmd/encrypt/encrypt.go
package encrypt

import (
	"github.com/GoFurry/gf-file-tool/cmd"
	"github.com/GoFurry/gf-file-tool/core/crypto"
	"github.com/GoFurry/gf-file-tool/utils"
	"github.com/GoFurry/gf-file-tool/utils/compress"
	"github.com/GoFurry/gf-file-tool/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// encryptCmd 压缩命令实例
var encryptCmd = &cobra.Command{
	Use:   "encrypt [source...]",
	Short: "加密文件/目录",
	Long: `加密文件/目录:
  gf-file-tool encrypt test.txt -k 123456 -o test.enc
  gf-file-tool encrypt ./docs -k 123456 --algorithm aes`,
	Args: cobra.MinimumNArgs(1),
	Run: func(c *cobra.Command, args []string) {
		// 解析参数
		outputPath, _ := c.Flags().GetString("output")
		algorithm, _ := c.Flags().GetString("algorithm")
		key, _ := c.Flags().GetString("key")
		keyLength, _ := c.Flags().GetInt("key-length")
		salt, _ := c.Flags().GetString("salt")

		// 校验密钥
		if key == "" {
			log.Error("必须指定加密密钥 (--key/-k)")
			return
		}

		// 批量获取文件
		var sourcePaths []string
		for _, src := range args {
			files, err := compress.GetFileList(src)
			if err != nil {
				log.Warn("跳过无效路径:", src, ", 错误:", err)
				continue
			}
			sourcePaths = append(sourcePaths, files...)
		}
		if len(sourcePaths) == 0 {
			log.Error("无有效加密文件")
			return
		}

		// 处理密钥
		if keyLength == 0 {
			if algorithm == "des" {
				keyLength = compress.DESKeyLength
			} else {
				keyLength = compress.AES256KeyLength
			}
		}
		paddedKey, err := compress.PadKey(algorithm, key)
		if err != nil {
			log.Error("密钥处理失败:", err)
			return
		}

		// 逐个加密文件
		for _, src := range sourcePaths {
			dst := outputPath
			if dst == "" {
				dst = src + ".enc"
			}

			// 构建加密配置
			opts := crypto.CryptoOptions{
				Algorithm:  algorithm,
				Key:        paddedKey,
				KeyLength:  keyLength,
				Salt:       salt,
				IsEncrypt:  true,
				SourcePath: src,
				OutputPath: dst,
			}

			// 执行加密
			if err := crypto.RunCrypto(opts); err != nil {
				log.Error("加密失败:", src, ", 错误:", err)
				continue
			}
			if utils.VerboseMode() {
				log.Success("加密成功:", src, "→", dst)
			}
		}
	},
}

// InitEncrypt 初始化命令
func InitEncrypt() {
	cmd.GetRootCmd().AddCommand(encryptCmd)

	// 注册参数
	encryptCmd.Flags().StringP("output", "o", "", "输出加密文件路径 (批量时为目录)")
	encryptCmd.Flags().StringP("algorithm", "a", "aes", "加密算法 (aes/des/3des/aes-ctr/rc4/chacha20)")
	encryptCmd.Flags().StringP("key", "k", "", "加密密钥 (必填)")
	encryptCmd.Flags().IntP("key-length", "l", 32, "密钥长度(AES 16/24/32) (DES 8)")
	encryptCmd.Flags().StringP("salt", "s", "", "加密盐值 (为空自动生成)")

	// 绑定 Viper
	_ = viper.BindPFlag("encrypt.algorithm", encryptCmd.Flags().Lookup("algorithm"))
	_ = viper.BindPFlag("encrypt.key-length", encryptCmd.Flags().Lookup("key-length"))
	_ = viper.BindPFlag("encrypt.salt", encryptCmd.Flags().Lookup("salt"))
}
