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
	"io/fs"
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
	parents []string
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
	return strings.Contains(path, ".ugit") || strings.Contains(path, ".git") || filepath.Base(path) == "ugit"
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
	if mpoid := getRef(MERGE_HEAD, true).value; mpoid != "" {
		commitMessage += fmt.Sprintf("parent %s\n", mpoid)
		deleteRefs(MERGE_HEAD, false)
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
			commit.parents = append(commit.parents, values[1])
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
func checkout(name string) error {
	oid := getOid(name)
	commit, err := getCommit(oid)
	if err != nil {
		return err
	}
	readTree(commit.tree)
	var head RefValue
	if isBranch(name) {
		head = RefValue{symbolic: true, value: filepath.Join("refs", "heads", name)}
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
	ref := filepath.Join("refs", "tags", name)
	updateRef(ref, RefValue{symbolic: false, value: oid}, true)
}

// prettier-ignore
/*******************************************************************************
  @Function name    : getOid
  @Description      : 遍历寻找对应的tag、head、branch或是oid，有名字返回名字，无名字返回oid
  @Params           :
  @Return           :
********************************************************************************/
func getOid(name string) string {
	if name == "@" {
		name = HEAD
	}
	refsToTry := []string{
		filepath.Join(name),
		filepath.Join("refs", name),
		filepath.Join("refs", "tags", name),
		filepath.Join("refs", "heads", name),
	}
	for _, ref := range refsToTry {
		if res := getRef(ref, false).value; res != "" {
			return getRef(ref, true).value
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
			if len(commit.parents) > 0 {
				stack = append(stack, commit.parents[:1]...)
				stack = append(commit.parents[1:], stack...)
			}
		}
	}()
	return ch
}

// prettier-ignore
/*******************************************************************************
  @Function name    : createBranch
  @Description      : 创建分支，创建引用指向oid
  @Params           :
	-oid			: commit对应的oid
	-name			: branch对应的name
  @Return           :
********************************************************************************/
func createBranch(oid string, name string) {
	updateRef(filepath.Join("refs", "heads", name), RefValue{symbolic: false, value: oid}, true)
}

// prettier-ignore
/*******************************************************************************
  @Function name    : isBranch
  @Description      : 判断是不是分支名称，判断是否存在
  @Params           :
	-branch			: 分支名称
  @Return           :
********************************************************************************/
func isBranch(branch string) bool {
	return getRef(filepath.Join("refs", "heads", branch), true).value != ""
}

// prettier-ignore
/*******************************************************************************
  @Function name    : baseInit
  @Description      : init的初始调用，创建仓库并且将head指向master
  @Params           :
  @Return           :
********************************************************************************/
func baseInit() {
	DoInit()
	// 默认创建master分支
	updateRef(HEAD, RefValue{symbolic: true, value: filepath.Join("refs", "heads", "master")}, true)
}

// prettier-ignore
/*******************************************************************************
  @Function name    : getBranchName
  @Description      : 获取当前分支名称
  @Params           :
  @Return           :
********************************************************************************/
func getBranchName() string {
	head := getRef(HEAD, false)
	if !head.symbolic {
		return ""
	}
	if !strings.HasPrefix(head.value, filepath.Join("refs", "heads")) {
		return ""
	}
	branch, err := filepath.Rel(filepath.Join("refs", "heads"), head.value)
	if err != nil {
		return ""
	}
	return branch
}

// prettier-ignore
/*******************************************************************************
  @Function name    : iterBranchName
  @Description      : 遍历分支名称
  @Params           :
  @Return           :
********************************************************************************/
func iterBranchName() <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for refmap := range iterRefs(filepath.Join("refs", "heads"), true) {
			if !strings.HasPrefix(refmap.refname, filepath.Join("refs", "heads")) {
				continue
			}
			branch, err := filepath.Rel(filepath.Join("refs", "heads"), refmap.refname)
			if err != nil {
				continue
			}
			ch <- branch
		}
	}()
	return ch
}

// prettier-ignore
/*******************************************************************************
  @Function name    : reset
  @Description      : 更新引用到具体的commitId
  @Params           :
  @Return           :
********************************************************************************/
func reset(commitId string) {
	updateRef(HEAD, RefValue{symbolic: false, value: commitId}, true)
}

