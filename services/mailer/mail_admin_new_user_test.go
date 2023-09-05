// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package mailer

import (
	"context"
	"strconv"
	"strings"
	"testing"

	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/setting"

	"github.com/stretchr/testify/assert"
)

func getTestUsers() []*user_model.User {
	admin := new(user_model.User)
	admin.Name = "admin"
	admin.IsAdmin = true
	admin.Language = "en_US"
	admin.Email = "admin@forgejo.org"

	newUser := new(user_model.User)
	newUser.Name = "new_user"
	newUser.Language = "en_US"
	newUser.IsAdmin = false
	newUser.Email = "new_user@forgejo.org"
	newUser.LastLoginUnix = 1693648327
	newUser.CreatedUnix = 1693648027

	user_model.CreateUser(admin)
	user_model.CreateUser(newUser)

	users := make([]*user_model.User, 0)
	users = append(users, admin)
	users = append(users, newUser)

	return users
}

func TestAdminNotificationMail_test(t *testing.T) {
	mailService := setting.Mailer{
		From:     "test@forgejo.org",
		Protocol: "dummy",
	}

	setting.MailService = &mailService
	setting.Domain = "localhost"
	setting.AppSubURL = "http://localhost"

	users := getTestUsers()
	oldSendAsyncs := sa
	defer func() {
		sa = oldSendAsyncs
	}()

	sa = func(msgs []*Message) {
		assert.Equal(t, len(msgs), 1, "Test provides only one admin user, so only one email must be sent")
		assert.Equal(t, msgs[0].To, users[0].Email, "checks if the recipient is the admin of the instance")
		manageUserURL := "/admin/users/" + strconv.FormatInt(users[1].ID, 10)
		assert.True(t, strings.ContainsAny(msgs[0].Body, manageUserURL), "checks if the message contains the link to manage the newly created user from the admin panel")
	}
	ctx := context.Background()
	NewContext(ctx)
	MailNewUser(ctx, users[1])
}
