// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/test"
	"testing"

	"code.gitea.io/gitea/models"

	"github.com/stretchr/testify/assert"
)

func TestCleanUploadName(t *testing.T) {
	models.PrepareTestEnv(t)

	var kases = map[string]string{
		".git/refs/master":               "",
		"/root/abc":                      "root/abc",
		"./../../abc":                    "abc",
		"a/../.git":                      "",
		"a/../../../abc":                 "abc",
		"../../../acd":                   "acd",
		"../../.git/abc":                 "",
		"..\\..\\.git/abc":               "..\\..\\.git/abc",
		"..\\../.git/abc":                "",
		"..\\../.git":                    "",
		"abc/../def":                     "def",
		".drone.yml":                     ".drone.yml",
		".abc/def/.drone.yml":            ".abc/def/.drone.yml",
		"..drone.yml.":                   "..drone.yml.",
		"..a.dotty...name...":            "..a.dotty...name...",
		"..a.dotty../.folder../.name...": "..a.dotty../.folder../.name...",
	}
	for k, v := range kases {
		assert.EqualValues(t, cleanUploadFileName(k), v)
	}
}

func TestGetUniquePatchBranchName(t *testing.T) {
	models.PrepareTestEnv(t)
	ctx := test.MockContext(t, "user2/repo1")
	ctx.SetParams(":id", "1")
	test.LoadRepo(t, ctx, 1)
	test.LoadRepoCommit(t, ctx)
	test.LoadUser(t, ctx, 2)
	test.LoadGitRepo(t, ctx)
	expectedBranchName := "user2-patch-1"
	branchName := GetUniquePatchBranchName(ctx)
	assert.Equal(t, expectedBranchName, branchName)
}

func TestGetClosestParentWithFiles(t *testing.T) {
	models.PrepareTestEnv(t)
	ctx := test.MockContext(t, "user2/repo1")
	ctx.SetParams(":id", "1")
	test.LoadRepo(t, ctx, 1)
	test.LoadRepoCommit(t, ctx)
	test.LoadUser(t, ctx, 2)
	test.LoadGitRepo(t, ctx)
	repo := ctx.Repo.Repository
	branch := repo.DefaultBranch
	gitRepo, _ := git.OpenRepository(repo.RepoPath())
	commit, _ := gitRepo.GetBranchCommit(branch)
	expectedTreePath := ""
	treePath := GetClosestParentWithFiles("dir/dir/dir", commit)
	assert.Equal(t, expectedTreePath, treePath)
}
