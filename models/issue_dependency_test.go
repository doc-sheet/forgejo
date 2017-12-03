package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateIssueDependency(t *testing.T) {
	// Prepare
	assert.NoError(t, PrepareTestDatabase())

	user1, err := GetUserByID(1)
	assert.NoError(t, err)

	issue1, err := GetIssueByID(1)
	assert.NoError(t, err)
	issue2, err := GetIssueByID(2)
	assert.NoError(t, err)

	// Create a dependency and check if it was successfull
	exists, circular, err := CreateIssueDependency(user1, issue1, issue2)
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.False(t, circular)

	// Do it again to see if it will check if the dependency already exists
	exists, _, err = CreateIssueDependency(user1, issue1, issue2)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check for circular dependencies
	_, circular, err = CreateIssueDependency(user1, issue2, issue1)
	assert.NoError(t, err)
	assert.True(t, exists)

	_ = AssertExistsAndLoadBean(t, &Comment{Type: CommentTypeAddDependency, PosterID: user1.ID, IssueID: issue1.ID})

	// Check if dependencies left is correct
	left, err := IssueNoDependenciesLeft(issue1)
	assert.NoError(t, err)
	assert.False(t, left)

	// Close #2 and check again
	err = issue2.ChangeStatus(user1, issue2.Repo, true)
	assert.NoError(t, err)

	left, err = IssueNoDependenciesLeft(issue1)
	assert.NoError(t, err)
	assert.True(t, left)

	// Test removing the depencency
	err = RemoveIssueDependency(user1, issue1, issue2, DependencyTypeBlockedBy)
	assert.NoError(t, err)
}
