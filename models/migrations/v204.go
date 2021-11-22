// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migrations

import (
	"xorm.io/xorm"
)

func setOwnersTeamToSeePrivateIssues(x *xorm.Engine) error {
	type Team struct {
		ID                  int64 `xorm:"pk autoincr"`
		CanSeePrivateIssues bool  `xorm:"NOT NULL DEFAULT false"`
	}

	if err := x.Sync2(new(Team)); err != nil {
		return err
	}

	_, err := x.Exec("UPDATE `team` SET `can_see_private_issues` = ? WHERE `name`=?",
		true, "Owners")
	return err
}
