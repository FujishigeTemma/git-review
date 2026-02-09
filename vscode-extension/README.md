# Git Review

A VSCode extension for commit review workflows with AI Agent teams.

## Features

- **Inline Comments** — Add comments on specific lines/ranges via the gutter `+` icon, GitHub PR-style
- **Status Bar** — Progress display in `Review 2/5 (3 💬)` format. Click to advance to the next commit
- **Context Menu** — Right-click with text selected to `Add Comment on Selection`
- **Command Palette** — Full access to all operations: start, comment, advance, abort

## Commands

| Command                                | Description                                                                                                    | Available When              |
| -------------------------------------- | -------------------------------------------------------------------------------------------------------------- | --------------------------- |
| `Git Review: Start Review`             | Start a review. Specify base ref (leave empty to auto-detect from main/master/develop)                         | When no review is active    |
| `Git Review: Next Commit`              | Advance to next commit (CLI auto-commits staged changes). Auto-finishes and returns to original branch on last | During review               |
| `Git Review: Add Comment`              | Add a comment on the current file/selection                                                                    | During review               |
| `Git Review: Add Comment on Selection` | Add a comment on the selected text range (via context menu)                                                    | During review, text selected |
| `Git Review: Show All Comments`        | Display all comments in Markdown                                                                               | Always                      |
| `Git Review: Show Status`              | Show progress (reviewed commit count, comment count)                                                           | During review               |
| `Git Review: Abort Review`             | Abort the review. Returns to the original branch and deletes the review branch                                 | During review               |

## UI Elements

### Status Bar

During a review, progress is displayed on the left side of the status bar.

```
$(git-compare) Review 2/5 (3 💬)
```

- **Display**: Current commit number / total commits, comment count
- **Click action**: Executes `Next Commit`
- **Tooltip**: Detailed commit progress and comment count

### Inline Comments

GitHub PR-style inline comment functionality using the VSCode Comments API.

- During review, a `+` icon appears in the gutter of all files
- Click to open a comment input field, specifying line/range for the comment
- Added comments are displayed as threads
- Comments are automatically restored for the relevant commit when switching commits

### Context Menu

During a review, the following items are added to the editor right-click menu:

- **Add Comment on Selection** — Shown only when text is selected
- **Add Comment** — Always shown (during review)

### Comment Highlight

Customize the background color of commented lines.

| Color ID                       | Description                                   | Default (dark) | Default (light) |
| ------------------------------ | --------------------------------------------- | --------------- | ---------------- |
| `gitReview.commentHighlight`   | Background highlight for commented lines      | `#ffd70020`     | `#ffd70030`      |

Example configuration (`settings.json`):

```json
{
    "workbench.colorCustomizations": {
        "gitReview.commentHighlight": "#ff000030"
    }
}
```

## Architecture: CLI Delegation

This extension uses the Go CLI tool `git review` as its backend. All review operations (start/next/finish/abort) are delegated to `git review <cmd>`, while the extension handles UI display and comment management.

```
VSCode Extension                    git review CLI
┌─────────────────┐               ┌──────────────────┐
│ Command Palette  │──start──────→│ git review start  │
│ Status Bar       │──next───────→│ git review next   │
│ Inline Comments  │──finish─────→│ git review finish │
│                  │──abort──────→│ git review abort  │
└────────┬────────┘               └──────────────────┘
         │                                  │
         │  UI state read                   │  git ops
         ▼                                  ▼
    .git/review/                   cherry-pick, notes,
    (shared state)                 branch management
```

Install the CLI:

```bash
go install github.com/FujishigeTemma/git-review@latest
```

## Development

```bash
cd vscode-extension
pnpm install

pnpm run dev             # watch mode (development build)
pnpm run build           # production build
pnpm run test            # vitest
pnpm run lint            # oxlint
pnpm run format          # oxfmt
pnpm run check           # lint + format + test (all checks)
pnpm run package         # create .vsix package
```

Launch Extension Development Host with F5, open a test repository, and use the Command Palette.

## Requirements

- VSCode ^1.85.0
- Git
- `git review` CLI (`go install github.com/FujishigeTemma/git-review@latest`)

## License

MIT
