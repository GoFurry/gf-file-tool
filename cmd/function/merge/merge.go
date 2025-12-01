// Package merge /cmd/function/merge/merge.go
package merge

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/GoFurry/gf-file-tool/cmd"
	"github.com/GoFurry/gf-file-tool/utils/compress"
	"github.com/GoFurry/gf-file-tool/utils/log"
	"github.com/spf13/cobra"
)

// mergeCmd 合并分卷压缩包命令实例
var mergeCmd = &cobra.Command{
	Use:   "merge [base-file]",
	Short: "合并分卷压缩包",
	Example: `gf-file-tool merge ./test/split_big.zip
gf-file-tool merge ./test/split_big.zip -o merged.zip -r`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		basePath := args[0]
		outputPath, _ := cmd.Flags().GetString("output")
		verify, _ := cmd.Flags().GetBool("verify")

		// 默认输出路径
		if outputPath == "" {
			outputPath = basePath + "_merged"
		}

		// 查找所有分卷文件
		var volumes []string
		dir := filepath.Dir(basePath)
		baseName := filepath.Base(basePath)

		files, _ := os.ReadDir(dir)
		for _, f := range files {
			if !f.IsDir() && strings.HasPrefix(f.Name(), baseName+".") {
				volumes = append(volumes, filepath.Join(dir, f.Name()))
			}
		}
		if len(volumes) == 0 {
			log.Error("未找到分卷文件")
			return
		}

		// 合并分卷
		log.Info("合并", len(volumes), "个分卷:", volumes)
		outFile, err := os.Create(outputPath)
		if err != nil {
			log.Error("创建合并文件失败:", err)
			return
		}
		defer outFile.Close()

		for _, vol := range volumes {
			inFile, err := os.Open(vol)
			if err != nil {
				log.Error("打开分卷", vol, "失败:", err)
				return
			}
			defer inFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				log.Error("合并分卷", vol, "失败", err)
				return
			}
		}

		// 验证 CRC32
		if verify {
			// 从分卷说明文件读取预期 CRC32
			manifestPath := basePath + ".split"
			if compress.CheckPathExist(manifestPath) {
				content, _ := os.ReadFile(manifestPath)
				lines := strings.Split(string(content), "\n")
				var expectedCRC string
				for _, line := range lines {
					if strings.HasPrefix(line, "CRC32:") {
						expectedCRC = strings.TrimSpace(strings.TrimPrefix(line, "CRC32:"))
						break
					}
				}

				if expectedCRC != "" {
					actualCRC, err := compress.CalculateCRC32(outputPath)
					if err != nil {
						log.Error("校验 CRC32 失败:", err)
						return
					}
					if actualCRC == expectedCRC {
						log.Success("CRC32 校验通过:", actualCRC)
					} else {
						log.Error("CRC32 校验失败: 预期", expectedCRC, ", 实际", actualCRC)
					}
				}
			} else {
				log.Warn("未找到分卷说明文件, 跳过 CRC32 校验")
			}
		}
		log.Success("合并完成, 输出文件:", outputPath)
	},
}

// InitMerge 初始化命令
func InitMerge() {
	cmd.GetRootCmd().AddCommand(mergeCmd)
	mergeCmd.Flags().StringP("output", "o", "", "合并后输出路径")
	mergeCmd.Flags().BoolP("verify", "r", false, "验证合并后文件的 CRC32")
}
