# git-review

Commit review workflow for AI Agent collaboration.

AI Agentがブランチ上でTDD実装・コミットした後、AI Agentチームが役割別（security, architecture等）に1コミットずつレビュー。必要に応じて人間が追加レビュー。AI Agentはコメントを読み改善を行う。

## Architecture

```
git-review (Go CLI)       ←→  .git/review/comments.json  ←→  VSCode Extension
  git review                                                    Command Palette
  git review add "msg"                                          Inline Comments (GitHub-like)
  git review next                                               Status Bar
  git review list
```

CLI と VSCode Extension は `.git/review/` ディレクトリを共有し、どちらからでも操作可能。

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

## Workflow

```
AI Agent: TDDで実装、feature branchに複数コミットを作成
    ↓
AI Agentチーム: git review で役割別にレビュー
  (例: -a security, -a architecture, -a code-quality)
    ↓
(任意) Human: CLI / VSCode Extensionで追加レビュー
    ↓
(任意) AI Agentチーム: 人間のフィードバックを踏まえ再レビュー
    ↓
AI Agent: git review list でコメントを確認、改善を実施
```

CLIコマンドの詳細・レビューワークフローガイドは `git review skill` (= [SKILL.md](SKILL.md)) を参照。

## VSCode Extension

See [vscode-extension/README.md](vscode-extension/README.md)

## Development

### Try it locally

#### CLI

```bash
go install .                     # install from source
./create-test-repo.sh    # テスト用リポを ./test-repo/ に作成

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

F5 で Extension Development Host を起動し、`./test-repo/` を開いて Command Palette から操作。

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

## How It Works

核心は `git cherry-pick --no-commit` です。各コミットの差分（親→そのコミット）をstagedな状態で適用し、VSCode上でdiffとして確認可能にします。レビュー中にコードを修正した場合、次のcherry-pickでconflictが出るため、そこで整合性を解決する流れになります。

レビュー完了時（`finish`）にはコメントがgit notesとして元コミットに書き込まれ、ブランチやリポジトリの履歴に残ります。

## License

MIT
