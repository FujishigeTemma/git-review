# git-review

Commit review workflow for AI Agent collaboration.

After an AI Agent implements features with TDD and commits to a branch, an AI Agent team reviews each commit by role (security, architecture, etc.). Humans can add additional reviews as needed. The AI Agent then reads the comments and makes improvements.

## Install

### CLI

```bash
go install github.com/FujishigeTemma/git-review@latest
```

### VSCode Extension

```bash
cd vscode-extension
pnpm install
pnpm run build

# Development: F5 to launch Extension Development Host
# Package: pnpm run package
```

See [vscode-extension/README.md](vscode-extension/README.md) for details.

## Workflow

```
AI Agent: Implement with TDD, create multiple commits on a feature branch
    |
AI Agent team: Review by role using git review
  (e.g., -a security, -a architecture, -a code-quality)
    |
(Optional) Human: Additional review via CLI / VSCode Extension
    |
(Optional) AI Agent team: Re-review incorporating human feedback
    |
AI Agent: Check comments with git review list, make improvements
```

See `git review skill` (= [SKILL.md](SKILL.md)) for detailed CLI commands and review workflow guide.

## Architecture

```
git-review (Go CLI)       <->  .git/review/review.db  <->  VSCode Extension
  git review                                                 Command Palette
  git review add "msg"                                       Inline Comments (GitHub-like)
  git review next                                            Status Bar
  git review list
```

The CLI and VSCode Extension share the `.git/review/` directory, allowing operations from either side.

## How It Works

When navigating commits (`next`, `jump`), git-review checks out the target commit's parent and then applies the target commit's tree via `git read-tree`, making the commit's changes visible as staged diffs. This gives reviewers full codebase context at each commit's state while showing the diff cleanly in editors.

Each reviewer with `-a <role>` gets their own git worktree, enabling independent parallel navigation. Review state and comments are stored in a shared SQLite database (WAL mode) at `.git/review/review.db`.

When the review is finished (`git review finish`), comments are written as git notes on the original commits, worktrees are removed, and `.git/review/` is cleaned up.

## Development

### Try it locally

#### CLI

```bash
go install .              # install from source
./create-test-repo.sh    # create test repo at ./test-repo/

cd test-repo
git review                       # start review
git review add "Good"            # add comment
git review next                  # next commit
git review list                  # show comments
git review abort                 # abort review
```

#### VSCode Extension

```bash
cd vscode-extension
pnpm install && pnpm run dev     # build (watch mode)
```

Launch Extension Development Host with F5, open `./test-repo/`, and use the Command Palette.

### Tests

```bash
# CLI
go test ./internal/...   # unit tests
go test ./tests/...      # e2e tests

# CLI: build
go build -o git-review .

# VSCode Extension
cd vscode-extension
pnpm install
pnpm run build           # production build
pnpm run dev             # watch mode
pnpm run lint            # oxlint (type-check included)
pnpm run format          # oxfmt
pnpm run test            # vitest
pnpm run check           # lint + format + test
pnpm run package         # create .vsix
```

## License

MIT
