# CODING.md

Design principles and coding style guide.

**Prioritize simplicity and clear errors over flexibility.** Minimize conditional branches and state management to keep code readable. Favor immutability, explicit dependencies, and localized side effects.

### Simplicity Guidelines

- **Don't create an interface if there's only one implementation.** Extract when needed (YAGNI)
- **Minimize state.** Prefer functions that complete with arguments and return values. Keep struct field counts minimal
- **Reduce branching.** Use early returns to keep nesting shallow. Don't use flag arguments to change behavior (split into separate functions)
- **Avoid implicit state transitions.** When state is needed, enumerate possible states with types and prevent invalid transitions at compile time
- **Never swallow errors — return messages that immediately reveal the cause.**
- **Abstraction trades off against readability.** Consider abstracting only when the same code is duplicated in 3+ places. If a "correct" abstraction makes code harder to follow, keeping the duplication is preferable (repeating straightforward implementations can be more readable than forced generalization)

## 0. Loose Coupling and High Cohesion

### Loose Coupling

Loosely coupled design provides the following benefits:
- Maintainability (resilience to change): Changes in one part are less likely to affect others, making maintenance tasks like modifications and feature additions easier. Loose coupling keeps changes localized even as the system grows, while tight coupling increases the risk of unexpected side effects.
- Reusability: Lower dependency between components allows creating more generic, context-independent components. Modules and classes become easier to reuse in different projects and contexts. Loosely coupled components can function independently, making reuse straightforward.
- Testability: In loosely coupled code, each component operates independently, making unit tests easier to write.
- Extensibility: New features can be added without major changes to existing code. In loosely coupled designs, extending functionality is as simple as implementing and swapping in an interface (aligning with the Open/Closed Principle). Tight coupling requires modifying existing code for new features, increasing the risk of introducing bugs.

| Coupling Type                | Summary                                          | Degree         |
| ---------------------------- | ------------------------------------------------ | -------------- |
| Content Coupling             | Directly modifies another module's internals     | Worst          |
| Common Coupling              | Shares global variables                          | High           |
| External Coupling            | Depends on external data formats or APIs         | Somewhat High  |
| Control Coupling             | Controls another's behavior via flags             | Somewhat High  |
| Stamp Coupling               | Passes more of a struct than needed              | Moderate       |
| Data Coupling                | Passes only necessary data as arguments          | Desirable      |
| Message Coupling             | Communicates only via messages/events (notification-based) | Ideal |

### High Cohesion

Cohesion measures how closely the functions within a class or module relate to a single, well-defined purpose.
- Low Cohesion: Different concerns are mixed together, reducing maintainability and readability.
- High Cohesion: Focused on a single responsibility, making changes easier and improving reusability.
Each function and module should focus on one clear responsibility, following the Single Responsibility Principle (SRP) to achieve high cohesion.
Highly cohesive designs are easier to test, modify, and reuse.

| Cohesion Type                    | Summary                                                                | Degree   |
| -------------------------------- | ---------------------------------------------------------------------- | -------- |
| Coincidental Cohesion            | Unrelated functions grouped in the same module                         | Worst    |
| Logical Cohesion                 | Multiple operations in one function selected by branching (switch, etc.) | Low    |
| Temporal Cohesion                | Operations grouped by execution timing (initialization, teardown, etc.) | Somewhat Low |
| Procedural Cohesion              | Operations grouped by procedural order with weak functional relation   | Moderate |
| Communicational Cohesion         | Operations grouped by accessing the same data                          | Somewhat Good |
| Sequential Cohesion              | A chain where one operation's output feeds the next                    | Good     |
| Functional Cohesion              | Only operations necessary for a single clear function                  | Ideal    |

## 1. Module Design and Dependency Injection

Each module has a single responsibility, and dependencies between modules are minimized.

**Principles:**
- Utility layers must not depend on domain logic
- Interfaces are limited to **external boundaries** (Git, filesystem, network, time)
- Internal logic calls other internal logic directly. Don't create interfaces with only one implementation
- The criterion for introducing an interface is whether external dependencies need to be swapped out during testing

### Go

#### NG

```go
// Unnecessary interface for internal logic
type Formatter interface {
    Format(c Commit) string
}

type defaultFormatter struct{}

func (f defaultFormatter) Format(c Commit) string {
    return fmt.Sprintf("%s %s", c.Hash[:7], c.Subject)
}
```

