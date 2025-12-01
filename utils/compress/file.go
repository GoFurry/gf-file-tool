package compress

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/GoFurry/gf-file-tool/progress"
	"github.com/GoFurry/gf-file-tool/utils"
	"github.com/GoFurry/gf-file-tool/utils/log"
)

// GetSystemTempDir 获取程序专属系统临时目录
func GetSystemTempDir() string {
	tempDir := os.TempDir()
	gfTempDir := filepath.Join(tempDir, "gf-file-tool")
	if err := MkdirIfNotExist(gfTempDir); err != nil {
		return tempDir
	}
	return gfTempDir
}

// GetFileList 批量获取文件列表
// src: 输入路径
// return: 所有文件的绝对路径列表、错误
func GetFileList(src string) ([]string, error) {
	var fileList []string

	// 校验路径是否存在
	info, err := os.Stat(src)
	if err != nil {
		return nil, fmt.Errorf("路径不存在: %s, 错误: %v", src, err)
	}

	// 单文件
	if info.Mode().IsRegular() {
		absPath, _ := filepath.Abs(src) // 绝对路径
		return []string{absPath}, nil
	}

	// 遍历目录所有文件
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("遍历路径失败: %s, 错误: %v", path, err)
		}
		// 跳过目录, 只保留文件
		if info.Mode().IsRegular() {
			absPath, _ := filepath.Abs(path)
			fileList = append(fileList, absPath)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历目录失败: %s, 错误: %v", src, err)
	}

	return fileList, nil
}

// CheckPathExist 校验路径是否存在
func CheckPathExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// MkdirIfNotExist 目录不存在则创建
func MkdirIfNotExist(dir string) error {
	if !CheckPathExist(dir) {
		// 其他用户不可写
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// GetDir 获取文件所在目录
// path: 文件路径
func GetDir(path string) string {
	return filepath.Dir(path)
}

// isSplitFile 优化分卷判断
func IsSplitFile(path string) bool {
	ext := filepath.Ext(path)
	// 匹配 .001/.002 等分卷后缀
	if len(ext) == 4 && ext[0] == '.' {
		for _, c := range ext[1:] {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}
	// 匹配分卷说明文件
	if ext == ".split" {
		return true
	}
	return false
}

// MergeSplitFiles 优化分卷合并
func MergeSplitFiles(firstSplitPath string, outputPath string) error {
	// 解析分卷基础名
	var base string
	if IsSplitFile(firstSplitPath) {
		ext := filepath.Ext(firstSplitPath)
		if len(ext) == 4 && ext[0] == '.' {
			base = firstSplitPath[:len(firstSplitPath)-4]
		} else if ext == ".split" {
			base = firstSplitPath[:len(firstSplitPath)-6] // 去掉 .split
		}
	} else {
		base = firstSplitPath
	}

	// 查找所有分卷
	var splitPaths []string
	for i := 1; ; i++ {
		splitPath := fmt.Sprintf("%s.%03d", base, i)
		if !CheckPathExist(splitPath) {
			break
		}
		splitPaths = append(splitPaths, splitPath)
	}
	if len(splitPaths) == 0 {
		return fmt.Errorf("未找到分卷文件: %s", base)
	}

	if utils.VerboseMode() {
		log.Info("合并", len(splitPaths), "个分卷:", splitPaths)
	}

	// 合并分卷
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建合并文件失败: %v", err)
	}
	defer outFile.Close()

	// 批量进度条
	batchBar := progress.NewBatchProgressBar(len(splitPaths))
	defer progress.FinishProgress(batchBar)

	buf := make([]byte, 4*1024*1024) // 4MB 缓冲区
	for _, splitPath := range splitPaths {
		progress.UpdateProgress(batchBar, 1)

		inFile, err := os.Open(splitPath)
		if err != nil {
			return fmt.Errorf("打开分卷 %s 失败: %v", splitPath, err)
		}
		defer inFile.Close()

		_, err = io.CopyBuffer(outFile, inFile, buf)
		if err != nil {
			return fmt.Errorf("拷贝分卷 %s 失败: %v", splitPath, err)
		}
	}

	return nil
}
