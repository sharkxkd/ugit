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
	"os/exec"
	"strings"
)

var type_ string
var message string
var tagName string

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
	catFileCmd.StringVar(&type_, "t", "blob", "文件类型")
	catFileCmd.StringVar(&type_, "type", "blob", "文件类型")

	writeTreeCmd := flag.NewFlagSet("write-tree", flag.ExitOnError)

	readTreeCmd := flag.NewFlagSet("read-tree", flag.ExitOnError)

	commitCmd := flag.NewFlagSet("commit", flag.ExitOnError)
	commitCmd.StringVar(&message, "m", "Default Message Empty", "提交消息")
	commitCmd.StringVar(&message, "message", "Default Message Empty", "提交消息")

	logCmd := flag.NewFlagSet("log", flag.ExitOnError)

	checkoutCmd := flag.NewFlagSet("checkout", flag.ExitOnError)

	tagCmd := flag.NewFlagSet("tag", flag.ExitOnError)
	tagCmd.StringVar(&tagName, "name", "Default_Name", "标签名称")

	kCmd := flag.NewFlagSet("k", flag.ExitOnError)
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
	case "read-tree":
		readTreeCmd.Parse(os.Args[2:])
		remainingArgs := readTreeCmd.Args()
		runReadTree(remainingArgs)
	case "commit":
		commitCmd.Parse(os.Args[2:])
		remainingArgs := commitCmd.Args()
		runCommit(remainingArgs)
	case "log":
		logCmd.Parse(os.Args[2:])
		remainingArgs := logCmd.Args()
		runLog(remainingArgs)
	case "checkout":
		checkoutCmd.Parse(os.Args[2:])
		remainingArgs := checkoutCmd.Args()
		runCheckout(remainingArgs)
	case "tag":
		tagCmd.Parse(os.Args[2:])
		remainingArgs := tagCmd.Args()
		runTag(remainingArgs)
	case "k":
		kCmd.Parse(os.Args[2:])
		remainingArgs := kCmd.Args()
		runK(remainingArgs)
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
	treeOid := getOid(args[0])
	content, err := DoRunCatFile(treeOid, type_)
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

// prettier-ignore
/*******************************************************************************
  @Function name    : read-tree
  @Description      : 将指定的OID的tree读入到工作目录
  @Params           : 命令中输入对应的oid
  @Return           :
********************************************************************************/
func runReadTree(args []string) {
	if len(args) < 1 {
		fmt.Println()
		os.Exit(1)
	}
	args[0] = getOid(args[0])
	if err := readTree(args[0]); err != nil {
		log.Fatal(err)
		fmt.Printf("fatal read tree %s", args[0])
	}

}

// prettier-ignore
/*******************************************************************************
  @Function name    : commit
  @Description      : 提交命令，创建一个提交对象
  @Params           :
  @Return           :
********************************************************************************/
func runCommit(args []string) {
	content, err := commit(message)
	if err != nil {
		fmt.Printf("fatal commit with error %s", err.Error())
		os.Exit(1)
	}
	fmt.Printf("%s", content)
}

// prettier-ignore
/*******************************************************************************
  @Function name    : funtion
  @Description      :
  @Params           :
  @Return           :
********************************************************************************/
func runLog(args []string) {
	if len(args) < 1 {
		args = append(args, "@")
	}
	oid := getOid(args[0])
	for oid != "" {
		commit, err := getCommit(oid)
		if err != nil {
			fmt.Printf("fatal print log with error %s", err.Error())
		}

		fmt.Printf("commit %s\n", oid)
		commit.message = "    " + commit.message
		fmt.Println(strings.ReplaceAll(commit.message, "\n", "\n    "))

		oid = commit.parent
	}
}

// prettier-ignore
/*******************************************************************************
  @Function name    : function
  @Description      :
  @Params           :
  @Return           :
********************************************************************************/
func runCheckout(args []string) {
	if len(args) < 1 {
		fmt.Println("error checkout without an oid")
		os.Exit(1)
	}
	args[0] = getOid(args[0])
	if err := checkout(args[0]); err != nil {
		fmt.Printf("Error happened while checkout with %s, description %s", args[0], err)
	}
}

// prettier-ignore
/*******************************************************************************
  @Function name    : tag
  @Description      : 为oid取名字
  @Params           :
  @Return           :
********************************************************************************/
func runTag(args []string) {
	if len(args) < 1 {
		args = append(args, "@")
	}
	oid := getOid(args[0])
	createTag(tagName, oid)
}

// prettier-ignore
/*******************************************************************************
  @Function name    : function
  @Description      :
  @Params           :
  @Return           :
********************************************************************************/
func runK(args []string) {
	dot := "digraph commits {\n"

	oids := []string{}
	for ref := range iterRefs() {
		dot += fmt.Sprintf("\"%s\" [shape=note]\n", ref.refname)
		dot += fmt.Sprintf("\"%s\" -> \"%s\"", ref.refname, ref.oid)
		oids = append(oids, ref.oid)
	}
	for oid := range iterCommitsAndParents(oids) {
		commit, err := getCommit(oid)
		if err != nil {
			fmt.Println("error with drawing graph")
			os.Exit(1)
		}
		dot += fmt.Sprintf("\"%s\" [shape=box style=filled label=\"%s\"]", oid, oid[:10])
		if commit.parent != "" {
			dot += fmt.Sprintf("\"%s\" -> \"%s\"", oid, commit.parent)
		}
	}
	dot += "}"
	fmt.Print(dot)
	cmd := exec.Command("dot", "-Tx11", "/dev/stdin")
	cmd.Stdin = strings.NewReader(dot)
	// 建议把标准错误重定向出来，如果 dot 报错（比如没有安装 GTK 插件），你能看到
	cmd.Stderr = os.Stderr

	// 启动命令并等待结束
	// 这相当于 Python 的 with ... as proc 以及 communicate 的组合效果
	fmt.Println("Opening visualization window...")
	err := cmd.Run()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
