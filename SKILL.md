# git-review: AI Agent Workflow Guide

## Workflow Overview

```
AI Agent: Implement with TDD, create commits on a feature branch
    ↓
AI Agent team leader: Delegate review to teammates
    ↓
AI Agent teammates: Each starts review independently
  - Each gets own worktree, navigates commits freely
  - Blind Review: Review all commits from own perspective
  - Cross-Review: Read others' comments, reply to discuss
    ↓
(Optional) Human: Additional review via CLI or VSCode Extension
    ↓
AI Agent: Read comments, make improvements, resolve addressed threads
    ↓
Leader: git review finish
```

## Reviewer Agent Guide

### Starting a Review

```bash
git review start main -a security       # explicit base ref
git review start -a architecture        # auto-detect base (main/master/develop)
git review start HEAD~5 -a performance  # review last 5 commits
git review start main                   # single reviewer (no worktree, checkout in current tree)
```

`-a <role>` creates a worktree at `.git/review/worktrees/<role>/`. The role name must be unique across the review session — starting with a role already in use is an error.

Output:

```
Review started: feature/new-auth (3 commits)
  abc1234 Add user model
  def5678 Add auth endpoint
  ghi9012 Add rate limiting
Worktree: .git/review/worktrees/security/
```

### Navigating Commits

```bash
git review next          # move to next commit (changes shown as staged)
git review jump abc1234  # jump to specific commit (hash prefix)
git review status        # show progress: current position, comment counts
```

`next` and `jump` set the worktree to the target commit's state, with the commit's changes visible as staged changes (`git diff --staged`). This prints commit info:

```
Commit 2/3: def5678 Add auth endpoint
Files changed:
  M handlers/auth.go
  A handlers/auth_test.go
```

On the last commit, `next` prints a summary instead of advancing:

```
All commits reviewed. 5 comments added.
Use `jump <hash>` to revisit, or `list` to see all comments.
```

### Adding Comments

Comments are always attached to the reviewer's current commit:

```bash
# General comment on the current commit
git review add "Overall approach looks good"

# File-specific comment
git review add -f src/auth.ts "Missing error handling"

# Line-specific comment
git review add -f src/auth.ts -l 42 "Use bcrypt instead of md5"

# Range-specific comment
git review add -f src/api.ts -l 10,25 "Split this function"
```

### Replying to Comments

Reply to create threaded discussions:

```bash
git review add -r <comment-id> "Agreed, also consider using argon2"
```

ID prefix matching is supported (e.g. `-r 019516c0` matches full UUID).

### Viewing Comments

```bash
git review list                             # all comments across all commits
git review list <id>                        # show a specific thread (walks up to root)
git review list --commit abc1234            # filter by commit (hash prefix)
git review list --unresolved                # show only unresolved threads
git review list --creator security          # filter by creator role
git review list --file src/auth.ts          # filter by file path
git review list --top-level                 # show only top-level comments (no replies)
```

Filters can be combined (ANDed together):

```bash
git review list --commit abc1234 --unresolved              # unresolved on a specific commit
git review list --creator security --file src/auth.ts      # security comments on a file
```

### Deleting Comments

```bash
git review delete <id>    # ID prefix match supported
```

Delete behavior:

- **Hard delete**: the comment is removed from the database
- **Non-root comment deleted**: children are re-parented to the deleted comment's parent
- **Root comment deleted** (`parentId` is `null`): the entire thread is deleted (all descendants cascade)

### Example Review Perspectives

| Role           | Focus                                                                |
| -------------- | -------------------------------------------------------------------- |
| `security`     | Authentication, authorization, input validation, secrets, injection  |
| `architecture` | Layering, separation of concerns, dependency direction, patterns     |
| `code-quality` | Readability, naming, duplication, complexity, test coverage          |
| `performance`  | Algorithmic complexity, memory allocation, caching, query efficiency |
| `api-design`   | Interface consistency, backwards compatibility, error handling       |

## Multi-Reviewer Workflow (Agent Teams)

### Recommended Flow

