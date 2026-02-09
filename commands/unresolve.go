package commands

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/FujishigeTemma/git-review/internal"
	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/newmo-oss/ergo"
)

type UnresolveCmd struct {
	ID string `arg:"" help:"ID (or prefix) of the thread to unresolve."`
}

func (c *UnresolveCmd) Run(g *git.Git, repo *repository.Repository, out *output.Output) error {
	if err := requireActive(repo); err != nil {
		return err
	}

	ctx := context.Background()
	q := repo.Queries()

	comment, err := q.FindCommentByPrefix(ctx, sql.NullString{String: c.ID, Valid: true})
	if err != nil {
		return ergo.New("comment not found", slog.String("comment_id", c.ID))
	}

	if comment.ParentID.Valid {
		return ergo.New("only root comments can be unresolved", slog.String("comment_id", c.ID))
	}

	if !comment.ResolvedAt.Valid {
		return ergo.New("thread is not resolved")
	}

	if err := q.UnresolveComment(ctx, comment.ID); err != nil {
		return ergo.Wrap(err, "failed to unresolve comment")
	}

	out.Ok(fmt.Sprintf("Unresolved [%s]", internal.ShortID(comment.ID)))

	return nil
}
