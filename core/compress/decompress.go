// Package compress /core/compress/decompress.go
package compress

import (
	"fmt"

	"github.com/GoFurry/gf-file-tool/utils"
	"github.com/GoFurry/gf-file-tool/utils/compress"
	"github.com/GoFurry/gf-file-tool/utils/log"
)

// DecompressOptions 解压缩配置
type DecompressOptions struct {
	SourcePath  string // 压缩包路径
	OutputDir   string // 输出目录
	Format      string // 压缩格式 zip/7z/targz
	Encrypt     bool   // 是否加密
	Key         []byte // 解密密钥
	Verify      bool   // 校验完整性
	EncryptSalt string // 解密盐值
	ExpectedCRC string // 预期 CRC32
}

// Decompressor 解压缩器接口
type Decompressor interface {
	Decompress(opts DecompressOptions) error
}

// NewDecompressor 创建解压缩器
func NewDecompressor(format string) (Decompressor, error) {
	switch format {
	case "zip":
		return &ZipDecompressor{}, nil
	case "targz":
		return &TarGzDecompressor{}, nil
	case "7z":
		return nil, fmt.Errorf("7z 格式暂未实现")
	default:
		return nil, fmt.Errorf("不支持的解压缩格式: %s", format)
	}
}

// RunDecompress 统一解压缩入口
func RunDecompress(opts DecompressOptions) error {
	// 参数校验
	if opts.SourcePath == "" || !compress.CheckPathExist(opts.SourcePath) {
		return fmt.Errorf("压缩包不存在: %s", opts.SourcePath)
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "."
	}
	if err := compress.MkdirIfNotExist(opts.OutputDir); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}
	// 加密参数校验
	if opts.Encrypt && len(opts.Key) == 0 {
		return fmt.Errorf("解密模式必须指定有效密钥")
	}

	// 创建解压缩器
	decompressor, err := NewDecompressor(opts.Format)
	if err != nil {
		return err
	}

	if utils.VerboseMode() {
		log.Info("开始解压缩:", opts.SourcePath, "→", opts.OutputDir)
		if opts.Encrypt {
			log.Info("启用解密模式")
		}
		if opts.Verify {
			log.Success("启用完整性校验")
		}
	}

	// 执行解压缩
	err = decompressor.Decompress(opts)
	if err != nil {
		return fmt.Errorf("解压缩失败: %v", err)
	}

	// 整体压缩包校验
	if opts.Verify && opts.ExpectedCRC != "" {
		if ok, err := compress.VerifyFileCRC32(opts.SourcePath, opts.ExpectedCRC); err != nil {
			log.Warn("校验压缩包失败:", err)
		} else if !ok {
			log.Error("压缩包", opts.SourcePath, "CRC32 不匹配")
		} else if utils.VerboseMode() {
			log.Success("压缩包", opts.SourcePath, "CRC32 校验通过")
		}
	}

	if utils.VerboseMode() && !utils.QuietMode() {
		log.Success("解压缩完成, 输出目录:", opts.OutputDir)
	}
	return nil
}
