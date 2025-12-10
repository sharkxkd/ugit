// prettier-ignore
/*******************************************************************************
  * FILENAME    : cli.go
  * Date        : 2025/12/09 11:05:39
  * Author      : zc
  * Version     : v1.0.0
  * Decription  : 实现命令的解析与分发
********************************************************************************/
package utils

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var type_ string

// prettier-ignore
/*******************************************************************************
  @Function name    : parseArgs
  @Description      : 定义子命令，并解析参数，命令的入口
  @Params           :
  @Return           :
********************************************************************************/
func ParseArgs() {
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	hashObjectCmd := flag.NewFlagSet("hash-object", flag.ExitOnError)
	hashObjectCmd.StringVar(&type_, "type", "blob", "文件类型")
	catFileCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)
	writeTreeCmd := flag.NewFlagSet("write-tree", flag.ExitOnError)
	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])
		runInit()
	case "hash-object":
		hashObjectCmd.Parse(os.Args[2:])
		remainingArgs := hashObjectCmd.Args()
		runHashObject(remainingArgs)
	case "cat-file":
		catFileCmd.Parse(os.Args[2:])
		remainingArgs := catFileCmd.Args()
		runCatFile(remainingArgs)
	case "write-tree":
		writeTreeCmd.Parse(os.Args[2:])
		remainingArgs := writeTreeCmd.Args()
		runWriteTree(remainingArgs)
	default:
		os.Exit(1)
	}

}

// prettier-ignore
/*******************************************************************************
  @Function name    : init
  @Description      : 初始化.Ggit目录
  @Params           :
  @Return           :
********************************************************************************/
func runInit() {
	DoInit()
	// 获取当前路径并处理，输出成功初始化目录
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Initialized empty Ggit repository in %s/%s\n", wd, GIT_DIR)
}

// prettier-ignore
/*******************************************************************************
  @Function name    : hash-object
  @Description      : 实现计算一个文件的SHA-1并且存储内容到object目录
  @Params           :
	-fileName		: 给定文件名字，目标文件
  @Return           : 打印文件的SHA-1值
1. Get the path of the file to store.
2. Read the file.
3. Hash the content of the file using SHA-1.
4. Store the file under ".ugit/objects/{the SHA-1 hash}".
********************************************************************************/
func runHashObject(args []string) {
	if len(args) < 1 {
		fmt.Println()
		os.Exit(1)
	}
	content, err := os.ReadFile(args[0])
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	oid, err := DoHashObject(content, type_)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(oid)
}

// prettier-ignore
/*******************************************************************************
  @Function name    : cat-file
  @Description      : 输入给定文件名，查看其信息
  @Params           :
  	-fileName		: 给定文件名字，目标文件
  @Return           : 控制台输出文件内容
********************************************************************************/
func runCatFile(args []string) {
	if len(args) < 1 {
		fmt.Println()
		os.Exit(1)
	}
	content, err := DoRunCatFile(args[0], "blob")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Println(string(content))
}

// prettier-ignore
/*******************************************************************************
  @Function name    : write-tree
  @Description      : 输出指定目录下的所有子文件
  @Params           : 指定的目录，默认当前命令的目录
  @Return           : 控制台打印输出的文件路径
********************************************************************************/
func runWriteTree(args []string) {
	oid, err := writeTree("")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Println(oid)
}
