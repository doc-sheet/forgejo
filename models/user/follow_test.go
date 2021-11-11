// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"testing"

	"code.gitea.io/gitea/models/db"
	"github.com/stretchr/testify/assert"
)

func TestIsFollowing(t *testing.T) {
	assert.NoError(t, db.PrepareTestDatabase())
	assert.True(t, IsFollowing(4, 2))
	assert.False(t, IsFollowing(2, 4))
	assert.False(t, IsFollowing(5, db.NonexistentID))
	assert.False(t, IsFollowing(db.NonexistentID, 5))
	assert.False(t, IsFollowing(db.NonexistentID, db.NonexistentID))
}
