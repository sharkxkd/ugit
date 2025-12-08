package utils

import (
	"flag"
	"fmt"
	"log"
	"os"
)

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
	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])
		runInit()
	case "hash-object":
		hashObjectCmd.Parse(os.Args[2:])
		remainingArgs := hashObjectCmd.Args()
		if len(remainingArgs) < 1 {
			fmt.Println()
			os.Exit(1)
		}
		runHashObject(remainingArgs)
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
  @Return           : 打印文件的SHA-1值
1. Get the path of the file to store.
2. Read the file.
3. Hash the content of the file using SHA-1.
4. Store the file under ".ugit/objects/{the SHA-1 hash}".
********************************************************************************/
func runHashObject(args []string) {
	file, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	oid, err := DoHashObject(*file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(oid)
}
