// Copyright 2023, 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package activitypub

import (
	"fmt"
	"net/http"
	"strings"

	"code.gitea.io/gitea/modules/activitypub"
	"code.gitea.io/gitea/modules/forgefed"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/web"
	"code.gitea.io/gitea/services/context"
	"code.gitea.io/gitea/services/federation"

	ap "github.com/go-ap/activitypub"
	"github.com/go-ap/jsonld"
)

// Repository function returns the Repository actor for a repo
func Repository(ctx *context.APIContext) {
	// swagger:operation GET /activitypub/repository-id/{repository-id} activitypub activitypubRepository
	// ---
	// summary: Returns the Repository actor for a repo
	// produces:
	// - application/json
	// parameters:
	// - name: repository-id
	//   in: path
	//   description: repository ID of the repo
	//   type: integer
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/ActivityPub"

	link := fmt.Sprintf("%s/api/v1/activitypub/repository-id/%d", strings.TrimSuffix(setting.AppURL, "/"), ctx.Repo.Repository.ID)
	repo := forgefed.RepositoryNew(ap.IRI(link))

	repo.Name = ap.NaturalLanguageValuesNew()
	err := repo.Name.Set("en", ap.Content(ctx.Repo.Repository.Name))
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "Set Name", err)
		return
	}

	repo.PublicKey.ID = ap.IRI(link + "#main-key")
	repo.PublicKey.Owner = ap.IRI(link)

	publicKeyPem, err := activitypub.GetRepoPublicKey(ctx, ctx.Repo.Repository)
	if err != nil {
		ctx.ServerError("GetPublicKey", err)
		return
	}
	repo.PublicKey.PublicKeyPem = publicKeyPem

	binary, err := jsonld.WithContext(
		jsonld.IRI(ap.ActivityBaseURI),
		jsonld.IRI(ap.SecurityContextURI),
	).Marshal(repo)
	if err != nil {
		ctx.ServerError("MarshalJSON", err)
		return
	}
	ctx.Resp.Header().Add("Content-Type", activitypub.ActivityStreamsContentType)
	ctx.Resp.WriteHeader(http.StatusOK)
	if _, err = ctx.Resp.Write(binary); err != nil {
		log.Error("write to resp err: %v", err)
	}
}

// PersonInbox function handles the incoming data for a repository inbox
func RepositoryInbox(ctx *context.APIContext) {
	// swagger:operation POST /activitypub/repository-id/{repository-id}/inbox activitypub activitypubRepositoryInbox
	// ---
	// summary: Send to the inbox
	// produces:
	// - application/json
	// parameters:
	// - name: repository-id
	//   in: path
	//   description: repository ID of the repo
	//   type: integer
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/ForgeLike"
	// responses:
	//   "204":
	//     "$ref": "#/responses/empty"

	repository := ctx.Repo.Repository
	log.Info("RepositoryInbox: repo: %v", repository)

	form := web.GetForm(ctx)
	httpStatus, title, err := federation.ProcessLikeActivity(ctx, form, repository.ID)
	if err != nil {
		ctx.Error(httpStatus, title, err)
	}
	ctx.Status(http.StatusNoContent)
}
