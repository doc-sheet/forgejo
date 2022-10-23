// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migrations

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	base "code.gitea.io/gitea/modules/migration"

	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
)

func TestGitlabDownloadRepo(t *testing.T) {
	// Skip tests if Gitlab token is not found
	gitlabPersonalAccessToken := os.Getenv("GITLAB_READ_TOKEN")
	if gitlabPersonalAccessToken == "" {
		t.Skip("skipped test because GITLAB_READ_TOKEN was not in the environment")
	}

	resp, err := http.Get("https://gitlab.com/gitea/test_repo")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skipf("Can't access test repo, skipping %s", t.Name())
	}

	downloader, err := NewGitlabDownloader(context.Background(), "https://gitlab.com", "gitea/test_repo", "", "", gitlabPersonalAccessToken)
	if err != nil {
		t.Fatalf("NewGitlabDownloader is nil: %v", err)
	}
	repo, err := downloader.GetRepoInfo()
	assert.NoError(t, err)
	// Repo Owner is blank in Gitlab Group repos
	assertRepositoryEqual(t, &base.Repository{
		Name:          "test_repo",
		Owner:         "",
		Description:   "Test repository for testing migration from gitlab to gitea",
		CloneURL:      "https://gitlab.com/gitea/test_repo.git",
		OriginalURL:   "https://gitlab.com/gitea/test_repo",
		DefaultBranch: "master",
	}, repo)

	topics, err := downloader.GetTopics()
	assert.NoError(t, err)
	assert.True(t, len(topics) == 2)
	assert.EqualValues(t, []string{"migration", "test"}, topics)

	milestones, err := downloader.GetMilestones()
	assert.NoError(t, err)
	assertMilestonesEqual(t, []*base.Milestone{
		{
			Title:   "1.1.0",
			Created: time.Date(2019, 11, 28, 8, 42, 44, 575000000, time.UTC),
			Updated: timePtr(time.Date(2019, 11, 28, 8, 42, 44, 575000000, time.UTC)),
			State:   "active",
		},
		{
			Title:   "1.0.0",
			Created: time.Date(2019, 11, 28, 8, 42, 30, 301000000, time.UTC),
			Updated: timePtr(time.Date(2019, 11, 28, 15, 57, 52, 401000000, time.UTC)),
			Closed:  timePtr(time.Date(2019, 11, 28, 15, 57, 52, 401000000, time.UTC)),
			State:   "closed",
		},
	}, milestones)

	labels, err := downloader.GetLabels()
	assert.NoError(t, err)
	assertLabelsEqual(t, []*base.Label{
		{
			Name:  "bug",
			Color: "d9534f",
		},
		{
			Name:  "confirmed",
			Color: "d9534f",
		},
		{
			Name:  "critical",
			Color: "d9534f",
		},
		{
			Name:  "discussion",
			Color: "428bca",
		},
		{
			Name:  "documentation",
			Color: "f0ad4e",
		},
		{
			Name:  "duplicate",
			Color: "7f8c8d",
		},
		{
			Name:  "enhancement",
			Color: "5cb85c",
		},
		{
			Name:  "suggestion",
			Color: "428bca",
		},
		{
			Name:  "support",
			Color: "f0ad4e",
		},
	}, labels)

	releases, err := downloader.GetReleases()
	assert.NoError(t, err)
	assertReleasesEqual(t, []*base.Release{
		{
			TagName:         "v0.9.99",
			TargetCommitish: "0720a3ec57c1f843568298117b874319e7deee75",
			Name:            "First Release",
			Body:            "A test release",
			Created:         time.Date(2019, 11, 28, 9, 9, 48, 840000000, time.UTC),
			PublisherID:     1241334,
			PublisherName:   "lafriks",
		},
	}, releases)

	issues, isEnd, err := downloader.GetIssues(1, 2)
	assert.NoError(t, err)
	assert.False(t, isEnd)

	assertIssuesEqual(t, []*base.Issue{
		{
			Number:     1,
			Title:      "Please add an animated gif icon to the merge button",
			Content:    "I just want the merge button to hurt my eyes a little. :stuck_out_tongue_closed_eyes:",
			Milestone:  "1.0.0",
			PosterID:   1241334,
			PosterName: "lafriks",
			State:      "closed",
			Created:    time.Date(2019, 11, 28, 8, 43, 35, 459000000, time.UTC),
			Updated:    time.Date(2019, 11, 28, 8, 46, 23, 304000000, time.UTC),
			Labels: []*base.Label{
				{
					Name: "bug",
				},
				{
					Name: "discussion",
				},
			},
			Reactions: []*base.Reaction{
				{
					UserID:   1241334,
					UserName: "lafriks",
					Content:  "thumbsup",
				},
				{
					UserID:   1241334,
					UserName: "lafriks",
					Content:  "open_mouth",
				},
			},
			Closed: timePtr(time.Date(2019, 11, 28, 8, 46, 23, 275000000, time.UTC)),
		},
		{
			Number:     2,
			Title:      "Test issue",
			Content:    "This is test issue 2, do not touch!",
			Milestone:  "1.1.0",
			PosterID:   1241334,
			PosterName: "lafriks",
			State:      "closed",
			Created:    time.Date(2019, 11, 28, 8, 44, 46, 277000000, time.UTC),
			Updated:    time.Date(2019, 11, 28, 8, 45, 44, 987000000, time.UTC),
			Labels: []*base.Label{
				{
					Name: "duplicate",
				},
			},
			Reactions: []*base.Reaction{
				{
					UserID:   1241334,
					UserName: "lafriks",
					Content:  "thumbsup",
				},
				{
					UserID:   1241334,
					UserName: "lafriks",
					Content:  "thumbsdown",
				},
				{
					UserID:   1241334,
					UserName: "lafriks",
					Content:  "laughing",
				},
				{
					UserID:   1241334,
					UserName: "lafriks",
					Content:  "tada",
				},
				{
					UserID:   1241334,
					UserName: "lafriks",
					Content:  "confused",
				},
				{
					UserID:   1241334,
					UserName: "lafriks",
					Content:  "hearts",
				},
			},
			Closed: timePtr(time.Date(2019, 11, 28, 8, 45, 44, 959000000, time.UTC)),
		},
	}, issues)

	comments, _, err := downloader.GetComments(base.GetCommentOptions{
		Commentable: &base.Issue{
			Number:       2,
			ForeignIndex: 2,
			Context:      gitlabIssueContext{IsMergeRequest: false},
		},
	})
	assert.NoError(t, err)
	assertCommentsEqual(t, []*base.Comment{
		{
			IssueIndex: 2,
			PosterID:   1241334,
			PosterName: "lafriks",
			Created:    time.Date(2019, 11, 28, 8, 44, 52, 501000000, time.UTC),
			Content:    "This is a comment",
			Reactions:  nil,
		},
		{
			IssueIndex: 2,
			PosterID:   1241334,
			PosterName: "lafriks",
			Created:    time.Date(2019, 11, 28, 8, 45, 2, 329000000, time.UTC),
			Content:    "changed milestone to %2",
			Reactions:  nil,
		},
		{
			IssueIndex: 2,
			PosterID:   1241334,
			PosterName: "lafriks",
			Created:    time.Date(2019, 11, 28, 8, 45, 45, 7000000, time.UTC),
			Content:    "closed",
			Reactions:  nil,
		},
		{
			IssueIndex: 2,
			PosterID:   1241334,
			PosterName: "lafriks",
			Created:    time.Date(2019, 11, 28, 8, 45, 53, 501000000, time.UTC),
			Content:    "A second comment",
			Reactions:  nil,
		},
	}, comments)

	prs, _, err := downloader.GetPullRequests(1, 1)
	assert.NoError(t, err)
	assertPullRequestsEqual(t, []*base.PullRequest{
		{
			Number:     4,
			Title:      "Test branch",
			Content:    "do not merge this PR",
			Milestone:  "1.0.0",
			PosterID:   1241334,
			PosterName: "lafriks",
			State:      "opened",
			Created:    time.Date(2019, 11, 28, 15, 56, 54, 104000000, time.UTC),
			Labels: []*base.Label{
				{
					Name: "bug",
				},
			},
			Reactions: []*base.Reaction{{
				UserID:   4575606,
				UserName: "real6543",
				Content:  "thumbsup",
			}, {
				UserID:   4575606,
				UserName: "real6543",
				Content:  "tada",
			}},
			PatchURL: "https://gitlab.com/gitea/test_repo/-/merge_requests/2.patch",
			Head: base.PullRequestBranch{
				Ref:       "feat/test",
				CloneURL:  "https://gitlab.com/gitea/test_repo/-/merge_requests/2",
				SHA:       "9f733b96b98a4175276edf6a2e1231489c3bdd23",
				RepoName:  "test_repo",
				OwnerName: "lafriks",
			},
			Base: base.PullRequestBranch{
				Ref:       "master",
				SHA:       "",
				OwnerName: "lafriks",
				RepoName:  "test_repo",
			},
			Closed:         nil,
			Merged:         false,
			MergedTime:     nil,
			MergeCommitSHA: "",
			ForeignIndex:   2,
			Context:        gitlabIssueContext{IsMergeRequest: true},
		},
	}, prs)

	rvs, _, err := downloader.GetReviews(base.GetReviewOptions{
		Reviewable: &base.PullRequest{Number: 1, ForeignIndex: 1},
	})
	assert.NoError(t, err)
	assertReviewsEqual(t, []*base.Review{
		{
			IssueIndex:   1,
			ReviewerID:   4102996,
			ReviewerName: "zeripath",
			CreatedAt:    time.Date(2019, 11, 28, 16, 2, 8, 377000000, time.UTC),
			State:        "APPROVED",
		},
		{
			IssueIndex:   1,
			ReviewerID:   527793,
			ReviewerName: "axifive",
			CreatedAt:    time.Date(2019, 11, 28, 16, 2, 8, 377000000, time.UTC),
			State:        "APPROVED",
		},
	}, rvs)

	rvs, _, err = downloader.GetReviews(base.GetReviewOptions{
		Reviewable: &base.PullRequest{Number: 2, ForeignIndex: 2},
	})
	assert.NoError(t, err)
	assertReviewsEqual(t, []*base.Review{
		{
			IssueIndex:   2,
			ReviewerID:   4575606,
			ReviewerName: "real6543",
			CreatedAt:    time.Date(2020, 4, 19, 19, 24, 21, 108000000, time.UTC),
			State:        "APPROVED",
		},
	}, rvs)
}

