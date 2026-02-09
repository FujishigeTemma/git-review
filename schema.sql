-- 
-- PRAGMA journal_mode = WAL;
-- PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS session (
    base_ref   TEXT PRIMARY KEY,
    branch     TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS commits (
    sha      TEXT PRIMARY KEY,
    message  TEXT NOT NULL,
    position INTEGER NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS reviewers (
    name           TEXT PRIMARY KEY,
    current_sha    TEXT REFERENCES commits(sha)
);

CREATE TABLE IF NOT EXISTS comments (
    id             TEXT PRIMARY KEY,
    parent_id      TEXT REFERENCES comments(id) ON DELETE CASCADE,
    "commit"       TEXT NOT NULL REFERENCES commits(sha),
    file           TEXT,
    start_line     INTEGER,
    end_line       INTEGER,
    body           TEXT NOT NULL,
    resolved_at    TEXT,
    resolved_by    TEXT,
    created_at     TEXT NOT NULL,
    created_by     TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_comments_commit ON comments("commit");
CREATE INDEX IF NOT EXISTS idx_comments_parent ON comments(parent_id);
