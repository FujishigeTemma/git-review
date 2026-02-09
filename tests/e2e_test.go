package tests

import (
	"path/filepath"
	"testing"
)

func TestStart_CreatesDBAndShowsCommits(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)

	output := mustRunGR(t, dir)

	assertContains(t, "shows commit count", output, "3 commit(s)")
	assertContains(t, "shows first commit", output, "Add hello function")
	assertFileExists(t, filepath.Join(dir, ".git", "review", "review.db"))
}

func TestStart_Next_AdvancesToSecondCommit(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir) // start positions at commit 1

	output := mustRunGR(t, dir, "next")

	assertContains(t, "shows second commit", output, "Add goodbye function")
	assertContains(t, "shows position", output, "[2/3]")
}

func TestNext_AdvancesToThirdCommit(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")

	output := mustRunGR(t, dir, "next")

	assertContains(t, "shows third commit", output, "Add main entry")
	assertContains(t, "shows position", output, "[3/3]")
}

func TestNext_ShowsMessageWhenAllReviewed(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "next")

	output := mustRunGR(t, dir, "next")

	assertContains(t, "all reviewed message", output, "All commits reviewed")
}

func TestAdd_GeneralAndFileComments(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")

	mustRunGR(t, dir, "add", "Good approach")
	mustRunGR(t, dir, "add", "-f", "app.js", "-l", "1", "Use arrow function")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	fileComment := findCommentByBody(comments, "Use arrow function")
	if fileComment == nil {
		t.Fatal("file comment not found")
	}
	if fileComment["file"] != "app.js" {
		t.Errorf("file: got %v", fileComment["file"])
	}
}

func TestAdd_RangeComment(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")

	mustRunGR(t, dir, "add", "-f", "app.js", "-l", "10,25", "Split this function")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	c := comments[0]
	if c["startLine"].(float64) != 10 {
		t.Errorf("startLine: got %v, want 10", c["startLine"])
	}
	if c["endLine"].(float64) != 25 {
		t.Errorf("endLine: got %v, want 25", c["endLine"])
	}
}

func TestList_ShowsAllComments(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "Good approach")
	mustRunGR(t, dir, "add", "-f", "app.js", "-l", "1", "Use arrow function")
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "Separate entry point into index.js")

	output := mustRunGR(t, dir, "list")

	assertContains(t, "shows branch", output, "feature/test")
	assertContains(t, "shows general comment", output, "Good approach")
	assertContains(t, "shows file comment", output, "app.js")
	assertContains(t, "shows third commit comment", output, "Separate entry point")
}

func TestFinish_CleansUpAndCheckoutsOriginal(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "next")

	output := mustRunGR(t, dir, "finish")

	assertContains(t, "review complete", output, "Review Complete")
	assertContains(t, "back on branch", output, "Back on")
	assertDirNotExists(t, filepath.Join(dir, ".git", "review"))

	branch := gitCmd(t, dir, "branch", "--show-current")
	if branch != "feature/test" {
		t.Errorf("branch: got %q, want 'feature/test'", branch)
	}
}

func TestStart_BeginsNewReviewAfterCompletion(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "finish")

	output := mustRunGR(t, dir)
	assertContains(t, "can start after finish", output, "Review Started")
	mustRunGR(t, dir, "abort")
}

func TestAbort_RemovesStateAndRestoresBranch(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)

	output, err := runGR(t, dir, "abort")
	if err != nil {
		t.Fatalf("abort: %v\n%s", err, output)
	}
	assertContains(t, "abort message", output, "aborted")

	branch := gitCmd(t, dir, "branch", "--show-current")
	if branch != "feature/test" {
		t.Errorf("branch: got %q, want 'feature/test'", branch)
	}

	assertDirNotExists(t, filepath.Join(dir, ".git", "review"))
}

func TestStatus_ShowsProgressAndCommentCount(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "comment on first")

	output := mustRunGR(t, dir, "status")

	assertContains(t, "shows progress header", output, "Review Progress")
	assertContains(t, "shows current indicator", output, "→")
	assertContains(t, "shows comment count", output, "1 comment")
}

func TestNoArgs_ShowsStatusDuringReview(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)

	output := mustRunGR(t, dir)
	assertContains(t, "shows status on no args", output, "Review Progress")
}

func TestUnknownCommand_ReturnsErrorDuringReview(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)

	output, err := runGR(t, dir, "somebranch")
	if err == nil {
		t.Fatalf("expected error for unknown command, got:\n%s", output)
	}
	assertContains(t, "error on unknown command", output, "Review already in progress")
}

func TestAdd_CommentsHaveUUID(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "test comment")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	id, ok := comments[0]["id"].(string)
	if !ok || id == "" {
		t.Errorf("comment should have non-empty id, got %v", comments[0]["id"])
	}
	if comments[0]["parentId"] != nil {
		t.Errorf("top-level comment should have null parentId, got %v", comments[0]["parentId"])
	}
}

func TestAdd_CommentsHaveCreatedAt(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "test comment")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	createdAt, ok := comments[0]["createdAt"].(string)
	if !ok || createdAt == "" {
		t.Errorf("comment should have non-empty createdAt, got %v", comments[0]["createdAt"])
	}
}

