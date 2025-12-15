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

type Commit struct {
	tree    string
	parent  string
	message string
}

// prettier-ignore
/*******************************************************************************
  @Function name    : writeTree
  @Description      : 输出当前目录结构下的所有子文件的路径
  @Params           : 当前目录
  @Return           : 异常
  ===================tree Object===================
  tree
  Name			OID			Type
  文件名		对应的键	 类型
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
		if isIngnored(path) {
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
func isIngnored(path string) bool {
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
  @Description      : 获取tree文件内所有的文件和目录，建立path到name的映射关系
  @Params           :
	-oid			: 对应的tree的OID
	-basePath		: 基础路径
  @Return           : 映射关系
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
  @Function name    : readTree
  @Description      : 根据一个oid，将其对应的所有文件复制到暂存区
  @Params           : tree_oid
  @Return           : 错误
********************************************************************************/
func readTree(treeOid string) error {
	if err := emptyCurrentDirectory("./test"); err != nil {
		return err
	}
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

// prettier-ignore
/*******************************************************************************
  @Function name    : emptyCurrentDirectory
  @Description      : 清空暂存区的所有文件
  @Params           : 输入默认基础路径
  @Return           : 错误
********************************************************************************/
func emptyCurrentDirectory(basePath string) error {
	if basePath == "" {
		basePath = "./test"
	}
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(basePath, entry.Name())
		if isIngnored(path) {
			continue
		}
		if entry.IsDir() {
			if err := emptyCurrentDirectory(path); err != nil {
				return err
			}
		}
		os.Remove(path)
	}
	return err
}

// prettier-ignore
/*******************************************************************************
  @Function name    : commit
  @Description      : commit底层实现，写对应commit文本，并提交到objects数据库，为"commit"类型
  @Params           : 提交的消息
  @Return           :
	-oid			: commit对象OID
	-err			: error
********************************************************************************/
func commit(message string) (string, error) {
	oid, err := writeTree("")
	if err != nil {
		return "", err
	}
	commitMessage := fmt.Sprintf("tree %s\n", oid)
	if poid := getRef(HEAD, true).value; poid != "" {
		commitMessage += fmt.Sprintf("parent %s\n", poid)
	}
	commitMessage += fmt.Sprintln()
	commitMessage += fmt.Sprintln(message)
	oid, err = DoHashObject([]byte(commitMessage), "commit")
	if err != nil {
		return "", err
	}
	updateRef(HEAD, RefValue{symbolic: false, value: oid}, true)
	return oid, nil
}

// prettier-ignore
/*******************************************************************************
  @Function name    : getCommit
  @Description      : 获取提交ID对应的提交信息
  @Params           :
  @Return           :
********************************************************************************/
func getCommit(oid string) (Commit, error) {
	byteOfContent, err := DoRunCatFile(oid, "commit")
	if err != nil {
		return Commit{}, err
	}
	content := string(byteOfContent)
	var commit Commit
	var contents = strings.Split(content, "\n")
	var startIndex = len(contents)
	for i, line := range contents {
		if line == "" {
			startIndex = i + 1
			break
		}
		values := strings.Split(line, " ")
		if values[0] == "tree" {
			commit.tree = values[1]
		} else if values[0] == "parent" {
			commit.parent = values[1]
		} else {
			fmt.Printf("Unknown Field %s\n", values[0])
		}
	}
	if startIndex < len(contents) {
		commit.message = strings.Join(contents[startIndex:], "")
	}
	return commit, nil
}

// prettier-ignore
/*******************************************************************************
  @Function name    : checkout
  @Description      : 所携带的oid都需要是commit类型的oid，只能在提交之间切换
  @Params           : commit oid
  @Return           : error
********************************************************************************/
func checkout(oid string) error {
	commit, err := getCommit(oid)
	if err != nil {
		return err
	}
	readTree(commit.tree)
	var head RefValue
	if isBranch(oid) {
		head = RefValue{symbolic: true, value: fmt.Sprintf("refs/heads/%s", oid)}
	} else {
		head = RefValue{symbolic: false, value: oid}
	}
	updateRef(HEAD, head, false)
	return nil
}

// prettier-ignore
/*******************************************************************************
  @Function name    : createTag
  @Description      : 创建tag映射oid
  @Params           :
  @Return           :
********************************************************************************/
func createTag(name string, oid string) {
	ref := fmt.Sprintf("refs/tags/%s", name)
	updateRef(ref, RefValue{symbolic: false, value: oid}, true)
}

// prettier-ignore
/*******************************************************************************
  @Function name    : getOid
  @Description      : 遍历寻找对应的tag、head、branch或是oid
  @Params           :
  @Return           :
********************************************************************************/
func getOid(name string) string {
	if name == "@" {
		name = HEAD
	}
	refsToTry := []string{
		fmt.Sprint(name),
		fmt.Sprintf("refs/%s", name),
		fmt.Sprintf("refs/tags/%s", name),
		fmt.Sprintf("refs/heads/%s", name),
	}
	for _, ref := range refsToTry {
		if res := getRef(ref, true).value; res != "" {
			return res
		}
	}
	if common.IsSHA1(name) {
		return name
	}
	fmt.Printf("Either a tag name or a oid of %s\n", name)
	os.Exit(1)
	return ""
}

// prettier-ignore
/*******************************************************************************
  @Function name    : iterCommitsAndParents
  @Description      : 根据oids寻找其提交信息及其父提交
  @Params           :
  @Return           :
********************************************************************************/
func iterCommitsAndParents(oids []string) <-chan string {
	ch := make(chan string)
	stack := append([]string{}, oids...)
	visited := make(map[string]struct{})
	go func() {
		defer close(ch)
		for len(stack) > 0 {
			index := len(stack) - 1
			oid := stack[index]
			stack = stack[:index]

			if _, exists := visited[oid]; exists || oid == "" {
				continue
			}
			visited[oid] = struct{}{}
			ch <- oid
			commit, err := getCommit(oid)
			if err != nil {
				fmt.Println("error with itering commits and parents")
			}
			stack = append(stack, commit.parent)
		}
	}()
	return ch
}

// prettier-ignore
/*******************************************************************************
  @Function name    : createBranch
  @Description      : 创建分支，创建引用指向oid
  @Params           :
  @Return           :
********************************************************************************/
func createBranch(oid string, name string) {
	updateRef(fmt.Sprintf("refs/heads/%s", name), RefValue{symbolic: false, value: oid}, true)
}

func isBranch(oid string) bool {
	return getRef(oid, true).value != ""
}

func baseInit() {

}
