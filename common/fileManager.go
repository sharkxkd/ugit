// prettier-ignore
/*******************************************************************************
  * FILENAME    : fileManager.go
  * Date        : 2025/12/10 23:17:16
  * Author      : zc
  * Version     : v1.0.0
  * Decription  : 公共工具方法封装
********************************************************************************/
package common

import "os"

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
