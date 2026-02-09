package commands

import (
	"context"
	"fmt"

	"github.com/FujishigeTemma/git-review/internal"
	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/newmo-oss/ergo"
)

type StatusCmd struct{}

func (c *StatusCmd) Run(g *git.Git, repo *repository.Repository, out *output.Output) error {
	if err := requireActive(repo); err != nil {
		return err
	}
	return showStatus(g, repo, out)
}

func showStatus(g *git.Git, repo *repository.Repository, out *output.Output) error {
	ctx := context.Background()
	q := repo.Queries()

	session, err := q.GetSession(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to get session")
	}

	commits, err := q.ListCommits(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to list commits")
	}

	comments, err := q.ListAllComments(ctx)
	if err != nil {
		out.Warn(fmt.Sprintf("failed to load comments: %v", err))
	}

	reviewers, err := q.ListReviewers(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to list reviewers")
	}

	out.Printf("\n")
	out.Printf("%s  %s\n", out.Bold("Review Progress"), session.Branch)
	out.Printf("\n")

	// Show per-reviewer progress if multiple reviewers
	if len(reviewers) > 1 {
		for _, r := range reviewers {
			name := r.Name
			if name == "" {
				name = "(default)"
			}
			pos := "not started"
			if r.CurrentSha.Valid {
				if p := findCommitPosition(commits, r.CurrentSha.String); p >= 0 {
					pos = fmt.Sprintf("%d/%d", p+1, len(commits))
				}
			}
			out.Printf("  Reviewer %s: %s\n", name, pos)
		}
		out.Printf("\n")
	}

	// Build a map of commit SHA -> comment count
	commentCount := map[string]int{}
	for _, c := range comments {
		commentCount[c.Commit]++
	}

	// Determine current reviewer position for display
	var currentPos int64 = -1
	for _, r := range reviewers {
		if r.Name == g.Reviewer && r.CurrentSha.Valid {
			currentPos = findCommitPosition(commits, r.CurrentSha.String)
			break
		}
	}

	for _, cm := range commits {
		oneline, _ := g.Oneline(cm.Sha)

		badge := ""
		n := commentCount[cm.Sha]
		if n > 0 {
			badge = fmt.Sprintf(" (%d %s)", n, internal.Pluralize(n, "comment", "comments"))
		}

		line := fmt.Sprintf("%d. %s%s", cm.Position+1, oneline, badge)

		if cm.Position < currentPos {
			out.Printf("  %s %s\n", out.Green("✓"), out.Green(line))
		} else if cm.Position == currentPos {
			out.Printf("  %s %s\n", out.Yellow("→"), out.Yellow(line))
		} else {
			out.Printf("  ○ %s\n", line)
		}
	}
	out.Printf("\n")

	return nil
}
