// Copyright 2022 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package activitypub

import (
	"context"
	"strings"

	repo_model "code.gitea.io/gitea/models/repo"

	ap "github.com/go-ap/activitypub"
)

// Process a Like activity to star a repository
func ReceiveStar(ctx context.Context, like ap.Like) (err error) {
	user, err := PersonIRIToUser(ctx, like.Actor.GetLink())
	if err != nil {
		return
	}
	repo, err := RepositoryIRIToRepository(ctx, like.Object.GetLink())
	if err != nil || strings.Contains(repo.Name, "@") || repo.IsPrivate {
		return
	}
	return repo_model.StarRepo(user.ID, repo.ID, true)
}