func TestDelete_ByUUID(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "first comment")
	mustRunGR(t, dir, "add", "second comment")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	id := comments[0]["id"].(string)
	mustRunGR(t, dir, "delete", id)

	state = loadState(t, dir)
	remaining := stateComments(t, state)
	if len(remaining) != 1 {
		t.Fatalf("expected 1 comment after delete, got %d", len(remaining))
	}
	if remaining[0]["body"] != "second comment" {
		t.Errorf("remaining comment body: got %v", remaining[0]["body"])
	}
}

func TestDelete_NotFound(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)

	_, err := runGR(t, dir, "delete", "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent ID")
	}
}

func TestReplyTo_CreatesReply(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "-f", "app.js", "-l", "1", "Use arrow function")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	parentID := comments[0]["id"].(string)

	mustRunGR(t, dir, "add", "--reply-to", parentID, "Fixed!")

	state = loadState(t, dir)
	comments = stateComments(t, state)
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	reply := findCommentByBody(comments, "Fixed!")
	if reply == nil {
		t.Fatal("reply not found")
	}
	if reply["parentId"] != parentID {
		t.Errorf("reply parentId: got %v, want %q", reply["parentId"], parentID)
	}
	if reply["body"] != "Fixed!" {
		t.Errorf("reply body: got %v", reply["body"])
	}
	if reply["file"] != "app.js" {
		t.Errorf("reply file: got %v, want app.js", reply["file"])
	}
}

func TestReplyTo_InheritsCommitFromParent(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "parent comment on commit 0")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	parentID := comments[0]["id"].(string)
	parentCommit := comments[0]["commit"].(string)

	mustRunGR(t, dir, "next")

	// Reply from commit 1 to parent on commit 0 -- v2 inherits commit from parent
	mustRunGR(t, dir, "add", "--reply-to", parentID, "Reply from commit 1")

	state = loadState(t, dir)
	comments = stateComments(t, state)
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	reply := findCommentByBody(comments, "Reply from commit 1")
	if reply == nil {
		t.Fatal("reply not found")
	}
	// In v2, reply inherits commit from parent
	replyCommit := reply["commit"].(string)
	if replyCommit != parentCommit {
		t.Errorf("reply should inherit parent's commit, got %v, want %v", replyCommit, parentCommit)
	}
}

func TestReplyTo_NotFound(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")

	_, err := runGR(t, dir, "add", "--reply-to", "nonexistent", "reply text")
	if err == nil {
		t.Fatal("expected error for nonexistent parent ID")
	}
}

func TestDelete_HardDeleteRoot_CascadesChildren(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "parent comment")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	parentID := comments[0]["id"].(string)

	mustRunGR(t, dir, "add", "--reply-to", parentID, "reply 1")
	mustRunGR(t, dir, "add", "--reply-to", parentID, "reply 2")

	// Delete root → CASCADE deletes all children
	mustRunGR(t, dir, "delete", parentID)

	state = loadState(t, dir)
	remaining := stateComments(t, state)
	if len(remaining) != 0 {
		t.Errorf("expected 0 comments after root delete, got %d", len(remaining))
	}
}

func TestDelete_NonRoot_ReparentsChildren(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "parent comment")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	parentID := comments[0]["id"].(string)

	mustRunGR(t, dir, "add", "--reply-to", parentID, "middle reply")

	state = loadState(t, dir)
	comments = stateComments(t, state)
	middleReply := findCommentByBody(comments, "middle reply")
	middleID := middleReply["id"].(string)

	mustRunGR(t, dir, "add", "--reply-to", middleID, "grandchild")

	// Delete the middle reply → grandchild re-parented to root
	mustRunGR(t, dir, "delete", middleID)

	state = loadState(t, dir)
	remaining := stateComments(t, state)
	if len(remaining) != 2 {
		t.Fatalf("expected 2 comments after middle delete, got %d", len(remaining))
	}

	grandchild := findCommentByBody(remaining, "grandchild")
	if grandchild == nil {
		t.Fatal("grandchild not found after re-parent")
	}
	if grandchild["parentId"] != parentID {
		t.Errorf("grandchild should be re-parented to root, got parentId=%v", grandchild["parentId"])
	}
}

func TestDelete_ReplyOnly(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "parent comment")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	parentID := comments[0]["id"].(string)

	mustRunGR(t, dir, "add", "--reply-to", parentID, "reply")

	state = loadState(t, dir)
	comments = stateComments(t, state)
	reply := findCommentByBody(comments, "reply")
	replyID := reply["id"].(string)

	mustRunGR(t, dir, "delete", replyID)

	state = loadState(t, dir)
	remaining := stateComments(t, state)
	if len(remaining) != 1 {
		t.Fatalf("expected 1 comment after deleting reply, got %d", len(remaining))
	}
	if remaining[0]["body"] != "parent comment" {
		t.Errorf("remaining should be parent, got %v", remaining[0]["body"])
	}
}

