// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"fmt"

	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/services/f3/util"

	"github.com/urfave/cli"
	"lab.forgefriends.org/friendlyforgeformat/gof3"
	f3_common "lab.forgefriends.org/friendlyforgeformat/gof3/forges/common"
	f3_format "lab.forgefriends.org/friendlyforgeformat/gof3/format"
)

var CmdF3 = cli.Command{
	Name:        "f3",
	Usage:       "Friendly Forge Format (F3) format export/import.",
	Description: "Import or export a repository from or to the Friendly Forge Format (F3) format.",
	Action:      runF3,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "directory",
			Value: "./f3",
			Usage: "Path of the directory where the F3 dump is stored",
		},
		cli.StringFlag{
			Name:  "user",
			Value: "",
			Usage: "The name of the user who owns the repository",
		},
		cli.StringFlag{
			Name:  "repository",
			Value: "",
			Usage: "The name of the repository",
		},
		cli.BoolFlag{
			Name:  "no-pull-request",
			Usage: "Do not dump pull requests",
		},
		cli.BoolFlag{
			Name:  "import",
			Usage: "Import from the directory",
		},
		cli.BoolFlag{
			Name:  "export",
			Usage: "Export to the directory",
		},
	},
}

func runF3(ctx *cli.Context) error {
	stdCtx, cancel := installSignals()
	defer cancel()

	if err := initDB(stdCtx); err != nil {
		return err
	}

	if err := git.InitSimple(stdCtx); err != nil {
		return err
	}

	return RunF3(stdCtx, ctx)
}

func RunF3(stdCtx context.Context, ctx *cli.Context) error {
	doer, err := user_model.GetAdminUser()
	if err != nil {
		return err
	}

	features := gof3.AllFeatures
	if ctx.Bool("no-pull-request") {
		features.PullRequests = false
	}

	forgejo := util.ForgejoForgeRoot(features, doer)
	f3 := util.F3ForgeRoot(features, ctx.String("directory"))

	if ctx.Bool("export") {
		forgejo.Forge.Users.List(stdCtx)
		user := forgejo.Forge.Users.GetFromFormat(stdCtx, &f3_format.User{UserName: ctx.String("user")})
		if user.IsNil() {
			return fmt.Errorf("%s is not a known user", ctx.String("user"))
		}

		user.Projects.List(stdCtx)
		project := user.Projects.GetFromFormat(stdCtx, &f3_format.Project{Name: ctx.String("repository")})
		if project.IsNil() {
			return fmt.Errorf("%s/%s is not a known repository", ctx.String("user"), ctx.String("repository"))
		}

		options := f3_common.NewMirrorOptionsRecurse(user, project)
		f3.Forge.Mirror(stdCtx, forgejo.Forge, options)
		fmt.Println("exported")
	} else if ctx.Bool("import") {
		options := f3_common.NewMirrorOptionsRecurse()
		forgejo.Forge.Mirror(stdCtx, f3.Forge, options)
		fmt.Println("imported")
	} else {
		return fmt.Errorf("either --import or --export must be specified")
	}

	return nil
}