#### OK

```go
// Internal logic is just a function
func FormatCommit(c Commit) string {
    return fmt.Sprintf("%s %s", c.Hash[:7], c.Subject)
}

// Interfaces are defined only for external boundaries
type GitExecutor interface {
    Log(args ...string) (string, error)
    Diff(base, head string) (string, error)
}
```

### TypeScript

#### NG

```typescript
// Interface with only one implementation
interface Formatter {
    format(commit: Commit): string;
}

class DefaultFormatter implements Formatter {
    format(commit: Commit): string {
        return `${commit.hash.slice(0, 7)} ${commit.subject}`;
    }
}
```

#### OK

```typescript
// Internal logic is just a function
function formatCommit(commit: Commit): string {
    return `${commit.hash.slice(0, 7)} ${commit.subject}`;
}

// Interfaces are defined only for external boundaries
interface GitExecutor {
    log(args: string[]): Promise<string>;
    diff(base: string, head: string): Promise<string>;
}
```

---

## 2. Immutability

Never mutate inputs — always return new values.

### Go

#### NG

```go
func normalize(items []Item) {
    for i := range items {
        items[i].Name = strings.ToLower(items[i].Name) // Mutates the caller's slice
    }
}
```

#### OK

```go
func normalize(items []Item) []Item {
    result := make([]Item, len(items))
    for i, item := range items {
        result[i] = Item{Name: strings.ToLower(item.Name)}
    }
    return result
}
```

### TypeScript

#### NG

```typescript
function normalize(items: Item[]): void {
    for (const item of items) {
        item.name = item.name.toLowerCase(); // Mutates the caller's array
    }
}
```

#### OK

```typescript
function normalize(items: readonly Item[]): Item[] {
    return items.map(item => ({ ...item, name: item.name.toLowerCase() }));
}
```

In Go, be careful with destructive changes to slices and maps. Use value receivers for structs; when mutation is needed, return a new struct. In TS, use `readonly` modifiers and the spread operator.

---

## 3. Fail-Fast Validation

Validate inputs with guard clauses at the beginning of functions, rejecting invalid inputs immediately.

**Principles:**
- Check all preconditions before entering the main logic
- Error messages must include **both the expected and actual values**
- Go: Return errors as `error` values. Don't panic
- TS: Use type guards and early returns for safety

### Go

#### NG

```go
func resolve(hash string, repo *Repository) (*Commit, error) {
    // No validation -> obscure errors later
    return repo.Lookup(hash)
}
```

#### OK

```go
func resolve(hash string, repo *Repository) (*Commit, error) {
    if hash == "" {
        return nil, errors.New("commit hash must not be empty")
    }
    if !isValidHash(hash) {
        return nil, fmt.Errorf("invalid commit hash format: %q", hash)
    }
    return repo.Lookup(hash)
}
```

---

## 4. Types and Contracts

Leverage Go's static type system and TS's type system to their fullest.

### Go

- Add godoc comments to exported functions and types
- Define interfaces with the minimum required methods (Interface Segregation)
- Minimize use of `any` / `interface{}`
- Annotate struct fields with comments explaining their meaning

```go
// ReviewState represents the current state of a sequential review session.
type ReviewState struct {
    SessionID string   // unique identifier for this review session
    Commits   []string // ordered list of commit hashes to review
    Current   int      // index of the commit currently under review
}
```

### TypeScript

- Enable `strict: true`
- Use discriminated unions to represent state in a type-safe manner
- Prefer `unknown` over `any`

```typescript
type ReviewEvent =
    | { type: "started"; sessionId: string }
    | { type: "resolved"; commitHash: string }
    | { type: "skipped"; commitHash: string; reason: string };
```

---

## 5. Naming Conventions

### Go

- Exported: `PascalCase` (`RunReview`, `CommitState`)
- Unexported: `camelCase` (`parseArgs`, `validateInput`)
- Package names: short `lowercase` words. Avoid plurals and generic names (`review` OK, `utils` NG, `common` NG)
- Receiver names: 1-2 character abbreviation of the type name (`r` for `Review`, `cs` for `CommitState`)
- Interface names: `-er` suffix for single-method interfaces (`Reader`, `Formatter`)
- Error variables: `Err` prefix (`ErrNotFound`, `ErrInvalidState`)

### TypeScript

