// Copyright 2016 The Gitea Authors. All rights reserved.
// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"time"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/models/forgefed"
	user_model "code.gitea.io/gitea/models/user"

	//"code.gitea.io/gitea/modules/activitypub"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/timeutil"
)

// Star represents a starred repo by an user.
type Star struct {
	ID          int64              `xorm:"pk autoincr"`
	UID         int64              `xorm:"UNIQUE(s)"`
	RepoID      int64              `xorm:"UNIQUE(s)"`
	CreatedUnix timeutil.TimeStamp `xorm:"INDEX created"`
}

func init() {
	db.RegisterModel(new(Star))
}

func StarRepo(ctx context.Context, doer user_model.User, repoID int64, star bool) error {
	if err := starLocalRepo(ctx, doer.ID, repoID, star); err != nil {
		return err
	}

	if star && setting.Federation.Enabled {
		if err := sendLikeActivities(ctx, doer, repoID); err != nil {
			return err
		}
	}

	return nil
}

// StarRepo or unstar repository.
func starLocalRepo(ctx context.Context, userID, repoID int64, star bool) error {
	ctx, committer, err := db.TxContext(ctx)
	if err != nil {
		return err
	}
	defer committer.Close()
	staring := IsStaring(ctx, userID, repoID)

	if star {
		if staring {
			return nil
		}

		if err := db.Insert(ctx, &Star{UID: userID, RepoID: repoID}); err != nil {
			return err
		}
		if _, err := db.Exec(ctx, "UPDATE `repository` SET num_stars = num_stars + 1 WHERE id = ?", repoID); err != nil {
			return err
		}
		if _, err := db.Exec(ctx, "UPDATE `user` SET num_stars = num_stars + 1 WHERE id = ?", userID); err != nil {
			return err
		}
	} else {
		if !staring {
			return nil
		}

		if _, err := db.DeleteByBean(ctx, &Star{UID: userID, RepoID: repoID}); err != nil {
			return err
		}
		if _, err := db.Exec(ctx, "UPDATE `repository` SET num_stars = num_stars - 1 WHERE id = ?", repoID); err != nil {
			return err
		}
		if _, err := db.Exec(ctx, "UPDATE `user` SET num_stars = num_stars - 1 WHERE id = ?", userID); err != nil {
			return err
		}
	}

	return committer.Commit()
}

// ToDo: Move to federation service or simillar
func sendLikeActivities(ctx context.Context, doer user_model.User, repoID int64) error {

	federatedRepos, err := FindFederatedReposByRepoID(ctx, repoID)
	log.Info("Federated Repos is: %v", federatedRepos)
	if err != nil {
		return err
	}

	/*apclient, err := activitypub.NewClient(ctx, &doer, doer.APAPIURL())
	if err != nil {
		return err
	}*/

	for _, federatedRepo := range federatedRepos {
		target := federatedRepo.Uri + "/inbox/" // A like goes to the inbox of the federated repo
		log.Info("Federated Repo URI is: %v", target)
		likeActivity, err := forgefed.NewForgeLike(doer.APAPIURL(), target, time.Now())
		if err != nil {
			return err
		}
		log.Info("Like Activity: %v", likeActivity)
		/*json, err := likeActivity.MarshalJSON()
		if err != nil {
			return err
		}*/

		// TODO: decouple loading & creating activities from sending them - use two loops.
		// TODO: set timeouts for outgoing request in oder to mitigate DOS by slow lories
		// TODO: Check if we need to respect rate limits
		// ToDo: Change this to the standalone table of FederatedRepos
		//apclient.Post([]byte(json), target)
	}

	return nil
}

// IsStaring checks if user has starred given repository.
func IsStaring(ctx context.Context, userID, repoID int64) bool {
	has, _ := db.GetEngine(ctx).Get(&Star{UID: userID, RepoID: repoID})
	return has
}

// GetStargazers returns the users that starred the repo.
func GetStargazers(ctx context.Context, repo *Repository, opts db.ListOptions) ([]*user_model.User, error) {
	sess := db.GetEngine(ctx).Where("star.repo_id = ?", repo.ID).
		Join("LEFT", "star", "`user`.id = star.uid")
	if opts.Page > 0 {
		sess = db.SetSessionPagination(sess, &opts)

		users := make([]*user_model.User, 0, opts.PageSize)
		return users, sess.Find(&users)
	}

	users := make([]*user_model.User, 0, 8)
	return users, sess.Find(&users)
}

// ClearRepoStars clears all stars for a repository and from the user that starred it.
// Used when a repository is set to private.
func ClearRepoStars(ctx context.Context, repoID int64) error {
	if _, err := db.Exec(ctx, "UPDATE `user` SET num_stars=num_stars-1 WHERE id IN (SELECT `uid` FROM `star` WHERE repo_id = ?)", repoID); err != nil {
		return err
	}

	if _, err := db.Exec(ctx, "UPDATE `repository` SET num_stars = 0 WHERE id = ?", repoID); err != nil {
		return err
	}

	return db.DeleteBeans(ctx, Star{RepoID: repoID})
}
