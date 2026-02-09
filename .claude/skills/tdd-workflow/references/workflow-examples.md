# TDD Workflow Examples

## Table of Contents
- [Example 1: FizzBuzz — Full TDD Cycle](#example-1-fizzbuzz--full-tdd-cycle)
- [Example 2: Design Pattern Terminology in Refactoring](#example-2-design-pattern-terminology-in-refactoring)
- [Example 3: Anti-Pattern Comparison](#example-3-anti-pattern-comparison)

## Example 1: FizzBuzz — Full TDD Cycle

Requirement: Given an integer n, return "Fizz" for multiples of 3, "Buzz" for multiples of 5, "FizzBuzz" for multiples of both, otherwise the number as a string.

### LIST Phase

```
Test scenarios for FizzBuzz:
- [ ] Returns "1" for input 1
- [ ] Returns "2" for input 2
- [ ] Returns "Fizz" for input 3
- [ ] Returns "Buzz" for input 5
- [ ] Returns "Fizz" for input 6 (another multiple of 3)
- [ ] Returns "Buzz" for input 10 (another multiple of 5)
- [ ] Returns "FizzBuzz" for input 15
- [ ] Returns "FizzBuzz" for input 30
```

### Cycle 1: "Returns '1' for input 1"

**RED:**
```python
def test_returns_1_for_1():
    assert fizzbuzz(1) == "1"
```
Run tests: `NameError: name 'fizzbuzz' is not defined` — FAIL (expected).

**GREEN:**
```python
def fizzbuzz(n):
    return "1"
```
Run tests: 1 passed — GREEN.

Note: Returning a hardcoded "1" is intentional. Only one test demands this behavior. The Transformation Priority Premise says `(nil → constant)` is simpler than `(nil → variable)`.

**REFACTOR:** No refactoring needed. Skip.

```
- [x] Returns "1" for input 1
- [ ] Returns "2" for input 2
...
```

### Cycle 2: "Returns '2' for input 2"

**RED:**
```python
def test_returns_2_for_2():
    assert fizzbuzz(2) == "2"
```
Run tests: `AssertionError: assert '1' == '2'` — FAIL (expected).

**GREEN:**
```python
def fizzbuzz(n):
    return str(n)
```
Run tests: 2 passed — GREEN.

Now we generalize: two tests demand number-to-string conversion, so `(constant → variable)` transformation is justified.

**REFACTOR:** No refactoring needed. Skip.

### Cycle 3: "Returns 'Fizz' for input 3"

**RED:**
```python
def test_returns_fizz_for_3():
    assert fizzbuzz(3) == "Fizz"
```
Run tests: `AssertionError: assert '3' == 'Fizz'` — FAIL (expected).

**GREEN:**
```python
def fizzbuzz(n):
    if n % 3 == 0:
        return "Fizz"
    return str(n)
```
Run tests: 3 passed — GREEN.

Transformation: `(unconditional → conditional)`.

**REFACTOR:** No refactoring needed yet. The code is still simple.

### Cycle 4: "Returns 'Buzz' for input 5"

**RED:**
```python
def test_returns_buzz_for_5():
    assert fizzbuzz(5) == "Buzz"
```
Run tests: `AssertionError: assert '5' == 'Buzz'` — FAIL (expected).

**GREEN:**
```python
def fizzbuzz(n):
    if n % 3 == 0:
        return "Fizz"
    if n % 5 == 0:
        return "Buzz"
    return str(n)
```
Run tests: 4 passed — GREEN.

**REFACTOR:** No refactoring needed yet.

### Cycle 5 (skipping confirmation tests): "Returns 'FizzBuzz' for input 15"

**RED:**
```python
def test_returns_fizzbuzz_for_15():
    assert fizzbuzz(15) == "FizzBuzz"
```
Run tests: `AssertionError: assert 'Fizz' == 'FizzBuzz'` — FAIL (expected).

Note: 15 is divisible by 3, so it hits the first condition and returns "Fizz". The test correctly identifies that the order of checks matters.

**GREEN:**
```python
def fizzbuzz(n):
    if n % 15 == 0:
        return "FizzBuzz"
    if n % 3 == 0:
        return "Fizz"
    if n % 5 == 0:
        return "Buzz"
    return str(n)
```
Run tests: 5 passed — GREEN.

**REFACTOR:** The `n % 15 == 0` check is equivalent to `n % 3 == 0 and n % 5 == 0`. Both are valid. The current form is clear enough — skip refactoring.

### Final test list:
```
- [x] Returns "1" for input 1
- [x] Returns "2" for input 2
- [x] Returns "Fizz" for input 3
- [x] Returns "Buzz" for input 5
- [x] Returns "Fizz" for input 6
- [x] Returns "Buzz" for input 10
- [x] Returns "FizzBuzz" for input 15
- [x] Returns "FizzBuzz" for input 30
```

All items checked. Feature complete.

## Example 2: Design Pattern Terminology in Refactoring

t_wada's insight: design pattern names are a highly compressed communication protocol between human and AI agent.

### Scenario: Order processing with accumulated conditionals

After multiple TDD cycles, the order processing code has grown conditional branches:

```python
class Order:
    def process(self):
        if self.status == "pending":
            self._validate()
            self.status = "confirmed"
        elif self.status == "confirmed":
            self._charge_payment()
            self.status = "paid"
        elif self.status == "paid":
            self._ship()
            self.status = "shipped"
        elif self.status == "cancelled":
            raise InvalidOperationError("Cannot process cancelled order")
```

### REFACTOR with pattern terminology

Instead of describing every structural change, say:

> "Refactor to State pattern"

This single phrase communicates:
1. Create state classes: `PendingState`, `ConfirmedState`, `PaidState`, `ShippedState`, `CancelledState`
2. Move conditional behavior into state-specific `process()` methods
3. Replace the conditional chain with polymorphic dispatch
4. Each state object handles its own transitions

The AI agent produces:

```python
class Order:
    def __init__(self):
        self._state = PendingState()

    def process(self):
        self._state = self._state.process(self)

class PendingState:
    def process(self, order):
        order._validate()
        return ConfirmedState()

class ConfirmedState:
    def process(self, order):
        order._charge_payment()
        return PaidState()
# ... etc.
```

All existing tests still pass — the refactoring changed structure, not behavior.

### Other effective pattern phrases:

| Phrase | What it communicates |
|---|---|
| "Extract Method" | Pull a code block into a named function |
| "Replace Conditional with Polymorphism" | Use inheritance instead of if/switch |
| "Introduce Parameter Object" | Group related parameters into a class |
| "Replace Magic Number with Named Constant" | Give meaning to literal values |
| "Move Method to [class]" | Relocate behavior to where the data lives |

## Example 3: Anti-Pattern Comparison

### Skipping RED verification

**Wrong:**
```python
# Agent writes test AND implementation together
def test_add():
    assert add(2, 3) == 5

def add(a, b):
    return a + b
# "All tests pass!" — but was the test ever seen to fail?
```

**Right:**
```python
# Step 1 (RED): Write test only, run it
def test_add():
    assert add(2, 3) == 5
# Run: NameError: name 'add' is not defined — FAIL. Good.

# Step 2 (GREEN): Write minimal implementation, run again
def add(a, b):
    return a + b
# Run: 1 passed — GREEN. Good.
```

### Assertion weakening

**Wrong:**
```python
# Original test with precise assertion
def test_calculate_total():
    assert calculate_total([10, 20, 30]) == 60

# AI changes test to make buggy code "pass"
def test_calculate_total():
    result = calculate_total([10, 20, 30])
    assert isinstance(result, int)  # weakened — hides the bug
```

**Right:**
```python
# Keep the original precise assertion
def test_calculate_total():
    assert calculate_total([10, 20, 30]) == 60
# If this fails, fix calculate_total — not the test.
```

### Computed value copying

**Wrong:**
```python
# AI runs the code, gets 59.99999, pastes it as expected
def test_calculate_total():
    assert calculate_total([10, 20, 30]) == 59.99999
# The expected value was copied from the implementation output.
# The requirement says the total should be 60.
```

**Right:**
```python
# Expected value comes from the REQUIREMENT, not from running the code
def test_calculate_total():
    assert calculate_total([10, 20, 30]) == 60
```

### Batch test writing (not TDD)

**Wrong:**
```python
# All tests written at once before any implementation
def test_add(): assert add(2, 3) == 5
def test_add_zero(): assert add(0, 5) == 5
def test_add_negative(): assert add(-1, 1) == 0
def test_add_large(): assert add(1000000, 1) == 1000001

# Then implement everything at once
def add(a, b):
    return a + b
```

**Right:**
```python
# Test 1 → implement → test 2 → implement → ...
# Each test drives one incremental design decision.
# See Example 1 (FizzBuzz) for the full cycle.
```
