import * as vscode from "vscode";
import { createGit } from "./git";
import { ReviewState } from "./types";

export type ReviewManager = ReturnType<typeof createReviewManager>;

export function createReviewManager(workspaceRoot: string) {
  const git = createGit(workspaceRoot);
  const stateEmitter = new vscode.EventEmitter<ReviewState | null>();

  function isActive(): boolean {
    const output = git.tryRun("review", "state");
    return output !== null && output.trim() !== "null";
  }

  function getState(): ReviewState | null {
    try {
      const output = git.review("state");
      const parsed = JSON.parse(output);
      if (parsed === null) return null;
      return parsed as ReviewState;
    } catch {
      return null;
    }
  }

  function fireState(): void {
    stateEmitter.fire(getState());
  }

  function runCLI(...args: string[]): void {
    try {
      git.review(...args);
    } catch (e: any) {
      throw new Error(e.stderr?.trim() || e.message);
    }
  }

  return {
    onDidChangeState: stateEmitter.event,
    isActive,
    getState,

    start(baseRef?: string): ReviewState {
      runCLI("start", ...(baseRef ? [baseRef] : []));
      const state = getState()!;
      stateEmitter.fire(state);
      return state;
    },

    next(): { finished: boolean } {
      if (!isActive()) throw new Error("No active review.");
      const before = getState();
      runCLI("next");
      const after = getState();
      if (!after) {
        stateEmitter.fire(null);
        return { finished: true };
      }
      // CLI next at end: position unchanged
      const atEnd =
        before !== null &&
        after.current !== null &&
        before.current === after.current &&
        after.current === after.commits.length - 1;
      stateEmitter.fire(after);
      return { finished: atEnd };
    },

    jump(hash: string): void {
      runCLI("jump", hash);
      fireState();
    },

    resolveComment(id: string): void {
      runCLI("resolve", id);
      fireState();
    },

    unresolveComment(id: string): void {
      runCLI("unresolve", id);
      fireState();
    },

    getCommitSubject(sha: string): string {
      return git.tryRun("log", "--format=%s", "-1", sha) ?? sha.substring(0, 7);
    },

    addComment(body: string, file?: string, startLine?: number, endLine?: number): void {
      const args: string[] = [];
      if (file) args.push("-f", file);
      if (startLine !== undefined) {
        args.push(
          "-l",
          endLine !== undefined && endLine !== startLine
            ? `${startLine},${endLine}`
            : String(startLine),
        );
      }
      args.push(body);
      runCLI("add", ...args);
      fireState();
    },

    replyToComment(parentId: string, body: string): void {
      runCLI("add", "--reply-to", parentId, body);
      fireState();
    },

    deleteComment(id: string): void {
      runCLI("delete", id);
      fireState();
    },

    finish(): void {
      if (!getState()) return;
      runCLI("finish");
      stateEmitter.fire(null);
    },

    abort(): void {
      runCLI("abort");
      stateEmitter.fire(null);
    },

    getListMarkdown(): string {
      try {
        return git.review("list");
      } catch {
        return "# No review data found.\n";
      }
    },
  };
}
