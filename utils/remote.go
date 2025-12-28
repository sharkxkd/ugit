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
	"fmt"
	"maps"
	"path/filepath"
	"slices"
)

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

// prettier-ignore
/*******************************************************************************
  @Function name    : getRemoteRefs
  @Description      :
  @Params           :
	-remotePath		:
	-prefix			: 默认值为""
  @Return           :
********************************************************************************/
func getRemoteRefs(remotePath string, prefix string) map[string]string {
	res := make(map[string]string)
	ChangeGitDir(remotePath, func() {
		for ref := range iterRefs(prefix, true) {
			res[ref.refname] = ref.ref.value
		}
	})
	return res
}

func push(remotePath string, refname string) {
	remoteRefs := getRemoteRefs(remotePath, "")
	remoteRef := remoteRefs[refname]
	localRef := getRef(refname, true).value
	if localRef == "" {
		fmt.Printf("error without refname of %s\n", refname)
		return
	}
	// Don't allow force push
	if remoteRef != "" || isAncestorOf(localRef, remoteRef) {
		fmt.Println("Don't allow force push")
		return
	}
	// feat：新增过滤一些存在的文件
	var knownRemoteRefs []string
	for _, value := range remoteRefs {
		if objectsExsits(value) {
			knownRemoteRefs = append(knownRemoteRefs, value)
		}
	}
	remoteObjects := make(map[string]bool)
	for ref := range iterObjectsAndCommits(knownRemoteRefs...) {
		remoteObjects[ref] = true
	}
	var objectsToPush []string
	for localObject := range iterObjectsAndCommits(localRef) {
		if !remoteObjects[localObject] {
			objectsToPush = append(objectsToPush, localObject)
		}
	}
	for _, oid := range objectsToPush {
		pushObject(oid, remotePath)
	}
	ChangeGitDir(remotePath, func() {
		updateRef(refname, RefValue{false, localRef}, true)
	})
}
