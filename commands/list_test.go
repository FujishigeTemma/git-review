package commands

import (
	"testing"

	"github.com/FujishigeTemma/git-review/internal/db"
	"github.com/google/uuid"
	"github.com/guregu/null/v6"
)

func TestFilterComments_NoFilter(t *testing.T) {
	comments := []db.Comment{
		newComment(uuid.Must(uuid.NewV7()), uuid.NullUUID{}, "abc", "c1", "", null.String{}, null.Int{}, null.Int{}),
		newComment(uuid.Must(uuid.NewV7()), uuid.NullUUID{}, "def", "c2", "", null.String{}, null.Int{}, null.Int{}),
	}
	idMap := buildIDMap(comments)
	got := filterComments(comments, nil, idMap, "", false, "", "")
	if len(got) != 2 {
		t.Errorf("expected 2 comments, got %d", len(got))
	}
}

func TestFilterComments_ByCommitPrefix(t *testing.T) {
	id1 := uuid.Must(uuid.NewV7())
	id2 := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(id1, uuid.NullUUID{}, "abc123", "match", "", null.String{}, null.Int{}, null.Int{}),
		newComment(id2, uuid.NullUUID{}, "def456", "no match", "", null.String{}, null.Int{}, null.Int{}),
	}
	commits := []db.Commit{{Sha: "abc123"}, {Sha: "def456"}}
	idMap := buildIDMap(comments)
	got := filterComments(comments, commits, idMap, "abc", false, "", "")
	if len(got) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(got))
	}
	if got[0].Body != "match" {
		t.Errorf("got body %q, want %q", got[0].Body, "match")
	}
}

func TestFilterComments_ByUnresolved(t *testing.T) {
	id1 := uuid.Must(uuid.NewV7())
	id2 := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(id1, uuid.NullUUID{}, "abc", "open", "", null.String{}, null.Int{}, null.Int{}),
		{ID: id2, Commit: "abc", Body: "resolved", ResolvedAt: null.StringFrom("2024-01-01T00:00:00Z")},
	}
	idMap := buildIDMap(comments)
	got := filterComments(comments, nil, idMap, "", true, "", "")
	if len(got) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(got))
	}
	if got[0].Body != "open" {
		t.Errorf("got body %q, want %q", got[0].Body, "open")
	}
}

func TestFilterComments_ByCreator(t *testing.T) {
	id1 := uuid.Must(uuid.NewV7())
	id2 := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(id1, uuid.NullUUID{}, "abc", "by alice", "alice", null.String{}, null.Int{}, null.Int{}),
		newComment(id2, uuid.NullUUID{}, "abc", "by bob", "bob", null.String{}, null.Int{}, null.Int{}),
	}
	idMap := buildIDMap(comments)
	got := filterComments(comments, nil, idMap, "", false, "alice", "")
	if len(got) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(got))
	}
	if got[0].Body != "by alice" {
		t.Errorf("got body %q", got[0].Body)
	}
}

func TestFilterComments_ByFile(t *testing.T) {
	id1 := uuid.Must(uuid.NewV7())
	id2 := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(id1, uuid.NullUUID{}, "abc", "on file", "", null.StringFrom("main.go"), null.Int{}, null.Int{}),
		newComment(id2, uuid.NullUUID{}, "abc", "general", "", null.String{}, null.Int{}, null.Int{}),
	}
	idMap := buildIDMap(comments)
	got := filterComments(comments, nil, idMap, "", false, "", "main.go")
	if len(got) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(got))
	}
	if got[0].Body != "on file" {
		t.Errorf("got body %q", got[0].Body)
	}
}

func TestFilterComments_IncludesDescendantsOfMatchingRoot(t *testing.T) {
	rootID := uuid.Must(uuid.NewV7())
	childID := uuid.Must(uuid.NewV7())
	otherID := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(rootID, uuid.NullUUID{}, "abc", "root", "alice", null.String{}, null.Int{}, null.Int{}),
		newComment(childID, uuid.NullUUID{UUID: rootID, Valid: true}, "abc", "reply", "bob", null.String{}, null.Int{}, null.Int{}),
		newComment(otherID, uuid.NullUUID{}, "abc", "other root", "charlie", null.String{}, null.Int{}, null.Int{}),
	}
	idMap := buildIDMap(comments)
	got := filterComments(comments, nil, idMap, "", false, "alice", "")
	if len(got) != 2 {
		t.Fatalf("expected 2 (root + reply), got %d", len(got))
	}
}

func TestFilterComments_NoMatchingCommit(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(id, uuid.NullUUID{}, "abc", "c1", "", null.String{}, null.Int{}, null.Int{}),
	}
	commits := []db.Commit{{Sha: "abc"}}
	idMap := buildIDMap(comments)
	got := filterComments(comments, commits, idMap, "zzz", false, "", "")
	if got != nil {
		t.Errorf("expected nil, got %d comments", len(got))
	}
}

func TestFilterComments_CombinedFilters(t *testing.T) {
	id1 := uuid.Must(uuid.NewV7())
	id2 := uuid.Must(uuid.NewV7())
	id3 := uuid.Must(uuid.NewV7())
	comments := []db.Comment{
		newComment(id1, uuid.NullUUID{}, "abc", "match both", "alice", null.StringFrom("main.go"), null.Int{}, null.Int{}),
		newComment(id2, uuid.NullUUID{}, "abc", "wrong creator", "bob", null.StringFrom("main.go"), null.Int{}, null.Int{}),
		newComment(id3, uuid.NullUUID{}, "abc", "wrong file", "alice", null.StringFrom("other.go"), null.Int{}, null.Int{}),
	}
	idMap := buildIDMap(comments)
	got := filterComments(comments, nil, idMap, "", false, "alice", "main.go")
	if len(got) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(got))
	}
	if got[0].Body != "match both" {
		t.Errorf("got body %q", got[0].Body)
	}
}

func TestDescendants_BuildsTree(t *testing.T) {
	root := uuid.Must(uuid.NewV7())
	child1 := uuid.Must(uuid.NewV7())
	child2 := uuid.Must(uuid.NewV7())
	grandchild := uuid.Must(uuid.NewV7())

	comments := []db.Comment{
		{ID: root, Body: "root"},
		{ID: child1, ParentID: uuid.NullUUID{UUID: root, Valid: true}, Body: "child1"},
		{ID: child2, ParentID: uuid.NullUUID{UUID: root, Valid: true}, Body: "child2"},
		{ID: grandchild, ParentID: uuid.NullUUID{UUID: child1, Valid: true}, Body: "grandchild"},
	}

	childrenMap := buildChildrenMap(comments)
	desc := descendants(childrenMap, root)
	if len(desc) != 3 {
		t.Fatalf("expected 3 descendants, got %d", len(desc))
	}
}

func TestFindRoot_WalksUpChain(t *testing.T) {
	root := uuid.Must(uuid.NewV7())
	child := uuid.Must(uuid.NewV7())
	grandchild := uuid.Must(uuid.NewV7())

	comments := []db.Comment{
		{ID: root, Body: "root"},
		{ID: child, ParentID: uuid.NullUUID{UUID: root, Valid: true}, Body: "child"},
		{ID: grandchild, ParentID: uuid.NullUUID{UUID: child, Valid: true}, Body: "grandchild"},
	}

	idMap := buildIDMap(comments)
	found := findRoot(idMap, comments[2])
	if found.ID != root {
		t.Errorf("expected root %s, got %s", root, found.ID)
	}
}
