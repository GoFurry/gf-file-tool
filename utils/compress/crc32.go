package compress

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/klauspost/crc32"
)

// CalculateCRC32 计算文件 CRC32 值
func CalculateCRC32(filePath string) (string, error) {
	if !CheckPathExist(filePath) {
		return "", fmt.Errorf("文件不存在: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 初始化 CRC32 计算器
	hash := crc32.NewIEEE()
	reader := bufio.NewReader(file)
	buf := make([]byte, 4096)

	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("读取文件失败: %v", err)
		}
		if n == 0 {
			break
		}
		hash.Write(buf[:n])
	}

	// 返回十六进制字符串
	return fmt.Sprintf("%08x", hash.Sum32()), nil
}

// VerifyFileCRC32 校验文件 CRC32 是否匹配
func VerifyFileCRC32(filePath, expectedCRC string) (bool, error) {
	actualCRC, err := CalculateCRC32(filePath)
	if err != nil {
		return false, err
	}
	return actualCRC == expectedCRC, nil
}
