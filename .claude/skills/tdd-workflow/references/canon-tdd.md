# Canon TDD

## Table of Contents
- [Kent Beck's Canonical Definition](#kent-becks-canonical-definition)
- [t_wada's Four-Phase Interpretation](#t_wadas-four-phase-interpretation)
- [Semantic Diffusion of TDD](#semantic-diffusion-of-tdd)
- [TDD as a Guardrail for AI](#tdd-as-a-guardrail-for-ai)
- [TDD and Design](#tdd-and-design)
- [Transformation Priority Premise](#transformation-priority-premise)

## Kent Beck's Canonical Definition

Kent Beck's Canon TDD (2024) defines TDD as a specific programming workflow, not a vague philosophy. The canonical steps:

1. **Test List**: Write a list of test scenarios covering expected behavior changes, edge cases, and existing functionality that should not break.
2. **Write One Test**: Convert exactly one item from the list into a concrete, runnable test — with setup, invocation, and assertions.
3. **Make It Pass**: Change the code to make the test (and all previous tests) pass. Add newly discovered test scenarios to the list.
4. **Optional Refactoring**: Improve implementation design decisions only after all tests pass.
5. **Repeat**: Continue until the test list is empty.

Key distinction: interface design decisions (how behavior is invoked) happen during test writing. Implementation design decisions (how the system achieves the behavior) happen during refactoring. These must not be conflated.

Source: Kent Beck, "Canon TDD" (Tidy First? Substack, 2024)

## t_wada's Four-Phase Interpretation

t_wada (和田卓人) — Japan's foremost TDD practitioner and translator of Kent Beck's "Test Driven Development: By Example" — interprets the canonical cycle as four explicit phases:

**List → Red → Green → Refactor**

The critical insight: the **List** phase is not a precursor to TDD — it IS part of TDD. Without it, developers jump into writing tests without analyzing the requirement, leading to incomplete coverage and directionless implementation.

t_wada's teaching emphasizes three essences of TDD:
1. **Build small, defect-free features in small cycles** — working functionality grows incrementally through tests
2. **Decompose uncertainty** — break large features into small TODOs, progressing step-by-step through test-driven implementation
3. **Enable fearless refactoring** — test code eliminates fear that changes break functionality

Source: t_wada, "テスト駆動開発の定義" (translation of Canon TDD, 2024)

## Semantic Diffusion of TDD

t_wada identifies a critical problem: TDD has suffered "semantic diffusion" (意味の希薄化). Three distinct concepts are routinely conflated:

| Concept | Definition | Relationship |
|---|---|---|
| **Automated Testing** | Writing code that automatically verifies behavior | Foundation — necessary but not TDD |
| **Test-First** | Writing tests before production code | Stronger — includes automated testing but is not TDD |
| **Test-Driven Development** | List-Red-Green-Refactor cycle with incremental design | Strongest — includes test-first plus design discipline |

When someone says "do TDD," they may mean any of these three. AI agents instructed to "do TDD" often default to writing a batch of tests first (Test-First at best), which misses the incremental design cycle that defines actual TDD.

## TDD as a Guardrail for AI

t_wada's 2025 thesis: TDD has evolved from an enthusiast practice to a critical guardrail for AI-era development.

**Why AI needs TDD guardrails:**
- AI agents can modify existing code while fixing bugs, inadvertently breaking other behaviors
- Without automated tests, these regressions are invisible until production
- Tests give AI agents "eyes" — the ability to see whether their changes maintain correctness
- The TDD cycle forces incremental progress, preventing the "generate a huge chunk and hope it works" pattern

**Why TDD adoption costs are dropping:**
- AI dramatically reduces the learning curve for TDD
- AI can generate test boilerplate, reducing the setup cost that previously hindered adoption
- The "tedious" parts of TDD (writing repetitive test code) are exactly what AI excels at

**The shift in feedback loops:**
- Classic TDD: feedback comes from test execution cycles
- AI-era TDD: feedback should also come from the design review phase (earlier in the process)
- Document design decisions in Architecture Decision Records (ADR) to preserve the "why" behind AI-assisted design choices

Source: t_wada, "AIエージェント時代に、TDDはガードレールになる" (Agile Journey, 2025)

## TDD and Design

TDD is not primarily about testing — it drives design. Kent Beck and t_wada both emphasize:

- The **test** is a design document — it specifies the interface before implementation exists
- The **refactoring phase** is where design emerges — not upfront, not as an afterthought
- Design pattern names (GoF, etc.) serve as an efficient vocabulary for refactoring intent
- Separating behavioral changes (RED→GREEN) from structural changes (REFACTOR) keeps each step verifiable

Kent Beck's "Augmented Coding" approach (2025): embed TDD philosophy in AI agent instructions via system prompts. The key is maintaining discipline — active monitoring for AI "cheating" behaviors such as deleting tests or weakening assertions to achieve green status.

Source: Kent Beck, "Augmented Coding: Beyond the Vibes" (Tidy First? Substack, 2025)

## Transformation Priority Premise

When choosing the "simplest code" during the GREEN phase, follow Kent Beck's Transformation Priority Premise — a guide for selecting implementations in order of increasing complexity:

1. `({} → nil)` — no code → return nil/null
2. `(nil → constant)` — return a hardcoded constant
3. `(constant → variable)` — replace constant with a variable or parameter
4. `(statement → statements)` — add more unconditional statements
5. `(unconditional → conditional)` — add if/switch branching
6. `(scalar → collection)` — use a list, map, or set
7. `(statement → recursion)` — introduce recursion
8. `(value → mutated value)` — introduce state mutation

Prefer transformations higher in the list. When two approaches would both make the test pass, choose the simpler transformation. This prevents over-engineering during the GREEN phase and defers design decisions to the REFACTOR phase where they belong.
