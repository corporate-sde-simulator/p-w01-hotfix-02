# Learning Guide - Go (Golang)

> **Welcome to Product-Track Week 1, Hotfix 2!**
> This is a **hotfix task** - a single file that needs urgent bug fixes.
> Hotfixes simulate real production emergencies where you need to fix code quickly.

---

## What You Need To Do (Summary)

1. **Read the comments** at the top of `counter.go` - they describe the problem
2. **Read** this guide to learn the Go (Golang) syntax you'll need
3. **Find the bugs** (search for `BUG` comments in the code)
4. **Fix each bug** using the hints provided
5. **Run the tests** (if included at the bottom of the file)

---

## Go (Golang) Quick Reference

### Variables and Types
```go
name := "Alice"                  // Short declaration (type inferred)
var count int = 42               // Explicit type
price := 19.99                   // float64
items := []int{1, 2, 3}         // slice (like a dynamic array)
config := map[string]string{     // map (dictionary)
    "key": "value",
}
isActive := true                 // bool
```

### Functions
```go
func greet(name string) string {
    return "Hello, " + name + "!"
}

// Multiple return values (very common in Go!)
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("cannot divide by zero")
    }
    return a / b, nil   // nil = no error
}

// Calling:
result, err := divide(10, 3)
if err != nil {
    fmt.Println("Error:", err)
}
```

### Structs (like classes)
```go
type Calculator struct {
    history []int                // field (lowercase = private)
}

// Constructor function (Go doesn't have constructors)
func NewCalculator() *Calculator {
    return &Calculator{history: []int{}}
}

// Method (function attached to a struct)
func (c *Calculator) Add(a, b int) int {
    result := a + b
    c.history = append(c.history, result)
    return result
}

func (c *Calculator) GetHistory() []int {
    return c.history
}

// Using it:
calc := NewCalculator()
calc.Add(2, 3)
fmt.Println(calc.GetHistory())  // [5]
```

### Maps (Key-Value Storage)
```go
user := map[string]string{"name": "Alice"}
user["name"]                     // Access: "Alice"
user["email"] = "alice@test.com" // Add/update
value, ok := user["name"]       // Check if exists (ok = true/false)
delete(user, "email")           // Remove
len(user)                       // Length
```

### Slices (Dynamic Arrays)
```go
items := []int{1, 2, 3}
items = append(items, 4)        // Add: [1, 2, 3, 4]
len(items)                      // Length: 4
for i, item := range items {   // Loop with index
    fmt.Println(i, item)
}
```

### Error Handling (Go uses explicit error returns)
```go
result, err := someFunction()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
// Use result safely here
```

### Concurrency (Goroutines & Mutexes)
```go
import "sync"

var mu sync.Mutex

func safeIncrement(counter *int) {
    mu.Lock()           // Lock before writing
    *counter++
    mu.Unlock()         // Unlock after writing
}

// Or use defer:
func safeRead(counter *int) int {
    mu.Lock()
    defer mu.Unlock()   // Automatically unlocks when function returns
    return *counter
}
```

### How to Run Tests
```bash
# From the task folder:
go test -v ./...

# With race detector:
go test -race -v ./...
```

### How to Add a Test
```go
func TestSomethingSpecific(t *testing.T) {
    obj := NewProcessor()
    result, err := obj.Process(input)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

---

## Project Structure

This is a **hotfix** - everything is in one file:

| File | Purpose |
|------|---------|
| `counter.go` | The code with bugs - **fix this file** |
| `GUIDE.md` | This learning guide |

---

## Bugs to Fix

### Bug #1
**What's wrong:** Not thread-safe â€” concurrent goroutines cause lost writes.

**How to find it:** Search for `BUG` in `counter.go` - the comments around each bug explain what's broken.

### Bug #2
**What's wrong:** Reading without synchronization â€” may see stale or torn values.

**How to find it:** Search for `BUG` in `counter.go` - the comments around each bug explain what's broken.

### Bug #3
**What's wrong:** Read and write are not atomic â€” another goroutine could

**How to find it:** Search for `BUG` in `counter.go` - the comments around each bug explain what's broken.

### Bug #4
**What's wrong:** Map access is not synchronized â€” concurrent goroutines will

**How to find it:** Search for `BUG` in `counter.go` - the comments around each bug explain what's broken.


---

## How to Approach This

1. **Read the top comment block** in `counter.go` carefully - it has:
   - The JIRA ticket description (what's happening in production)
   - Slack thread (discussion about the problem)
   - Acceptance criteria (checklist of what needs to work)
2. **Search for `BUG`** in the file to find each bug location
3. **Read the surrounding code** to understand what it's trying to do
4. **Fix the logic** based on the bug description
5. **Check the tests** at the bottom of the file and make sure they pass

---

## Common Mistakes to Avoid

- Don't change the structure of the code - only fix the buggy logic
- Read **all** the bugs before starting - sometimes fixing one helps you understand another
- Pay attention to the Slack thread comments - they often contain hints about the root cause
