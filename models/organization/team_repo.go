// Copyright 2022 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package organization

import (
	"context"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/models/perm"
)

// TeamRepo represents an team-repository relation.
type TeamRepo struct {
	ID     int64 `xorm:"pk autoincr"`
	OrgID  int64 `xorm:"INDEX"`
	TeamID int64 `xorm:"UNIQUE(s)"`
	RepoID int64 `xorm:"UNIQUE(s)"`
}

// HasTeamRepo returns true if given repository belongs to team.
func HasTeamRepo(ctx context.Context, orgID, teamID, repoID int64) bool {
	has, _ := db.GetEngine(ctx).
		Where("org_id=?", orgID).
		And("team_id=?", teamID).
		And("repo_id=?", repoID).
		Get(new(TeamRepo))
	return has
}

// AddTeamRepo addes a repo for an organization's team
func AddTeamRepo(ctx context.Context, orgID, teamID, repoID int64) error {
	_, err := db.GetEngine(ctx).Insert(&TeamRepo{
		OrgID:  orgID,
		TeamID: teamID,
		RepoID: repoID,
	})
	return err
}

// RemoveTeamRepo remove repository from team
func RemoveTeamRepo(ctx context.Context, teamID, repoID int64) error {
	_, err := db.DeleteByBean(ctx, &TeamRepo{
		TeamID: teamID,
		RepoID: repoID,
	})
	return err
}

// GetTeamsWithAccessToRepo returns all teams in an organization that have given access level to the repository.
func GetTeamsWithAccessToRepo(orgID, repoID int64, mode perm.AccessMode) ([]*Team, error) {
	teams := make([]*Team, 0, 5)
	return teams, db.GetEngine(db.DefaultContext).Where("team.authorize >= ?", mode).
		Join("INNER", "team_repo", "team_repo.team_id = team.id").
		And("team_repo.org_id = ?", orgID).
		And("team_repo.repo_id = ?", repoID).
		Find(&teams)
}
