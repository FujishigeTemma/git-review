---
name: tdd-workflow
description: >-
  Test-Driven Development workflow based on t_wada and Kent Beck's Canon TDD.
  Enforce the List-Red-Green-Refactor cycle for AI agent development.
  Use when: (1) implementing features with TDD, (2) AI agent is delegated to
  write code with tests, (3) user requests TDD or test-first development,
  (4) ensuring correctness of AI-generated code through test-driven discipline.
---

# TDD Workflow

Strict Test-Driven Development workflow for AI agents. Follow the **List-Red-Green-Refactor** cycle — four phases, not three. The List phase is mandatory.

This workflow is based on Kent Beck's Canon TDD and t_wada's (和田卓人) interpretation for AI-era development. TDD serves as a **guardrail** for AI-generated code — without it, AI agents produce code that appears to work but accumulates defects and technical debt at unprecedented speed.

## Core Cycle

### 1. LIST — Build the Test Scenario List

Before writing any code or tests, analyze the requirement and produce a test scenario list.

```
Test scenarios for [feature]:
- [ ] [simplest base case]
- [ ] [next simplest variation]
- [ ] [edge case]
- [ ] [error case]
- [ ] [boundary condition]
```

Rules:
- Decompose the requirement into individually testable behaviors
- Order by implementation dependency — simplest and most foundational first
- Each item is one sentence describing expected input/output or behavior
- Do NOT think about implementation yet — focus on observable behavior
- Present the list to the user before proceeding (in collaboration mode)
- Revisit and update the list after each Red-Green-Refactor cycle

### 2. RED — Write Exactly One Failing Test

Pick the next unchecked `[ ]` item from the test list and translate it into a concrete, runnable test.

Steps:
1. Write one test with setup, invocation, and assertion
2. Run the test suite
3. Confirm the new test **fails** with an expected error message
4. If the test passes without code changes, either the behavior already exists or the test is not testing what you think — investigate before proceeding

The failure message must describe the expected vs actual mismatch. A `NameError` or `ImportError` is an acceptable first failure (the function does not exist yet).

### 3. GREEN — Make It Pass with Minimal Code

Write the simplest code that makes the failing test (and all previous tests) pass.

Rules:
- "Simplest" means: no generalization, no elegance, no future-proofing
- Hardcoded return values are acceptable if only one test demands the behavior
- Do NOT add code for behaviors not yet covered by a test
- Run the full test suite — **ALL** tests must pass
- If any existing test breaks, fix it before proceeding
- Do NOT refactor yet

### 4. REFACTOR — Improve Design Under Green Tests

Only after all tests are green, improve the code's design without changing its behavior.

Rules:
- Structural changes only — no new behavior
- Use design pattern terminology for clear intent (e.g., "Extract Method," "Replace Conditional with Polymorphism," "State pattern")
- Run tests after each refactoring step — if any test breaks, revert immediately
- Separate refactoring commits from behavioral commits
- If no meaningful refactoring is needed, explicitly skip this phase

### 5. LOOP

1. Mark the completed item `[x]` in the test list
2. If new scenarios emerged during implementation, add them to the list
3. If unchecked items remain, return to **RED**
4. When the list is empty, the feature is complete

## AI Agent Operating Rules

These rules are non-negotiable when following this workflow:

1. **Never delete or modify existing tests to make new code pass.** If an existing test fails, the new code has a bug — fix the code, not the test.
2. **Never weaken assertions** (e.g., changing `assertEquals(42, result)` to `assertTrue(result > 0)`).
3. **Never copy computed output values into expected values.** Expected values come from the requirement, not from running the code.
4. **Never write all tests before implementation.** Write one test, make it pass, then write the next.
5. **Never skip RED verification.** Always run the test and confirm failure before writing implementation.
6. **Never skip GREEN verification.** Always run the full test suite and confirm all tests pass before refactoring.
7. **Never mix refactoring with making a test pass.** These are separate phases with separate commits.
8. **Never add behavior not demanded by a failing test.** The test list is the scope boundary.
9. **When delegated a task, follow strict TDD — no shortcuts.** Every RED must be verified, every GREEN must be verified.
10. **When uncertain about the test list, present it to the user for review.**

## Communication Patterns

### Use Design Pattern Terminology

Design pattern names are a compressed communication protocol. Use them:

- "Refactor to State pattern" (instead of describing the full structural change)
- "Extract Method for the validation logic"
- "Replace Conditional with Polymorphism"

This follows t_wada's insight: pattern terminology communicates intent precisely while saving tokens.

### Report Each Cycle

After each Red-Green-Refactor cycle, briefly report:
- Which test scenario was addressed
- Whether RED failed as expected
- What minimal code was added in GREEN
- What refactoring was done (or "no refactoring needed")

### Present the Test List

At the start of a TDD session, present the full test scenario list. After each cycle, show updated progress with `[x]` marks.

## Delegation vs Collaboration Mode

### Delegation Mode (AI works independently)

When the user delegates implementation entirely:
- Follow the full TDD cycle strictly with no shortcuts
- Verify every RED and GREEN phase by running tests
- Commit at each phase boundary when possible
- Complete the entire test list autonomously

### Collaboration Mode (human reviews at key points)

When working interactively with the user:
- Present the test list for approval before starting
- After RED: confirm the test captures the intended behavior
- After GREEN: ask if the minimal implementation direction is correct
- After REFACTOR: review design decisions together
- Flag any ambiguity in requirements immediately

Default to collaboration mode unless the user explicitly delegates.

## Table-Driven Tests in TDD

When building the test scenario list (LIST phase), evaluate whether scenarios share the same assertion structure. If they do, **prefer table-driven tests** — add each scenario as a table entry rather than a separate test function.

### When to Use Table-Driven Tests

Use table-driven tests when:
- Multiple scenarios test the same function with different inputs/outputs
- The setup → act → assert structure is identical across scenarios
- You are adding edge cases or boundary conditions incrementally

Do NOT use table-driven tests when:
- Scenarios require fundamentally different setup or assertion logic
- The test struct accumulates boolean flags controlling test behavior
- Each case is a distinct behavioral concept

### TDD Cycle with Table-Driven Tests

1. **LIST**: Group scenarios that share assertion structure. Mark them as candidates for a single table-driven test.
2. **RED**: Add one new table entry to the test table. Run tests — confirm only the new entry fails.
3. **GREEN**: Write minimal code to make the new entry (and all previous entries) pass.
4. **REFACTOR**: If duplicate test functions emerge, consolidate into a table-driven test. If the table's struct grows boolean flags, split back into separate tests.

This approach naturally fits TDD: each table entry is one RED step, and the table grows incrementally with the test list.

For detailed patterns, struct design, and anti-patterns for table-driven tests, invoke the `table-driven-test` skill.

## References

- **Canon TDD and t_wada's interpretation**: See [references/canon-tdd.md](references/canon-tdd.md) — read when questions arise about the theoretical foundation, why the List phase matters, or the relationship between TDD and design.
- **AI-specific patterns and anti-patterns**: See [references/ai-tdd-patterns.md](references/ai-tdd-patterns.md) — read when encountering edge cases, unsure if a practice is an anti-pattern, or for guidance on Augmented Coding vs Vibe Coding.
- **Concrete workflow examples**: See [references/workflow-examples.md](references/workflow-examples.md) — read when starting a TDD session for the first time or when unsure how a specific phase should look in practice.
