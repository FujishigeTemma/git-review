# AI-Specific TDD Patterns and Anti-Patterns

## Table of Contents
- [Anti-Patterns: What AI Agents Must Never Do](#anti-patterns-what-ai-agents-must-never-do)
- [Augmented Coding vs Vibe Coding](#augmented-coding-vs-vibe-coding)
- [Context Isolation](#context-isolation)
- [Commit Discipline](#commit-discipline)
- [Test Quality Guidelines](#test-quality-guidelines)
- [Architecture Decision Records](#architecture-decision-records)

## Anti-Patterns: What AI Agents Must Never Do

### 1. Test Deletion or Modification

**Symptom:** Test count decreases or existing assertions change after a code change.
**Why it happens:** AI finds it easier to make tests match the code than to make the code match the tests.
**Rule:** If an existing test fails after your code change, the code is wrong — not the test. Fix the implementation.

### 2. Assertion Weakening

**Symptom:** Strict assertions replaced with lenient ones.
```python
# WRONG: weakened to make buggy code "pass"
assert isinstance(result, int)       # was: assert result == 60

# WRONG: range check instead of exact value
assert 50 < result < 70              # was: assert result == 60
```
**Rule:** Assertions must match the requirement's precision. Never weaken them to accommodate implementation defects.

### 3. Computed Value Copying

**Symptom:** Expected test values match the code's output exactly — because they were copied from a test run.
```python
# WRONG: copied from running the buggy implementation
assert calculate_tax(100) == 7.9999999  # requirement says 8.0
```
**Rule:** Expected values come from the requirement specification, not from running the code. If you find yourself copying an output value into a test, stop — you are validating the implementation, not the requirement.

### 4. Batch Test Writing

**Symptom:** 10 tests written before any implementation code.
**Why it's wrong:** TDD's power comes from incremental design — each test drives a small design decision. Writing all tests first is "Test-First" at best, not TDD.
**Rule:** Write one test. Make it pass. Then write the next test.

### 5. Skipping RED Verification

**Symptom:** Test and implementation written simultaneously in the same step.
**Why it's wrong:** If you never see the test fail, you don't know if it CAN fail. A test that cannot fail provides no safety.
**Rule:** Always run the test suite after writing a new test and before writing implementation. Confirm the failure.

### 6. Mixing Behavioral and Structural Changes

**Symptom:** A single commit adds new behavior AND refactors existing code.
**Why it's wrong:** If tests break, you cannot determine which change caused the failure. This violates Kent Beck's "Tidy First" principle.
**Rule:** GREEN commits add behavior. REFACTOR commits improve structure. Never combine them.

### 7. Vibe Coding

**Symptom:** Large chunks of code generated with minimal oversight, then "hope it works."
**Why it's wrong:** Rapid accumulation of technical debt. Problems that typically emerge over months appear in weeks. Code review becomes impossible as throughput increases.
**Rule:** Follow the Augmented Coding approach — maintain understanding and verification at every step.

## Augmented Coding vs Vibe Coding

Kent Beck's critical distinction (2025):

| | Augmented Coding | Vibe Coding |
|---|---|---|
| **Process** | TDD + discipline + AI assistance | Prompt → generate → hope |
| **Code quality** | Clean, working, understood | Unknown quality, fragile |
| **Test role** | Guardrails driving design | Optional afterthought |
| **Human role** | Active monitor and decision-maker | Passive consumer |
| **Outcome** | Production-ready software | Throwaway prototypes |

**Key practices for Augmented Coding with TDD:**
- Embed TDD discipline in the workflow, not just in intent
- Watch intermediate results — intervene when development becomes unproductive
- Give specific guidance ("for the next test, add keys in reverse order") instead of vague requests ("add more tests")
- Detect and halt AI cheating immediately (test deletion, assertion weakening)

Source: Kent Beck, "Augmented Coding: Beyond the Vibes" (Tidy First? Substack, 2025)

## Context Isolation

A fundamental problem with AI TDD: when the same context writes both the test and the implementation, knowledge of the implementation bleeds into the test, defeating the purpose of test-first design.

**Principle:** Write the test based on the specification, not on the code.

**In practice:**
- When writing a test (RED), focus on the requirement and the expected interface — do not look at existing implementation details
- When writing implementation (GREEN), focus on making the test pass — do not pre-optimize for future tests
- If modifying existing code, read the specification/requirement first, not the source code
- This prevents the "make the test match the code" trap

**For multi-agent systems:** Separate test-writing and implementation into isolated contexts. The test writer should not see the implementer's plan, and vice versa.

## Commit Discipline

Recommended commit message pattern for TDD phases:

```
red: add failing test for [behavior description]
green: make [behavior description] test pass
refactor: [design change description]
```

Rules:
- Each commit should be at a phase boundary (after RED, after GREEN, or after REFACTOR)
- Never combine RED+GREEN or GREEN+REFACTOR in one commit
- GREEN commits must have all tests passing
- REFACTOR commits must have all tests passing
- RED commits have exactly one new failing test (all others still pass)

When committing is impractical for every phase (e.g., very rapid micro-cycles), at minimum commit after each complete Red-Green-Refactor cycle:

```
feat: [behavior description] (TDD cycle)
```

## Test Quality Guidelines

### Naming
Test names describe behavior, not implementation:
```python
# Good
def test_rejects_negative_amounts():
def test_applies_discount_for_premium_members():

# Bad
def test_validate_method():
def test_calculate_returns_value():
```

### Structure
Use Arrange-Act-Assert (AAA):
```python
def test_applies_10_percent_discount():
    # Arrange
    order = Order(amount=100, member_type="premium")

    # Act
    total = order.calculate_total()

    # Assert
    assert total == 90
```

### Independence
- Each test runs independently — no shared mutable state between tests
- No dependency on test execution order
- Each test sets up its own prerequisites

### Assertion Clarity
- One logical assertion per test (multiple physical assertions are fine if they verify one behavior)
- Assertion messages should explain intent when the framework supports them

## Architecture Decision Records

When AI-assisted TDD leads to significant design decisions during the REFACTOR phase, document them in Architecture Decision Records (ADR).

**When to write an ADR:**
- A design pattern is chosen over alternatives (e.g., State pattern vs Strategy pattern)
- A significant structural refactoring is performed
- The AI suggested a design direction that was accepted or rejected
- Trade-offs were made between competing design goals

**Minimal ADR format:**
```markdown
## ADR: [Decision Title]
**Status:** Accepted
**Context:** [What prompted this decision]
**Decision:** [What was decided]
**Consequences:** [What changes as a result]
```

ADRs preserve the "why" behind design decisions — especially important when AI agents make suggestions that humans accept. Without ADRs, the rationale is lost when the conversation context disappears.

Source: t_wada, "AI時代のソフトウェア開発を考える" (Agile Japan, 2025)
