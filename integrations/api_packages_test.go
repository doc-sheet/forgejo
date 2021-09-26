// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/models/db"
	api "code.gitea.io/gitea/modules/structs"

	"github.com/stretchr/testify/assert"
)

func TestPackageAPI(t *testing.T) {
	defer prepareTestEnv(t)()
	repository := db.AssertExistsAndLoadBean(t, &models.Repository{ID: 15}).(*models.Repository)
	user := db.AssertExistsAndLoadBean(t, &models.User{ID: repository.OwnerID}).(*models.User)
	session := loginUser(t, user.Name)
	token := getTokenForLoggedInUser(t, session)

	packageName := "test-package"
	packageVersion := "1.0.3"
	filename := "file.bin"

	url := fmt.Sprintf("/api/v1/repos/%s/%s/packages/generic/%s/%s/%s?token=%s", user.Name, repository.Name, packageName, packageVersion, filename, token)
	req := NewRequestWithBody(t, "PUT", url, bytes.NewReader([]byte{}))
	MakeRequest(t, req, http.StatusCreated)

	var packageID int64

	t.Run("ListPackages", func(t *testing.T) {
		defer PrintCurrentTest(t)()

		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/packages?token=%s", user.Name, repository.Name, token))
		resp := MakeRequest(t, req, http.StatusOK)

		var packages []*api.Package
		DecodeJSON(t, resp, &packages)

		assert.Len(t, packages, 1)
		assert.Equal(t, "Generic", packages[0].Type)
		assert.Equal(t, packageName, packages[0].Name)
		assert.Equal(t, packageVersion, packages[0].Version)
		assert.NotNil(t, packages[0].Creator)
		assert.Equal(t, user.Name, packages[0].Creator.UserName)

		packageID = packages[0].ID
	})

	t.Run("GetPackage", func(t *testing.T) {
		defer PrintCurrentTest(t)()

		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/packages/%d?token=%s", user.Name, repository.Name, 123456, token))
		MakeRequest(t, req, http.StatusNotFound)

		req = NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/packages/%d?token=%s", user.Name, repository.Name, packageID, token))
		resp := MakeRequest(t, req, http.StatusOK)

		var p *api.Package
		DecodeJSON(t, resp, &p)

		assert.Equal(t, "Generic", p.Type)
		assert.Equal(t, packageName, p.Name)
		assert.Equal(t, packageVersion, p.Version)
		assert.NotNil(t, p.Creator)
		assert.Equal(t, user.Name, p.Creator.UserName)
	})

	t.Run("ListPackageFiles", func(t *testing.T) {
		defer PrintCurrentTest(t)()

		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/packages/%d/files?token=%s", user.Name, repository.Name, 123456, token))
		MakeRequest(t, req, http.StatusNotFound)

		req = NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/packages/%d/files?token=%s", user.Name, repository.Name, packageID, token))
		resp := MakeRequest(t, req, http.StatusOK)

		var files []*api.PackageFile
		DecodeJSON(t, resp, &files)

		assert.Len(t, files, 1)
		assert.Equal(t, int64(0), files[0].Size)
		assert.Equal(t, filename, files[0].Name)
		assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", files[0].HashMD5)
		assert.Equal(t, "da39a3ee5e6b4b0d3255bfef95601890afd80709", files[0].HashSHA1)
		assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", files[0].HashSHA256)
		assert.Equal(t, "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", files[0].HashSHA512)
	})

	t.Run("DeletePackage", func(t *testing.T) {
		defer PrintCurrentTest(t)()

		req := NewRequest(t, "DELETE", fmt.Sprintf("/api/v1/repos/%s/%s/packages/%d?token=%s", user.Name, repository.Name, 123456, token))
		MakeRequest(t, req, http.StatusNotFound)

		req = NewRequest(t, "DELETE", fmt.Sprintf("/api/v1/repos/%s/%s/packages/%d?token=%s", user.Name, repository.Name, packageID, token))
		MakeRequest(t, req, http.StatusNoContent)
	})
}