```bash
# === Implementation Phase ===
# AI Agent creates feature branch with TDD commits
# c1: abc1234 Add user model + tests
# c2: def5678 Add auth endpoint + tests
# c3: ghi9012 Add rate limiting + tests

# === Leader delegates review via Agent Teams messaging ===
# "Review feature/new-auth against main.
#  Security and architecture perspectives."

# === Blind Review (parallel, independent) ===
# Each reviewer works through all commits from their own perspective.
# Do NOT read others' comments during this phase (avoid anchoring bias).

# Security Reviewer
git review start main -a security
git review next
# Commit 1/3: abc1234 Add user model
# (worktree checked out at this commit — full codebase available)
git review add -f models/user.go -l 15 "Hash password before storing"
git review next
# Commit 2/3: def5678 Add auth endpoint
git review add -f handlers/auth.go -l 30 "Add CSRF token validation"
git review add -f handlers/auth.go -l 45,52 "Rate limit login attempts"
git review next
# Commit 3/3: ghi9012 Add rate limiting
git review add "Rate limiter implementation looks solid"
git review next
# All commits reviewed. 4 comments added.

# Architecture Reviewer (parallel, in own worktree)
git review start main -a architecture
git review next
git review add -f models/user.go "Consider separating domain model from DB model"
git review next
git review add "Clean layering, good separation of concerns"
git review next
git review add -f middleware/ratelimit.go -l 8,20 "Extract config to environment"
git review next
# All commits reviewed. 3 comments added.

# === Cross-Review (parallel, collaborative) ===
# After completing own review, read others' comments and discuss via reply.

# Security Reviewer reads architecture comments
git review list
# Sees architecture's comment on models/user.go
git review add -r <arch-comment-id> "Good point, separate model also helps with input sanitization"
# Jump back to re-examine
git review jump ghi9012
git review add -f middleware/ratelimit.go -l 12 "Config values need validation too"

# Architecture Reviewer reads security comments
git review list
git review add -r <security-comment-id> "Agreed, also consider argon2 over bcrypt"

# === Notify leader via Agent Teams messaging ===
# "Review complete."

# === (Optional) Human Review via VSCode or CLI ===

# === Improvement Phase ===
# Implementer agent reads all feedback:
git review list
# Address each comment, reply to acknowledge
git review add -r <comment-id> -a implementer "Fixed: switched to argon2"
# Resolve addressed threads
git review resolve <comment-id>
# Commit fixes on the same branch

# === Leader finalizes ===
git review finish
```

### Guidelines for Agent Teams

1. **Blind Review first**: Complete your own review of all commits before running `git review list`. This prevents anchoring bias and ensures independent perspectives.

2. **Reply for code discussion**: Any discussion that the implementer should read belongs in replies (`git review add -r`), not in Agent Teams direct messages. Replies are persistent, attached to code, and part of the review record.

3. **Direct messaging for coordination only**: Use Agent Teams messaging to tell the leader "review complete" or "found critical blocker". Do not discuss code via messages.

4. **One role per reviewer**: Each reviewer should focus on a single perspective for clear separation of concerns.

5. **Use `jump` for re-examination**: During Cross-Review, use `jump <hash>` to revisit commits where other reviewers left interesting comments.

## Implementer Agent Guide

### Reading Review Feedback

```bash
git review list
```

Output (Markdown format):

```markdown
# Review Comments

Branch: feature/new-auth
Commits: 3

---

## Commit 1/3 abc1234: Add user authentication

[019516c0] Overall approach looks solid @security
  [019516c1] Thanks! @implementer
src/auth.ts
  [019516c2] L42: Use bcrypt instead of md5 @security
    [019516c3] Agreed, also consider argon2 @architecture
    [019516c4] Fixed: switched to argon2 @implementer
  [019516c5] Move validation to domain layer @architecture

---

## Commit 2/3 def5678: Add database schema

No comments
```

### Making Improvements

1. Read all comments: `git review list`
2. Address each comment by modifying the relevant file/line
3. Commit fixes on the same branch
4. Reply to comments acknowledging fixes: `git review add -r <id> -a implementer "Fixed"`

## CLI Quick Reference

| Command                                                | Description                                          |
| ------------------------------------------------------ | ---------------------------------------------------- |
| `git review start [base-ref] [-a role]`                | Start review (creates worktree if `-a` specified)    |
| `git review next`                                      | Move to next commit                                  |
| `git review jump <hash>`                               | Jump to specific commit                              |
| `git review add [-a author] [-f file] [-l line] "msg"` | Add comment                                          |
| `git review add -r <id> "msg"`                         | Reply to comment                                     |
| `git review list [<id>] [flags]`                       | Show comments (flags: `--commit`, `--unresolved`, `--creator`, `--file`, `--top-level`) |
| `git review status`                                    | Show review progress                                 |
| `git review delete <id>`                               | Delete comment (hard delete; re-parents or cascades) |
| `git review resolve [-a who] <id>`                     | Resolve a thread (root comment only)                 |
| `git review unresolve <id>`                            | Unresolve a thread                                   |
| `git review finish`                                    | Finish review, write git notes, clean up             |
| `git review abort`                                     | Cancel review, clean up                              |
| `git review state`                                     | Output review state as JSON (for VSCode extension)   |
| `git review skill`                                     | Show this guide                                      |

