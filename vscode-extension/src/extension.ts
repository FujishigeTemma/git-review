import * as vscode from "vscode";
import { createReviewManager } from "./reviewManager";
import { createCommentController } from "./commentController";
import { createStatusBar } from "./statusBar";
import type { ReviewState } from "./types";

export function activate(context: vscode.ExtensionContext) {
  const root = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
  if (!root) return;

  let manager: ReturnType<typeof createReviewManager>;
  try {
    manager = createReviewManager(root);
  } catch {
    return; // not a git repo
  }

  const commentCtrl = createCommentController(manager);
  const statusBar = createStatusBar();
  context.subscriptions.push(commentCtrl, statusBar);

  const setActive = (active: boolean) =>
    vscode.commands.executeCommand("setContext", "gitReview.active", active);

  const showList = () => {
    vscode.workspace
      .openTextDocument({ content: manager.getListMarkdown(), language: "markdown" })
      .then((doc) => vscode.window.showTextDocument(doc, { preview: true }));
  };

  const syncState = (state: ReviewState | null) => {
    const active = state !== null && manager.isActive();
    setActive(active);
    statusBar.update(state);
    if (active) {
      commentCtrl.restoreComments(state, root);
    } else {
      commentCtrl.clearThreads();
    }
  };

  manager.onDidChangeState(syncState);

  // Command registration helper
  const reg = (id: string, fn: (...args: any[]) => any) =>
    context.subscriptions.push(vscode.commands.registerCommand(id, fn));

  reg("gitReview.start", async () => {
    if (manager.isActive()) {
      vscode.commands.executeCommand("gitReview.status");
      return;
    }

    const baseRef = await vscode.window.showInputBox({
      prompt: "Base ref (leave empty to auto-detect from main/master)",
      placeHolder: "main, HEAD~5, abc1234...",
    });
    if (baseRef === undefined) return;

    try {
      const state = manager.start(baseRef || undefined);
      vscode.window.showInformationMessage(
        `Git Review: started — ${state.commits.length} commit(s)`,
      );
      vscode.commands.executeCommand("workbench.view.scm");
    } catch (e: any) {
      vscode.window.showErrorMessage(`Git Review: ${e.message}`);
    }
  });

  reg("gitReview.next", async () => {
    try {
      if (!manager.isActive()) {
        vscode.window.showWarningMessage("Git Review: No active review.");
        return;
      }
      const { finished } = manager.next();
      if (finished) {
        const action = await vscode.window.showInformationMessage(
          "Git Review: All commits reviewed.",
          "Finish Review",
          "Show Comments",
        );
        if (action === "Finish Review") {
          vscode.commands.executeCommand("gitReview.finish");
        } else if (action === "Show Comments") {
          showList();
        }
        return;
      }
      const state = manager.getState()!;
      if (state.current !== null) {
        const progress = `[${state.current + 1}/${state.commits.length}]`;
        vscode.window.showInformationMessage(`Git Review ${progress}: Staged changes ready.`);
      }
      vscode.commands.executeCommand("workbench.view.scm");
    } catch (e: any) {
      vscode.window.showErrorMessage(`Git Review: ${e.message}`);
    }
  });

  reg("gitReview.addComment", () => addComment());
  reg("gitReview.addCommentFromSelection", () => addComment(true));
  reg("gitReview.list", showList);

  reg("gitReview.jump", async () => {
    const state = manager.getState();
    if (!state) {
      vscode.window.showWarningMessage("Git Review: No active review.");
      return;
    }

    const items = state.commits.map((sha, i) => {
      const short = sha.substring(0, 7);
      const subject = manager.getCommitSubject(sha);
      const prefix = state.current !== null && i === state.current ? "-> " : "";
      return {
        label: `${prefix}[${i + 1}/${state.commits.length}] ${short}`,
        description: subject,
        sha,
      };
    });

    const pick = await vscode.window.showQuickPick(items, {
      placeHolder: "Jump to commit...",
    });
    if (!pick) return;

    try {
      manager.jump(pick.sha);
      vscode.commands.executeCommand("workbench.view.scm");
    } catch (e: any) {
      vscode.window.showErrorMessage(`Git Review: ${e.message}`);
    }
  });

  reg("gitReview.status", () => {
    const state = manager.getState();
    if (!state) {
      vscode.window.showInformationMessage("Git Review: No review data.");
      return;
    }

    const lines: string[] = ["# Git Review Status", ""];
    lines.push(`**Base:** \`${state.baseRef}\`  `);
    lines.push(`**Branch:** \`${state.branch}\`  `);
    lines.push(`**Comments:** ${state.comments.length}`, "");
    lines.push("## Commits", "");

    for (let i = 0; i < state.commits.length; i++) {
      const sha = state.commits[i];
      const short = sha.substring(0, 7);
      const subject = manager.getCommitSubject(sha);
      if (state.current !== null && i === state.current) {
        lines.push(`- -> \`${short}\` ${subject}`);
      } else if (state.current !== null && i < state.current) {
        lines.push(`- [x] \`${short}\` ${subject}`);
      } else {
        lines.push(`- [ ] \`${short}\` ${subject}`);
      }
    }

    vscode.workspace
      .openTextDocument({ content: lines.join("\n"), language: "markdown" })
      .then((doc) => vscode.window.showTextDocument(doc, { preview: true }));
  });

  reg("gitReview.finish", () => {
    try {
      const listMarkdown = manager.getListMarkdown();
      manager.finish();
      vscode.window.showInformationMessage("Git Review: Complete!");
      vscode.workspace
        .openTextDocument({ content: listMarkdown, language: "markdown" })
        .then((doc) => vscode.window.showTextDocument(doc, { preview: true }));
    } catch (e: any) {
      vscode.window.showErrorMessage(`Git Review: ${e.message}`);
    }
  });

  reg("gitReview.abort", async () => {
    const pick = await vscode.window.showWarningMessage(
      "Abort the current review? All comments will be lost.",
      { modal: true },
      "Abort",
    );
    if (pick !== "Abort") return;

    try {
      manager.abort();
      vscode.window.showInformationMessage("Git Review aborted.");
    } catch (e: any) {
      vscode.window.showErrorMessage(`Git Review: ${e.message}`);
    }
  });

  syncState(manager.getState());

  // Watch .git/HEAD for external changes
  const headWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(root, ".git/HEAD"),
  );
  headWatcher.onDidChange(() => syncState(manager.getState()));
  context.subscriptions.push(headWatcher);

  // Watch .git/review/review.db for state changes from CLI
  const dbWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(root, ".git/review/review.db"),
  );
  dbWatcher.onDidChange(() => syncState(manager.getState()));
  dbWatcher.onDidCreate(() => syncState(manager.getState()));
  dbWatcher.onDidDelete(() => syncState(manager.getState()));
  context.subscriptions.push(dbWatcher);

  async function addComment(requireSelection = false): Promise<void> {
    const editor = vscode.window.activeTextEditor;
    if (requireSelection && (!editor || editor.selection.isEmpty)) {
      vscode.window.showWarningMessage("Select some lines first.");
      return;
    }

    const file = editor ? vscode.workspace.asRelativePath(editor.document.uri) : undefined;
    const selection = editor && !editor.selection.isEmpty ? editor.selection : undefined;
    const startLine = selection ? selection.start.line + 1 : undefined;
    const endLine = selection ? selection.end.line + 1 : undefined;

    const loc = file ? formatLocation(file, startLine, endLine) : undefined;
    const message = await vscode.window.showInputBox({
      prompt: loc ? `Comment on ${loc}` : "General review comment",
      placeHolder: "Your review comment...",
    });
    if (!message) return;

    try {
      manager.addComment(message, file, startLine, endLine === startLine ? undefined : endLine);
      vscode.window.showInformationMessage(`Comment added${loc ? ` on ${loc}` : ""}`);
    } catch (e: any) {
      vscode.window.showErrorMessage(`Git Review: ${e.message}`);
    }
  }
}

function formatLocation(file: string, start?: number, end?: number): string {
  if (!start) return file;
  return end && end !== start ? `${file}:${start}-${end}` : `${file}:${start}`;
}

export function deactivate() {}
