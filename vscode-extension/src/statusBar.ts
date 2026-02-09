import * as vscode from "vscode";
import { ReviewState } from "./types";

export function createStatusBar() {
  const item = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
  item.command = "gitReview.next";

  return {
    update(state: ReviewState | null): void {
      if (!state) {
        item.hide();
        return;
      }
      const total = state.commits.length;
      const comments = state.comments.length;

      if (state.current === null) {
        item.text = `$(git-compare) Review (no position)${comments > 0 ? ` (${comments} comments)` : ""}`;
        item.tooltip = `Git Review: No position\nComments: ${comments}\n\nClick to advance to next commit`;
      } else {
        const n = state.current + 1;
        item.text = `$(git-compare) Review ${n}/${total}${comments > 0 ? ` (${comments} comments)` : ""}`;
        item.tooltip = `Git Review: Commit ${n} of ${total}\nComments: ${comments}\n\nClick to advance to next commit`;
      }
      item.show();
    },
    dispose(): void {
      item.dispose();
    },
  };
}
