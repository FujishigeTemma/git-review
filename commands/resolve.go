package commands

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/FujishigeTemma/git-review/internal"
	"github.com/FujishigeTemma/git-review/internal/db"
	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/guregu/null/v6"
	"github.com/newmo-oss/ergo"
)

type ResolveCmd struct {
	ID   string `arg:"" help:"ID (or prefix) of the thread to resolve."`
	Name string `short:"a" help:"Who resolved it (default: worktree name)."`
}

func (c *ResolveCmd) Run(g *git.Git, repo *repository.Repository, out *output.Output) error {
	if err := requireActive(repo); err != nil {
		return err
	}

	ctx := context.Background()
	q := repo.Queries()

	name := c.Name
	if name == "" {
		name = g.Reviewer
	}

	comment, err := q.FindCommentByPrefix(ctx, sql.NullString{String: c.ID, Valid: true})
	if err != nil {
		return ergo.New("comment not found", slog.String("comment_id", c.ID))
	}

	if comment.ParentID.Valid {
		return ergo.New("only root comments can be resolved", slog.String("comment_id", c.ID))
	}

	if comment.ResolvedAt.Valid {
		return ergo.New("thread is already resolved")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if err := q.ResolveComment(ctx, db.ResolveCommentParams{
		ResolvedAt: null.StringFrom(now),
		ResolvedBy: null.StringFrom(name),
		ID:         comment.ID,
	}); err != nil {
		return ergo.Wrap(err, "failed to resolve comment")
	}

	out.Ok(fmt.Sprintf("Resolved [%s]", internal.ShortID(comment.ID)))

	return nil
}
