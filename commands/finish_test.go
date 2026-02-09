package commands

import (
	"testing"

	"github.com/FujishigeTemma/git-review/internal/db"
	"github.com/google/uuid"
	"github.com/guregu/null/v6"
)

func newComment(id uuid.UUID, parentID uuid.NullUUID, commit, body, createdBy string, file null.String, startLine, endLine null.Int) db.Comment {
	return db.Comment{
		ID:        id,
		ParentID:  parentID,
		Commit:    commit,
		Body:      body,
		CreatedBy: createdBy,
		File:      file,
		StartLine: startLine,
		EndLine:   endLine,
	}
}

func TestBuildCommitNotes_NoComments(t *testing.T) {
	childrenMap := buildChildrenMap(nil)
	got := buildCommitNotes(nil, childrenMap, "abc123")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestBuildCommitNotes_GeneralComment(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(id, uuid.NullUUID{}, "abc123", "Good work", "alice", null.String{}, null.Int{}, null.Int{}),
	}
	childrenMap := buildChildrenMap(comments)
	got := buildCommitNotes(comments, childrenMap, "abc123")
	if got != "Good work @alice" {
		t.Errorf("got %q, want %q", got, "Good work @alice")
	}
}

func TestBuildCommitNotes_FileComment(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(id, uuid.NullUUID{}, "abc123", "Fix this", "bob",
			null.StringFrom("main.go"), null.IntFrom(10), null.IntFrom(10)),
	}
	childrenMap := buildChildrenMap(comments)
	got := buildCommitNotes(comments, childrenMap, "abc123")
	want := "main.go:10 -- Fix this @bob"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildCommitNotes_FileCommentWithRange(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(id, uuid.NullUUID{}, "abc123", "Split this", "bob",
			null.StringFrom("main.go"), null.IntFrom(5), null.IntFrom(12)),
	}
	childrenMap := buildChildrenMap(comments)
	got := buildCommitNotes(comments, childrenMap, "abc123")
	want := "main.go:5-12 -- Split this @bob"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildCommitNotes_WithReplies(t *testing.T) {
	parentID := uuid.Must(uuid.NewV7())
	childID := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(parentID, uuid.NullUUID{}, "abc123", "Issue here", "alice", null.String{}, null.Int{}, null.Int{}),
		newComment(childID, uuid.NullUUID{UUID: parentID, Valid: true}, "abc123", "Fixed!", "bob", null.String{}, null.Int{}, null.Int{}),
	}
	childrenMap := buildChildrenMap(comments)
	got := buildCommitNotes(comments, childrenMap, "abc123")
	want := "Issue here @alice\n  Fixed! @bob"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildCommitNotes_CrossCommitReply(t *testing.T) {
	parentID := uuid.Must(uuid.NewV7())
	childID := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(parentID, uuid.NullUUID{}, "abc123", "Issue", "alice", null.String{}, null.Int{}, null.Int{}),
		newComment(childID, uuid.NullUUID{UUID: parentID, Valid: true}, "def456", "Reply from other commit", "bob", null.String{}, null.Int{}, null.Int{}),
	}
	childrenMap := buildChildrenMap(comments)
	got := buildCommitNotes(comments, childrenMap, "abc123")
	want := "Issue @alice\n  (def456) Reply from other commit @bob"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildCommitNotes_EmptyAuthor(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(id, uuid.NullUUID{}, "abc123", "Anonymous comment", "", null.String{}, null.Int{}, null.Int{}),
	}
	childrenMap := buildChildrenMap(comments)
	got := buildCommitNotes(comments, childrenMap, "abc123")
	if got != "Anonymous comment" {
		t.Errorf("got %q, want %q", got, "Anonymous comment")
	}
}

func TestBuildCommitNotes_SkipsOtherCommits(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(id, uuid.NullUUID{}, "other", "Not this one", "alice", null.String{}, null.Int{}, null.Int{}),
	}
	childrenMap := buildChildrenMap(comments)
	got := buildCommitNotes(comments, childrenMap, "abc123")
	if got != "" {
		t.Errorf("expected empty for other commit, got %q", got)
	}
}
