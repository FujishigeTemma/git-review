-- Session

-- name: InsertSession :exec
INSERT INTO session (base_ref, branch, created_at) VALUES (?, ?, ?);

-- name: GetSession :one
SELECT base_ref, branch, created_at FROM session LIMIT 1;

-- name: SessionExists :one
SELECT COUNT(*) FROM session;

-- name: DeleteSession :exec
DELETE FROM session;

-- Commits

-- name: InsertCommit :exec
INSERT INTO commits (sha, message, position) VALUES (?, ?, ?);

-- name: ListCommits :many
SELECT sha, message, position FROM commits ORDER BY position;

-- name: GetCommitByPosition :one
SELECT sha, message, position FROM commits WHERE position = ?;

-- name: GetCommitBySHA :one
SELECT sha, message, position FROM commits WHERE sha = ?;

-- name: FindCommitBySHAPrefix :one
SELECT sha, message, position FROM commits WHERE sha LIKE ?||'%';

-- name: CountCommits :one
SELECT COUNT(*) FROM commits;

-- name: DeleteCommits :exec
DELETE FROM commits;

-- Reviewers

-- name: InsertReviewer :exec
INSERT INTO reviewers (name, current_sha) VALUES (?, ?);

-- name: GetReviewer :one
SELECT name, current_sha FROM reviewers WHERE name = ?;

-- name: ListReviewers :many
SELECT name, current_sha FROM reviewers;

-- name: UpdateReviewerCurrent :exec
UPDATE reviewers SET current_sha = ? WHERE name = ?;

-- name: DeleteReviewers :exec
DELETE FROM reviewers;

-- Comments

-- name: InsertComment :exec
INSERT INTO comments (id, parent_id, "commit", file, start_line, end_line, body, resolved_at, resolved_by, created_at, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetComment :one
SELECT id, parent_id, "commit", file, start_line, end_line, body, resolved_at, resolved_by, created_at, created_by
FROM comments WHERE id = ?;

-- name: FindCommentByPrefix :one
SELECT id, parent_id, "commit", file, start_line, end_line, body, resolved_at, resolved_by, created_at, created_by
FROM comments WHERE id LIKE ?||'%';

-- name: ListAllComments :many
SELECT id, parent_id, "commit", file, start_line, end_line, body, resolved_at, resolved_by, created_at, created_by
FROM comments;

-- name: ListCommentsByCommit :many
SELECT id, parent_id, "commit", file, start_line, end_line, body, resolved_at, resolved_by, created_at, created_by
FROM comments WHERE "commit" = ?;

-- name: ReparentChildren :exec
UPDATE comments SET parent_id = ? WHERE parent_id = ?;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = ?;

-- name: DeleteAllComments :exec
DELETE FROM comments;

-- Resolve

-- name: ResolveComment :exec
UPDATE comments SET resolved_at = ?, resolved_by = ? WHERE id = ? AND parent_id IS NULL;

-- name: UnresolveComment :exec
UPDATE comments SET resolved_at = NULL, resolved_by = NULL WHERE id = ?;

-- Filtered list queries

-- name: ListUnresolvedRoots :many
SELECT id, parent_id, "commit", file, start_line, end_line, body, resolved_at, resolved_by, created_at, created_by
FROM comments WHERE parent_id IS NULL AND resolved_at IS NULL;

-- name: ListCommentsByCreator :many
SELECT id, parent_id, "commit", file, start_line, end_line, body, resolved_at, resolved_by, created_at, created_by
FROM comments WHERE created_by = ?;

-- name: ListCommentsByFile :many
SELECT id, parent_id, "commit", file, start_line, end_line, body, resolved_at, resolved_by, created_at, created_by
FROM comments WHERE file = ?;
