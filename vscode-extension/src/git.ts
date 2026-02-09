import { execaSync } from "execa";

export type Git = ReturnType<typeof createGit>;

export function createGit(cwd: string) {
  const run = (...args: string[]): string =>
    execaSync("git", args, { cwd, timeout: 15_000 }).stdout;

  const tryRun = (...args: string[]): string | null => {
    try {
      return run(...args);
    } catch {
      return null;
    }
  };

  return {
    run,
    tryRun,
    gitDir: (): string => run("rev-parse", "--git-dir"),
    review: (...args: string[]): string => run("review", ...args),
  };
}

// ──────────── In-source tests ────────────

if (import.meta.vitest) {
  const { describe, it, expect, vi, beforeEach } = import.meta.vitest;

  vi.mock("execa", () => ({
    execaSync: vi.fn(),
  }));

  const mockedExecaSync = vi.mocked(execaSync);

  describe("createGit", () => {
    let git: Git;

    beforeEach(() => {
      vi.clearAllMocks();
      git = createGit("/test/repo");
    });

    describe("run", () => {
      it("calls execaSync with correct arguments", () => {
        mockedExecaSync.mockReturnValue({ stdout: "result" } as any);
        const result = git.run("status", "--short");
        expect(mockedExecaSync).toHaveBeenCalledWith("git", ["status", "--short"], {
          cwd: "/test/repo",
          timeout: 15_000,
        });
        expect(result).toBe("result");
      });

      it("throws on non-zero exit", () => {
        mockedExecaSync.mockImplementation(() => {
          throw new Error("exit code 1");
        });
        expect(() => git.run("bad-command")).toThrow("exit code 1");
      });
    });

    describe("tryRun", () => {
      it("returns stdout on success", () => {
        mockedExecaSync.mockReturnValue({ stdout: "ok" } as any);
        expect(git.tryRun("status")).toBe("ok");
      });

      it("returns null on failure", () => {
        mockedExecaSync.mockImplementation(() => {
          throw new Error("fail");
        });
        expect(git.tryRun("bad")).toBeNull();
      });
    });

    describe("review", () => {
      it("delegates to run with review subcommand", () => {
        mockedExecaSync.mockReturnValue({ stdout: "ok" } as any);
        const result = git.review("start", "HEAD~3");
        expect(mockedExecaSync).toHaveBeenCalledWith("git", ["review", "start", "HEAD~3"], {
          cwd: "/test/repo",
          timeout: 15_000,
        });
        expect(result).toBe("ok");
      });
    });
  });
}
