package commands

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/FujishigeTemma/git-review/internal"
	"github.com/FujishigeTemma/git-review/internal/db"
	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/google/uuid"
	"github.com/guregu/null/v6"
	"github.com/newmo-oss/ergo"
)

type AddCmd struct {
	File    string `short:"f" help:"File path for the comment."`
	Line    string `short:"l" help:"Line or range (e.g. 42, 10,35)."`
	ReplyTo string `short:"r" name:"reply-to" help:"ID of parent comment to reply to."`
	Author  string `short:"a" help:"Author name (default: worktree name)."`
	Message string `arg:"" help:"Comment message."`
}

func parseLineRange(raw string) (start, end null.Int, err error) {
	if raw == "" {
		return null.Int{}, null.Int{}, nil
	}
	if i := strings.IndexByte(raw, ','); i >= 0 {
		s, err := strconv.ParseInt(raw[:i], 10, 64)
		if err != nil {
			return null.Int{}, null.Int{}, ergo.New("invalid line range", slog.String("range", raw))
		}
		e, err := strconv.ParseInt(raw[i+1:], 10, 64)
		if err != nil {
			return null.Int{}, null.Int{}, ergo.New("invalid line range", slog.String("range", raw))
		}
		if s > e {
			return null.Int{}, null.Int{}, ergo.New("invalid line range: start must not exceed end", slog.String("range", raw))
		}
		return null.IntFrom(s), null.IntFrom(e), nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return null.Int{}, null.Int{}, ergo.New("invalid line number", slog.String("line", raw))
	}
	return null.IntFrom(n), null.IntFrom(n), nil
}

func (c *AddCmd) Run(g *git.Git, repo *repository.Repository, out *output.Output) error {
	if err := requireActive(repo); err != nil {
		return err
	}

	ctx := context.Background()
	q := repo.Queries()
	now := time.Now().UTC().Format(time.RFC3339)
	newID := uuid.Must(uuid.NewV7())

	author := c.Author
	if author == "" {
		author = g.Reviewer
	}

	var params db.InsertCommentParams

	if c.ReplyTo != "" {
		// Reply mode: find parent, inherit commit from parent
		parent, err := q.FindCommentByPrefix(ctx, sql.NullString{String: c.ReplyTo, Valid: true})
		if err != nil {
			return ergo.New("comment not found", slog.String("reply_to", c.ReplyTo))
		}

		params = db.InsertCommentParams{
			ID:        newID,
			ParentID:  uuid.NullUUID{UUID: parent.ID, Valid: true},
			Commit:    parent.Commit,
			File:      parent.File,
			StartLine: parent.StartLine,
			EndLine:   parent.EndLine,
			Body:      c.Message,
			CreatedAt: now,
			CreatedBy: author,
		}
	} else {
		// Non-reply: get reviewer's current commit
		reviewer, err := q.GetReviewer(ctx, g.Reviewer)
		if err != nil {
			return ergo.Wrap(err, "failed to get reviewer")
		}

		if !reviewer.CurrentSha.Valid {
			return ergo.New("No commit selected. Run 'git review next' first.")
		}
		commitSHA := reviewer.CurrentSha.String

		startLine, endLine, err := parseLineRange(c.Line)
		if err != nil {
			return err
		}

		var file null.String
		if c.File != "" {
			file = null.StringFrom(c.File)
		}

		params = db.InsertCommentParams{
			ID:        newID,
			Commit:    commitSHA,
			File:      file,
			StartLine: startLine,
			EndLine:   endLine,
			Body:      c.Message,
			CreatedAt: now,
			CreatedBy: author,
		}
	}

	if err := q.InsertComment(ctx, params); err != nil {
		return ergo.Wrap(err, "failed to save comment")
	}

	idStr := internal.ShortID(newID)
	if c.ReplyTo != "" {
		out.Ok(fmt.Sprintf("[%s] %s", idStr, c.Message))
	} else if c.File != "" {
		loc := c.File
		if lr := internal.FormatLineRange(params.StartLine, params.EndLine); lr != "" {
			loc += ":" + lr
		}
		out.Ok(fmt.Sprintf("[%s] %s %s", idStr, loc, c.Message))
	} else {
		out.Ok(fmt.Sprintf("[%s] %s", idStr, c.Message))
	}

	return nil
}
