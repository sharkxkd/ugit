// prettier-ignore
/*******************************************************************************
  * FILENAME    : base.go
  * Date        : 2025/12/09 11:05:28
  * Author      : zc
  * Version     : v1.0.0
  * Decription  : ugit的上层部分实现，主要负责objects数据库
********************************************************************************/
package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// prettier-ignore
/*******************************************************************************
  @Function name    : writeTree
  @Description      : 输出当前目录结构下的所有子文件的路径
  @Params           : 当前目录
  @Return           : 异常
********************************************************************************/
func writeTree(directory string) error {
	if directory == "" {
		directory = "."
	}
	var files []os.DirEntry
	var err error
	if files, err = os.ReadDir(directory); err != nil {
		return err
	}
	for _, file := range files {
		path := filepath.Join(directory, file.Name())
		if file.IsDir() {
			if err := writeTree(path); err != nil {
				return err
			}
		} else {
			fmt.Println(path)
		}
	}
	return nil
}
