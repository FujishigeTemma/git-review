package commands

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/newmo-oss/ergo"
)

type JumpCmd struct {
	Hash string `arg:"" help:"Commit hash (or prefix) to jump to."`
}

func (c *JumpCmd) Run(g *git.Git, repo *repository.Repository, out *output.Output) error {
	if err := requireActive(repo); err != nil {
		return err
	}

	ctx := context.Background()
	q := repo.Queries()

	target, err := q.FindCommitBySHAPrefix(ctx, sql.NullString{String: c.Hash, Valid: true})
	if err != nil {
		return ergo.New("commit not found", slog.String("hash", c.Hash))
	}

	if err := jumpTo(g, repo, g.Reviewer, target); err != nil {
		return err
	}

	commits, err := q.ListCommits(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to list commits")
	}

	oneline, _ := g.Oneline(target.Sha)
	stat, _ := g.DiffStagedStat()
	out.Printf("\n")
	out.Printf("  %s [%d/%d] %s\n", out.Bold("→"), target.Position+1, int64(len(commits)), oneline)
	if stat != "" {
		out.Printf("\n%s\n", stat)
	}

	return nil
}
