// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package doctor

import (
	"context"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/models/migrations"
	"code.gitea.io/gitea/modules/log"
)

func checkDBConsistency(logger log.Logger, autofix bool) error {
	// make sure DB version is uptodate
	if err := models.NewEngine(context.Background(), migrations.EnsureUpToDate); err != nil {
		logger.Critical("Model version on the database does not match the current Gitea version. Model consistency will not be checked until the database is upgraded")
		return err
	}

	// find labels without existing repo or org
	count, err := models.CountOrphanedLabels()
	if err != nil {
		logger.Critical("Error: %v whilst counting orphaned labels")
		return err
	}

	if count > 0 {
		if autofix {
			if err = models.DeleteOrphanedLabels(); err != nil {
				logger.Critical("Error: %v whilst deleting orphaned labels")
				return err
			}
			logger.Info("%d labels without existing repository/organisation deleted", count)
		} else {
			logger.Warn("%d labels without existing repository/organisation", count)
		}
	}

	// find issues without existing repository
	count, err = models.CountOrphanedIssues()
	if err != nil {
		logger.Critical("Error: %v whilst counting orphaned issues")
		return err
	}
	if count > 0 {
		if autofix {
			if err = models.DeleteOrphanedIssues(); err != nil {
				logger.Critical("Error: %v whilst deleting orphaned issues")
				return err
			}
			logger.Info("%d issues without existing repository deleted", count)
		} else {
			logger.Warn("%d issues without existing repository", count)
		}
	}

	// find pulls without existing issues
	count, err = models.CountOrphanedObjects("pull_request", "issue", "pull_request.issue_id=issue.id")
	if err != nil {
		logger.Critical("Error: %v whilst counting orphaned objects")
		return err
	}
	if count > 0 {
		if autofix {
			if err = models.DeleteOrphanedObjects("pull_request", "issue", "pull_request.issue_id=issue.id"); err != nil {
				logger.Critical("Error: %v whilst deleting orphaned objects")
				return err
			}
			logger.Info("%d pull requests without existing issue deleted", count)
		} else {
			logger.Warn("%d pull requests without existing issue", count)
		}
	}

	// find tracked times without existing issues/pulls
	count, err = models.CountOrphanedObjects("tracked_time", "issue", "tracked_time.issue_id=issue.id")
	if err != nil {
		logger.Critical("Error: %v whilst counting orphaned objects")
		return err
	}
	if count > 0 {
		if autofix {
			if err = models.DeleteOrphanedObjects("tracked_time", "issue", "tracked_time.issue_id=issue.id"); err != nil {
				logger.Critical("Error: %v whilst deleting orphaned objects")
				return err
			}
			logger.Info("%d tracked times without existing issue deleted", count)
		} else {
			logger.Warn("%d tracked times without existing issue", count)
		}
	}

	// find null archived repositories
	count, err = models.CountNullArchivedRepository()
	if err != nil {
		logger.Critical("Error: %v whilst counting null archived repositories")
		return err
	}
	if count > 0 {
		if autofix {
			updatedCount, err := models.FixNullArchivedRepository()
			if err != nil {
				logger.Critical("Error: %v whilst fixing null archived repositories")
				return err
			}
			logger.Info("%d repositories with null is_archived updated", updatedCount)
		} else {
			logger.Warn("%d repositories with null is_archived", count)
		}
	}

	// TODO: function to recalc all counters

	return nil
}

func init() {
	Register(&Check{
		Title:     "Check consistency of database",
		Name:      "check-db-consistency",
		IsDefault: false,
		Run:       checkDBConsistency,
		Priority:  3,
	})
}
