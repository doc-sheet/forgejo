// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/sync"
	"code.gitea.io/gitea/modules/util"

	"github.com/Unknwon/com"
)

var (
	reservedWikiNames = []string{"_pages", "_new", "_edit", "_delete", "raw"}
	wikiWorkingPool   = sync.NewExclusivePool()
)

// NormalizeWikiName normalizes a wiki name
func NormalizeWikiName(name string) string {
	return strings.Replace(name, "-", " ", -1)
}

// WikiNameToSubURL converts a wiki name to its corresponding sub-URL. This will escape dangerous letters.
func WikiNameToSubURL(name string) string {
	// remove path up
	re1 := regexp.MustCompile(`(\.\.\/)`)
	name = re1.ReplaceAllString(name, "")
	// trim whitespace and /
	name = strings.Trim(name, "\n\r\t /")
	name = url.QueryEscape(name)
	//restore spaces
	re3 := regexp.MustCompile(`(?m)(%20|\+)`)
	return re3.ReplaceAllString(name, "%20")
}

// WikiNameToFilename converts a wiki name to its corresponding filename.
func WikiNameToFilename(name string) string {
	name = strings.Replace(name, " ", "-", -1)
	return url.QueryEscape(name) + ".md"
}

// WikiNameToPathFilename converts a wiki name to its corresponding filename, keep directory paths.
func WikiNameToPathFilename(name string) string {
	var restore = [1][2]string{
		{`(\.\.\/)`, ""}, // remove path up
	}
	for _, kv := range restore {
		loopRe := regexp.MustCompile(kv[0])
		name = loopRe.ReplaceAllString(name, kv[1])
	}
	name = strings.Trim(name, "\n\r\t ./") // trim whitespace and / .
	return name + ".md"
}

// FilenameToPathFilename converts a wiki filename to filename with filepath.
func FilenameToPathFilename(name string) string {
	// restore spaces and slashes
	var restore = [4][2]string{
		{`(?m)%2F`, "/"},      //recover slashes /
		{`(?m)(%20|\+)`, " "}, //restore spaces
		{`(?m)(%25)`, "%"},    //restore %
		{`(?m)(%26)`, "&"},    //restore &
	}
	for _, kv := range restore {
		loopRe := regexp.MustCompile(kv[0])
		name = loopRe.ReplaceAllString(name, kv[1])
	}
	return name
}

// WikiNameToRawPrefix Get raw file path inside wiki, removes last path element and returns
func WikiNameToRawPrefix(repositoryName string, wikiPage string) string {
	a := strings.Split(wikiPage, "/")
	a = a[:len(a)-1]
	return util.URLJoin(repositoryName, "wiki", "raw", strings.Join(a, "/"))
}

// WikiFilenameToName converts a wiki filename to its corresponding page name.
func WikiFilenameToName(filename string) (string, string, error) {
	if !strings.HasSuffix(filename, ".md") {
		return "", "", ErrWikiInvalidFileName{filename}
	}
	basename := filename[:len(filename)-3]
	unescaped, err := url.QueryUnescape(basename)
	if err != nil {
		return basename, basename, err
	}
	return unescaped, basename, nil
}

// WikiCloneLink returns clone URLs of repository wiki.
func (repo *Repository) WikiCloneLink() *CloneLink {
	return repo.cloneLink(x, true)
}

// WikiPath returns wiki data path by given user and repository name.
func WikiPath(userName, repoName string) string {
	return filepath.Join(UserPath(userName), strings.ToLower(repoName)+".wiki.git")
}

// WikiPath returns wiki data path for given repository.
func (repo *Repository) WikiPath() string {
	return WikiPath(repo.MustOwnerName(), repo.Name)
}

// HasWiki returns true if repository has wiki.
func (repo *Repository) HasWiki() bool {
	return com.IsDir(repo.WikiPath())
}

// InitWiki initializes a wiki for repository,
// it does nothing when repository already has wiki.
func (repo *Repository) InitWiki() error {
	if repo.HasWiki() {
		return nil
	}

	if err := git.InitRepository(repo.WikiPath(), true); err != nil {
		return fmt.Errorf("InitRepository: %v", err)
	} else if err = createDelegateHooks(repo.WikiPath()); err != nil {
		return fmt.Errorf("createDelegateHooks: %v", err)
	}
	return nil
}

