// Package compress /core/compress/zip.go
package compress

import (
	"archive/zip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/GoFurry/gf-file-tool/progress"
	"github.com/GoFurry/gf-file-tool/utils"
	"github.com/GoFurry/gf-file-tool/utils/compress"
	"github.com/GoFurry/gf-file-tool/utils/log"
)

// 加密会破坏冗余数据导致压缩失效, 所以加密正确的逻辑应该放在压缩之后而不是压缩之前.
// 此处故意留作为不影响工具正常功能的缺陷, 大家可以思考优化方式.

// ============================== Zip 压缩部分 ==============================

// ZipCompressor Zip 压缩器
type ZipCompressor struct{}

// Compress Zip 压缩逻辑
func (z *ZipCompressor) Compress(opts *CompressOptions) error {
	// 不分卷逻辑
	if opts.SplitSize <= 0 {
		return z.compressSingleFile(*opts)
	}

	// 分卷压缩逻辑
	return z.compressSplitFiles(opts)
}

// compressSingleFile 单文件压缩
func (z *ZipCompressor) compressSingleFile(opts CompressOptions) error {
	// 创建输出文件
	outFile, err := os.Create(opts.OutputPath)
	if err != nil {
		return fmt.Errorf("创建压缩包失败: %v", err)
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			fmt.Printf("关闭文件写入器失败: %v\n", err)
		}
	}()

	// 初始化 Zip Writer
	zipWriter := zip.NewWriter(outFile)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			fmt.Printf("关闭 Zip 写入器失败: %v\n", err)
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
			return fmt.Errorf("获取文件信息失败:%s, 错误: %v", srcPath, err)
		}

		// 计算相对路径
		relPath := srcPath
		if len(opts.SourcePaths) > 1 {
			// 取相对于第一个文件目录的路径
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

		// 创建 Zip 文件头
		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return fmt.Errorf("创建文件头失败: %s, 错误: %v", srcPath, err)
		}
		header.Name = relPath
		header.Method = zip.Deflate // 启用压缩
		header.SetMode(fileInfo.Mode())

		// 创建 Zip 写入器
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("创建 Zip 写入器失败: %s, 错误: %v", srcPath, err)
		}

		// 单个文件进度条
		fileBar := progress.NewFileProgressBar(fileInfo.Size(), relPath)

		// 分块拷贝
		buf := make([]byte, 4*1024*1024) // 4MB 缓冲区
		totalWritten := int64(0)

		if opts.Encrypt {
			// AES-GCM 自定义加密写入
			block, err := aes.NewCipher(opts.Key)
			if err != nil {
				return fmt.Errorf("初始化 AES 加密失败: %v", err)
			}

			gcm, err := cipher.NewGCM(block)
			if err != nil {
				return fmt.Errorf("初始化 GCM 模式失败: %v", err)
			}

			// 生成随机 nonce
			nonce := make([]byte, gcm.NonceSize())
			if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
				return fmt.Errorf("生成 nonce 失败: %v", err)
			}

			// 先写长度, 再写盐值
			// 写入 nonce
			if _, err := writer.Write(nonce); err != nil {
				return fmt.Errorf("写入 nonce 失败: %v", err)
			}

			// 写入盐值
			saltBytes := []byte(opts.EncryptSalt)
			saltLenBuf := make([]byte, 4)
			binary.BigEndian.PutUint32(saltLenBuf, uint32(len(saltBytes)))
			// 写入盐值长度
			if _, err := writer.Write(saltLenBuf); err != nil {
				return fmt.Errorf("写入盐值长度失败: %v", err)
			}
			// 写入盐值内容
			if _, err := writer.Write(saltBytes); err != nil {
				return fmt.Errorf("写入盐值失败: %v", err)
			}

			// 分块加密写入
			buf = make([]byte, 4*1024*1024) // 4MB 缓冲区
			totalWritten = int64(0)
			blockIndex := uint64(0) // 固定块索引

			for {
				n, err := file.Read(buf)
				if err != nil && err != io.EOF {
					return fmt.Errorf("读取文件失败: %s, 错误: %v", srcPath, err)
				}
				if n == 0 {
					break
				}

				// 标准库 GCM 加密: Seal (nil, nonce, 明文, 附加数据)
				subNonce := make([]byte, len(nonce))
				copy(subNonce, nonce)
				binary.BigEndian.PutUint64(subNonce[4:], blockIndex) // 用固定块索引

				// 加密当前块
				cipherText := gcm.Seal(nil, subNonce, buf[:n], nil)

				// 先写入块长度
				lenBuf := make([]byte, 8)
				binary.BigEndian.PutUint64(lenBuf, uint64(len(cipherText)))
				if _, err := writer.Write(lenBuf); err != nil {
					return fmt.Errorf("写入加密块长度失败: %v", err)
				}

				// 写入加密数据
				if _, err := writer.Write(cipherText); err != nil {
					return fmt.Errorf("加密写入失败: %s, 错误: %v", srcPath, err)
				}

				totalWritten += int64(n)
				blockIndex++ // 块索引递增
				if fileBar != nil {
					_ = fileBar.Set64(totalWritten)
				}
			}
		} else {
			// 非加密写入
			for {
				n, err := file.Read(buf)
				if err != nil && err != io.EOF {
					return fmt.Errorf("读取文件失败: %s, 错误: %v", srcPath, err)
				}
				if n == 0 {
					break
				}

				if _, err := writer.Write(buf[:n]); err != nil {
					return fmt.Errorf("写入 Zip 失败: %s, 错误: %v", srcPath, err)
				}

				totalWritten += int64(n)
				if fileBar != nil {
					_ = fileBar.Set64(totalWritten)
				}
			}
		}

		// 手动关闭防止泄露
		progress.FinishProgress(fileBar)
		err = file.Close()
		if err != nil {
			log.Warn("文件关闭失败")
		}

		// 打印成功日志
		if utils.VerboseMode() {
			log.Success("已压缩:", relPath, "/", totalWritten, "字节")
		}
	}

	return nil
}

