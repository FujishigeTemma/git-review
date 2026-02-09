package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/FujishigeTemma/git-review/internal/db"
	"github.com/FujishigeTemma/git-review/internal"
	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/newmo-oss/ergo"
)

type FinishCmd struct{}

func (c *FinishCmd) Run(g *git.Git, repo *repository.Repository, out *output.Output) error {
	if err := requireMainWorktree(g); err != nil {
		return err
	}
	if err := requireActive(repo); err != nil {
		return err
	}
	return finishReview(g, repo, out)
}

func finishReview(g *git.Git, repo *repository.Repository, out *output.Output) error {
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

	total := len(commits)
	nComments := len(comments)

	// Write comments to git notes on original commits
	childrenMap := buildChildrenMap(comments)
	for _, cm := range commits {
		if note := buildCommitNotes(comments, childrenMap, cm.Sha); note != "" {
			if err := g.NotesAppend(cm.Sha, note); err != nil {
				out.Warn(fmt.Sprintf("failed to write notes for %s: %v", internal.ShortSHA(cm.Sha), err))
			}
		}
	}

	cleanupReview(g, repo, out, session)

	out.Printf("\n")
	out.Ok("══ Review Complete ══")
	out.Printf("\n")
	out.Info(fmt.Sprintf("  Comments : %d across %d commits", nComments, total))
	out.Info(fmt.Sprintf("  Back on  : %s", session.Branch))
	out.Printf("\n")
	out.Printf("  Comments written to git notes on original commits.\n")

	return nil
}

// buildCommitNotes builds a git notes string for all comments on a given commit SHA.
func buildCommitNotes(allComments []db.Comment, childrenMap map[string][]db.Comment, commitSHA string) string {
	// Collect top-level comments for this commit
	var topLevel []db.Comment
	for _, c := range allComments {
		if c.Commit == commitSHA && !c.ParentID.Valid {
			topLevel = append(topLevel, c)
		}
	}

	var notes []string
	for _, c := range topLevel {
		authorTag := authorSuffix(c.CreatedBy)
		if c.File.Valid {
			loc := c.File.String
			if lr := internal.FormatLineRange(c.StartLine, c.EndLine); lr != "" {
				loc += ":" + lr
			}
			notes = append(notes, fmt.Sprintf("%s -- %s%s", loc, c.Body, authorTag))
		} else {
			notes = append(notes, fmt.Sprintf("%s%s", c.Body, authorTag))
		}
		for _, r := range descendants(childrenMap, c.ID) {
			rAuthorTag := authorSuffix(r.CreatedBy)
			commitTag := ""
			if r.Commit != commitSHA {
				commitTag = "(" + internal.ShortSHA(r.Commit) + ") "
			}
			notes = append(notes, fmt.Sprintf("  %s%s%s", commitTag, r.Body, rAuthorTag))
		}
	}
	return strings.Join(notes, "\n")
}
