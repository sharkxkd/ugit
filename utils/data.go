// prettier-ignore
/*******************************************************************************
  * FILENAME    : data.go
  * Date        : 2025/12/09 11:06:20
  * Author      : zc
  * Version     : v1.0.0
  * Decription  : 实现ugit的底层命令
********************************************************************************/
package utils

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"ugit/common"
)

const GIT_DIR = ".ugit"
const OBJECTS = "objects"
const HEAD = "HEAD"
const REFS = ".ugit/refs"

type refMap struct {
	refname string
	oid     string
}

// prettier-ignore
/*******************************************************************************
   @Function name    : DoInit
   @Description      : init的具体实现，初始化ugit及其子目录
   @Params           :
   @Return           :
*******************************************************************************/
func DoInit() {
	err := os.Mkdir(GIT_DIR, 0755)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Mkdir(filepath.Join(GIT_DIR, OBJECTS), 0755)
	if err != nil {
		log.Fatal(err)
	}
}

// prettier-ignore
/*******************************************************************************
  @Function name    : DoHashObject
  @Description      : 根据文件内容计算文件的SHA-1值，并写入到对应的数据库文件
  @Params           :
	-content		: 文件内容
	-fileType		: 文件类型
  @Return           :
	-oid			: SHA-1文件名，对应数据库的键
	-error			: 异常
********************************************************************************/
func DoHashObject(content []byte, fileType string) (string, error) {
	obj := []byte(fileType)
	obj = append(obj, 0)
	obj = append(obj, content...)
	hash := sha1.Sum(obj)
	oid := hex.EncodeToString(hash[:])

	err := os.WriteFile(filepath.Join(GIT_DIR, OBJECTS, oid), obj, 0644)
	if err != nil {
		return "", err
	}
	return oid, nil
}

// prettier-ignore
/*******************************************************************************
  @Function name    : DoRunCatFile
  @Description      : 查看指定哈希值的文件内容信息
  @Params           :
	-oid			: 文件对应的键
	-expected		: 期望的类型, 默认 "blob"
  @Return           :
	-content		: 文件的二进制内容
	-error			: 中途遇到的一些异常
********************************************************************************/
func DoRunCatFile(oid string, expected string) ([]byte, error) {
	obj, err := os.ReadFile(filepath.Join(GIT_DIR, OBJECTS, oid))
	if err != nil {
		return []byte{}, err
	}
	fileType, content, found := bytes.Cut(obj, []byte{0})
	if !found {
		// 如果没找到分隔符，需要处理
		return []byte{}, err
	}
	if string(fileType) != expected {
		return []byte{}, fmt.Errorf("expected type %s but got %s", expected, string(fileType))
	}
	return content, nil
}

func updateRef(ref string, oid string) {
	path := filepath.Join(GIT_DIR, ref)
	if _, err := common.PathExists(filepath.Dir(path)); err != nil {
		log.Fatal(err)
		fmt.Print(err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, []byte(oid), 0644); err != nil {
		fmt.Printf("error with writing %s file\n", ref)
	}
}

// prettier-ignore
/*******************************************************************************
  @Function name    : getRef
  @Description      : 解析ref文件, 如果包含有ref: 需要递归解析
  @Params           :
  @Return           :
********************************************************************************/
func getRef(ref string) string {
	path := filepath.Join(GIT_DIR, ref)
	if fileInfo, err := os.Stat(path); err != nil || fileInfo.IsDir() {
		if os.IsNotExist(err) {
			return ""
		}
		fmt.Printf("error with opening %s file\n", ref)
		return ""
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		fmt.Printf("error with opening %s file\n", ref)
	}
	oid := string(content)
	oid = strings.TrimSpace(oid)
	if oid != "" && strings.Contains(oid, "ref:") {
		oid = strings.TrimSpace(strings.Split(oid, ":")[1])
		oid = getRef(oid)
	}
	return oid
}

// prettier-ignore
/*******************************************************************************
  @Function name    : function
  @Description      :
  @Params           :
  @Return           :
********************************************************************************/
func iterRefs() <-chan refMap {
	ch := make(chan refMap)
	go func() {
		defer close(ch)

		// 先处理 HEAD
		ch <- refMap{refname: HEAD, oid: getRef(HEAD)}

		// 遍历 refs 下的所有引用文件
		_ = filepath.WalkDir(REFS, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(GIT_DIR, path)
			if err != nil {
				rel = path
			}
			ch <- refMap{refname: rel, oid: getRef(rel)}
			return nil
		})
	}()
	return ch
}
