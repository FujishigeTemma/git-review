import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { ReviewManager } from "./reviewManager";
import { ReviewComment, ReviewState } from "./types";

export function createCommentController(manager: ReviewManager) {
  const controller = vscode.comments.createCommentController("git-review-comments", "Git Review");
  const threads = new Map<string, vscode.CommentThread>();
  const threadToId = new Map<vscode.CommentThread, string>();

  controller.commentingRangeProvider = {
    provideCommentingRanges: (doc: vscode.TextDocument) =>
      manager.isActive() ? [new vscode.Range(0, 0, doc.lineCount - 1, 0)] : [],
  };

  controller.options = {
    placeHolder: "Add review comment...",
    prompt: "Add Comment",
  };

  const commandDisposables = [
    vscode.commands.registerCommand("gitReview.createNote", (reply: vscode.CommentReply) => {
      const ws = vscode.workspace.workspaceFolders?.[0];
      if (!ws) return;

      const thread = reply.thread;

      // Determine if this is a reply to an existing comment or a new thread
      const existingComments = thread.comments as NoteComment[];
      if (existingComments.length > 0 && existingComments[0].commentId) {
        // Reply to existing thread — find the top-level comment ID
        const topComment = existingComments.find((c) => c.isTopLevel);
        const parentId = topComment?.commentId ?? existingComments[0].commentId;
        try {
          manager.replyToComment(parentId, reply.text);
        } catch (e: any) {
          vscode.window.showErrorMessage(`Git Review: ${e.message}`);
        }
      } else {
        // New thread
        const file = path.relative(ws.uri.fsPath, thread.uri.fsPath);
        const start = thread.range ? thread.range.start.line + 1 : 1;
        const end = thread.range ? thread.range.end.line + 1 : 1;

        thread.dispose();
        try {
          manager.addComment(reply.text, file, start, end === start ? undefined : end);
        } catch (e: any) {
          vscode.window.showErrorMessage(`Git Review: ${e.message}`);
        }
      }
    }),

    vscode.commands.registerCommand("gitReview.deleteNote", (comment: NoteComment) => {
      if (!comment.commentId) return;
      try {
        manager.deleteComment(comment.commentId);
      } catch (e: any) {
        vscode.window.showErrorMessage(`Git Review: ${e.message}`);
      }
    }),

    vscode.commands.registerCommand("gitReview.resolveThread", (thread: vscode.CommentThread) => {
      const id = threadToId.get(thread);
      if (!id) return;
      try {
        manager.resolveComment(id);
      } catch (e: any) {
        vscode.window.showErrorMessage(`Git Review: ${e.message}`);
      }
    }),

    vscode.commands.registerCommand("gitReview.unresolveThread", (thread: vscode.CommentThread) => {
      const id = threadToId.get(thread);
      if (!id) return;
      try {
        manager.unresolveComment(id);
      } catch (e: any) {
        vscode.window.showErrorMessage(`Git Review: ${e.message}`);
      }
    }),
  ];

  function restoreComments(state: ReviewState, workspaceRoot: string): void {
    clearThreads();

    if (state.current === null) return;

    const currentCommit = state.commits[state.current];

    // Build a map from comment ID to root parent ID for all comments
    const commentMap = new Map<string, ReviewComment>();
    for (const c of state.comments) {
      commentMap.set(c.id, c);
    }

    function getRootParentId(c: ReviewComment): string {
      let current = c;
      while (current.parentId !== null) {
        const parent = commentMap.get(current.parentId);
        if (!parent) break;
        current = parent;
      }
      return current.id;
    }

    // Get top-level comments for current commit that have a file
    const currentTopLevel = state.comments.filter(
      (c) => c.commit === currentCommit && c.file !== null && c.parentId === null,
    );

    // Build reply map keyed by root parent ID, collecting all descendants regardless of commit
    const replyMap = new Map<string, ReviewComment[]>();
    for (const c of state.comments) {
      if (c.parentId === null) continue;
      const rootId = getRootParentId(c);
      // Only include if the root parent is one of our current top-level comments
      if (!currentTopLevel.some((tl) => tl.id === rootId)) continue;
      const replies = replyMap.get(rootId) ?? [];
      replies.push(c);
      replyMap.set(rootId, replies);
    }

    for (const tc of currentTopLevel) {
      const start = tc.startLine ?? 0;
      const end = tc.endLine ?? start;

      // Skip if file doesn't exist on disk
      const filePath = path.join(workspaceRoot, tc.file!);
      if (!fs.existsSync(filePath)) continue;

      const uri = vscode.Uri.file(filePath);
      const range = new vscode.Range(Math.max(start - 1, 0), 0, Math.max(end - 1, 0), 0);

      const thread = controller.createCommentThread(uri, range, []);
      thread.canReply = true;
      thread.label = formatLabel(tc.file!, start, end);
      thread.contextValue = tc.resolvedAt ? "reviewThread-resolved" : "reviewThread-unresolved";

      const notes: NoteComment[] = [
        createNote(
          tc.body,
          vscode.CommentMode.Preview,
          thread,
          tc.id,
          true,
          tc.createdAt,
          tc.createdBy,
        ),
      ];
      const replies = replyMap.get(tc.id) ?? [];
      for (const r of replies) {
        notes.push(
          createNote(
            r.body,
            vscode.CommentMode.Preview,
            thread,
            r.id,
            false,
            r.createdAt,
            r.createdBy,
          ),
        );
      }
      thread.comments = notes;

      threads.set(tc.id, thread);
      threadToId.set(thread, tc.id);
    }
  }

  function clearThreads(): void {
    for (const t of threads.values()) t.dispose();
    threads.clear();
    threadToId.clear();
  }

  return {
    restoreComments,
    clearThreads,
    dispose(): void {
      clearThreads();
      controller.dispose();
      commandDisposables.forEach((d) => d.dispose());
    },
  };
}

function formatLabel(file: string, start: number, end: number): string {
  if (!start) return `Review: ${file}`;
  const loc = end && end !== start ? `${file}:${start}-${end}` : `${file}:${start}`;
  return `Review: ${loc}`;
}

interface NoteComment extends vscode.Comment {
  commentId: string;
  isTopLevel: boolean;
  parent?: vscode.CommentThread;
}

function createNote(
  body: string,
  mode: vscode.CommentMode,
  parent: vscode.CommentThread,
  commentId: string,
  isTopLevel: boolean,
  createdAt: string,
  createdBy: string,
): NoteComment {
  return {
    body,
    mode,
    author: { name: createdBy || "Reviewer" },
    contextValue: isTopLevel ? "reviewNote" : "reviewReply",
    parent,
    commentId,
    isTopLevel,
    timestamp: new Date(createdAt),
  };
}
