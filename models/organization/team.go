// Copyright 2018 The Gitea Authors. All rights reserved.
// Copyright 2016 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package organization

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/models/perm"
	repo_model "code.gitea.io/gitea/models/repo"
	"code.gitea.io/gitea/models/unit"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/log"

	"xorm.io/builder"
)

// ___________
// \__    ___/___ _____    _____
//   |    |_/ __ \\__  \  /     \
//   |    |\  ___/ / __ \|  Y Y  \
//   |____| \___  >____  /__|_|  /
//              \/     \/      \/

// ErrTeamAlreadyExist represents a "TeamAlreadyExist" kind of error.
type ErrTeamAlreadyExist struct {
	OrgID int64
	Name  string
}

// IsErrTeamAlreadyExist checks if an error is a ErrTeamAlreadyExist.
func IsErrTeamAlreadyExist(err error) bool {
	_, ok := err.(ErrTeamAlreadyExist)
	return ok
}

func (err ErrTeamAlreadyExist) Error() string {
	return fmt.Sprintf("team already exists [org_id: %d, name: %s]", err.OrgID, err.Name)
}

// ErrTeamNotExist represents a "TeamNotExist" error
type ErrTeamNotExist struct {
	OrgID  int64
	TeamID int64
	Name   string
}

// IsErrTeamNotExist checks if an error is a ErrTeamNotExist.
func IsErrTeamNotExist(err error) bool {
	_, ok := err.(ErrTeamNotExist)
	return ok
}

func (err ErrTeamNotExist) Error() string {
	return fmt.Sprintf("team does not exist [org_id %d, team_id %d, name: %s]", err.OrgID, err.TeamID, err.Name)
}

// OwnerTeamName return the owner team name
const OwnerTeamName = "Owners"

// Team represents a organization team.
type Team struct {
	ID                      int64 `xorm:"pk autoincr"`
	OrgID                   int64 `xorm:"INDEX"`
	LowerName               string
	Name                    string
	Description             string
	AccessMode              perm.AccessMode          `xorm:"'authorize'"`
	Repos                   []*repo_model.Repository `xorm:"-"`
	Members                 []*user_model.User       `xorm:"-"`
	NumRepos                int
	NumMembers              int
	Units                   []*TeamUnit `xorm:"-"`
	IncludesAllRepositories bool        `xorm:"NOT NULL DEFAULT false"`
	CanCreateOrgRepo        bool        `xorm:"NOT NULL DEFAULT false"`
}

func init() {
	db.RegisterModel(new(Team))
	db.RegisterModel(new(TeamUser))
	db.RegisterModel(new(TeamRepo))
	db.RegisterModel(new(TeamUnit))
}

// SearchTeamOptions holds the search options
type SearchTeamOptions struct {
	db.ListOptions
	UserID      int64
	Keyword     string
	OrgID       int64
	IncludeDesc bool
}

// SearchMembersOptions holds the search options
type SearchMembersOptions struct {
	db.ListOptions
}

// SearchTeam search for teams. Caller is responsible to check permissions.
func SearchTeam(opts *SearchTeamOptions) ([]*Team, int64, error) {
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PageSize == 0 {
		// Default limit
		opts.PageSize = 10
	}

	cond := builder.NewCond()

	if len(opts.Keyword) > 0 {
		lowerKeyword := strings.ToLower(opts.Keyword)
		var keywordCond builder.Cond = builder.Like{"lower_name", lowerKeyword}
		if opts.IncludeDesc {
			keywordCond = keywordCond.Or(builder.Like{"LOWER(description)", lowerKeyword})
		}
		cond = cond.And(keywordCond)
	}

	cond = cond.And(builder.Eq{"org_id": opts.OrgID})

	sess := db.GetEngine(db.DefaultContext)

	count, err := sess.
		Where(cond).
		Count(new(Team))
	if err != nil {
		return nil, 0, err
	}

	sess = sess.Where(cond)
	if opts.PageSize == -1 {
		opts.PageSize = int(count)
	} else {
		sess = sess.Limit(opts.PageSize, (opts.Page-1)*opts.PageSize)
	}

	teams := make([]*Team, 0, opts.PageSize)
	if err = sess.
		OrderBy("lower_name").
		Find(&teams); err != nil {
		return nil, 0, err
	}

	return teams, count, nil
}

// ColorFormat provides a basic color format for a Team
func (t *Team) ColorFormat(s fmt.State) {
	if t == nil {
		log.ColorFprintf(s, "%d:%s (OrgID: %d) %-v",
			log.NewColoredIDValue(0),
			"<nil>",
			log.NewColoredIDValue(0),
			0)
		return
	}
	log.ColorFprintf(s, "%d:%s (OrgID: %d) %-v",
		log.NewColoredIDValue(t.ID),
		t.Name,
		log.NewColoredIDValue(t.OrgID),
		t.AccessMode)
}

// GetUnits return a list of available units for a team
func (t *Team) GetUnits() error {
	return t.getUnits(db.DefaultContext)
}

func (t *Team) getUnits(ctx context.Context) (err error) {
	if t.Units != nil {
		return nil
	}

	t.Units, err = getUnitsByTeamID(ctx, t.ID)
	return err
}

// GetUnitNames returns the team units names
func (t *Team) GetUnitNames() (res []string) {
	if t.AccessMode >= perm.AccessModeAdmin {
		return unit.AllUnitKeyNames()
	}

	for _, u := range t.Units {
		res = append(res, unit.Units[u.Type].NameKey)
	}
	return
}

// GetUnitsMap returns the team units permissions
func (t *Team) GetUnitsMap() map[string]string {
	m := make(map[string]string)
	if t.AccessMode >= perm.AccessModeAdmin {
		for _, u := range unit.Units {
			m[u.NameKey] = t.AccessMode.String()
		}
	} else {
		for _, u := range t.Units {
			m[u.Unit().NameKey] = u.AccessMode.String()
		}
	}
	return m
}

