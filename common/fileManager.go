// prettier-ignore
/*******************************************************************************
  * FILENAME    : fileManager.go
  * Date        : 2025/12/10 23:17:16
  * Author      : zc
  * Version     : v1.0.0
  * Decription  : 公共工具方法封装
********************************************************************************/
package common

import (
	"fmt"
	"os"
)

// prettier-ignore
/*******************************************************************************
  @Function name    : PathExists
  @Description      : 判断路径是否存在，不存在则创建
  @Params           : 文件路径
  @Return           :
********************************************************************************/
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // 已存在
	}
	if os.IsNotExist(err) {
		// 创建多级目录
		if mkErr := os.MkdirAll(path, os.ModePerm); mkErr != nil {
			return false, mkErr
		}
		return true, nil
	}
	return false, err // 其他错误
}

// prettier-ignore
/*******************************************************************************
  @Function name    : IsSHA1
  @Description      : IsSHA1 判断字符串是否为合法的 SHA1 哈希
  @Params           : 字符串s
  @Return           : 是否为SHA1
********************************************************************************/
// IsSHA1 判断字符串是否为合法的 SHA1 哈希
func IsSHA1(s string) bool {
	// 1. 长度必须是 40
	if len(s) != 40 {
		return false
	}

	// 2. 遍历检查每个字符是否为十六进制 (0-9, a-f, A-F)
	for i := 0; i < len(s); i++ {
		c := s[i]
		isHex := (c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F')

		if !isHex {
			return false
		}
	}

	return true
}

func OenpTmp(content []byte) (*os.File, error) {
	f, err := os.CreateTemp("", "diff-blob-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}
	// 强制刷盘，确保外部命令 diff 能读到完整内容
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}
	
	return f, nil
}
