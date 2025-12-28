// prettier-ignore
/*******************************************************************************
  * FILENAME    : remote.go
  * Date        : 2025/12/28 17:03:18
  * Author      : zc
  * Version     : v1.0.0
  * Decription  :
********************************************************************************/
package utils

import (
	"maps"
	"path/filepath"
	"slices"
)

// "fmt"

const REMOTE_REF_BASE = "refs/heads/"
const LOCAL_REF_BASE = "refs/remote/"

func fetch(remotePath string) {
	refs := getRemoteRefs(remotePath, REMOTE_REF_BASE)
	for oid := range iterObjectsAndCommits(slices.Collect(maps.Values(refs))...) {
		fetchObjectIfMissing(oid, remotePath)
	}
	for remoteName, value := range refs {
		refname := filepath.Base(remoteName)
		updateRef(filepath.Join(LOCAL_REF_BASE, refname), RefValue{false, value}, true)
	}
}

func getRemoteRefs(remotePath string, prefix string) map[string]string {
	res := make(map[string]string)
	ChangeGitDir(remotePath, func() {
		for ref := range iterRefs(prefix, true) {
			res[ref.refname] = ref.ref.value
		}
	})
	return res
}
