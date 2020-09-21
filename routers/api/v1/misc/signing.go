// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package misc

import (
	"fmt"
	"net/http"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
)

// GpgEmptyResponse return it when user not has gpg public key.
const GpgEmptyResponse = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Note: This user hasn't uploaded any GPG keys.


=twTO
-----END PGP PUBLIC KEY BLOCK-----`

// SigningKey returns the public key of the default signing key if it exists
func SigningKey(ctx *context.APIContext) {
	// swagger:operation GET /signing-key.gpg miscellaneous getSigningKey
	// ---
	// summary: Get default signing-key.gpg
	// produces:
	//     - text/plain
	// responses:
	//   "200":
	//     description: "GPG armored public key"
	//     schema:
	//       type: string

	// swagger:operation GET /repos/{owner}/{repo}/signing-key.gpg repository repoSigningKey
	// ---
	// summary: Get signing-key.gpg for given repository
	// produces:
	//     - text/plain
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     description: "GPG armored public key"
	//     schema:
	//       type: string

	path := ""
	if ctx.Repo != nil && ctx.Repo.Repository != nil {
		path = ctx.Repo.Repository.RepoPath()
	}

	content, err := models.PublicSigningKey(path)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "gpg export", err)
		return
	}

	if len(content) == 0 {
		_, err = ctx.Write([]byte(GpgEmptyResponse))
	} else {
		_, err = ctx.Write([]byte(content))
	}
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "gpg export", fmt.Errorf("Error writing key content %v", err))
	}
}
