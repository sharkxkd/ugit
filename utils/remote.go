// prettier-ignore
/*******************************************************************************
  * FILENAME    : remote.go
  * Date        : 2025/12/28 17:03:18
  * Author      : zc
  * Version     : v1.0.0
  * Decription  :
********************************************************************************/
package utils

import "fmt"

func fetch(remotePath string) {
	fmt.Println("Will fetch the following refs:")
	for refname := range getRemoteRefs(remotePath, "refs/heads") {
		fmt.Printf("- %s\n", refname)
	}
}

func getRemoteRefs(remotePath string, prefix string) []string {
	var res []string
	ChangeGitDir(remotePath, func() {
		for ref := range iterRefs(prefix, true) {
			res = append(res, ref.refname)
		}
	})
	return res
}
