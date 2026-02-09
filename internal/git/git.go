package git

import (
	"errors"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/newmo-oss/ergo"
)

// Git wraps git commands executed in a specific working directory.
type Git struct {
	WorkDir   string
	CommonDir string // Absolute path to shared .git directory.
	Reviewer  string // Worktree name. Empty string for main worktree.
}

// New creates a Git instance, resolving CommonDir and Reviewer at construction time.
func New(workDir string) (*Git, error) {
	g := &Git{WorkDir: workDir}

	commonDir, err := g.Run("rev-parse", "--git-common-dir")
	if err != nil {
		return nil, ergo.Wrap(err, "failed to resolve git common dir",
			slog.String("work_dir", workDir))
	}
	if !filepath.IsAbs(commonDir) {
		commonDir, err = filepath.Abs(filepath.Join(workDir, commonDir))
		if err != nil {
			return nil, ergo.Wrap(err, "failed to resolve absolute path for common dir",
				slog.String("common_dir", commonDir))
		}
	}
	g.CommonDir = commonDir

	reviewer, err := g.worktreeName(commonDir)
	if err != nil {
		return nil, ergo.Wrap(err, "failed to get worktree name",
			slog.String("common_dir", commonDir))
	}
	g.Reviewer = reviewer

	return g, nil
}

// ForWorktree returns a new Git for a linked worktree, inheriting CommonDir.
func (g *Git) ForWorktree(name, path string) *Git {
	return &Git{
		WorkDir:   path,
		CommonDir: g.CommonDir,
		Reviewer:  name,
	}
}

// Run executes a git command and returns trimmed stdout.
func (g *Git) Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.WorkDir
	out, err := cmd.Output()
	if err != nil {
		return "", ergo.Wrap(err, "git command failed",
			slog.String("args", strings.Join(args, " ")),
			slog.String("work_dir", g.WorkDir))
	}
	return strings.TrimSpace(string(out)), nil
}

// RunSilent executes a git command, ignoring output. Returns error if non-zero exit.
func (g *Git) RunSilent(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.WorkDir
	if err := cmd.Run(); err != nil {
		return ergo.Wrap(err, "git command failed",
			slog.String("args", strings.Join(args, " ")),
			slog.String("work_dir", g.WorkDir))
	}
	return nil
}

func (g *Git) GitDir() (string, error) {
	return g.Run("rev-parse", "--absolute-git-dir")
}

func (g *Git) CurrentBranch() (string, error) {
	return g.Run("branch", "--show-current")
}

func (g *Git) RefExists(ref string) bool {
	return g.RunSilent("rev-parse", "--verify", ref) == nil
}

func (g *Git) IsClean() (bool, error) {
	if err := g.RunSilent("diff", "--cached", "--quiet"); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return false, nil // staged changes exist
		}
		return false, err
	}
	if err := g.RunSilent("diff", "--quiet"); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return false, nil // unstaged changes exist
		}
		return false, err
	}
	return true, nil
}

func (g *Git) MergeBase(ref1, ref2 string) (string, error) {
	return g.Run("merge-base", ref1, ref2)
}

// RevList returns commit SHAs in reverse chronological order (oldest first).
func (g *Git) RevList(rangeSpec string) ([]string, error) {
	out, err := g.Run("rev-list", "--reverse", rangeSpec)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

func (g *Git) Oneline(ref string) (string, error) {
	return g.Run("log", "--oneline", "-1", ref)
}

func (g *Git) Subject(ref string) (string, error) {
	return g.Run("log", "-1", "--format=%s", ref)
}

func (g *Git) FullMessage(ref string) (string, error) {
	return g.Run("log", "-1", "--format=%B", ref)
}

func (g *Git) Checkout(ref string) error {
	return g.RunSilent("checkout", ref, "--quiet")
}

func (g *Git) CheckoutForce(ref string) error {
	return g.RunSilent("checkout", "--force", ref, "--quiet")
}

// NotesAppend appends a message to git notes for the given SHA.
// Falls back to "notes add" if "notes append" fails (no existing notes).
func (g *Git) NotesAppend(sha, message string) error {
	if err := g.RunSilent("notes", "append", "-m", message, sha); err != nil {
		return g.RunSilent("notes", "add", "-m", message, sha)
	}
	return nil
}

func (g *Git) WorktreeAdd(path string) error {
	return g.RunSilent("worktree", "add", path, "--detach")
}

func (g *Git) WorktreeRemove(path string) error {
	return g.RunSilent("worktree", "remove", path, "--force")
}

func (g *Git) ReadTreeReset(ref string) error {
	return g.RunSilent("read-tree", "-u", "--reset", ref)
}

func (g *Git) DiffStagedStat() (string, error) {
	return g.Run("diff", "--staged", "--stat")
}

// worktreeName returns the worktree name if running inside a linked worktree,
// or "" if in the main worktree. commonDir is passed from New() to avoid
// re-running "rev-parse --git-common-dir".
func (g *Git) worktreeName(commonDir string) (string, error) {
	gitDir, err := g.Run("rev-parse", "--absolute-git-dir")
	if err != nil {
		return "", err // already wrapped by Run()
	}
	if gitDir == commonDir {
		return "", nil // main worktree
	}
	return filepath.Base(gitDir), nil
}
