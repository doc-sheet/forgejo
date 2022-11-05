// Copyright 2016 The Gogs Authors. All rights reserved.
// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"net/http"

	project_model "code.gitea.io/gitea/models/project"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/convert"
	"code.gitea.io/gitea/modules/setting"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/modules/web"
)

func GetProject(ctx *context.APIContext) {
	// swagger:operation GET /projects/{id} project projectGetProject
	// ---
	// summary: Get project
	// produces:
	// - application/json
	// parameters:
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/Project"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"
	project, err := project_model.GetProjectByID(ctx, ctx.ParamsInt64(":id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.NotFound()
		} else {
			ctx.Error(http.StatusInternalServerError, "GetProjectByID", err)
		}
		return
	}
	ctx.JSON(http.StatusOK, convert.ToAPIProject(project))
}

func UpdateProject(ctx *context.APIContext) {
	// swagger:operation PATCH /projects/{id} project projectUpdateProject
	// ---
	// summary: Update project
	// produces:
	// - application/json
	// consumes:
	// - application/json
	// parameters:
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: string
	//   required: true
	// - name: project
	//   in: body
	//   required: true
	//   schema: { "$ref": "#/definitions/UpdateProjectPayload" }
	// responses:
	//   "200":
	//     "$ref": "#/responses/Project"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"
	form := web.GetForm(ctx).(*api.UpdateProjectPayload)
	project, err := project_model.GetProjectByID(ctx, ctx.ParamsInt64(":id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.NotFound()
		} else {
			ctx.Error(http.StatusInternalServerError, "GetProjectByID", err)
		}
		return
	}
	if form.Title != project.Title {
		project.Title = form.Title
	}
	if form.Description != project.Description {
		project.Description = form.Description
	}
	err = project_model.UpdateProject(ctx, project)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "UpdateProject", err)
		return
	}

	ctx.JSON(http.StatusOK, project)
}

func DeleteProject(ctx *context.APIContext) {
	// swagger:operation DELETE /projects/{id} project projectDeleteProject
	// ---
	// summary: Delete project
	// parameters:
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: string
	//   required: true
	// responses:
	//   "204":
	//     "description": "Deleted the project"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"
	err := project_model.DeleteProjectByIDCtx(ctx, ctx.ParamsInt64(":id"))
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "DeleteProject", err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func CreateRepositoryProject(ctx *context.APIContext) {
	// swagger:operation POST /repos/{owner}/{repo}/projects project projectCreateRepositoryProject
	// ---
	// summary: Create a repository project
	// produces:
	// - application/json
	// consumes:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: repo
	//   type: string
	//   required: true
	// - name: project
	//   in: body
	//   required: true
	//   schema: { "$ref": "#/definitions/NewProjectPayload" }
	// responses:
	//   "201":
	//     "$ref": "#/responses/Project"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"
	form := web.GetForm(ctx).(*api.NewProjectPayload)
	project := &project_model.Project{
		RepoID:      ctx.Repo.Repository.ID,
		Title:       form.Title,
		Description: form.Description,
		CreatorID:   ctx.Doer.ID,
		BoardType:   project_model.BoardType(form.BoardType),
		Type:        project_model.TypeRepository,
	}

	var err error
	if err = project_model.NewProject(project); err != nil {
		ctx.Error(http.StatusInternalServerError, "NewProject", err)
		return
	}

	project, err = project_model.GetProjectByID(ctx, project.ID)

	if err != nil {
		ctx.Error(http.StatusInternalServerError, "GetProjectByID", err)
		return
	}

	ctx.JSON(http.StatusCreated, convert.ToAPIProject(project))
}

func ListRepositoryProjects(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/projects project projectListRepositoryProjects
	// ---
	// summary: List repository projects
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repository
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: repo
	//   type: string
	//   required: true
	// - name: closed
	//   in: query
	//   description: include closed issues or not
	//   type: boolean
	// - name: page
	//   in: query
	//   description: page number of results to return (1-based)
	//   type: integer
	// - name: limit
	//   in: query
	//   description: page size of results
	//   type: integer
	// responses:
	//   "200":
	//     "$ref": "#/responses/ProjectList"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"
	fmt.Print(ctx.FormOptionalBool("closed"))
	projects, count, err := project_model.GetProjects(ctx, project_model.SearchOptions{
		RepoID:   ctx.Repo.Repository.ID,
		Page:     ctx.FormInt("page"),
		IsClosed: ctx.FormOptionalBool("closed"),
		Type:     project_model.TypeRepository,
	})
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "Projects", err)
		return
	}

	ctx.SetLinkHeader(int(count), setting.UI.IssuePagingNum)
	ctx.SetTotalCountHeader(count)

	apiProjects, err := convert.ToAPIProjectList(projects)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}

	ctx.JSON(http.StatusOK, apiProjects)
}