// nameAllowed checks if a wiki name is allowed
func nameAllowed(name string) error {
	for _, reservedName := range reservedWikiNames {
		if name == reservedName {
			return ErrWikiReservedName{name}
		}
	}
	return nil
}

// checkNewWikiFilename check filename or file exists inside repository
func checkNewWikiFilename(repo *git.Repository, name string) (bool, error) {
	filesInIndex, err := repo.LsFiles(name)
	if err != nil {
		log.Error("%v", err)
		return false, err
	}
	for _, file := range filesInIndex {
		if file == name {
			return true, ErrWikiAlreadyExist{name}
		}
	}
	return false, nil
}

// updateWikiPage adds a new page to the repository wiki.
func (repo *Repository) updateWikiPage(doer *User, oldWikiName, newWikiName, content, message string, isNew bool) (err error) {
	if err = nameAllowed(newWikiName); err != nil {
		return err
	}
	wikiWorkingPool.CheckIn(com.ToStr(repo.ID))
	defer wikiWorkingPool.CheckOut(com.ToStr(repo.ID))

	if err = repo.InitWiki(); err != nil {
		return fmt.Errorf("InitWiki: %v", err)
	}

	hasMasterBranch := git.IsBranchExist(repo.WikiPath(), "master")

	basePath, err := CreateTemporaryPath("update-wiki")
	if err != nil {
		return err
	}
	defer func() {
		if err := RemoveTemporaryPath(basePath); err != nil {
			log.Error("Merge: RemoveTemporaryPath: %s", err)
		}
	}()

	cloneOpts := git.CloneRepoOptions{
		Bare:   true,
		Shared: true,
	}

	if hasMasterBranch {
		cloneOpts.Branch = "master"
	}

	if err := git.Clone(repo.WikiPath(), basePath, cloneOpts); err != nil {
		log.Error("Failed to clone repository: %s (%v)", repo.FullName(), err)
		return fmt.Errorf("Failed to clone repository: %s (%v)", repo.FullName(), err)
	}

	gitRepo, err := git.OpenRepository(basePath)

	fmt.Println(reflect.TypeOf(gitRepo))

	if err != nil {
		log.Error("Unable to open temporary repository: %s (%v)", basePath, err)
		return fmt.Errorf("Failed to open new temporary repository in: %s %v", basePath, err)
	}

	if hasMasterBranch {
		if err := gitRepo.ReadTreeToIndex("HEAD"); err != nil {
			log.Error("Unable to read HEAD tree to index in: %s %v", basePath, err)
			return fmt.Errorf("Unable to read HEAD tree to index in: %s %v", basePath, err)
		}
	}

	newWikiPath := WikiNameToFilename(newWikiName)
	newWikiDirPath := WikiNameToPathFilename(newWikiName)

	if isNew {
		// check file already exists - plain structure
		if _, err := checkNewWikiFilename(gitRepo, newWikiPath); err != nil {
			return err
		}

		// check file already exists - directory structure
		if _, err := checkNewWikiFilename(gitRepo, newWikiDirPath); err != nil {
			return err
		}
	} else {
		var found bool

		// check file already exists - plain structure
		oldWikiPath := WikiNameToFilename(oldWikiName)
		if found, err = checkNewWikiFilename(gitRepo, oldWikiPath); err != nil && !found {
			return err
		}
		if found {
			err := gitRepo.RemoveFilesFromIndex(oldWikiPath)
			if err != nil {
				log.Error("%v", err)
				return err
			}
		}

		// check file already exists - directory structure
		oldWikiDirPath := WikiNameToPathFilename(oldWikiName)
		if found, err = checkNewWikiFilename(gitRepo, oldWikiDirPath); err != nil && !found {
			return err
		}
		if found {
			err := gitRepo.RemoveFilesFromIndex(oldWikiDirPath)
			if err != nil {
				log.Error("%v", err)
				return err
			}
		}
	}

	newWikiDirPath = FilenameToPathFilename(newWikiDirPath)

	// FIXME: The wiki doesn't have lfs support at present - if this changes need to check attributes here

	objectHash, err := gitRepo.HashObject(strings.NewReader(content))
	if err != nil {
		log.Error("%v", err)
		return err
	}

	if err := gitRepo.AddObjectToIndex("100644", objectHash, newWikiDirPath); err != nil {
		log.Error("%v", err)
		return err
	}

	tree, err := gitRepo.WriteTree()
	if err != nil {
		log.Error("%v", err)
		return err
	}

	commitTreeOpts := git.CommitTreeOpts{
		Message: message,
	}
	if hasMasterBranch {
		commitTreeOpts.Parents = []string{"HEAD"}
	}
	commitHash, err := gitRepo.CommitTree(doer.NewGitSig(), tree, commitTreeOpts)
	if err != nil {
		log.Error("%v", err)
		return err
	}

	if err := git.Push(basePath, git.PushOptions{
		Remote: "origin",
		Branch: fmt.Sprintf("%s:%s%s", commitHash.String(), git.BranchPrefix, "master"),
		Env:    PushingEnvironment(doer, repo),
	}); err != nil {
		log.Error("%v", err)
		return fmt.Errorf("Push: %v", err)
	}

	return nil
}

