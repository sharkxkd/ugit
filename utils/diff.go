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

// type TreeComparison struct {
// 	path string
// 	fOid string
// 	tOid string
// }

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
	for res := range compareTrees(fromTree, toTree) {
		if res[1] != res[2] {
			output += diffBlobs(res[1], res[2], res[0])
		}
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
func compareTrees(trees ...map[string]string) <-chan []string {
	ch := make(chan []string)
	len := len(trees)
	go func() {
		defer close(ch)
		entries := make(map[string][]string)
		// 遍历树
		for i, tree := range trees {
			for path, oid := range tree {
				if _, ok := entries[path]; !ok {
					entries[path] = make([]string, len)
					for j := range entries[path] {
						entries[path][j] = ""
					}
				}
				// 填充当前树的 OID
				entries[path][i] = oid
			}
		}

		// 3. 发送结果（路径 + 各树的 OID）
		for path, oids := range entries {
			result := append([]string{path}, oids...)
			ch <- result
		}
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
		for res := range compareTrees(fromTree, toTree) {
			if res[1] != res[2] {
				if res[1] == "" {
					ch <- ChangedFileType{path: res[0], action: "Created"}
				} else if res[2] == "" {
					ch <- ChangedFileType{path: res[0], action: "Deleted"}
				} else {
					ch <- ChangedFileType{path: res[0], action: "Modified"}
				}
			}
		}
		close(ch)

	}()
	return ch
}

// prettier-ignore
/*******************************************************************************
  @Function name    : mergeTrees
  @Description      : 合并所有的文件并返回一个路径到内容的映射,三路合并,添加一个基础tree
  @Params           :
	-BaseTrees		:
	-HeadTrees		:
	-OtherTrees		:
  @Return           :
********************************************************************************/
func mergeTrees(baseTrees map[string]string, headTrees map[string]string, otherThrees map[string]string) map[string][]byte {
	tree := make(map[string][]byte)
	for res := range compareTrees(baseTrees, headTrees, otherThrees) {
		tree[res[0]] = mergeBlobs(res[1], res[2], res[3])
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
func mergeBlobs(boid string, foid string, toid string) []byte {
	// 写入两个临时文件进行合并
	bContent, err := DoRunCatFile(boid, "blob")
	if err != nil {
		fmt.Printf("error with cat file %s\n", boid)
		os.Exit(1)
	}
	base, err := common.OenpTmp(bContent)
	if err != nil {
		os.Exit(1)
	}
	defer base.Close()
	defer os.Remove(base.Name())
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
	cmd := exec.Command("diff3", "-m",
		"-L", "HEAD", from.Name(),
		"-L", "BASE", base.Name(),
		"-L", "MERGE_HEAD", to.Name(),
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
