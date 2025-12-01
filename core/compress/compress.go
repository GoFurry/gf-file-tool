// Package compress /core/compress/compress.go
package compress

import (
	"fmt"
	"os"

	"github.com/GoFurry/gf-file-tool/utils"
	"github.com/GoFurry/gf-file-tool/utils/compress"
	"github.com/GoFurry/gf-file-tool/utils/log"
)

// CompressOptions 压缩配置
type CompressOptions struct {
	SourcePaths  []string // 待压缩文件/目录列表
	OutputPath   string   // 输出压缩包路径
	Format       string   // 压缩格式 zip/7z/targz
	SplitSize    int64    // 分卷字节大小 =0不分卷
	Encrypt      bool     // 是否加密
	Key          []byte   // 加密密钥
	KeyLength    int      // 密钥长度 16/24/32
	EncryptSalt  string   // 加密盐值
	Verify       bool     // 是否校验完整性
	SplitSuffix  string   // 分卷后缀 如.001/.002
	TotalSize    int64    // 分卷文件总大小
	TempFilePath string   // 临时文件路径
}

// Compressor 压缩器接口
type Compressor interface {
	Compress(opts *CompressOptions) error
}

// NewCompressor 创建对应压缩器
func NewCompressor(format string) (Compressor, error) {
	switch format {
	case "zip":
		return &ZipCompressor{}, nil
	case "targz":
		return &TarGzCompressor{}, nil
	default:
		return nil, fmt.Errorf("不支持的压缩格式：%s，仅支持 zip/targz", format)
	}
}

// RunCompress 压缩入口
func RunCompress(opts CompressOptions) error {
	log.Info("压缩开始")

	// 参数校验
	if len(opts.SourcePaths) == 0 {
		return fmt.Errorf("待压缩文件列表为空")
	}
	if opts.OutputPath == "" {
		return fmt.Errorf("输出路径不能为空")
	}

	// 计算文件总大小
	var totalSize int64
	for _, src := range opts.SourcePaths {
		info, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf("获取文件大小失败：%s，错误：%v", src, err)
		}
		totalSize += info.Size()
	}
	opts.TotalSize = totalSize

	// 创建输出目录
	outputDir := compress.GetDir(opts.OutputPath)
	if err := compress.MkdirIfNotExist(outputDir); err != nil {
		return fmt.Errorf("创建输出目录失败：%v", err)
	}

	// 创建压缩器
	compressor, err := NewCompressor(opts.Format)
	if err != nil {
		return err
	}

	// 执行压缩
	if utils.VerboseMode() {
		log.Info("压缩任务信息:")
		fmt.Printf("   - 文件数量: %d\n", len(opts.SourcePaths))
		fmt.Printf("   - 总大小: %d 字节\n", totalSize)
		fmt.Printf("   - 格式: %s\n", opts.Format)
		fmt.Printf("   - 分卷大小: %d 字节\n", opts.SplitSize)
		fmt.Printf("   - 加密: %t\n", opts.Encrypt)
		fmt.Println("--------------------------------")
	}
	err = compressor.Compress(&opts)
	if err != nil {
		return fmt.Errorf("压缩失败: %v", err)
	}

	// 完整性校验
	if opts.Verify {
		if utils.VerboseMode() {
			log.Info("开始校验压缩包完整性...")
		}

		// 分卷场景校验临时完整包
		verifyPath := opts.OutputPath
		if opts.SplitSize > 0 {
			verifyPath = opts.TempFilePath // 使用临时文件路径
			if !compress.CheckPathExist(verifyPath) {
				log.Warn("分卷压缩: 临时包已清理, 跳过 CRC32 校验")
			} else {
				// 计算 CRC32
				crc, err := compress.CalculateCRC32(verifyPath)
				if err != nil {
					// 校验失败清理临时文件
					removeErr := os.Remove(opts.TempFilePath)
					if removeErr != nil {
						log.Warn("临时文件清理失败:", removeErr)
					}
					return fmt.Errorf("计算 CRC32 失败：%v", err)
				}
				if utils.VerboseMode() {
					log.Success("压缩包 CRC32:", crc)
				}
				// 校验完成后立即清理临时文件
				removeErr := os.Remove(opts.TempFilePath)
				if removeErr != nil {
					log.Warn("临时文件清理失败:", removeErr)
				}
			}
		} else {
			// 非分卷场景
			crc, err := compress.CalculateCRC32(verifyPath)
			if err != nil {
				return fmt.Errorf("计算 CRC32 失败:%v", err)
			}
			if utils.VerboseMode() {
				log.Success("压缩包 CRC32:", crc)
			}
		}
	} else {
		// 不校验时直接清理临时文件
		if opts.SplitSize > 0 && compress.CheckPathExist(opts.TempFilePath) {
			removeErr := os.Remove(opts.TempFilePath)
			if removeErr != nil {
				log.Warn("临时文件清理失败:", removeErr)
			}
		}
	}

	// 兜底清理
	if opts.SplitSize > 0 && compress.CheckPathExist(opts.TempFilePath) {
		removeErr := os.Remove(opts.TempFilePath)
		if removeErr != nil {
			log.Warn("临时文件清理失败:", removeErr)
		}
		if utils.VerboseMode() {
			log.Info("清理临时文件:", opts.TempFilePath)
		}
	}
	return nil
}