// AddWikiPage adds a new wiki page with a given wikiPath.
func (repo *Repository) AddWikiPage(doer *User, wikiName, content, message string) error {
	return repo.updateWikiPage(doer, "", wikiName, content, message, true)
}

// EditWikiPage updates a wiki page identified by its wikiPath,
// optionally also changing wikiPath.
func (repo *Repository) EditWikiPage(doer *User, oldWikiName, newWikiName, content, message string) error {
	return repo.updateWikiPage(doer, oldWikiName, newWikiName, content, message, false)
}

// DeleteWikiPage deletes a wiki page identified by its path.
func (repo *Repository) DeleteWikiPage(doer *User, wikiName string) (err error) {
	wikiWorkingPool.CheckIn(com.ToStr(repo.ID))
	defer wikiWorkingPool.CheckOut(com.ToStr(repo.ID))

	if err = repo.InitWiki(); err != nil {
		return fmt.Errorf("InitWiki: %v", err)
	}

	basePath, err := CreateTemporaryPath("update-wiki")
	if err != nil {
		return err
	}
	defer func() {
		if err := RemoveTemporaryPath(basePath); err != nil {
			log.Error("Merge: RemoveTemporaryPath: %s", err)
		}
	}()

	if err := git.Clone(repo.WikiPath(), basePath, git.CloneRepoOptions{
		Bare:   true,
		Shared: true,
		Branch: "master",
	}); err != nil {
		log.Error("Failed to clone repository: %s (%v)", repo.FullName(), err)
		return fmt.Errorf("Failed to clone repository: %s (%v)", repo.FullName(), err)
	}

	gitRepo, err := git.OpenRepository(basePath)
	if err != nil {
		log.Error("Unable to open temporary repository: %s (%v)", basePath, err)
		return fmt.Errorf("Failed to open new temporary repository in: %s %v", basePath, err)
	}

	if err := gitRepo.ReadTreeToIndex("HEAD"); err != nil {
		log.Error("Unable to read HEAD tree to index in: %s %v", basePath, err)
		return fmt.Errorf("Unable to read HEAD tree to index in: %s %v", basePath, err)
	}

	var found bool

	// check file exists - plain structure
	wikiPath := WikiNameToFilename(wikiName)
	if found, err = checkNewWikiFilename(gitRepo, wikiName); err != nil && !found {
		return err
	}
	if found {
		err := gitRepo.RemoveFilesFromIndex(wikiPath)
		if err != nil {
			return err
		}
	} else {
		// check file exists - plain structure
		wikiDirPath := WikiNameToPathFilename(wikiName)
		if found, err = checkNewWikiFilename(gitRepo, wikiDirPath); err != nil && !found {
			return err
		}
		if found {
			err := gitRepo.RemoveFilesFromIndex(wikiDirPath)
			if err != nil {
				return err
			}
		} else {
			return os.ErrNotExist
		}
	}

	// FIXME: The wiki doesn't have lfs support at present - if this changes need to check attributes here

	tree, err := gitRepo.WriteTree()
	if err != nil {
		return err
	}
	message := "Delete page '" + wikiName + "'"

	commitHash, err := gitRepo.CommitTree(doer.NewGitSig(), tree, git.CommitTreeOpts{
		Message: message,
		Parents: []string{"HEAD"},
	})
	if err != nil {
		return err
	}

	if err := git.Push(basePath, git.PushOptions{
		Remote: "origin",
		Branch: fmt.Sprintf("%s:%s%s", commitHash.String(), git.BranchPrefix, "master"),
		Env:    PushingEnvironment(doer, repo),
	}); err != nil {
		return fmt.Errorf("Push: %v", err)
	}

	return nil
}
