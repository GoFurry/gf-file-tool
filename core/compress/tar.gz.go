// Package compress /core/compress/tar.gz.go
package compress

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/GoFurry/gf-file-tool/progress"
	"github.com/GoFurry/gf-file-tool/utils"
	"github.com/GoFurry/gf-file-tool/utils/compress"
	"github.com/GoFurry/gf-file-tool/utils/log"
)

// ============================== tar.gz 压缩部分 ==============================

// TarGzCompressor tar.gz 压缩器 (仅支持基础压缩, 无分卷/加密)
type TarGzCompressor struct{}

// Compress tar.gz 压缩逻辑
func (t *TarGzCompressor) Compress(opts *CompressOptions) error {
	// 禁用分卷
	if opts.SplitSize > 0 {
		return fmt.Errorf("tar.gz 格式不支持分卷压缩, 请使用 zip 格式")
	}

	// 禁用加密
	if opts.Encrypt {
		return fmt.Errorf("tar.gz 格式不支持加密压缩, 请使用 zip 格式")
	}

	// 执行基础压缩逻辑
	return t.compressSingleFile(*opts)
}

// compressSingleFile tar.gz 单文件压缩
func (t *TarGzCompressor) compressSingleFile(opts CompressOptions) error {
	// 创建输出文件
	outFile, err := os.Create(opts.OutputPath)
	if err != nil {
		return fmt.Errorf("创建 tar.gz 文件失败: %v", err)
	}
	defer outFile.Close()

	// 初始化 gzip Writer
	gzWriter, err := gzip.NewWriterLevel(outFile, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("初始化 gzip 写入器失败: %v", err)
	}
	defer func() {
		if err := gzWriter.Close(); err != nil {
			log.Warn("关闭 gzip 写入器失败:", err)
		}
	}()

	// 初始化 tar Writer
	tarWriter := tar.NewWriter(gzWriter)
	defer func() {
		if err := tarWriter.Close(); err != nil {
			log.Warn("关闭 tar 写入器失败:", err)
		}
	}()

	// 批量进度条
	batchBar := progress.NewBatchProgressBar(len(opts.SourcePaths))
	defer progress.FinishProgress(batchBar)

	// 遍历文件压缩
	for _, srcPath := range opts.SourcePaths {
		progress.UpdateProgress(batchBar, 1)

		// 打开源文件
		file, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("打开文件失败: %s, 错误: %v", srcPath, err)
		}

		// 获取文件信息
		fileInfo, err := file.Stat()
		if err != nil {
			return fmt.Errorf("获取文件信息失败: %s, 错误: %v", srcPath, err)
		}

		// 计算相对路径
		relPath := srcPath
		if len(opts.SourcePaths) > 1 {
			baseDir := filepath.Dir(opts.SourcePaths[0])
			rel, err := filepath.Rel(baseDir, srcPath)
			if err == nil {
				relPath = rel
			} else {
				relPath = filepath.Base(srcPath)
			}
		} else {
			relPath = filepath.Base(srcPath)
		}

		// 创建 Tar 头
		header, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return fmt.Errorf("创建 tar 文件头失败: %s, 错误: %v", srcPath, err)
		}
		header.Name = relPath
		header.Mode = int64(fileInfo.Mode().Perm())
		header.Size = fileInfo.Size()

		// 写入 Tar 头
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("写入 tar 头失败: %s, 错误: %v", srcPath, err)
		}

		// 单个文件进度条
		fileBar := progress.NewFileProgressBar(fileInfo.Size(), relPath)

		// 分块拷贝
		buf := make([]byte, 4*1024*1024)
		totalWritten := int64(0)
		for {
			n, err := file.Read(buf)
			if err != nil && err != io.EOF {
				return fmt.Errorf("读取文件失败: %s, 错误: %v", srcPath, err)
			}
			if n == 0 {
				break
			}

			if _, err := tarWriter.Write(buf[:n]); err != nil {
				return fmt.Errorf("写入 tar 失败: %s, 错误: %v", srcPath, err)
			}

			totalWritten += int64(n)
			if fileBar != nil {
				_ = fileBar.Set64(totalWritten)
			}
		}

		// 手动关闭, 防止泄露
		closeErr := file.Close()
		if closeErr != nil {
			log.Warn("文件关闭失败")
		}
		progress.FinishProgress(fileBar)

		if utils.VerboseMode() {
			log.Success("已压缩:", relPath, "总计", totalWritten, "字节")
		}
	}

	return nil
}

// ============================== tar.gz 解压缩部分 ==============================

