// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package convert

import (
	"strings"

	"code.gitea.io/gitea/models"
	api "code.gitea.io/gitea/modules/structs"
)

// ToAPIIssue converts an Issue to API format
// it assumes some fields assigned with values:
// Required - Poster, Labels,
// Optional - Milestone, Assignee, PullRequest
func ToAPIIssue(issue *models.Issue) *api.Issue {
	if err := issue.LoadLabels(); err != nil {
		return &api.Issue{}
	}
	if err := issue.LoadPoster(); err != nil {
		return &api.Issue{}
	}
	if err := issue.LoadRepo(); err != nil {
		return &api.Issue{}
	}

	apiIssue := &api.Issue{
		ID:       issue.ID,
		URL:      issue.APIURL(),
		HTMLURL:  issue.HTMLURL(),
		Index:    issue.Index,
		Poster:   issue.Poster.APIFormat(),
		Title:    issue.Title,
		Body:     issue.Content,
		Labels:   ToLabelList(issue.Labels),
		State:    issue.State(),
		Comments: issue.NumComments,
		Created:  issue.CreatedUnix.AsTime(),
		Updated:  issue.UpdatedUnix.AsTime(),
	}

	apiIssue.Repo = &api.RepositoryMeta{
		ID:       issue.Repo.ID,
		Name:     issue.Repo.Name,
		Owner:    issue.Repo.OwnerName,
		FullName: issue.Repo.FullName(),
	}

	if issue.ClosedUnix != 0 {
		apiIssue.Closed = issue.ClosedUnix.AsTimePtr()
	}

	if err := issue.LoadMilestone(); err != nil {
		return &api.Issue{}
	}
	if issue.Milestone != nil {
		apiIssue.Milestone = issue.Milestone.APIFormat()
	}

	if err := issue.LoadAssignees(); err != nil {
		return &api.Issue{}
	}
	if len(issue.Assignees) > 0 {
		for _, assignee := range issue.Assignees {
			apiIssue.Assignees = append(apiIssue.Assignees, assignee.APIFormat())
		}
		apiIssue.Assignee = issue.Assignees[0].APIFormat() // For compatibility, we're keeping the first assignee as `apiIssue.Assignee`
	}
	if issue.IsPull {
		if err := issue.LoadPullRequest(); err != nil {
			return &api.Issue{}
		}
		apiIssue.PullRequest = &api.PullRequestMeta{
			HasMerged: issue.PullRequest.HasMerged,
		}
		if issue.PullRequest.HasMerged {
			apiIssue.PullRequest.Merged = issue.PullRequest.MergedUnix.AsTimePtr()
		}
	}
	if issue.DeadlineUnix != 0 {
		apiIssue.Deadline = issue.DeadlineUnix.AsTimePtr()
	}

	return apiIssue
}

// ToAPIIssueList converts an IssueList to API format
func ToAPIIssueList(il models.IssueList) []*api.Issue {
	result := make([]*api.Issue, len(il))
	for i := range il {
		result[i] = ToAPIIssue(il[i])
	}
	return result
}

// ToTrackedTime converts TrackedTime to API format
func ToTrackedTime(t *models.TrackedTime) (apiT *api.TrackedTime) {
	apiT = &api.TrackedTime{
		ID:       t.ID,
		IssueID:  t.IssueID,
		UserID:   t.UserID,
		UserName: t.User.Name,
		Time:     t.Time,
		Created:  t.Created,
	}
	if t.Issue != nil {
		apiT.Issue = ToAPIIssue(t.Issue)
	}
	if t.User != nil {
		apiT.UserName = t.User.Name
	}
	return
}

// ToTrackedTimeList converts TrackedTimeList to API format
func ToTrackedTimeList(tl models.TrackedTimeList) api.TrackedTimeList {
	result := make([]*api.TrackedTime, 0, len(tl))
	for _, t := range tl {
		result = append(result, ToTrackedTime(t))
	}
	return result
}

// ToLabel converts Label to API format
func ToLabel(label *models.Label) (apiT *api.Label) {
	return &api.Label{
		ID:          label.ID,
		Name:        label.Name,
		Color:       strings.TrimLeft(label.Color, "#"),
		Description: label.Description,
	}
}

// ToLabelList converts list of Label to API format
func ToLabelList(labels []*models.Label) (apiT []*api.Label) {
	result := make([]*api.Label, len(labels))
	for i := range labels {
		result[i] = ToLabel(labels[i])
	}
	return result
}
