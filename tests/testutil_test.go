package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build binary
	tmp, err := os.MkdirTemp("", "git-review-test-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp)

	binaryPath = filepath.Join(tmp, "git-review")
	cmd := exec.Command("go", "build", "-o", binaryPath, "..")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build: " + err.Error())
	}

	os.Exit(m.Run())
}

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	gitCmd := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	gitCmd("init", "-b", "main")
	gitCmd("config", "user.email", "test@test.com")
	gitCmd("config", "user.name", "Test")

	writeFile(t, dir, "README.md", "# Project\n")
	gitCmd("add", ".")
	gitCmd("commit", "-m", "Initial commit")

	gitCmd("checkout", "-b", "feature/test")

	writeFile(t, dir, "app.js", "function hello() { return \"hello\"; }\n")
	gitCmd("add", ".")
	gitCmd("commit", "-m", "Add hello function")

	writeFile(t, dir, "app.js", "function hello() { return \"hello\"; }\nfunction goodbye() { return \"bye\"; }\n")
	gitCmd("add", ".")
	gitCmd("commit", "-m", "Add goodbye function")

	writeFile(t, dir, "app.js", "function hello() { return \"hello\"; }\nfunction goodbye() { return \"bye\"; }\nconsole.log(hello());\n")
	gitCmd("add", ".")
	gitCmd("commit", "-m", "Add main entry")

	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runGR(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@test.com",
		"TERM=dumb", // Disable colors
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func mustRunGR(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := runGR(t, dir, args...)
	if err != nil {
		t.Fatalf("git-review %v: %v\n%s", args, err, out)
	}
	return out
}

// loadState runs `git review state` and parses the JSON output.
func loadState(t *testing.T, dir string) map[string]interface{} {
	t.Helper()
	out := mustRunGR(t, dir, "state")
	out = strings.TrimSpace(out)
	if out == "null" {
		return nil
	}
	var state map[string]interface{}
	if err := json.Unmarshal([]byte(out), &state); err != nil {
		t.Fatalf("unmarshal state: %v\noutput: %s", err, out)
	}
	return state
}

// stateComments extracts the "comments" array from state JSON.
func stateComments(t *testing.T, state map[string]interface{}) []map[string]interface{} {
	t.Helper()
	raw, ok := state["comments"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]interface{})
	if !ok {
		t.Fatalf("comments is not an array: %T", raw)
	}
	var result []map[string]interface{}
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			t.Fatalf("comment is not a map: %T", item)
		}
		result = append(result, m)
	}
	return result
}

func assertContains(t *testing.T, label, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("%s: expected to contain %q, got:\n%s", label, needle, haystack)
	}
}

func assertNotContains(t *testing.T, label, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("%s: expected NOT to contain %q, got:\n%s", label, needle, haystack)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %s", path)
	}
}

func assertDirNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected directory to not exist: %s", path)
	}
}

func gitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func findCommentByBody(comments []map[string]interface{}, body string) map[string]interface{} {
	for _, c := range comments {
		if c["body"] == body {
			return c
		}
	}
	return nil
}

