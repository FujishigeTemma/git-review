package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/FujishigeTemma/git-review/commands"
	"github.com/FujishigeTemma/git-review/internal"
	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/alecthomas/kong"
	"github.com/newmo-oss/ergo"
)

//go:embed SKILL.md
var skill string

//go:embed schema.sql
var schema string

// CLI defines the kong command structure for git-review.
type CLI struct {
	Start     commands.StartCmd     `cmd:"" default:"withargs" help:"Start review (auto-detects base if omitted)."`
	Add       commands.AddCmd       `cmd:"" help:"Add comment to current commit."`
	Next      commands.NextCmd      `cmd:"" help:"Move to next commit."`
	Jump      commands.JumpCmd      `cmd:"" help:"Jump to a specific commit."`
	List      commands.ListCmd      `cmd:"" help:"Show all comments (Markdown)."`
	Status    commands.StatusCmd    `cmd:"" help:"Show review progress."`
	Delete    commands.DeleteCmd    `cmd:"" help:"Delete a comment by ID."`
	Resolve   commands.ResolveCmd   `cmd:"" help:"Resolve a thread."`
	Unresolve commands.UnresolveCmd `cmd:"" help:"Unresolve a thread."`
	Finish    commands.FinishCmd    `cmd:"" help:"Finish review and write git notes."`
	Abort     commands.AbortCmd     `cmd:"" help:"Cancel review and clean up."`
	State     commands.StateCmd     `cmd:"" hidden:""`
	Skill     commands.SkillCmd     `cmd:"" help:"Show AI Agent workflow guide."`

	repo *repository.Repository
}

// AfterApply runs after flag parsing, before Run().
// Binds shared dependencies to Kong context for injection into Run().
func (c *CLI) AfterApply(ctx *kong.Context) error {
	ctx.Bind(output.New())

	if ctx.Selected().Name == "skill" {
		return nil
	}

	g, err := git.New(".")
	if err != nil {
		return ergo.WithCode(
			ergo.New("not in a git repository"),
			internal.ErrCodeNotInRepo)
	}
	ctx.Bind(g)

	dbPath := filepath.Join(g.CommonDir, "review", "review.db")
	var repo *repository.Repository
	if ctx.Selected().Name == "start" {
		repo, err = repository.Create(dbPath, schema)
	} else {
		repo, err = repository.Open(dbPath)
	}
	if err != nil {
		if ctx.Selected().Name == "state" {
			// state outputs "null" when no review exists
			ctx.Bind((*repository.Repository)(nil))
			return nil
		}
		return err
	}
	c.repo = repo
	ctx.Bind(repo)

	return nil
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("git-review"),
		kong.Description("Commit review workflow for AI Agent collaboration"),
		kong.UsageOnError(),
		kong.Bind(commands.SkillMarkdown(skill)),
	)
	defer func() {
		if cli.repo != nil {
			cli.repo.Close()
		}
	}()

	if err := ctx.Run(); err != nil {
		// ergo.WithCode wraps the message as "CodeName: message".
		// Strip the code prefix to show a clean user-facing message.
		msg := err.Error()
		if code := ergo.CodeOf(err); !code.IsZero() {
			prefix := code.String() + ": "
			if len(msg) > len(prefix) && msg[:len(prefix)] == prefix {
				msg = msg[len(prefix):]
			}
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
		os.Exit(1)
	}
}
