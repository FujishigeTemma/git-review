/** A single review comment stored in review.db. */
export interface ReviewComment {
  /** UUID v7 identifier. */
  id: string;
  /** Parent comment ID for replies, or null for top-level comments. */
  parentId: string | null;
  /** Full SHA of the original commit being reviewed. */
  commit: string;
  /** Workspace-relative file path, or null for general comments. */
  file: string | null;
  /** 1-based start line, or null if not line-specific. */
  startLine: number | null;
  /** 1-based end line (inclusive), or null. */
  endLine: number | null;
  /** The review comment body. */
  body: string;
  /** ISO 8601 resolve timestamp, or null if unresolved. */
  resolvedAt: string | null;
  /** Who resolved the thread, or null if unresolved. */
  resolvedBy: string | null;
  /** ISO 8601 creation timestamp. */
  createdAt: string;
  /** Creator name (reviewer role). */
  createdBy: string;
}

/** In-memory representation of the full review state. */
export interface ReviewState {
  baseRef: string;
  branch: string;
  commits: string[];
  current: number | null;
  comments: ReviewComment[];
}
