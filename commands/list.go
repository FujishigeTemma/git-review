package commands

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/FujishigeTemma/git-review/internal"
	"github.com/FujishigeTemma/git-review/internal/db"
	"github.com/FujishigeTemma/git-review/internal/git"
	"github.com/FujishigeTemma/git-review/internal/output"
	"github.com/FujishigeTemma/git-review/internal/repository"
	"github.com/newmo-oss/ergo"
)

type ListCmd struct {
	ID         string `arg:"" optional:"" help:"Comment ID to show specific thread."`
	Commit     string `help:"Filter by commit hash prefix." name:"commit"`
	Unresolved bool   `help:"Show only unresolved threads." name:"unresolved"`
	Creator    string `help:"Filter by creator." name:"creator"`
	File       string `help:"Filter by file path." name:"file"`
	TopLevel   bool   `help:"Show only top-level comments (no replies)." name:"top-level"`
}

func (c *ListCmd) Run(g *git.Git, repo *repository.Repository, out *output.Output) error {
	if err := requireActive(repo); err != nil {
		return err
	}

	ctx := context.Background()
	q := repo.Queries()

	// If ID specified, show that thread only
	if c.ID != "" {
		return c.showThread(ctx, q, out)
	}

	session, err := q.GetSession(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to get session")
	}

	commits, err := q.ListCommits(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to list commits")
	}

	allComments, err := q.ListAllComments(ctx)
	if err != nil {
		out.Warn(fmt.Sprintf("failed to load comments: %v", err))
	}

	// Build lookup maps once for efficient tree operations
	childrenMap := buildChildrenMap(allComments)
	idMap := buildIDMap(allComments)

	// Apply filters to get the set of relevant root comment IDs
	comments := filterComments(allComments, commits, idMap, c.Commit, c.Unresolved, c.Creator, c.File)

	total := len(commits)

	out.Printf("\n")
	out.Printf("# Review Comments\n")
	out.Printf("\n")
	out.Printf("Branch: %s\n", session.Branch)
	out.Printf("Commits: %d\n", total)

	for _, cm := range commits {
		out.Printf("\n")
		out.Printf("---\n")
		out.Printf("\n")
		out.Printf("## Commit %d/%d %s: %s\n", cm.Position+1, total, internal.ShortSHA(cm.Sha), cm.Message)
		out.Printf("\n")

		// Collect top-level comments for this commit from filtered set
		var commitTopLevel []db.Comment
		for _, cc := range comments {
			if cc.Commit == cm.Sha && !cc.ParentID.Valid {
				commitTopLevel = append(commitTopLevel, cc)
			}
		}

		if len(commitTopLevel) == 0 {
			out.Printf("No comments\n")
			continue
		}

		// General comments (no file)
		for _, tc := range commitTopLevel {
			if tc.File.Valid {
				continue
			}
			if c.TopLevel {
				printCommentLine(out, tc, cm.Sha, "")
			} else {
				printThreadFlat(out, childrenMap, tc, cm.Sha)
			}
		}

		// File-specific comments grouped by file
		type fileEntry struct {
			file     string
			comments []db.Comment
		}
		seen := map[string]int{}
		var fileEntries []fileEntry
		for _, tc := range commitTopLevel {
			if !tc.File.Valid {
				continue
			}
			f := tc.File.String
			if idx, ok := seen[f]; ok {
				fileEntries[idx].comments = append(fileEntries[idx].comments, tc)
			} else {
				seen[f] = len(fileEntries)
				fileEntries = append(fileEntries, fileEntry{file: f, comments: []db.Comment{tc}})
			}
		}

		for _, fe := range fileEntries {
			out.Printf("%s\n", fe.file)
			for _, tc := range fe.comments {
				if c.TopLevel {
					printCommentLine(out, tc, cm.Sha, "  ")
				} else {
					printFileThreadFlat(out, childrenMap, tc, cm.Sha)
				}
			}
		}
	}
	out.Printf("\n")

	return nil
}

// showThread displays a single thread (root + all descendants).
func (c *ListCmd) showThread(ctx context.Context, q *db.Queries, out *output.Output) error {
	root, err := q.FindCommentByPrefix(ctx, sql.NullString{String: c.ID, Valid: true})
	if err != nil {
		return ergo.New("comment not found", slog.String("comment_id", c.ID))
	}

	// Walk up to find the thread root
	current := root
	for current.ParentID.Valid {
		parent, err := q.GetComment(ctx, current.ParentID.UUID)
		if err != nil {
			break
		}
		current = parent
	}
	root = current

	allComments, err := q.ListAllComments(ctx)
	if err != nil {
		return ergo.Wrap(err, "failed to load comments")
	}

	childrenMap := buildChildrenMap(allComments)
	out.Printf("\n")
	printThreadFlat(out, childrenMap, root, root.Commit)
	out.Printf("\n")

	return nil
}

