// prettier-ignore
/*******************************************************************************
  * FILENAME    : diff.go
  * Date        : 2025/12/20 13:44:19
  * Author      : zc
  * Version     : v1.0.0
  * Decription  : 对比前后版本文件的差异
********************************************************************************/
package utils

import (
	"fmt"
	"os"
	"os/exec"
	"ugit/common"
)

type TreeComparison struct {
	path string
	fOid string
	tOid string
}

type ChangedFileType struct {
	path   string
	action string
}

// prettier-ignore
/*******************************************************************************
  @Function name    : diffTrees
  @Description      : 比较两次提交的异同
  @Params           :
  @Return           :
********************************************************************************/
func diffTrees(fromTree map[string]string, toTree map[string]string) string {
	output := ""
	for treeComparsion := range compareTrees(fromTree, toTree) {
		output += diffBlobs(treeComparsion.fOid, treeComparsion.tOid, treeComparsion.path)
	}
	return output
}

// prettier-ignore
/*******************************************************************************
  @Function name    : compareTrees
  @Description      : 比较两个树的差异
  @Params           :
  @Return           :
********************************************************************************/
func compareTrees(fromTree map[string]string, toTree map[string]string) <-chan TreeComparison {
	ch := make(chan TreeComparison)
	go func() {

		// 比较异同
		for fPath, fOid := range fromTree {
			if toTree[fPath] == "" {
				ch <- TreeComparison{path: fPath, fOid: fOid}
			} else if fOid != toTree[fPath] {
				ch <- TreeComparison{path: fPath, fOid: fOid, tOid: toTree[fPath]}
			}
		}

		for tPath, tOid := range toTree {
			if fromTree[tPath] == "" {
				ch <- TreeComparison{path: tPath, tOid: tOid}
			}
		}
		close(ch)
	}()
	return ch
}

// prettier-ignore
/*******************************************************************************
  @Function name    : diffBlobs
  @Description      : 调用diff命令检查两个文件的异同
  @Params           :
  @Return           : 输出diff命令的输出
********************************************************************************/
func diffBlobs(fromOid string, toOid string, path string) string {
	if path == "" {
		path = "blob"
	}
	fContent, err := DoRunCatFile(fromOid, "blob")
	if err != nil {
		fmt.Printf("error with cat file %s\n", fromOid)
		return ""
	}
	from, err := common.OenpTmp(fContent)
	if err != nil {
		return ""
	}
	defer from.Close()
	defer os.Remove(from.Name())
	tContent, err := DoRunCatFile(toOid, "blob")
	if err != nil {
		fmt.Printf("error with cat file %s\n", toOid)
		return ""
	}

	to, err := common.OenpTmp(tContent)
	if err != nil {
		return ""
	}
	defer to.Close()
	defer os.Remove(to.Name())
	// 执行命令
	cmd := exec.Command("diff",
		"--unified",
		"--show-c-function",
		"--label", "a/"+path, from.Name(),
		"--label", "b/"+path, to.Name(),
	)
	output, err := cmd.CombinedOutput()
	// 处理 diff 的特定退出码
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return string(output) // 1 表示有差异，是正常结果
			}
		}
		fmt.Println("error:", err)
		return ""
	}
	return string(output)
}

// prettier-ignore
/*******************************************************************************
  @Function name    : iterchangedFiles
  @Description      : 迭代展示被修改的文件，包括删除/更新/新增
  @Params           :
  @Return           :
********************************************************************************/
func iterChangedFiles(fromTree map[string]string, toTree map[string]string) <-chan ChangedFileType {
	ch := make(chan ChangedFileType)
	go func() {
		for comparison := range compareTrees(fromTree, toTree) {
			if comparison.fOid == "" {
				ch <- ChangedFileType{path: comparison.path, action: "Created"}
			} else if comparison.tOid == "" {
				ch <- ChangedFileType{path: comparison.path, action: "Deleted"}
			} else {
				ch <- ChangedFileType{path: comparison.path, action: "Modified"}
			}
		}
		close(ch)

	}()
	return ch
}

// prettier-ignore
/*******************************************************************************
  @Function name    : mergeTrees
  @Description      : 合并所有的文件并返回一个路径到内容的映射
  @Params           :
  @Return           :
********************************************************************************/
func mergeTrees(headTrees map[string]string, otherThrees map[string]string) map[string][]byte {
	tree := make(map[string][]byte)
	for treeComparsion := range compareTrees(headTrees, otherThrees) {
		tree[treeComparsion.path] = mergeBlobs(treeComparsion.fOid, treeComparsion.tOid)
	}
	return tree
}

// prettier-ignore
/*******************************************************************************
  @Function name    : mergeBlobs
  @Description      : 合并两个文件并返回字节流
  @Params           :
  @Return           :
********************************************************************************/
func mergeBlobs(foid string, toid string) []byte {
	// 写入两个临时文件进行合并
	fContent, err := DoRunCatFile(foid, "blob")
	if err != nil {
		fmt.Printf("error with cat file %s\n", foid)
		os.Exit(1)
	}
	from, err := common.OenpTmp(fContent)
	if err != nil {
		os.Exit(1)
	}
	defer from.Close()
	defer os.Remove(from.Name())
	tContent, err := DoRunCatFile(toid, "blob")
	if err != nil {
		fmt.Printf("error with cat file %s\n", toid)
		os.Exit(1)
	}

	to, err := common.OenpTmp(tContent)
	if err != nil {
		os.Exit(1)
	}
	defer to.Close()
	defer os.Remove(to.Name())
	// 执行命令
	cmd := exec.Command("diff",
		"-DHEAD",
		from.Name(),
		to.Name(),
	)
	output, err := cmd.CombinedOutput()
	// 处理 diff 的特定退出码
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return output // 1 表示有差异，是正常结果
			}
		}
		fmt.Println("error:", err)
		os.Exit(1)
	}
	return output
}