func gitlabClientMockSetup(t *testing.T) (*http.ServeMux, *httptest.Server, *gitlab.Client) {
	// mux is the HTTP request multiplexer used with the test server.
	mux := http.NewServeMux()

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(mux)

	// client is the Gitlab client being tested.
	client, err := gitlab.NewClient("", gitlab.WithBaseURL(server.URL))
	if err != nil {
		server.Close()
		t.Fatalf("Failed to create client: %v", err)
	}

	return mux, server, client
}

func gitlabClientMockTeardown(server *httptest.Server) {
	server.Close()
}

type reviewTestCase struct {
	repoID, prID, reviewerID int
	reviewerName             string
	createdAt, updatedAt     *time.Time
	expectedCreatedAt        time.Time
}

func convertTestCase(t reviewTestCase) (func(w http.ResponseWriter, r *http.Request), base.Review) {
	var updatedAtField string
	if t.updatedAt == nil {
		updatedAtField = ""
	} else {
		updatedAtField = `"updated_at": "` + t.updatedAt.Format(time.RFC3339) + `",`
	}

	var createdAtField string
	if t.createdAt == nil {
		createdAtField = ""
	} else {
		createdAtField = `"created_at": "` + t.createdAt.Format(time.RFC3339) + `",`
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
{
  "id": 5,
  "iid": `+strconv.Itoa(t.prID)+`,
  "project_id": `+strconv.Itoa(t.repoID)+`,
  "title": "Approvals API",
  "description": "Test",
  "state": "opened",
  `+createdAtField+`
  `+updatedAtField+`
  "merge_status": "cannot_be_merged",
  "approvals_required": 2,
  "approvals_left": 1,
  "approved_by": [
    {
      "user": {
        "name": "Administrator",
        "username": "`+t.reviewerName+`",
        "id": `+strconv.Itoa(t.reviewerID)+`,
        "state": "active",
        "avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80\u0026d=identicon",
        "web_url": "http://localhost:3000/root"
      }
    }
  ]
}`)
	}
	review := base.Review{
		IssueIndex:   int64(t.prID),
		ReviewerID:   int64(t.reviewerID),
		ReviewerName: t.reviewerName,
		CreatedAt:    t.expectedCreatedAt,
		State:        "APPROVED",
	}

	return handler, review
}

func TestGitlabGetReviews(t *testing.T) {
	mux, server, client := gitlabClientMockSetup(t)
	defer gitlabClientMockTeardown(server)

	repoID := 1324

	downloader := &GitlabDownloader{
		ctx:    context.Background(),
		client: client,
		repoID: repoID,
	}

	createdAt := time.Date(2020, 4, 19, 19, 24, 21, 0, time.UTC)

	for _, testCase := range []reviewTestCase{
		{
			repoID:            repoID,
			prID:              1,
			reviewerID:        801,
			reviewerName:      "someone1",
			createdAt:         nil,
			updatedAt:         &createdAt,
			expectedCreatedAt: createdAt,
		},
		{
			repoID:            repoID,
			prID:              2,
			reviewerID:        802,
			reviewerName:      "someone2",
			createdAt:         &createdAt,
			updatedAt:         nil,
			expectedCreatedAt: createdAt,
		},
		{
			repoID:            repoID,
			prID:              3,
			reviewerID:        803,
			reviewerName:      "someone3",
			createdAt:         nil,
			updatedAt:         nil,
			expectedCreatedAt: time.Now(),
		},
	} {
		mock, review := convertTestCase(testCase)
		mux.HandleFunc(fmt.Sprintf("/api/v4/projects/%d/merge_requests/%d/approvals", testCase.repoID, testCase.prID), mock)

		id := int64(testCase.prID)
		rvs, _, err := downloader.GetReviews(base.GetReviewOptions{
			Reviewable: &base.Issue{Number: id, ForeignIndex: id},
		})
		assert.NoError(t, err)
		assertReviewsEqual(t, []*base.Review{&review}, rvs)
	}
}