func TestList_ShowsThreadedComments(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "Good approach")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	parentID := comments[0]["id"].(string)

	mustRunGR(t, dir, "add", "--reply-to", parentID, "Thanks!")

	output := mustRunGR(t, dir, "list")
	assertContains(t, "shows parent", output, "Good approach")
	assertContains(t, "shows reply", output, "Thanks!")
}

func TestList_ShowsNoEmoji(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "Good approach")
	mustRunGR(t, dir, "add", "-f", "app.js", "-l", "1", "Fix this")

	output := mustRunGR(t, dir, "list")
	assertNotContains(t, "no emoji in list", output, "\U0001f4ac")
	assertNotContains(t, "no emoji in list", output, "\u21a9")
	assertNotContains(t, "no emoji in list", output, "\U0001f4c4")
	assertNotContains(t, "no emoji in list", output, "\U0001f4dd")
}

func TestList_ShowsIDsInBrackets(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "Test comment")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	shortID := comments[0]["id"].(string)[:8]

	output := mustRunGR(t, dir, "list")
	assertContains(t, "shows ID in brackets", output, "["+shortID+"]")
}

func TestResolve_ResolvesRootComment(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "Need to fix this")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	id := comments[0]["id"].(string)

	output := mustRunGR(t, dir, "resolve", id, "-a", "reviewer")
	assertContains(t, "resolved message", output, "Resolved")

	state = loadState(t, dir)
	comments = stateComments(t, state)
	if comments[0]["resolvedAt"] == nil {
		t.Error("comment should have resolvedAt set")
	}
	if comments[0]["resolvedBy"] != "reviewer" {
		t.Errorf("resolvedBy: got %v, want 'reviewer'", comments[0]["resolvedBy"])
	}
}

func TestResolve_ErrorOnNonRoot(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "parent")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	parentID := comments[0]["id"].(string)

	mustRunGR(t, dir, "add", "--reply-to", parentID, "child reply")

	state = loadState(t, dir)
	comments = stateComments(t, state)
	replyID := findCommentByBody(comments, "child reply")["id"].(string)

	_, err := runGR(t, dir, "resolve", replyID)
	if err == nil {
		t.Fatal("expected error resolving non-root comment")
	}
}

func TestUnresolve_UnresolvesComment(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "Issue found")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	id := comments[0]["id"].(string)

	mustRunGR(t, dir, "resolve", id, "-a", "reviewer")
	mustRunGR(t, dir, "unresolve", id)

	state = loadState(t, dir)
	comments = stateComments(t, state)
	if comments[0]["resolvedAt"] != nil {
		t.Error("comment should have null resolvedAt after unresolve")
	}
}

func TestState_OutputsJSON(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "test comment")

	state := loadState(t, dir)
	if state == nil {
		t.Fatal("state should not be null")
	}

	if state["branch"] != "feature/test" {
		t.Errorf("branch: got %v", state["branch"])
	}

	commits, ok := state["commits"].([]interface{})
	if !ok || len(commits) != 3 {
		t.Errorf("expected 3 commits, got %v", state["commits"])
	}

	comments := stateComments(t, state)
	if len(comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(comments))
	}
}

func TestState_NullWhenNoReview(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)

	out := mustRunGR(t, dir, "state")
	assertContains(t, "null output", out, "null")
}

func TestAdd_WorksAfterStart(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)

	// start sets current_sha to first commit — add should work immediately
	mustRunGR(t, dir, "add", "comment after start")
}

func TestJump_ToSpecificCommit(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")

	// Get the second commit SHA from state
	state := loadState(t, dir)
	commits := state["commits"].([]interface{})
	secondSHA := commits[1].(string)

	// Jump using prefix
	output := mustRunGR(t, dir, "jump", secondSHA[:7])
	assertContains(t, "shows jumped commit", output, "Add goodbye function")
	assertContains(t, "shows position", output, "[2/3]")
}

func TestJump_NotFound(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)

	_, err := runGR(t, dir, "jump", "deadbeef")
	if err == nil {
		t.Fatal("expected error for nonexistent commit hash")
	}
}

func TestList_FilterByUnresolved(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "resolved issue")
	mustRunGR(t, dir, "add", "open issue")

	state := loadState(t, dir)
	comments := stateComments(t, state)
	resolvedID := findCommentByBody(comments, "resolved issue")["id"].(string)

	mustRunGR(t, dir, "resolve", resolvedID, "-a", "reviewer")

	output := mustRunGR(t, dir, "list", "--unresolved")
	assertContains(t, "shows unresolved", output, "open issue")
	assertNotContains(t, "hides resolved", output, "resolved issue")
}

func TestFinish_WritesGitNotes(t *testing.T) {
	t.Parallel()
	dir := setupTestRepo(t)
	mustRunGR(t, dir)
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "add", "Good function naming")
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "next")
	mustRunGR(t, dir, "finish")

	// Check git notes on the first commit (Add hello function)
	notes := gitCmd(t, dir, "log", "--notes", "--format=%N", "main..feature/test")
	assertContains(t, "notes contain comment", notes, "Good function naming")
}
