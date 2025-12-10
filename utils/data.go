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
	"log"
	"os"
	"path/filepath"
)

const GIT_DIR = ".ugit"
const OBJECTS = "objects"

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
