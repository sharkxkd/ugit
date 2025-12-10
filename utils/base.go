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
	"strings"
)
type dirType struct{
		name string
		oid string
		fType string
}

// prettier-ignore
/*******************************************************************************
  @Function name    : writeTree
  @Description      : 输出当前目录结构下的所有子文件的路径
  @Params           : 当前目录
  @Return           : 异常
********************************************************************************/
func writeTree(directory string) (string, error) {
	if directory == "" {
		directory = "."
	}
	var files []os.DirEntry
	var err error
	if files, err = os.ReadDir(directory); err != nil {
		return "", err
	}
	var dirs []dirType
	for _, file := range files {
		path := filepath.Join(directory, file.Name())
		// 忽略自己的子目录
		if isUgit(path) {
			continue
		}
		if file.IsDir() {
			oid, err := writeTree(path)
			if err != nil {
				return "",err
			}
			dirs = append(dirs, dirType{file.Name(), oid, "tree"})
		} else {
			content, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			foid, err := DoHashObject(content, "blob")
			if err != nil {
				return "",err;
			}
			fmt.Printf("%s: %s\n",foid, path)
		}
	}
	var content string
	for _, d := range dirs {
		content += fmt.Sprintf("%s %s %s\n",d.name,d.oid,d.fType)
	}
	oid, err := DoHashObject([]byte(content), "tree")
	if err != nil {
		return "", err
	}
	return oid ,nil
}

// prettier-ignore
/*******************************************************************************
  @Function name    : isUgit
  @Description      : 判断是否存在.ugit目录
  @Params           : 完整路径
  @Return           : 返回是否存在.ugit
********************************************************************************/
func isUgit(path string) bool {
	return strings.Contains(path, ".ugit") || strings.Contains(path, ".git")
}
