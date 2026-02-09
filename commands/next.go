package commands

import (
	"context"

	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/newmo-oss/ergo"
)

type NextCmd struct{}

func (c *NextCmd) Run(g *git.Git, repo *repository.Repository, out *output.Output) error {
	if err := requireActive(repo); err != nil {
		return err
	}

	ctx := context.Background()
	q := repo.Queries()

	reviewer, err := q.GetReviewer(ctx, g.Reviewer)
	if err != nil {
		return ergo.Wrap(err, "failed to get reviewer")
	}

	commits, err := q.ListCommits(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to list commits")
	}

	total := len(commits)

	// Determine next commit
	var nextIdx int64
	if !reviewer.CurrentSha.Valid {
		nextIdx = 0
	} else {
		pos := findCommitPosition(commits, reviewer.CurrentSha.String)
		if pos < 0 {
			return ergo.New("current commit not found in commit list")
		}
		nextIdx = pos + 1
	}

	if nextIdx >= int64(total) {
		out.Printf("\n")
		out.Ok("All commits reviewed.")
		out.Printf("\n")
		out.Printf("  git review finish    Complete the review\n")
		out.Printf("  git review list      View all comments\n")
		return nil
	}

	target := commits[nextIdx]
	if err := jumpTo(g, repo, g.Reviewer, target); err != nil {
		return err
	}

	oneline, _ := g.Oneline(target.Sha)
	stat, _ := g.DiffStagedStat()
	out.Printf("\n")
	out.Printf("  %s [%d/%d] %s\n", out.Bold("→"), nextIdx+1, total, oneline)
	if stat != "" {
		out.Printf("\n%s\n", stat)
	}

	return nil
}
