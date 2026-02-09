package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/FujishigeTemma/git-review/internal"
	"github.com/FujishigeTemma/git-review/internal/db"
	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/guregu/null/v6"
	"github.com/newmo-oss/ergo"
)

// requireMainWorktree checks that the command is not running from a linked worktree.
// Commands like finish and abort must run from the main worktree (original repo).
func requireMainWorktree(g *git.Git) error {
	if g.Reviewer != "" {
		return ergo.WithCode(
			ergo.New("This command must be run from the main worktree, not from a reviewer worktree.\n  cd to the original repository and retry."),
			internal.ErrCodeWrongWorktree)
	}
	return nil
}

// requireActive checks that a review session exists.
func requireActive(repo *repository.Repository) error {
	count, err := repo.Queries().SessionExists(context.Background())
	if err != nil {
		return ergo.Wrap(err, "failed to check session")
	}
	if count == 0 {
		return ergo.WithCode(
			ergo.New("No review in progress. Start with: git review"),
			internal.ErrCodeNoReview)
	}
	return nil
}

// jumpTo performs the checkout-parent + read-tree-target dance and updates the reviewer position.
func jumpTo(g *git.Git, repo *repository.Repository, reviewerName string, target db.Commit) error {
	ctx := context.Background()
	q := repo.Queries()

	// Determine parent: if position==0, use session.base_ref; else commits[position-1]
	var parentRef string
	if target.Position == 0 {
		session, err := q.GetSession(ctx)
		if err != nil {
			return ergo.Wrap(err, "failed to get session")
		}
		parentRef = session.BaseRef
	} else {
		parent, err := q.GetCommitByPosition(ctx, target.Position-1)
		if err != nil {
			return ergo.Wrap(err, "failed to get parent commit")
		}
		parentRef = parent.Sha
	}

	if err := g.Checkout(parentRef); err != nil {
		return ergo.Wrap(err, "failed to checkout parent")
	}
	if err := g.ReadTreeReset(target.Sha); err != nil {
		return ergo.Wrap(err, "failed to read-tree target")
	}

	if err := q.UpdateReviewerCurrent(ctx, db.UpdateReviewerCurrentParams{
		CurrentSha: null.StringFrom(target.Sha),
		Name:       reviewerName,
	}); err != nil {
		return ergo.Wrap(err, "failed to update reviewer position")
	}

	return nil
}

// cleanupReview removes worktrees, checks out the original branch, closes the DB,
// and removes the review directory. Shared by finish and abort.
func cleanupReview(g *git.Git, repo *repository.Repository, out *output.Output, session db.Session) {
	ctx := context.Background()
	q := repo.Queries()

	reviewers, err := q.ListReviewers(ctx)
	if err != nil {
		out.Warn(fmt.Sprintf("failed to list reviewers: %v", err))
	}

	for _, r := range reviewers {
		if r.Name == "" {
			continue
		}
		worktreePath := filepath.Join(g.CommonDir, "review", "worktrees", r.Name)
		if _, statErr := os.Stat(worktreePath); os.IsNotExist(statErr) {
			continue
		}
		if err := g.WorktreeRemove(worktreePath); err != nil {
			out.Warn(fmt.Sprintf("failed to remove worktree %s: %v", r.Name, err))
		}
	}

	if err := g.CheckoutForce(session.Branch); err != nil {
		out.Warn(fmt.Sprintf("failed to checkout %s: %v", session.Branch, err))
	}

	repo.Close()
	reviewDir := filepath.Join(g.CommonDir, "review")
	if err := os.RemoveAll(reviewDir); err != nil {
		out.Warn(fmt.Sprintf("failed to clean up review directory: %v", err))
	}
}

// findCommitPosition returns the position of a commit with the given SHA, or -1 if not found.
func findCommitPosition(commits []db.Commit, sha string) int64 {
	for _, cm := range commits {
		if cm.Sha == sha {
			return cm.Position
		}
	}
	return -1
}
