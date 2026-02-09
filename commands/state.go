package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FujishigeTemma/git-review/internal/db"
	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/guregu/null/v6"
	"github.com/newmo-oss/ergo"
)

type StateCmd struct{}

type stateOutput struct {
	BaseRef  string         `json:"baseRef"`
	Branch   string         `json:"branch"`
	Commits  []string       `json:"commits"`
	Current  null.Int       `json:"current"`
	Comments []stateComment `json:"comments"`
}

type stateComment struct {
	ID         string      `json:"id"`
	ParentID   null.String `json:"parentId"`
	Commit     string      `json:"commit"`
	File       null.String `json:"file"`
	StartLine  null.Int    `json:"startLine"`
	EndLine    null.Int    `json:"endLine"`
	Body       string      `json:"body"`
	ResolvedAt null.String `json:"resolvedAt"`
	ResolvedBy null.String `json:"resolvedBy"`
	CreatedAt  string      `json:"createdAt"`
	CreatedBy  string      `json:"createdBy"`
}

func (c *StateCmd) Run(g *git.Git, repo *repository.Repository, out *output.Output) error {
	if repo == nil {
		fmt.Fprintln(out.Stdout, "null")
		return nil
	}

	ctx := context.Background()

	count, err := repo.Queries().SessionExists(ctx)
	if err != nil || count == 0 {
		fmt.Fprintln(out.Stdout, "null")
		return nil
	}

	q := repo.Queries()

	session, err := q.GetSession(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to get session")
	}

	commits, err := q.ListCommits(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to list commits")
	}

	commitSHAs := make([]string, len(commits))
	for i, c := range commits {
		commitSHAs[i] = c.Sha
	}

	// Determine current position from worktree reviewer
	var current null.Int
	reviewer, err := q.GetReviewer(ctx, g.Reviewer)
	if err == nil && reviewer.CurrentSha.Valid {
		if pos := findCommitPosition(commits, reviewer.CurrentSha.String); pos >= 0 {
			current = null.IntFrom(pos)
		}
	}

	comments, err := q.ListAllComments(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to list comments")
	}

	stateComments := make([]stateComment, len(comments))
	for i, c := range comments {
		stateComments[i] = toStateComment(c)
	}

	s := stateOutput{
		BaseRef:  session.BaseRef,
		Branch:   session.Branch,
		Commits:  commitSHAs,
		Current:  current,
		Comments: stateComments,
	}

	enc := json.NewEncoder(out.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func toStateComment(c db.Comment) stateComment {
	sc := stateComment{
		ID:         c.ID.String(),
		Commit:     c.Commit,
		File:       c.File,
		StartLine:  c.StartLine,
		EndLine:    c.EndLine,
		Body:       c.Body,
		ResolvedAt: c.ResolvedAt,
		ResolvedBy: c.ResolvedBy,
		CreatedAt:  c.CreatedAt,
		CreatedBy:  c.CreatedBy,
	}
	if c.ParentID.Valid {
		sc.ParentID = null.StringFrom(c.ParentID.UUID.String())
	}
	return sc
}
