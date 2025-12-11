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
	"log"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"ugit/common"
)

type dirType struct {
	name  string
	oid   string
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
				return "", err
			}
			dirs = append(dirs, dirType{file.Name(), oid, "tree"})
		} else {
			content, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			foid, err := DoHashObject(content, "blob")
			if err != nil {
				return "", err
			}
			fmt.Printf("%s: %s\n", foid, path)
			dirs = append(dirs, dirType{file.Name(), foid, "blob"})
		}
	}
	// tree对应的内容
	var content string
	for _, d := range dirs {
		content += fmt.Sprintf("%s %s %s\n", d.name, d.oid, d.fType)
	}
	oid, err := DoHashObject([]byte(content), "tree")
	if err != nil {
		return "", err
	}
	return oid, nil
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

// prettier-ignore
/*******************************************************************************
  @Function name    : iterTreeEntries
  @Description      : 遍历格式为tree的文件，解析并通过通道返回数据
  @Params           : oid
  @Return           : 流式返回name, oid, type
********************************************************************************/
func iterTreeEntries(oid string) <-chan dirType {
	ch := make(chan dirType)
	content, err := DoRunCatFile(oid, "tree")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	go func() {
		defer close(ch)
		for _, dir := range strings.Split(string(content), "\n") {
			entry := strings.Split(dir, " ")
			if len(entry) != 3 {
				continue
			}
			ch <- dirType{entry[0], entry[1], entry[2]}
		}
	}()
	return ch
}

// prettier-ignore
/*******************************************************************************
  @Function name    : getTree
  @Description      :
  @Params           :
  @Return           :
********************************************************************************/
func getTree(oid string, basePath string) map[string]string {
	result := make(map[string]string)
	for entry := range iterTreeEntries(oid) {
		path := filepath.Join(basePath, entry.name)
		if entry.fType == "blob" {
			result[path] = entry.oid
		} else if entry.fType == "tree" {
			maps.Copy(result, getTree(entry.oid, path))
		} else {
			panic(fmt.Sprintf("不支持该类型%s", entry.fType))
		}
	}
	return result
}

// prettier-ignore
/*******************************************************************************
  @Function name    : fucntion
  @Description      :
  @Params           :
  @Return           :
********************************************************************************/
func readTree(treeOid string) error {
	for path, oid := range getTree(treeOid, "./test") {
		fmt.Printf("path: %s  oid: %s\n", path, oid)
		if _, err := common.PathExists(filepath.Dir(path)); err != nil {
			return err
		}
		content, err := DoRunCatFile(oid, "blob")
		if err != nil {
			fmt.Print("err")
			return err
		}
		if err := os.WriteFile(path, content, 0644); err != nil {
			fmt.Print("err")
			return err
		}
	}
	return nil
}
