---
name: table-driven-test
description: >-
  Table-driven test pattern for writing clean, maintainable, and comprehensive tests.
  Separate test data (table entries) from test logic to eliminate copy-paste duplication.
  Use when: (1) writing Go tests with multiple input/output variations,
  (2) testing pure functions with various combinations, (3) refactoring duplicated test code,
  (4) adding edge cases and boundary conditions to existing tests,
  (5) implementing TDD test scenarios that share the same assertion structure.
  Covers Go-specific patterns (t.Run subtests, parallel execution, cmp.Diff)
  and language-agnostic principles applicable to any parameterized testing.
---

# Table-Driven Test

Write test data once as a table, test logic once as a loop. Each table entry is a complete test case with inputs and expected results.

Based on: [Go Wiki: TableDrivenTests](https://go.dev/wiki/TableDrivenTests), Dave Cheney, Mitchell Hashimoto, Fatih Arslan.

## When to Use Table-Driven Tests

**Use** when:
- Multiple test cases share the same setup → act → assert structure
- Copy-paste appears in tests (the primary signal)
- Testing a function with various input/output combinations
- Adding edge cases, boundary conditions, or error cases incrementally
- TDD test list items share the same assertion pattern

**Avoid** when:
- Each test case requires fundamentally different setup or assertion logic
- The test struct accumulates boolean flags (`shouldError`, `skipValidation`, `useAlternateSetup`) — this signals the tests should be separate functions
- A single, one-off test is sufficient
- The table definition becomes harder to read than separate test functions

Table complexity is a **design smell**: if the table is convoluted, the function under test likely has too many responsibilities.

## Core Pattern (Go)

### Basic: Slice of Structs with t.Run

```go
func TestParseFlag(t *testing.T) {
    tests := []struct {
        name string
        in   string
        want string
    }{
        {name: "simple percent", in: "%a", want: "[%a]"},
        {name: "left-align", in: "%-a", want: "[%-a]"},
        {name: "with width and precision", in: "%1.2a", want: "[%1.2a]"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ParseFlag(tt.in)
            if got != tt.want {
                t.Errorf("ParseFlag(%q) = %q, want %q", tt.in, got, tt.want)
            }
        })
    }
}
```

### Map-Based: Keys as Test Names

```go
func TestReverse(t *testing.T) {
    tests := map[string]struct {
        input string
        want  string
    }{
        "empty string":    {input: "", want: ""},
        "single char":     {input: "x", want: "x"},
        "multi-byte":      {input: "🥳🎉", want: "🎉🥳"},
        "ascii string":    {input: "abc", want: "cba"},
    }

    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            if got := Reverse(tt.input); got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}
```

Map advantages: undefined iteration order **exposes order-dependent bugs**; key naturally serves as test name.

### Parallel Execution

```go
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // Go 1.22+: loop variables are per-iteration, no capture needed
        // Go <1.22: add `tt := tt` before t.Run
        got := Expensive(tt.input)
        if diff := cmp.Diff(tt.want, got); diff != "" {
            t.Errorf("mismatch (-want +got):\n%s", diff)
        }
    })
}
```

Call `t.Parallel()` on both the parent test and each subtest.

## Best Practices

### 1. Name every test case

Descriptive names make failure output self-documenting. Prefer behavior descriptions over input descriptions.

```go
// Good
{name: "rejects negative amount", ...}
{name: "applies discount for premium member", ...}

// Acceptable (input as name for simple transformations)
{name: "empty string", ...}
{name: "unicode input", ...}
```

### 2. Use t.Errorf for pure functions, t.Fatalf for preconditions

`t.Errorf` continues execution — collects all failures. Use `t.Fatalf` only when a failed precondition makes subsequent checks meaningless (e.g., nil pointer guard).

### 3. Use cmp.Diff for complex comparisons

Replace `reflect.DeepEqual` with `github.com/google/go-cmp/cmp`:

```go
if diff := cmp.Diff(tt.want, got); diff != "" {
    t.Errorf("mismatch (-want +got):\n%s", diff)
}
```

Produces human-readable diffs for structs, slices, and maps.

### 4. Include error case fields naturally

```go
tests := []struct {
    name    string
    input   int
    want    int
    wantErr bool
}{
    {name: "valid input", input: 5, want: 25, wantErr: false},
    {name: "negative input", input: -1, want: 0, wantErr: true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := Square(tt.input)
        if (err != nil) != tt.wantErr {
            t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
        }
        if got != tt.want {
            t.Errorf("got %d, want %d", got, tt.want)
        }
    })
}
```

### 5. Keep tables lean — use functional modifiers for complex structs

When test inputs are large structs where only one or two fields change, use Fatih Arslan's functional modifier pattern:

```go
tests := []struct {
    name    string
    modify  func(cfg *Config)
    wantErr bool
}{
    {name: "valid config", modify: func(cfg *Config) {}, wantErr: false},
    {name: "missing host", modify: func(cfg *Config) { cfg.Host = "" }, wantErr: true},
    {name: "invalid port", modify: func(cfg *Config) { cfg.Port = -1 }, wantErr: true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        cfg := validConfig() // base valid struct
        tt.modify(cfg)
        err := cfg.Validate()
        if (err != nil) != tt.wantErr {
            t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
        }
    })
}
```

### 6. Do not force-fit different test shapes into one table

If some cases need different setup, assertion logic, or mocking — write separate test functions. A table with boolean flags controlling branches in the test body is an anti-pattern.

## Anti-Patterns

| Anti-Pattern | Signal | Fix |
|---|---|---|
| Boolean flags in struct | `skipValidation`, `useAlternate` fields | Split into separate test functions |
| Massive struct literals | Each entry spans 20+ lines | Use functional modifiers or separate tests |
| Shared mutable state | Tests pass only in specific order | Isolate each case; use maps to detect order dependency |
| Computed expected values | Expected value copied from running the code | Derive expected values from the requirement |
| Single-case table | Only one entry in the table | Use a regular test function |

## TDD Integration

When following the List-Red-Green-Refactor cycle:

1. **LIST**: Identify test scenarios. If multiple scenarios share the same assertion structure, plan them as table entries.
2. **RED**: Add one new table entry. Run tests — confirm the new entry fails.
3. **GREEN**: Write minimal code to make the new entry pass (and all previous entries).
4. **REFACTOR**: If you notice duplicated test functions during refactoring, consolidate into a table-driven test.

Table-driven tests are a natural fit for TDD: each new table entry is one RED step, and the table grows incrementally with the test list.

## References

- [Go Wiki: TableDrivenTests](https://go.dev/wiki/TableDrivenTests)
- [Dave Cheney: Prefer table driven tests (2019)](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Fatih Arslan: Functional table-driven tests in Go (2022)](https://arslan.io/2022/12/04/functional-table-driven-tests-in-go/)
- [Mitchell Hashimoto: Advanced Testing with Go (GopherCon 2017)](https://speakerdeck.com/mitchellh/advanced-testing-with-go)
- [Learn Go with Tests: Anti-patterns](https://quii.gitbook.io/learn-go-with-tests/meta/anti-patterns)
