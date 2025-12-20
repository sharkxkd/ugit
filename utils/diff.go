// prettier-ignore
/*******************************************************************************
  * FILENAME    : diff.go
  * Date        : 2025/12/20 13:44:19
  * Author      : zc
  * Version     : v1.0.0
  * Decription  : 对比前后版本文件的差异
********************************************************************************/
package utils

import "fmt"

type TreeComparison struct {
	path string
	fOid string
	tOid string
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
		if treeComparsion.fOid == "" {
			output += fmt.Sprintf("Add: %s\n", treeComparsion.path)
		} else if treeComparsion.tOid == "" {
			output += fmt.Sprintf("Delete: %s\n", treeComparsion.path)
		} else if treeComparsion.fOid != treeComparsion.tOid {
			output += fmt.Sprintf("Changed: %s\n", treeComparsion.path)
		}
	}
	return output
}

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
