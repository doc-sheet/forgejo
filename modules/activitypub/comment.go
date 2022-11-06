// Copyright 2022 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package activitypub

import (
	"context"
	"strconv"
	"strings"

	"code.gitea.io/gitea/models/issues"
	repo_model "code.gitea.io/gitea/models/repo"

	ap "github.com/go-ap/activitypub"
)

// Create a comment
func Comment(ctx context.Context, note *ap.Note) error {
	actorUser, err := PersonIRIToUser(ctx, note.AttributedTo.GetLink())
	if err != nil {
		return err
	}

	// TODO: Move IRI processing stuff to iri.go
	context := note.Context.GetLink()
	contextSplit := strings.Split(context.String(), "/")
	username := contextSplit[3]
	reponame := contextSplit[4]
	repo, _ := repo_model.GetRepositoryByOwnerAndNameCtx(ctx, username, reponame)

	idx, _ := strconv.ParseInt(contextSplit[len(contextSplit)-1], 10, 64)
	issue, _ := issues.GetIssueByIndex(repo.ID, idx)
	_, err = issues.CreateCommentCtx(ctx, &issues.CreateCommentOptions{
		Doer:    actorUser,
		Repo:    repo,
		Issue:   issue,
		Content: note.Content.String(),
	})
	return err
}