- Functions/variables: `camelCase`
- Types/interfaces: `PascalCase`
- Constants: `UPPER_SNAKE_CASE` or `PascalCase`
- Filenames: `camelCase.ts`

### Common

- Function names describe "what", not "how"
- Add comments for abbreviations that may be unclear to readers outside the domain

| OK            | NG                         |
| ------------- | -------------------------- |
| `ListCommits` | `FetchCommitsFromGitLog`   |
| `Resolve`     | `ResolveUsingStateMachine` |
| `FormatDiff`  | `FormatDiffWithANSIColors` |

---

## 6. Public API Design

### Go

- Hide implementation details with the `internal/` package
- Minimize the public surface area. Export only what's necessary
- Struct fields should be unexported by default; expose through methods as needed
- Constructors use the `New` prefix (`NewReview`, `NewSession`)

### TypeScript

- Only add `export` to public API items
- Barrel files (`index.ts`) contain only re-exports. No logic
- Don't export internal implementations

### Common

- Input normalization: Accept flexible user input, normalize it internally to a consistent form
- Pass options as structs/objects (avoid depending on argument order)

```go
type StartOptions struct {
    BaseBranch string
    Verbose    bool
}

func Start(opts StartOptions) error {
    // ...
}
```

---

## 7. Error Handling

