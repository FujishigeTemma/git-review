package commands

import (
	"context"

	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/newmo-oss/ergo"
)

type AbortCmd struct{}

func (c *AbortCmd) Run(g *git.Git, repo *repository.Repository, out *output.Output) error {
	if err := requireMainWorktree(g); err != nil {
		return err
	}
	if err := requireActive(repo); err != nil {
		return err
	}

	ctx := context.Background()
	q := repo.Queries()

	session, err := q.GetSession(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to get session")
	}

	cleanupReview(g, repo, out, session)
	out.Ok("Review aborted. Back on: " + session.Branch)

	return nil
}
