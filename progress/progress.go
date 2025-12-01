// Package progress /progress/progress.go
package progress

import (
	"fmt"
	"os"
	"time"

	"github.com/GoFurry/gf-file-tool/utils"
	"github.com/schollz/progressbar/v3"
)

// NewFileProgressBar 创建单个文件的进度条
// totalSize: 文件总字节大小
// fileName: 文件名
// return: 进度条实例
func NewFileProgressBar(totalSize int64, fileName string) *progressbar.ProgressBar {
	// 静默模式不显示进度条
	if utils.QuietMode() {
		return nil
	}

	// 进度条配置
	bar := progressbar.NewOptions64(
		totalSize,
		progressbar.OptionSetWriter(os.Stdout),   // 输出到标准输出
		progressbar.OptionEnableColorCodes(true), // 启用颜色
		progressbar.OptionShowBytes(true),        // 显示字节数
		progressbar.OptionSetWidth(50),           // 进度条宽度
		progressbar.OptionSetDescription( // 左侧描述
			fmt.Sprintf("处理文件: [red]%s [reset]", fileName),
		),
		progressbar.OptionSetTheme(progressbar.Theme{ // 进度条样式
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowElapsedTimeOnFinish(), // 完成后显示耗时
	)

	return bar
}

// NewBatchProgressBar 创建批量文件的进度条
// totalCount: 文件总数
// return: 进度条实例
func NewBatchProgressBar(totalCount int) *progressbar.ProgressBar {
	if utils.QuietMode() {
		return nil
	}

	bar := progressbar.NewOptions(
		totalCount,
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetDescription("[cyan]批量处理中[reset]"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowCount(),                    // 显示已处理/总数
		progressbar.OptionThrottle(100*time.Millisecond), // 限制更新频率
	)

	return bar
}

// UpdateProgress 更新进度条
// bar: 进度条实例
// n: 要更新的进度值
func UpdateProgress(bar *progressbar.ProgressBar, n int) {
	if bar != nil {
		_ = bar.Add(n)
	}
}

// FinishProgress 结束进度条
func FinishProgress(bar *progressbar.ProgressBar) {
	if bar != nil {
		_ = bar.Finish()
		fmt.Println()
	}
}
