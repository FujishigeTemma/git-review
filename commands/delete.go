package commands

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/FujishigeTemma/git-review/internal/db"
	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/google/uuid"
	"github.com/newmo-oss/ergo"
)

type DeleteCmd struct {
	ID string `arg:"" help:"ID (or prefix) of the comment to delete."`
}

func (c *DeleteCmd) Run(g *git.Git, repo *repository.Repository, out *output.Output) error {
	if err := requireActive(repo); err != nil {
		return err
	}

	ctx := context.Background()

	if err := repo.WithTx(ctx, func(q *db.Queries) error {
		target, err := q.FindCommentByPrefix(ctx, sql.NullString{String: c.ID, Valid: true})
		if err != nil {
			return ergo.New("comment not found", slog.String("comment_id", c.ID))
		}

		// If non-root: re-parent children to this comment's parent
		if target.ParentID.Valid {
			if err := q.ReparentChildren(ctx, db.ReparentChildrenParams{
				ParentID:   target.ParentID,
				ParentID_2: uuid.NullUUID{UUID: target.ID, Valid: true},
			}); err != nil {
				return ergo.Wrap(err, "failed to re-parent children")
			}
		}

		// Delete the comment (CASCADE handles root's children)
		if err := q.DeleteComment(ctx, target.ID); err != nil {
			return ergo.Wrap(err, "failed to delete comment")
		}

		return nil
	}); err != nil {
		return err
	}

	out.Ok("Comment deleted.")
	return nil
}