// prettier-ignore
/*******************************************************************************
  @Function name    : printCommit
  @Description      :
  @Params           :
  @Return           :
********************************************************************************/
func printCommit(oid string, commit Commit, refsstring string) {
	fmt.Printf("commit %s %s\n", oid, refsstring)
	commit.message = "    " + commit.message
	fmt.Println(strings.ReplaceAll(commit.message, "\n", "\n    "))
}

// prettier-ignore
/*******************************************************************************
  @Function name    : getWorkingTree
  @Description      : 获取当前工作目录的Tree
  @Params           :
  @Return           :
********************************************************************************/
func getWorkingTree() map[string]string {
	result := make(map[string]string)
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !isIngnored(path) && !d.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			oid, err := DoHashObject(content, "blob")
			if err != nil {
				return err
			}
			rel, err := filepath.Rel("", path)
			if err != nil {
				return err
			}
			result[rel] = oid
		}
		return nil
	})
	if err != nil {
		return map[string]string{}
	}
	return result
}

func merge(other string) {
	// 1. 获取当前分支头节点对应的oid
	headOid := getRef(HEAD, true).value
	mergeBase := getMergeBase(headOid, other)
	cOther, err := getCommit(other)
	if err != nil {
		fmt.Printf("Error with merging branches %s\n", other)
		os.Exit(1)
	}
	if mergeBase == headOid {
		readTree(cOther.tree)
		updateRef(HEAD, RefValue{false, other}, true)
		fmt.Println("Fast-forward merge, no need to commit")
		return
	}
	cHead, err := getCommit(headOid)
	if err != nil {
		fmt.Printf("Error with merging branches %s\n", other)
		os.Exit(1)
	}
	cBase, err := getCommit(mergeBase)
	if err != nil {
		fmt.Printf("Error with merging branches %s\n", cBase)
		os.Exit(1)
	}
	updateRef(MERGE_HEAD, RefValue{false, other}, true)
	// 2. 将两个分支的内容合并并读取到工作区
	readTreeMerged(cBase.tree, cHead.tree, cOther.tree)
	fmt.Println("Merged in working tree\nPlease commit")
}

func readTreeMerged(tBase string, tHead string, tOther string) {
	// 1. 清除当前工作区的所有内容
	emptyCurrentDirectory(".")
	// 2. 获取当前两个分支对应的所有子文件夹的映射情况并且进行合并
	for path, content := range mergeTrees(getTree(tBase, "."), getTree(tHead, "."), getTree(tOther, ".")) {
		if _, err := common.PathExists(filepath.Dir(path)); err != nil {
			fmt.Printf("Error with creating dir on %s\n", path)
			os.Exit(1)
		}
		if err := os.WriteFile(path, content, 0644); err != nil {
			fmt.Printf("Error with writing file on %s\n", path)
			os.Exit(1)
		}
	}
}

// prettier-ignore
/*******************************************************************************
  @Function name    : getMergeBase
  @Description      : 获取两个提交的第一个公共父节点
  @Params           :
  @Return           :
********************************************************************************/
func getMergeBase(oid1 string, oid2 string) string {
	parents1 := iterCommitsAndParents([]string{oid1})
	set := make(map[string]bool)
	for parent := range parents1 {
		set[parent] = true
	}
	for parent := range iterCommitsAndParents([]string{oid2}) {
		if _, ok := set[parent]; ok {
			return parent
		}
	}
	fmt.Printf("Cann't find a common parent with %s and %s\n", oid1, oid2)
	return ""
}

func iterObjectsAndCommits(oids ...string) <-chan string {

	ch := make(chan string)
	go func() {
		defer close(ch)
		visited := make(map[string]bool)

		var iterObjectsInTree func(oid string)
		iterObjectsInTree = func(oid string) {
			visited[oid] = true
			ch <- oid
			for entry := range iterTreeEntries(oid) {
				if !visited[entry.oid] {
					if entry.fType == "tree" {
						iterObjectsInTree(entry.oid)
					} else {
						visited[entry.oid] = true
						ch <- entry.oid
					}
				}
			}
		}

		for oid := range iterCommitsAndParents(oids) {
			ch <- oid
			commit, err := getCommit(oid)
			if err != nil {
				fmt.Printf("【iterObjectsAndCommits】error with getCommit of %s\n", oid)
				os.Exit(1)
			}
			if !visited[commit.tree] {
				iterObjectsInTree(commit.tree)
			}
		}
	}()
	return ch
}

func isAncestorOf(commit string, maybeAncestor string) bool {
	for oid := range iterCommitsAndParents([]string{commit}) {
		if maybeAncestor == oid {
			return true
		}
	}
	return false
}