## Concepts

### Worktrees

Each reviewer gets their own git worktree — a full copy of the repository linked to the same `.git` directory. This allows:

- **Independent navigation**: each reviewer checks out different commits without affecting others
- **Full codebase context**: reviewers can open any file, trace dependencies, understand architecture at each commit's state
- **Parallel work**: no shared mutable state for navigation

Worktrees are stored in `.git/review/worktrees/<role>/` and managed by git-review.

### Shared Comments

While worktrees are independent, `review.db` (SQLite WAL mode) is shared across all reviewers via the common `.git/review/` directory.

### Communication Channels

| Channel                            | Use for                                    | Properties                                                       |
| ---------------------------------- | ------------------------------------------ | ---------------------------------------------------------------- |
| **Reply** (`git review add -r`)    | Code discussion tied to specific locations | Persistent, visible to implementer, attached to commit/file/line |
| **Direct messaging** (Agent Teams) | Coordination with team leader              | Ephemeral, not part of review record                             |

Use reply for any discussion that the implementer should read. Use direct messaging only for coordination ("review complete", "found critical blocker").

## Data Storage

### Review Session

```
.git/review/
├── review.db             # SQLite (WAL mode): session, commits, reviewers, comments
└── worktrees/
    ├── security/         # git worktree for security reviewer
    └── architecture/     # git worktree for architecture reviewer
```

On finish, comments are written to git notes, worktrees are removed via `git worktree remove`, `review.db` is closed, and `.git/review/` is deleted.

### SQLite Schema

```sql
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE session (
    base_ref   TEXT PRIMARY KEY,
    branch     TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE commits (
    sha      TEXT PRIMARY KEY,
    message  TEXT NOT NULL,
    position INTEGER NOT NULL UNIQUE  -- 0-based display order
);

CREATE TABLE reviewers (
    name           TEXT PRIMARY KEY,
    current_sha    TEXT REFERENCES commits(sha)
);

CREATE TABLE comments (
    id             TEXT PRIMARY KEY,
    parent_id      TEXT REFERENCES comments(id) ON DELETE CASCADE,
    commit         TEXT NOT NULL REFERENCES commits(sha),
    file           TEXT,
    start_line     INTEGER,
    end_line       INTEGER,
    body           TEXT NOT NULL,
    resolved_at    TEXT,              -- NULL = unresolved, ISO 8601 = resolved
    resolved_by    TEXT,
    created_at     TEXT NOT NULL,
    created_by     TEXT NOT NULL
);

CREATE INDEX idx_comments_commit ON comments(commit);
CREATE INDEX idx_comments_parent ON comments(parent_id);
```

| Column        | Type              | Description                                          |
| ------------- | ----------------- | ---------------------------------------------------- |
| `id`          | `TEXT`            | UUID v7 identifier                                   |
| `parent_id`   | `TEXT \| NULL`    | Parent comment ID for replies. Top-level is `NULL`   |
| `commit`      | `TEXT`            | Full SHA of the reviewed commit                      |
| `file`        | `TEXT \| NULL`    | Workspace-relative path. `NULL` for general comments |
| `start_line`  | `INTEGER \| NULL` | 1-based start line. `NULL` if no line specified      |
| `end_line`    | `INTEGER \| NULL` | 1-based end line (inclusive). `NULL` if no range     |
| `body`        | `TEXT`            | Comment body text                                    |
| `resolved_at` | `TEXT \| NULL`    | `NULL` = unresolved, ISO 8601 = resolved             |
| `resolved_by` | `TEXT \| NULL`    | Who resolved the thread. `NULL` if unresolved        |
| `created_at`  | `TEXT`            | ISO 8601 creation timestamp                          |
| `created_by`  | `TEXT`            | Reviewer role name                                   |

Key fields for targeted improvements:

- `file` + `start_line`/`end_line`: exact location to fix
- `commit`: which commit the comment refers to
- `created_by`: which review perspective raised the issue
- `parent_id`: if non-null, this is a reply in a thread

### Delete Behavior

- **Hard delete**: comment is removed from the database
- **Non-root** (`parent_id` is non-null): children are re-parented (`parent_id` set to deleted comment's `parent_id`)
- **Root** (`parent_id` is `NULL`): entire thread is deleted (all descendants removed via CASCADE)