// TarGzDecompressor tar.gz 解压缩器
type TarGzDecompressor struct{}

// Decompress tar.gz 解压缩逻辑
func (t *TarGzDecompressor) Decompress(opts DecompressOptions) error {
	// 禁用分卷
	if compress.IsSplitFile(opts.SourcePath) {
		return fmt.Errorf("tar.gz 格式不支持分卷解压，请使用 zip 格式")
	}

	// 禁用加密
	if opts.Encrypt {
		return fmt.Errorf("tar.gz 格式不支持加密解密，请使用 zip 格式")
	}

	// 打开压缩包
	tarFile, err := os.Open(opts.SourcePath)
	if err != nil {
		return fmt.Errorf("打开 tar.gz 失败：%v", err)
	}
	// 解压流程结束再关闭 tarFile
	defer func() {
		if err := tarFile.Close(); err != nil && utils.VerboseMode() {
			log.Warn("关闭 tar.gz 文件失败:", err)
		}
	}()

	// 初始化 gzip 读取器
	gzReader, err := gzip.NewReader(tarFile)
	if err != nil {
		return fmt.Errorf("初始化 gzip 读取器失败：%v", err)
	}
	// 解压流程结束再关闭 gzReader
	defer func() {
		if err := gzReader.Close(); err != nil && utils.VerboseMode() {
			log.Warn("关闭 gzip 读取器失败:", err)
		}
	}()

	// 初始化 tar 读取器
	tarReader := tar.NewReader(gzReader)

	// 遍历 tar 内文件
	fileCount := 0
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取 tar 头失败: %v", err)
		}
		fileCount++

		// 构建输出路径
		outputPath := filepath.Join(opts.OutputDir, header.Name)
		if utils.VerboseMode() {
			fmt.Println()
			log.Info("解压文件:", header.Name, "→", outputPath)
		}

		// 处理目录
		if header.Typeflag == tar.TypeDir {
			if err := compress.MkdirIfNotExist(outputPath); err != nil {
				return fmt.Errorf("创建目录失败: %s, 错误: %v", outputPath, err)
			}
			continue
		}

		// 创建文件目录
		if err := compress.MkdirIfNotExist(filepath.Dir(outputPath)); err != nil {
			return fmt.Errorf("创建文件目录失败: %s, 错误: %v", filepath.Dir(outputPath), err)
		}

		// 创建输出文件
		dstFile, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("创建输出文件失败: %s, 错误: %v", outputPath, err)
		}
		// 写入后立即关闭
		defer dstFile.Close()

		// 单个文件进度条
		fileBar := progress.NewFileProgressBar(header.Size, header.Name)
		defer progress.FinishProgress(fileBar)

		// 分块拷贝
		buf := make([]byte, 4*1024*1024) // 4MB 缓冲区
		totalWritten := int64(0)
		for {
			n, err := tarReader.Read(buf)
			if err != nil && err != io.EOF {
				return fmt.Errorf("读取 tar 内文件失败: %s, 错误: %v", header.Name, err)
			}
			if n == 0 {
				break
			}

			if _, err := dstFile.Write(buf[:n]); err != nil {
				return fmt.Errorf("写入文件失败: %s, 错误: %v", outputPath, err)
			}

			totalWritten += int64(n)
			if fileBar != nil {
				_ = fileBar.Set64(totalWritten)
			}
		}

		// 主动刷新并关闭当前文件
		if err := dstFile.Sync(); err != nil && utils.VerboseMode() {
			log.Warn("刷新文件缓存失败:", outputPath, ", 错误:", err)
		}
		if err := dstFile.Close(); err != nil && utils.VerboseMode() {
			log.Warn("关闭输出文件失败:", outputPath, ", 错误:", err)
		}

		// 保留权限
		if err := os.Chmod(outputPath, os.FileMode(header.Mode)); err != nil && utils.VerboseMode() {
			log.Warn("设置文件权限失败:", outputPath, ", 错误:", err)
		}

		// 完整性校验
		if opts.Verify {
			fmt.Println()
			if ok, err := compress.VerifyFileCRC32(outputPath, opts.ExpectedCRC); err != nil {
				log.Warn("校验文件", outputPath, "失败:", err)
			} else if !ok {
				log.Error("文件", outputPath, "CRC32 不匹配")
			} else if utils.VerboseMode() {
				log.Success("文件", outputPath, "CRC32 校验通过")
			}
		}
	}

	if utils.VerboseMode() {
		fmt.Println()
		log.Success("共解压", fileCount, "个文件")
	}
	return nil
}