// IsOwnerTeam returns true if team is owner team.
func (t *Team) IsOwnerTeam() bool {
	return t.Name == OwnerTeamName
}

// IsMember returns true if given user is a member of team.
func (t *Team) IsMember(userID int64) bool {
	isMember, err := IsTeamMember(db.DefaultContext, t.OrgID, t.ID, userID)
	if err != nil {
		log.Error("IsMember: %v", err)
		return false
	}
	return isMember
}

// GetRepositoriesCtx returns paginated repositories in team of organization.
func (t *Team) GetRepositoriesCtx(ctx context.Context) error {
	if t.Repos != nil {
		return nil
	}
	return db.GetEngine(ctx).Join("INNER", "team_repo", "repository.id = team_repo.repo_id").
		Where("team_repo.team_id=?", t.ID).
		OrderBy("repository.name").
		Find(&t.Repos)
}

// GetRepositories returns paginated repositories in team of organization.
func (t *Team) GetRepositories(opts *SearchTeamOptions) error {
	if opts.Page == 0 {
		return t.GetRepositoriesCtx(db.DefaultContext)
	}

	return t.GetRepositoriesCtx(db.WithPaginator(db.DefaultContext, opts))
}

// GetMembersCtx returns paginated members in team of organization.
func (t *Team) GetMembersCtx(ctx context.Context) (err error) {
	t.Members, err = GetTeamMembers(ctx, t.ID)
	return err
}

// GetMembers returns paginated members in team of organization.
func (t *Team) GetMembers(opts *SearchMembersOptions) (err error) {
	if opts.Page == 0 {
		return t.GetMembersCtx(db.DefaultContext)
	}

	return t.GetMembersCtx(db.WithPaginator(db.DefaultContext, opts))
}

// UnitEnabled returns if the team has the given unit type enabled
func (t *Team) UnitEnabled(tp unit.Type) bool {
	return t.unitEnabled(db.DefaultContext, tp)
}

func (t *Team) unitEnabled(ctx context.Context, tp unit.Type) bool {
	return t.UnitAccessMode(ctx, tp) > perm.AccessModeNone
}

// UnitAccessMode returns if the team has the given unit type enabled
func (t *Team) UnitAccessMode(ctx context.Context, tp unit.Type) perm.AccessMode {
	if err := t.getUnits(ctx); err != nil {
		log.Warn("Error loading team (ID: %d) units: %s", t.ID, err.Error())
	}

	for _, unit := range t.Units {
		if unit.Type == tp {
			return unit.AccessMode
		}
	}
	return perm.AccessModeNone
}

// IsUsableTeamName tests if a name could be as team name
func IsUsableTeamName(name string) error {
	switch name {
	case "new":
		return db.ErrNameReserved{Name: name}
	default:
		return nil
	}
}

func getTeam(ctx context.Context, orgID int64, name string) (*Team, error) {
	t := &Team{
		OrgID:     orgID,
		LowerName: strings.ToLower(name),
	}
	has, err := db.GetByBean(ctx, t)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrTeamNotExist{orgID, 0, name}
	}
	return t, nil
}

// GetTeam returns team by given team name and organization.
func GetTeam(orgID int64, name string) (*Team, error) {
	return getTeam(db.DefaultContext, orgID, name)
}

// GetTeamIDsByNames returns a slice of team ids corresponds to names.
func GetTeamIDsByNames(orgID int64, names []string, ignoreNonExistent bool) ([]int64, error) {
	ids := make([]int64, 0, len(names))
	for _, name := range names {
		u, err := GetTeam(orgID, name)
		if err != nil {
			if ignoreNonExistent {
				continue
			} else {
				return nil, err
			}
		}
		ids = append(ids, u.ID)
	}
	return ids, nil
}

// GetOwnerTeam returns team by given team name and organization.
func GetOwnerTeam(ctx context.Context, orgID int64) (*Team, error) {
	return getTeam(ctx, orgID, OwnerTeamName)
}

// GetTeamByIDCtx returns team by given ID.
func GetTeamByIDCtx(ctx context.Context, teamID int64) (*Team, error) {
	t := new(Team)
	has, err := db.GetEngine(ctx).ID(teamID).Get(t)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrTeamNotExist{0, teamID, ""}
	}
	return t, nil
}

// GetTeamByID returns team by given ID.
func GetTeamByID(teamID int64) (*Team, error) {
	return GetTeamByIDCtx(db.DefaultContext, teamID)
}

// GetTeamNamesByID returns team's lower name from a list of team ids.
func GetTeamNamesByID(teamIDs []int64) ([]string, error) {
	if len(teamIDs) == 0 {
		return []string{}, nil
	}

	var teamNames []string
	err := db.GetEngine(db.DefaultContext).Table("team").
		Select("lower_name").
		In("id", teamIDs).
		Asc("name").
		Find(&teamNames)

	return teamNames, err
}

func getRepoTeams(e db.Engine, repo *repo_model.Repository) (teams []*Team, err error) {
	return teams, e.
		Join("INNER", "team_repo", "team_repo.team_id = team.id").
		Where("team.org_id = ?", repo.OwnerID).
		And("team_repo.repo_id=?", repo.ID).
		OrderBy("CASE WHEN name LIKE '" + OwnerTeamName + "' THEN '' ELSE name END").
		Find(&teams)
}

// GetRepoTeams gets the list of teams that has access to the repository
func GetRepoTeams(repo *repo_model.Repository) ([]*Team, error) {
	return getRepoTeams(db.GetEngine(db.DefaultContext), repo)
}