Never swallow errors. Propagate them with information that immediately identifies the source and cause. Use [ergo](https://github.com/newmo-oss/ergo) as the error library.

Reference: [Error handling in Go at newmo](https://tech.newmo.me/entry/2025/12/07/090000)

**Principles:**
- Return errors as values. Don't panic
- Use `ergo.New` / `ergo.Wrap` instead of `errors.New` / `fmt.Errorf`
- Don't embed values in error messages with format verbs (`%s`, `%d`); attach them as structured attributes via `slog.Attr`
- Stack traces are attached only once (`ergo.New` / `ergo.Wrap` automatically attach them when creating root errors)
- Use `ergo.NewSentinel` for sentinel errors (no stack trace attached)
- Never pass `nil` as the first argument to `ergo.Wrap`

#### Creating and Wrapping Errors — Add context and attributes at each layer

##### NG

```go
func getVehicle(ctx context.Context, id string) (*Vehicle, error) {
    v, err := db.Find(ctx, id)
    if err != nil {
        return nil, err // Passed through without context -> source is unclear
    }
    return v, nil
}
```

##### NG

```go
func getVehicle(ctx context.Context, id string) (*Vehicle, error) {
    v, err := db.Find(ctx, id)
    if err != nil {
        // fmt.Errorf -> no stack trace attached, values embedded in message
        return nil, fmt.Errorf("get vehicle %s: %w", id, err)
    }
    return v, nil
}
```

##### OK

```go
func getVehicle(ctx context.Context, vehicleID uuid.UUID) (*Vehicle, error) {
    v, err := db.Find(ctx, vehicleID)
    if err != nil {
        return nil, ergo.Wrap(err, "failed to get vehicle",
            slog.String("vehicle_id", vehicleID.String()))
    }
    return v, nil
}

func dispatch(ctx context.Context, rideID, vehicleID uuid.UUID) error {
    vehicle, err := getVehicle(ctx, vehicleID)
    if err != nil {
        return ergo.Wrap(err, "failed to dispatch",
            slog.String("ride_id", rideID.String()))
    }
    // ...
}
```

Context and attributes are added with `ergo.Wrap` at each higher layer. The message chain forms like `"failed to dispatch: failed to get vehicle: record not found"`, and attributes can be collected from the entire error chain via `ergo.AttrsAll(err)`.

#### Structured Attributes — Attach values via `slog.Attr`, not in messages

Embedding values in messages with `fmt.Errorf` reduces log searchability. Separating attributes with `slog.Attr` enables structured search in log infrastructure.

##### NG

```go
// Embedding values in the message -> requires regex for log search
return ergo.New("vehicle not found: " + vehicleID.String())
```

##### OK

```go
// Attaching structured attributes -> searchable as vehicle_id:xxx in log infrastructure
return ergo.New("vehicle not found",
    slog.String("vehicle_id", vehicleID.String()),
    slog.String("region", region))
```

```json
{
  "level": "ERROR",
  "msg": "request failed",
  "error": {
    "message": "failed to dispatch: failed to get vehicle: record not found",
    "stack": "..."
  },
  "attrs": {
    "ride_id": "...",
    "vehicle_id": "..."
  }
}
```

#### Sentinel Errors — Define with `ergo.NewSentinel`

Use `ergo.NewSentinel` for sentinel errors defined as package variables. Using `ergo.New` would attach a meaningless stack trace from package initialization time.

##### NG

```go
// ergo.New -> attaches a stack trace at package initialization
var ErrNotFound = ergo.New("not found")
```

##### OK

```go
// ergo.NewSentinel -> no stack trace
var (
    ErrNotFound     = ergo.NewSentinel("not found")
    ErrInvalidState = ergo.NewSentinel("invalid state")
)
```

#### Error Codes — Classify with `ergo.NewCode` / `ergo.CodeOf`

When error classifications grow large, switching on error codes is clearer than chaining `errors.Is`. Error codes can be declared without prefixes across multiple packages.

##### NG

```go
// Chain of errors.Is -> grows unwieldy as classifications increase
if errors.Is(err, vehicle.ErrNotFound) || errors.Is(err, driver.ErrNotFound) {
    return nil, status.Error(codes.NotFound, err.Error())
} else if errors.Is(err, vehicle.ErrInvalidID) || errors.Is(err, driver.ErrInvalidID) {
    return nil, status.Error(codes.InvalidArgument, err.Error())
}
```

##### OK

```go
// Error code definitions
var (
    CodeVehicleNotFound  = ergo.NewCode("VehicleNotFound", "vehicle not found")
    CodeInvalidVehicleID = ergo.NewCode("InvalidVehicleID", "invalid vehicle id")
)

// Switch on error codes in the handler layer
switch ergo.CodeOf(err) {
case CodeVehicleNotFound, CodeDriverNotFound:
    return nil, status.Error(codes.NotFound, err.Error())
case CodeInvalidVehicleID, CodeInvalidDriverID:
    return nil, status.Error(codes.InvalidArgument, err.Error())
}
```

---

## 8. Testability

Always design for testability.
Minimize use of mocks and stubs.

### Go

- Use table-driven tests as the standard pattern
- Inject external dependencies (Git, filesystem, time) via interfaces
- Place test fixtures in `testdata/` directories
- Add `t.Helper()` to test helpers

```go
func TestResolve(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {name: "valid hash", input: "abc123", want: "abc123"},
        {name: "empty hash", input: "", wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Resolve(tt.input)
            if (err != nil) != tt.wantErr {
                t.Fatalf("Resolve(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("Resolve(%q) = %q, want %q", tt.input, got, tt.want)
            }
        })
    }
}
```

### TypeScript

- Follow testing framework conventions
- Inject functions with side effects via interfaces, swapping in stubs during tests

---

## 9. CLI / Extension Design Principles

### CLI (Go)

- Follow the subcommand pattern for command structure
- Separate user-facing output from program output (stderr vs stdout)
- Communicate success/failure via exit codes (0: success, 1: error)
- Keep flag names short and clear. Provide aliases for long flags

### VSCode Extension (TypeScript)

- Confine VSCode API dependencies to a thin adapter layer
- Keep business logic independent of VSCode for unit testability
- Register commands declaratively

### Common

- Clearly define precedence for configuration: environment variables > config files > flags
- Make user-facing messages specific and actionable

---

## Summary

| #   | Principle                     | Summary                                                                      |
| --- | ----------------------------- | ---------------------------------------------------------------------------- |
| 0   | **Loose Coupling & High Cohesion** | Minimize dependencies between modules; each module focuses on a single responsibility |
| 1   | **Module Design & DI**        | Interfaces only at external boundaries. Internal logic calls directly        |
| 2   | **Immutability**              | Never mutate inputs. Always return new values                                |
| 3   | **Fail-Fast**                 | Validate early with guard clauses. Include expected and actual values in errors |
| 4   | **Types & Contracts**         | Maximize static types. Document interfaces with godoc / TSDoc               |
| 5   | **Naming Conventions**        | Follow Go / TS conventions. Function names describe results                  |
| 6   | **Public API**                | Minimize public surface. Use `internal/` and export control                  |
| 7   | **Error Handling**            | Wrap with ergo, structured attributes, error codes. Stack traces only once   |
| 8   | **Testability**               | Table-driven tests. Inject external dependencies via interfaces              |
| 9   | **CLI / Extension**           | Subcommand structure. Localize VSCode dependencies. Messages are actionable  |
