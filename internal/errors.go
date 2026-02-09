package internal

import "github.com/newmo-oss/ergo"

var (
	ErrCodeNotInRepo      = ergo.NewCode("NotInRepo", "not in a git repository")
	ErrCodeNoReview       = ergo.NewCode("NoReview", "no review in progress")
	ErrCodeReviewActive   = ergo.NewCode("ReviewActive", "review already in progress")
	ErrCodeDirtyWorkDir   = ergo.NewCode("DirtyWorkDir", "uncommitted changes")
	ErrCodeInvalidRef     = ergo.NewCode("InvalidRef", "invalid git ref")
	ErrCodeNoCommits      = ergo.NewCode("NoCommits", "no commits to review")
	ErrCodeDetachedHead   = ergo.NewCode("DetachedHead", "detached HEAD state")
	ErrCodeWrongWorktree  = ergo.NewCode("WrongWorktree", "must run from main worktree")
)

