package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
)

const GIT_DIR = ".ugit"
const OBJECTS = "objects"

// prettier-ignore
/*
*******************************************************************************
   @Function name    : DoInit
   @Description      : init的具体实现，初始化ugit及其子目录
   @Params           :
   @Return           :
*******************************************************************************
*/
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
/*
*******************************************************************************

  @Function name    : DoHashObject
  @Description      : 计算文件的SHA-1值，并写入到对应的数据库文件
  @Params           :
	-*file			: 文件指针
  @Return           :
	-oid			: SHA-1文件名，对应数据库的键
	-error			: 异常
********************************************************************************
*/
func DoHashObject(file os.File) (string, error) {
	hash := sha1.New()
	_, err := io.Copy(hash, &file)
	if err != nil {
		return "", err
	}
	// 重置文件指针到开头
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", err
	}
	oid := hex.EncodeToString(hash.Sum(nil))
	newFile, err := os.Create(filepath.Join(GIT_DIR, OBJECTS, oid))
	if err != nil {
		return "", err
	}
	defer newFile.Close()
	_, err = io.Copy(newFile, &file)
	if err != nil {
		return "", err
	}
	return oid, nil
}