// compressSplitFiles 分卷压缩
func (z *ZipCompressor) compressSplitFiles(opts *CompressOptions) error {
	// 创建临时压缩包
	tempZip := opts.OutputPath + ".tmp"
	tempOpts := *opts // 浅拷贝, 不修改指针内容
	tempOpts.OutputPath = tempZip
	tempOpts.SplitSize = 0 // 临时包不分卷

	// 压缩为完整包
	if err := z.compressSingleFile(tempOpts); err != nil {
		removeErr := os.Remove(tempZip)
		if removeErr != nil {
			log.Warn("tempZip 清理临时压缩包失败", removeErr)
		}
		return fmt.Errorf("创建临时压缩包失败: %v", removeErr)
	}

	// 打开临时包
	tempFile, err := os.Open(tempZip)
	if err != nil {
		removeErr := os.Remove(tempZip)
		if removeErr != nil {
			log.Warn("tempZip 清理临时压缩包失败", removeErr)
		}
		return fmt.Errorf("打开临时压缩包失败：%v", err)
	}

	// 获取临时包大小
	tempInfo, err := tempFile.Stat()
	if err != nil {
		closeErr := tempFile.Close()
		if closeErr != nil {
			log.Warn("tempFile 文件关闭失败", err)
		}
		removeErr := os.Remove(tempZip)
		if removeErr != nil {
			log.Warn("tempZip 清理临时压缩包失败", removeErr)
		}
		return fmt.Errorf("获取临时包信息失败: %v", closeErr)
	}
	tempSize := tempInfo.Size()

	// 计算分卷数
	splitCount := (tempSize + opts.SplitSize - 1) / opts.SplitSize
	if utils.VerboseMode() {
		log.Info("开始分卷: 总大小", tempSize, "字节, 分卷大小", opts.SplitSize, "字节, 共", splitCount, "卷")
	}

	// 分卷切割
	remaining := tempSize
	volumeNum := int64(1)
	splitBar := progress.NewBatchProgressBar(int(splitCount))
	defer progress.FinishProgress(splitBar)

	for remaining > 0 {
		progress.UpdateProgress(splitBar, 1)
		currentSize := opts.SplitSize
		if remaining < currentSize {
			currentSize = remaining
		}

		// 分卷文件名
		splitPath := fmt.Sprintf("%s.%03d", opts.OutputPath, volumeNum)
		if opts.SplitSuffix != "" {
			splitPath = fmt.Sprintf("%s%s", opts.OutputPath, fmt.Sprintf(opts.SplitSuffix, volumeNum))
		}

		// 创建分卷文件
		splitFile, err := os.Create(splitPath)
		if err != nil {
			closeErr := tempFile.Close()
			if closeErr != nil {
				log.Warn("tempFile 文件关闭失败", err)
			}
			removeErr := os.Remove(tempZip)
			if removeErr != nil {
				log.Warn("tempZip 清理临时压缩包失败", removeErr)
			}
			return fmt.Errorf("创建分卷 %s 失败：%v", splitPath, err)
		}

		// 写入分卷数据
		written, err := io.CopyN(splitFile, tempFile, currentSize)
		closeErr := splitFile.Close() // 立即关闭分卷文件句柄
		if closeErr != nil {
			log.Warn("splitFile 文件关闭失败", err)
		}

		if err != nil && err != io.EOF {
			removeErr := os.Remove(splitPath)
			if removeErr != nil {
				log.Warn("splitPath 清理临时压缩包失败", removeErr)
			}
			closeErr = tempFile.Close()
			if closeErr != nil {
				log.Warn("tempFile 文件关闭失败", err)
			}
			removeErr = os.Remove(tempZip)
			if removeErr != nil {
				log.Warn("tempZip 清理临时压缩包失败", removeErr)
			}
			return fmt.Errorf("写入分卷 %s 失败: %v", splitPath, err)
		}

		if utils.VerboseMode() {
			log.Success("生成分卷:", splitPath, written, "字节")
		}

		remaining -= written
		volumeNum++
	}

	// 切割完成后立即关闭临时文件句柄
	closeErr := tempFile.Close()
	if closeErr != nil {
		log.Warn("tempFile 文件关闭失败:", closeErr)
	}

	// 修改指针, 用于外层兜底清除临时文件
	opts.TempFilePath = tempZip

	// 生成分卷说明文件 TODO:新增一个可选参数
	//manifestPath := opts.OutputPath + ".split"
	//manifest, err := os.Create(manifestPath)
	//if err != nil {
	//	log.Warn("创建分卷说明文件失败:", err)
	//} else {
	//	_, _ = manifest.WriteString(fmt.Sprintf("Source: %s\n", opts.OutputPath))
	//	_, _ = manifest.WriteString(fmt.Sprintf("TotalSize: %d\n", tempSize))
	//	_, _ = manifest.WriteString(fmt.Sprintf("SplitSize: %d\n", opts.SplitSize))
	//	_, _ = manifest.WriteString(fmt.Sprintf("VolumeCount: %d\n", splitCount))
	//	_ = manifest.Close()
	//	if utils.VerboseMode() {
	//		log.Success("生成分卷说明:", manifestPath)
	//	}
	//}

	return nil
}

