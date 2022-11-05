// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package structs

import "time"

// swagger:model
type NewProjectPayload struct {
	// required:true
	Title string `json:"title" binding:"Required"`
	// required:true
	BoardType   uint8  `json:"board_type"`
	Description string `json:"description"`
}

// swagger:model
type UpdateProjectPayload struct {
	// required:true
	Title       string `json:"title" binding:"Required"`
	Description string `json:"description"`
}

type Project struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	BoardType   uint8  `json:"board_type"`
	IsClosed    bool   `json:"is_closed"`
	// swagger:strfmt date-time
	Created time.Time `json:"created_at"`
	// swagger:strfmt date-time
	Updated time.Time `json:"updated_at"`
	// swagger:strfmt date-time
	Closed time.Time `json:"closed_at"`

	Repo    *RepositoryMeta `json:"repository"`
	Creator *User           `json:"creator"`
}

type ProjectBoard struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title"`
	Default bool     `json:"default"`
	Color   string   `json:"color"`
	Sorting int8     `json:"sorting"`
	Project *Project `json:"project"`
	Creator *User    `json:"creator"`
	// swagger:strfmt date-time
	Created time.Time `json:"created_at"`
	// swagger:strfmt date-time
	Updated time.Time `json:"updated_at"`
}

// swagger:model
type NewProjectBoardPayload struct {
	// required:true
	Title   string `json:"title"`
	Default bool   `json:"default"`
	Color   string `json:"color"`
	Sorting int8   `json:"sorting"`
}

// swagger:model
type UpdateProjectBoardPayload struct {
	Title string `json:"title"`
	Color string `json:"color"`
}
