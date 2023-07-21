// Copyright 2023 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"testing"

	auth_model "code.gitea.io/gitea/models/auth"
	repo_model "code.gitea.io/gitea/models/repo"
	"code.gitea.io/gitea/models/unittest"
	user_model "code.gitea.io/gitea/models/user"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/tests"

	"github.com/stretchr/testify/assert"
)

func TestAPIUserBlock(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	user := "user4"
	session := loginUser(t, user)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteUser)

	t.Run("BlockUser", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequest(t, "PUT", fmt.Sprintf("/api/v1/user/block/user2?token=%s", token))
		MakeRequest(t, req, http.StatusNoContent)

		unittest.AssertExistsAndLoadBean(t, &user_model.BlockedUser{UserID: 4, BlockID: 2})
	})

	t.Run("ListBlocked", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/user/list_blocked?token=%s", token))
		resp := MakeRequest(t, req, http.StatusOK)

		// One user just got blocked and the other one is defined in the fixtures.
		assert.Equal(t, "2", resp.Header().Get("X-Total-Count"))

		var blockedUsers []api.BlockedUser
		DecodeJSON(t, resp, &blockedUsers)
		assert.Len(t, blockedUsers, 2)
		assert.EqualValues(t, 1, blockedUsers[0].BlockID)
		assert.EqualValues(t, 2, blockedUsers[1].BlockID)
	})

	t.Run("UnblockUser", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequest(t, "PUT", fmt.Sprintf("/api/v1/user/unblock/user2?token=%s", token))
		MakeRequest(t, req, http.StatusNoContent)

		unittest.AssertNotExistsBean(t, &user_model.BlockedUser{UserID: 4, BlockID: 2})
	})
}

func TestAPIOrgBlock(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	user := "user5"
	org := "user6"
	session := loginUser(t, user)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteOrganization)

	t.Run("BlockUser", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequest(t, "PUT", fmt.Sprintf("/api/v1/orgs/%s/block/user2?token=%s", org, token))
		MakeRequest(t, req, http.StatusNoContent)

		unittest.AssertExistsAndLoadBean(t, &user_model.BlockedUser{UserID: 6, BlockID: 2})
	})

	t.Run("ListBlocked", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/orgs/%s/list_blocked?token=%s", org, token))
		resp := MakeRequest(t, req, http.StatusOK)

		assert.Equal(t, "1", resp.Header().Get("X-Total-Count"))

		var blockedUsers []api.BlockedUser
		DecodeJSON(t, resp, &blockedUsers)
		assert.Len(t, blockedUsers, 1)
		assert.EqualValues(t, 2, blockedUsers[0].BlockID)
	})

	t.Run("UnblockUser", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequest(t, "PUT", fmt.Sprintf("/api/v1/orgs/%s/unblock/user2?token=%s", org, token))
		MakeRequest(t, req, http.StatusNoContent)

		unittest.AssertNotExistsBean(t, &user_model.BlockedUser{UserID: 6, BlockID: 2})
	})
}

// TestAPIBlock_AddCollaborator ensures that the doer and blocked user cannot
// add each others as collaborators via the API.
func TestAPIBlock_AddCollaborator(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	user1 := "user10"
	user2 := "user2"
	perm := "write"
	collabOption := &api.AddCollaboratorOption{Permission: &perm}

	// User1 blocks User2.
	session := loginUser(t, user1)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteUser, auth_model.AccessTokenScopeWriteRepository)

	req := NewRequest(t, "PUT", fmt.Sprintf("/api/v1/user/block/%s?token=%s", user2, token))
	MakeRequest(t, req, http.StatusNoContent)
	unittest.AssertExistsAndLoadBean(t, &user_model.BlockedUser{UserID: 10, BlockID: 2})

	t.Run("BlockedUser Add Doer", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 2, OwnerID: 2})
		session := loginUser(t, user2)
		token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

		req := NewRequestWithJSON(t, "PUT", fmt.Sprintf("/api/v1/repos/%s/%s/collaborators/%s?token=%s", user2, repo.Name, user1, token), collabOption)
		session.MakeRequest(t, req, http.StatusForbidden)
	})

	t.Run("Doer Add BlockedUser", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 7, OwnerID: 10})
		session := loginUser(t, user1)
		token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

		req := NewRequestWithJSON(t, "PUT", fmt.Sprintf("/api/v1/repos/%s/%s/collaborators/%s?token=%s", user1, repo.Name, user2, token), collabOption)
		session.MakeRequest(t, req, http.StatusForbidden)
	})
}