// ============================== Zip 解压缩部分 ==============================

// ZipDecompressor Zip 解压缩器
type ZipDecompressor struct{}

// Decompress 解压缩逻辑
func (z *ZipDecompressor) Decompress(opts DecompressOptions) error {
	// 检测分卷并合并
	if compress.IsSplitFile(opts.SourcePath) {
		// 合并分卷为完整压缩包
		mergedPath := opts.SourcePath + ".merged"
		if err := compress.MergeSplitFiles(opts.SourcePath, mergedPath); err != nil {
			return fmt.Errorf("合并分卷失败: %v", err)
		}
		// 替换为合并后的路径, 解压完成后删除临时文件
		oldSourcePath := opts.SourcePath
		opts.SourcePath = mergedPath
		defer func() {
			// 兜底删除临时文件
			if err := os.Remove(mergedPath); err != nil && utils.VerboseMode() {
				log.Warn("清理合并临时文件失败:", err)
			}
		}()
		if utils.VerboseMode() {
			log.Success("分卷合并完成:", oldSourcePath, "→", mergedPath)
		}
	}

	// 执行解压
	return z.decompressSingleFile(opts)
}

// decompressSingleFile Zip 解压缩
func (z *ZipDecompressor) decompressSingleFile(opts DecompressOptions) error {
	// 打开压缩包
	zipFile, err := os.Open(opts.SourcePath)
	if err != nil {
		return fmt.Errorf("打开压缩包失败: %v", err)
	}
	// 解压流程结束再关闭 zipFile
	defer func() {
		if err := zipFile.Close(); err != nil && utils.VerboseMode() {
			log.Warn("关闭压缩包文件失败:", err)
		}
	}()

	fileInfo, err := zipFile.Stat()
	if err != nil {
		return fmt.Errorf("获取压缩包信息失败: %v", err)
	}

	zipReader, err := zip.NewReader(zipFile, fileInfo.Size())
	if err != nil {
		return fmt.Errorf("初始化 Zip 读取器失败: %v", err)
	}

	// 临时文件列表
	var tempFiles []string
	defer func() {
		// 兜底清理临时文件
		for _, tempPath := range tempFiles {
			if err = os.Remove(tempPath); err != nil && utils.VerboseMode() {
				log.Warn("清理临时文件失败:", tempPath, ", 错误:", err)
			}
		}
	}()

	// 批量进度条
	batchBar := progress.NewBatchProgressBar(len(zipReader.File))
	defer progress.FinishProgress(batchBar)

	for _, file := range zipReader.File {
		progress.UpdateProgress(batchBar, 1)

		// 构建输出路径
		outputPath := filepath.Join(opts.OutputDir, file.Name)
		if utils.VerboseMode() {
			fmt.Println()
			log.Info("解压文件:", file.Name, "→", outputPath)
		}

		// 处理目录
		if file.FileInfo().IsDir() {
			if err = compress.MkdirIfNotExist(outputPath); err != nil {
				return fmt.Errorf("创建目录失败: %s, 错误: %v", outputPath, err)
			}
			continue
		}

		// 创建文件目录
		if err = compress.MkdirIfNotExist(filepath.Dir(outputPath)); err != nil {
			return fmt.Errorf("创建文件目录失败: %s, 错误: %v", filepath.Dir(outputPath), err)
		}

		// 打开压缩包内文件
		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("打开压缩包内文件失败: %s, 错误: %v", file.Name, err)
		}
		// 每个文件读取完立即关闭 srcFile
		defer func() {
			if err := srcFile.Close(); err != nil && utils.VerboseMode() {
				log.Warn("关闭压缩包内文件失败:", file.Name, ", 错误:", err)
			}
		}()

		// 处理加密文件
		if opts.Encrypt {
			// 初始化 AES-GCM
			block, err := aes.NewCipher(opts.Key)
			if err != nil {
				return fmt.Errorf("初始化 AES 解密失败: %v", err)
			}

			gcm, err := cipher.NewGCM(block)
			if err != nil {
				return fmt.Errorf("初始化 GCM 模式失败: %v", err)
			}

			// 读取 nonce
			nonce := make([]byte, gcm.NonceSize())
			if _, err := io.ReadFull(srcFile, nonce); err != nil {
				return fmt.Errorf("读取 Nonce 失败: %s, 错误: %v", file.Name, err)
			}

			// 读取盐值长度
			saltLenBuf := make([]byte, 4)
			if _, err := io.ReadFull(srcFile, saltLenBuf); err != nil {
				return fmt.Errorf("读取盐值长度失败: %s, 错误: %v", file.Name, err)
			}
			saltLen := binary.BigEndian.Uint32(saltLenBuf)

			// 读取盐值内容
			saltBytes := make([]byte, saltLen)
			if _, err := io.ReadFull(srcFile, saltBytes); err != nil {
				return fmt.Errorf("读取盐值失败: %s, 错误: %v", file.Name, err)
			}

			// 验证盐值
			if opts.EncryptSalt != "" && string(saltBytes) != opts.EncryptSalt {
				return fmt.Errorf("盐值不匹配: 预期 %s, 实际 %s", opts.EncryptSalt, string(saltBytes))
			}

			// 创建输出文件
			dstFile, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("创建输出文件失败: %s, 错误: %v", outputPath, err)
			}
			defer func() {
				if err := dstFile.Close(); err != nil && utils.VerboseMode() {
					log.Warn("关闭输出文件失败:", outputPath, ", 错误:", err)
				}
			}()

			// 分块解密读取
			fileBar := progress.NewFileProgressBar(int64(file.UncompressedSize64), file.Name)
			defer progress.FinishProgress(fileBar)

			totalWritten := int64(0)
			blockIndex := uint64(0) // 固定块索引

			for {
				lenBuf := make([]byte, 8)
				n, err := io.ReadFull(srcFile, lenBuf)
				if err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("读取加密块长度失败: %s, 错误: %v", file.Name, err)
				}
				if n != 8 {
					return fmt.Errorf("无效的加密块长度: %s", file.Name)
				}
				cipherLen := binary.BigEndian.Uint64(lenBuf)

				// 读取加密块数据
				cipherText := make([]byte, cipherLen)
				if _, err := io.ReadFull(srcFile, cipherText); err != nil {
					return fmt.Errorf("读取加密块数据失败: %s, 错误: %v", file.Name, err)
				}

				// 生成子 Nonce
				subNonce := make([]byte, len(nonce))
				copy(subNonce, nonce)
				binary.BigEndian.PutUint64(subNonce[4:], blockIndex) // 用固定块索引

				// 解密当前块
				plainText, err := gcm.Open(nil, subNonce, cipherText, nil)
				if err != nil {
					return fmt.Errorf("解密块失败: %s, 错误: %v (块索引: %d, 密码/盐值错误或文件损坏)", file.Name, err, blockIndex)
				}

				// 写入明文
				if _, err := dstFile.Write(plainText); err != nil {
					return fmt.Errorf("写入解密文件失败: %s, 错误: %v", outputPath, err)
				}

				totalWritten += int64(len(plainText))
				blockIndex++ // 块索引递增
				if fileBar != nil {
					_ = fileBar.Set64(totalWritten)
				}
			}

			// 保留权限
			if err := os.Chmod(outputPath, file.Mode()); err != nil && utils.VerboseMode() {
				log.Warn("设置文件权限失败:", outputPath, ", 错误:", err)
			}
		} else {
			dstFile, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("创建输出文件失败: %s, 错误: %v", outputPath, err)
			}
			defer func() {
				if err := dstFile.Close(); err != nil && utils.VerboseMode() {
					log.Warn("关闭输出文件失败:", outputPath, ", 错误:", err)
				}
			}()

			// 分块拷贝非加密文件
			fileBar := progress.NewFileProgressBar(int64(file.UncompressedSize64), file.Name)
			defer progress.FinishProgress(fileBar)

			buf := make([]byte, 4*1024*1024) // 4MB 缓冲区
			totalWritten := int64(0)
			for {
				n, err := srcFile.Read(buf)
				if err != nil && err != io.EOF {
					return fmt.Errorf("读取压缩包内文件失败: %s, 错误: %v", file.Name, err)
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

			// 保留权限
			if err := os.Chmod(outputPath, file.Mode()); err != nil && utils.VerboseMode() {
				log.Warn("设置文件权限失败:", outputPath, ", 错误:", err)
			}
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

	return nil
}
