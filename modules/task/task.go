// Copyright 2019 Gitea. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package task

import (
	"fmt"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/graceful"
	"code.gitea.io/gitea/modules/json"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/migrations/base"
	"code.gitea.io/gitea/modules/queue"
	repo_module "code.gitea.io/gitea/modules/repository"
	"code.gitea.io/gitea/modules/secret"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/modules/timeutil"
	"code.gitea.io/gitea/modules/util"
)

// taskQueue is a global queue of tasks
var taskQueue queue.Queue

// Run a task
func Run(t *models.Task) error {
	switch t.Type {
	case structs.TaskTypeMigrateRepo:
		return runMigrateTask(t)
	default:
		return fmt.Errorf("Unknown task type: %d", t.Type)
	}
}

// Init will start the service to get all unfinished tasks and run them
func Init() error {
	taskQueue = queue.CreateQueue("task", handle, &models.Task{})

	if taskQueue == nil {
		return fmt.Errorf("Unable to create Task Queue")
	}

	go graceful.GetManager().RunWithShutdownFns(taskQueue.Run)

	return nil
}

func handle(data ...queue.Data) {
	for _, datum := range data {
		task := datum.(*models.Task)
		if err := Run(task); err != nil {
			log.Error("Run task failed: %v", err)
		}
	}
}

// MigrateRepository add migration repository to task
func MigrateRepository(doer, u *models.User, opts base.MigrateOptions) (*models.Repository, error) {
	task, repo, err := createMigrationTask(doer, u, opts)
	if err != nil {
		return repo, err
	}

	return repo, taskQueue.Push(task)
}

// createMigrationTask creates a migrate task
func createMigrationTask(doer, u *models.User, opts base.MigrateOptions) (*models.Task, *models.Repository, error) {
	// encrypt credentials for persistence
	var err error
	opts.CloneAddrEncrypted, err = secret.EncryptSecret(setting.SecretKey, opts.CloneAddr)
	if err != nil {
		return nil, nil, err
	}
	opts.CloneAddr = util.NewStringURLSanitizer(opts.CloneAddr, true).Replace(opts.CloneAddr)
	opts.AuthPasswordEncrypted, err = secret.EncryptSecret(setting.SecretKey, opts.AuthPassword)
	if err != nil {
		return nil, nil, err
	}
	opts.AuthPassword = ""
	opts.AuthTokenEncrypted, err = secret.EncryptSecret(setting.SecretKey, opts.AuthToken)
	if err != nil {
		return nil, nil, err
	}
	opts.AuthToken = ""
	bs, err := json.Marshal(&opts)
	if err != nil {
		return nil, nil, err
	}

	var task = models.Task{
		DoerID:         doer.ID,
		OwnerID:        u.ID,
		Type:           structs.TaskTypeMigrateRepo,
		Status:         structs.TaskStatusQueue,
		PayloadContent: string(bs),
	}

	if err := models.CreateTask(&task); err != nil {
		return nil, nil, err
	}

	repo, err := repo_module.CreateRepository(doer, u, models.CreateRepoOptions{
		Name:           opts.RepoName,
		Description:    opts.Description,
		OriginalURL:    opts.CloneAddr,
		GitServiceType: opts.GitServiceType,
		IsPrivate:      opts.Private,
		IsMirror:       opts.Mirror,
		Status:         models.RepositoryBeingMigrated,
	})
	if err != nil {
		task.EndTime = timeutil.TimeStampNow()
		task.Status = structs.TaskStatusFailed
		err2 := task.UpdateCols("end_time", "status")
		if err2 != nil {
			log.Error("UpdateCols Failed: %v", err2.Error())
		}
		return nil, nil, err
	}

	task.RepoID = repo.ID
	if err = task.UpdateCols("repo_id"); err != nil {
		return nil, repo, err
	}

	return &task, repo, nil
}