// filterComments applies filters, returning only matching root comments and their descendants.
// Filters are ANDed together.
func filterComments(allComments []db.Comment, commits []db.Commit, idMap map[string]db.Comment, commit string, unresolved bool, creator string, file string) []db.Comment {
	hasFilter := commit != "" || unresolved || creator != "" || file != ""
	if !hasFilter {
		return allComments
	}

	// Build commit SHA lookup for prefix matching
	var matchCommitSHA string
	if commit != "" {
		for _, cm := range commits {
			if strings.HasPrefix(cm.Sha, commit) {
				matchCommitSHA = cm.Sha
				break
			}
		}
		if matchCommitSHA == "" {
			return nil // no matching commit
		}
	}

	// Build a set of root IDs that pass filters
	rootIDs := map[string]bool{}
	for _, cm := range allComments {
		if cm.ParentID.Valid {
			continue // only filter roots
		}

		if matchCommitSHA != "" && cm.Commit != matchCommitSHA {
			continue
		}
		if unresolved && cm.ResolvedAt.Valid {
			continue
		}
		if creator != "" && cm.CreatedBy != creator {
			continue
		}
		if file != "" && (!cm.File.Valid || cm.File.String != file) {
			continue
		}

		rootIDs[cm.ID.String()] = true
	}

	// Return comments that are either matching roots or descendants of matching roots
	var result []db.Comment
	for _, cm := range allComments {
		if !cm.ParentID.Valid {
			if rootIDs[cm.ID.String()] {
				result = append(result, cm)
			}
		} else {
			root := findRoot(idMap, cm)
			if rootIDs[root.ID.String()] {
				result = append(result, cm)
			}
		}
	}

	return result
}

// buildChildrenMap builds a parentID -> children lookup for efficient tree traversal.
func buildChildrenMap(allComments []db.Comment) map[string][]db.Comment {
	m := make(map[string][]db.Comment, len(allComments))
	for _, c := range allComments {
		if c.ParentID.Valid {
			key := c.ParentID.UUID.String()
			m[key] = append(m[key], c)
		}
	}
	return m
}

// buildIDMap builds an ID -> comment lookup for efficient parent chain walking.
func buildIDMap(allComments []db.Comment) map[string]db.Comment {
	m := make(map[string]db.Comment, len(allComments))
	for _, c := range allComments {
		m[c.ID.String()] = c
	}
	return m
}

// findRoot walks up the parent chain to find the root comment.
func findRoot(idMap map[string]db.Comment, c db.Comment) db.Comment {
	current := c
	for current.ParentID.Valid {
		parent, ok := idMap[current.ParentID.UUID.String()]
		if !ok {
			break
		}
		current = parent
	}
	return current
}

// descendants returns all descendants of a comment sorted by ID (UUIDv7 = chronological).
func descendants(childrenMap map[string][]db.Comment, id fmt.Stringer) []db.Comment {
	var result []db.Comment
	queue := []string{id.String()}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, c := range childrenMap[current] {
			result = append(result, c)
			queue = append(queue, c.ID.String())
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID.String() < result[j].ID.String()
	})
	return result
}

func printThreadFlat(out *output.Output, childrenMap map[string][]db.Comment, tc db.Comment, sectionCommit string) {
	printCommentLine(out, tc, sectionCommit, "")
	for _, d := range descendants(childrenMap, tc.ID) {
		printCommentLine(out, d, sectionCommit, "  ")
	}
}

func printFileThreadFlat(out *output.Output, childrenMap map[string][]db.Comment, tc db.Comment, sectionCommit string) {
	loc := ""
	if lr := internal.FormatLineRange(tc.StartLine, tc.EndLine); lr != "" {
		loc = "L" + lr + ": "
	}
	commitTag := crossCommitTag(tc, sectionCommit)
	suffix := authorSuffix(tc.CreatedBy)
	tag := resolvedTag(tc)
	out.Printf("  [%s] %s%s%s%s%s\n", internal.ShortID(tc.ID), commitTag, loc, tc.Body, suffix, tag)

	for _, d := range descendants(childrenMap, tc.ID) {
		printCommentLine(out, d, sectionCommit, "    ")
	}
}

func printCommentLine(out *output.Output, c db.Comment, sectionCommit string, indent string) {
	commitTag := crossCommitTag(c, sectionCommit)
	suffix := authorSuffix(c.CreatedBy)
	tag := resolvedTag(c)
	out.Printf("%s[%s] %s%s%s%s\n", indent, internal.ShortID(c.ID), commitTag, c.Body, suffix, tag)
}

// resolvedTag returns a " [resolved ...]" suffix for root comments, or "" for replies/unresolved.
func resolvedTag(c db.Comment) string {
	if c.ParentID.Valid || !c.ResolvedAt.Valid {
		return ""
	}
	tag := " [resolved"
	if c.ResolvedBy.Valid {
		tag += " by " + c.ResolvedBy.String
	}
	return tag + "]"
}

func crossCommitTag(c db.Comment, sectionCommit string) string {
	if c.Commit != sectionCommit {
		return "(" + internal.ShortSHA(c.Commit) + ") "
	}
	return ""
}

func authorSuffix(author string) string {
	if author == "" {
		return ""
	}
	return " @" + author
}

