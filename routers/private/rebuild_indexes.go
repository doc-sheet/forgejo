// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package private includes all internal routes. The package name internal is ideal but Golang is not allowed, so we use private as package name instead.
package private

import (
	"code.gitea.io/gitea/models"

	macaron "gopkg.in/macaron.v1"
)

// RebuildRepoIndex rebuild a repository index
func RebuildRepoIndex(ctx *macaron.Context) {
	repoID := ctx.ParamsInt64(":repoid")
	if err := models.RebuildRepoIndex(repoID); err != nil {
		ctx.JSON(500, map[string]interface{}{
			"err": err.Error(),
		})
		return
	}

	ctx.PlainText(200, []byte("success"))
}
