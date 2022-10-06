// Copyright 2019 The Gitea Authors. All rights reserved.
// Copyright 2018 Jonas Franz. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migration

// LabelPriority represents the priority type of a label
type LabelPriority string

const (
	// LabelPriorityLow represents the low priority of a label
	LabelPriorityLow LabelPriority = "low"
	// LabelPriorityMedium represents the medium priority of a label
	LabelPriorityMedium LabelPriority = "medium"
	// LabelPriorityHigh represents the high priority of a label
	LabelPriorityHigh LabelPriority = "high"
	// LabelPriorityCritical represents the critical priority of a label
	LabelPriorityCritical LabelPriority = "critical"
)

// Label defines a standard label information
type Label struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Priority    LabelPriority
	Description string `json:"description"`
}
