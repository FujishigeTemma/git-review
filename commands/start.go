package commands

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/FujishigeTemma/git-review/internal"
	"github.com/FujishigeTemma/git-review/internal/db"
	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/newmo-oss/ergo"
)

type StartCmd struct {
	Base string `arg:"" optional:"" help:"Base ref to review from (auto-detects if omitted)."`
	Name string `short:"a" help:"Reviewer role name."`
}

func (c *StartCmd) Run(g *git.Git, repo *repository.Repository, out *output.Output) error {
	ctx := context.Background()

	// Check if a session already exists
	count, err := repo.Queries().SessionExists(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to check session")
	}
	if count > 0 {
		if c.Name != "" {
			return c.joinExistingSession(g, repo, out)
		}
		if c.Base != "" {
			return ergo.WithCode(
				ergo.New("Review already in progress. Finish or abort first."),
				internal.ErrCodeReviewActive)
		}
		return showStatus(g, repo, out)
	}

	currentBranch, err := g.CurrentBranch()
	if err != nil || currentBranch == "" {
		return ergo.WithCode(
			ergo.New("Detached HEAD. Checkout a branch first."),
			internal.ErrCodeDetachedHead)
	}

	// Detect base
	var base string
	if c.Base != "" {
		base, err = g.Run("rev-parse", c.Base)
		if err != nil {
			return ergo.WithCode(
				ergo.New("invalid ref", slog.String("ref", c.Base)),
				internal.ErrCodeInvalidRef)
		}
	} else {
		for _, ref := range []string{"main", "master", "develop"} {
			if g.RefExists(ref) {
				base, err = g.MergeBase(ref, "HEAD")
				if err != nil {
					continue
				}
				oneline, _ := g.Oneline(base)
				out.Info(fmt.Sprintf("Base: %s (%s)", ref, oneline))
				break
			}
		}
		if base == "" {
			return ergo.WithCode(
				ergo.New("Cannot detect base branch. Specify: git review <base-ref>"),
				internal.ErrCodeInvalidRef)
		}
	}

	commits, err := g.RevList(base + "..HEAD")
	if err != nil || len(commits) == 0 {
		return ergo.WithCode(
			ergo.New("No commits to review between base and HEAD."),
			internal.ErrCodeNoCommits)
	}

	nCommits := len(commits)

	reviewerName := c.Name
	if reviewerName == "" {
		reviewerName = g.Reviewer
	}

	// Insert session, commits, and reviewer in a transaction
	if err := repo.WithTx(ctx, func(q *db.Queries) error {
		if err := q.InsertSession(ctx, db.InsertSessionParams{
			BaseRef:   base,
			Branch:    currentBranch,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}); err != nil {
			return ergo.Wrap(err, "failed to insert session")
		}

		for i, sha := range commits {
			msg, _ := g.Subject(sha)
			if err := q.InsertCommit(ctx, db.InsertCommitParams{
				Sha:      sha,
				Message:  msg,
				Position: int64(i),
			}); err != nil {
				return ergo.Wrap(err, "failed to insert commit",
					slog.String("sha", sha))
			}
		}

		if err := q.InsertReviewer(ctx, db.InsertReviewerParams{
			Name: reviewerName,
		}); err != nil {
			return ergo.Wrap(err, "failed to insert reviewer",
				slog.String("name", reviewerName))
		}
		return nil
	}); err != nil {
		return ergo.Wrap(err, "failed to initialize review")
	}

	// If -a is specified, create a worktree and jumpTo from there
	jumpGit := g
	if c.Name != "" {
		worktreePath := filepath.Join(g.CommonDir, "review", "worktrees", c.Name)
		if err := g.WorktreeAdd(worktreePath); err != nil {
			return ergo.Wrap(err, "failed to create worktree")
		}
		jumpGit = g.ForWorktree(c.Name, worktreePath)
	}

	// Jump to the first commit so that `add` works immediately after `start`
	firstCommit, err := repo.Queries().GetCommitByPosition(ctx, 0)
	if err != nil {
		return ergo.Wrap(err, "failed to get first commit")
	}
	if err := jumpTo(jumpGit, repo, reviewerName, firstCommit); err != nil {
		return ergo.Wrap(err, "failed to jump to first commit")
	}

	oneline, _ := g.Oneline(commits[0])
	out.Printf("\n")
	out.Ok(fmt.Sprintf("══ Review Started: %d commit(s) ══", nCommits))
	out.Printf("\n")
	out.Printf("  %s [1/%d] %s\n", out.Bold("→"), nCommits, oneline)
	out.Printf("\n")
	out.Printf("  Staged changes are ready for review.\n")
	out.Printf("\n")
	out.Printf("    git review add 'message'                Add comment\n")
	out.Printf("    git review add -f file -l N 'message'   Add comment on file:line\n")
	out.Printf("    git review next                         Next commit\n")

	return nil
}

// joinExistingSession adds a new reviewer to an existing session and creates a worktree.
func (c *StartCmd) joinExistingSession(g *git.Git, repo *repository.Repository, out *output.Output) error {
	ctx := context.Background()
	q := repo.Queries()

	// Insert the new reviewer
	if err := q.InsertReviewer(ctx, db.InsertReviewerParams{
		Name: c.Name,
	}); err != nil {
		return ergo.Wrap(err, "failed to add reviewer")
	}

	// Create worktree
	worktreePath := filepath.Join(g.CommonDir, "review", "worktrees", c.Name)
	if err := g.WorktreeAdd(worktreePath); err != nil {
		return ergo.Wrap(err, "failed to create worktree")
	}
	jumpGit := g.ForWorktree(c.Name, worktreePath)

	// Jump to the first commit
	firstCommit, err := q.GetCommitByPosition(ctx, 0)
	if err != nil {
		return ergo.Wrap(err, "failed to get first commit")
	}
	if err := jumpTo(jumpGit, repo, c.Name, firstCommit); err != nil {
		return ergo.Wrap(err, "failed to jump to first commit")
	}

	commits, err := q.ListCommits(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to list commits")
	}

	oneline, _ := g.Oneline(firstCommit.Sha)
	out.Printf("\n")
	out.Ok(fmt.Sprintf("══ Joined Review as %s: %d commit(s) ══", c.Name, len(commits)))
	out.Printf("\n")
	out.Printf("  %s [1/%d] %s\n", out.Bold("→"), len(commits), oneline)
	out.Printf("\n")
	out.Printf("  Worktree: %s\n", worktreePath)

	return nil
}
