# Go Learning Notes

Concepts encountered while building ReSEARCH Events. Each entry explains the concept,
why Go does it this way, and where it appears in this codebase.

---

## Table of contents

- [context.Context](#contextcontext)
- [Pointer receivers and pointer types](#pointer-receivers-and-pointer-types)
- [Interfaces](#interfaces)
- [Struct literals and the address-of operator](#struct-literals-and-the-address-of-operator)
- [Short variable declaration :=](#short-variable-declaration-)
- [make — initializing maps, slices, and channels](#make--initializing-maps-slices-and-channels)
- [Stack vs Heap](#stack-vs-heap)
- [How these concepts connect to Functional Programming](#how-these-concepts-connect-to-functional-programming)

---

## `context.Context`

**First seen in:** `internal/health/db.go` — `DatabaseChecker.Check(ctx context.Context)`

### What it is

`context.Context` is a Go interface that flows through your entire call stack, carrying
three things:

| What | Example |
|---|---|
| Deadline | "this operation must finish by 10:00:03" |
| Cancellation signal | "the caller gave up — stop what you're doing" |
| Request-scoped values | trace ID, authenticated user ID |

### Why it matters

Without context, if a user closes their browser mid-request, your DB query keeps running,
holding a connection, burning CPU — even though nobody is waiting for the result.
With context, the cancelled request propagates down the call stack and every I/O operation
can stop early.

```go
// ctx carries a 3-second deadline.
// If the DB doesn't respond in time, the query is cancelled automatically.
ctx, cancel := context.WithTimeout(parentCtx, 3*time.Second)
defer cancel() // always cancel to release resources, even if it finishes early

err := db.QueryRowContext(ctx, "SELECT 1").Scan(&n)
```

### The convention

Go convention: **always pass `ctx` as the first parameter** to any function that does I/O
(DB queries, HTTP calls, file reads). The compiler does not enforce this — it is a
community standard that every Go developer follows. Breaking it makes your code
incompatible with the ecosystem.

```go
// Correct — context first
func (c *DatabaseChecker) Check(ctx context.Context) CheckResult

// Wrong — context buried or missing
func (c *DatabaseChecker) Check(timeout int) CheckResult
```

### `context.Background()` vs `context.WithTimeout()`

| Function | Use |
|---|---|
| `context.Background()` | Root context — use at the top of `main()` or in tests |
| `context.WithTimeout(ctx, d)` | Creates a child context that auto-cancels after duration `d` |
| `context.WithCancel(ctx)` | Creates a child context you cancel manually |

**Where it appears in this project:**
- `DatabaseChecker.Check` uses `context.WithTimeout` to cap the `SELECT 1` at 3 seconds
- Every repository method accepts `ctx` and passes it to GORM via `.WithContext(ctx)`
- `main.go` creates the root context with `context.Background()`

---

## Pointer receivers and pointer types

**First seen in:** `internal/handler/health.go` — `func (h *HealthHandler) ServeHTTP(...)`

### Value vs pointer receiver

When you define a method on a struct, you choose how the method receives the struct:

```go
type Counter struct {
    count int
}

// Value receiver — gets a COPY of Counter.
// The original Counter is unchanged after this call.
func (c Counter) Value() int {
    return c.count
}

// Pointer receiver — gets the memory ADDRESS of Counter.
// Changes here affect the original struct.
func (c *Counter) Increment() {
    c.count++ // modifies the real Counter, not a copy
}
```

### When to use each

Use a **pointer receiver** when:
- The method needs to **modify** the struct's fields
- The struct is **large** — passing a pointer (8 bytes) is cheaper than copying the whole struct
- **Any** other method on the type uses a pointer receiver (mixed receivers cause subtle bugs — pick one and stick to it)

Use a **value receiver** when:
- The method only **reads** from the struct AND the struct is very small (1–2 fields)
- You want to make it explicit that the struct is immutable from this method's perspective

In practice, most structs in this codebase use pointer receivers because they hold
dependencies (DB connections, loggers) that are large or require consistent identity.

### Pointer types: `*HealthHandler`, `*sql.DB`

The `*` in a type position means "pointer to":

```go
var h *HealthHandler  // h is a pointer — it holds the memory address of a HealthHandler
var h  HealthHandler  // h is a value  — it holds the actual HealthHandler data
```

**Why `*sql.DB` is always a pointer:**

`sql.DB` is a connection pool. If you copied it, every copy would manage its own pool
independently — connections would multiply, state would diverge, transactions would break.
Passing `*sql.DB` ensures the whole application shares **one** pool.

```go
// One pool, shared by everyone via pointer.
// Every handler, every repository — same underlying connections.
db *sql.DB
```

**`nil` — the zero value for pointers:**

A pointer that has not been assigned points to nothing — its value is `nil`.
Calling a method on a nil pointer panics. This is why constructors like
`NewHealthHandler(...)` always return a fully initialized `*HealthHandler`, never nil.

```go
var h *HealthHandler     // nil — do not call methods on this
h = NewHealthHandler(…)  // now h points to a real HealthHandler — safe
```

**Where it appears in this project:**
- All handler, service, and repository structs use pointer receivers
- DB connections (`*gorm.DB`, `*sql.DB`) are always passed as pointers
- Constructors (`NewXxx`) always return a pointer to the created struct

---

## Interfaces

**First seen in:** `internal/health/health.go` — `type Checker interface`

### What they are

An interface in Go defines a set of method signatures. Any type that implements
all those methods automatically satisfies the interface — no `implements` keyword needed.
This is called **implicit (structural) implementation**.

```go
// The interface — defines the contract
type Checker interface {
    Name()  string
    Check(ctx context.Context) CheckResult
}

// DatabaseChecker satisfies Checker automatically
// because it has both Name() and Check() with matching signatures.
type DatabaseChecker struct { db *sql.DB }

func (c *DatabaseChecker) Name() string                      { return "database" }
func (c *DatabaseChecker) Check(ctx context.Context) CheckResult { … }
```

### Why Go chose implicit interfaces

In Java or C#, you write `class DatabaseChecker implements Checker`. This creates a
hard dependency: `DatabaseChecker` must know about `Checker` at compile time.

Go flips it: **the consumer defines the interface**. The `health` package defines
`Checker`, and any type anywhere — even in a third-party library — can satisfy it
without being modified. This is why new health checkers (cache, external API) can be
added to this project without changing the handler or the registry.

### Interfaces enable testing without real dependencies

Because `Checker` is an interface, tests can substitute a fake implementation
(a gomock mock) that returns controlled results — no real Postgres needed:

```go
// In production: DatabaseChecker hits real Postgres
registry.Register(health.NewDatabaseChecker(db))

// In tests: MockChecker returns whatever the test needs
mockChecker.EXPECT().Check(gomock.Any()).Return(health.CheckResult{Status: "unhealthy"})
```

### The Go interface convention

- **Keep interfaces small** — the smaller the interface, the easier it is to satisfy.
  `Checker` has two methods. `io.Reader` (standard library) has one. Small is better.
- **Define interfaces in the consumer package** — `health` defines `Checker` because
  `health` consumes it. Repository interfaces are defined in `repository` for the same
  reason. This prevents circular imports.
- **Accept interfaces, return concrete types** — functions should accept an interface
  (flexible) and return a concrete struct (predictable). Never return an interface
  unless you have a strong reason.

**Where it appears in this project:**
- `health.Checker` — the extensible checker pattern
- `service.EventService` — the business logic contract consumed by handlers
- `repository.EventRepository` — the DB contract consumed by services

---

## Goroutines

**First seen in:** `cmd/api/main.go` — `go func() { srv.ListenAndServe() }()`

### What they are

A goroutine is a lightweight concurrent function. The `go` keyword launches a function
and returns immediately — the caller does not wait for it to finish.

```go
go func() {
    // this runs concurrently with the rest of main()
    srv.ListenAndServe()
}()
// execution continues here immediately — the server is running in the background
```

### Why we need one for the server

`srv.ListenAndServe()` blocks forever — it loops accepting connections until the server
is shut down. Without a goroutine, `main()` would be stuck there and never reach the
signal-waiting code below. The goroutine lets the server run in the background while
`main()` continues to wait for a shutdown signal.

### Goroutines vs OS threads

| | Goroutine | OS Thread |
|---|---|---|
| Initial stack size | ~8 KB | 1–8 MB |
| Managed by | Go runtime | OS kernel |
| Switching cost | Very cheap | Expensive (kernel context switch) |
| Typical count | Thousands | Tens to hundreds |

The Go runtime multiplexes many goroutines onto a small pool of OS threads. This is
why Go servers can handle tens of thousands of concurrent connections cheaply.

---

## Channels

**First seen in:** `cmd/api/main.go` — `quit := make(chan os.Signal, 1)`

### What they are

A channel is a typed pipe for communication between goroutines. Values are sent into
a channel with `<-` and received from it with `<-`.

```go
ch := make(chan int)   // unbuffered: sender blocks until receiver is ready
ch <- 42              // send 42 into the channel
v := <-ch             // receive from the channel — blocks until a value arrives
```

### The shutdown pattern

```go
// make(chan os.Signal, 1) — buffered channel that holds 1 value.
// The buffer of 1 ensures the OS can deliver the signal even if we are not
// reading at the exact moment it arrives (prevents the signal from being dropped).
quit := make(chan os.Signal, 1)

// signal.Notify registers quit to receive SIGTERM and SIGINT from the OS.
signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

// <-quit blocks — main() sleeps here doing nothing until a signal arrives.
// This is the correct way to "wait" in Go: use a channel, not a sleep loop.
sig := <-quit
```

When Fly.io restarts the app or you press Ctrl+C, the OS sends SIGTERM or SIGINT.
`<-quit` unblocks, we log the signal, and the graceful shutdown begins.

### Buffered vs unbuffered channels

```go
make(chan os.Signal)    // unbuffered — sender blocks until someone reads
make(chan os.Signal, 1) // buffered(1) — sender can send 1 value without blocking
```

For signal handling, a buffer of 1 is standard practice — it prevents the unlikely
but possible case where the OS sends the signal before `<-quit` is reached.

---

## Multiple return values and nil as "no error"

**First seen in:** `internal/config/config.go` — `return Config{...}, nil`

### Go functions can return more than one value

The convention for any operation that can fail is to return `(result, error)`:

```go
func Load() (Config, error)
//            ^^^^^^  ^^^^^
//            result  nil means "everything went fine"
```

`error` is a built-in interface in Go:

```go
type error interface {
    Error() string
}
```

Because it is an interface, its zero value is `nil` — meaning "no error".

### The two paths

```go
// Success path — return the value AND nil (no error)
return Config{
    Port:        getEnv("PORT", "8080"),
    DatabaseURL: dbURL,
}, nil

// Failure path — return zero value AND a real error
if dbURL == "" {
    return Config{}, errors.New("DATABASE_URL is required but not set")
}
```

`Config{}` is the zero value of the struct — all fields empty. It is a placeholder;
the caller must check the error before using the struct.

### The caller always checks the error immediately

```go
cfg, err := config.Load()
if err != nil {     // "did something go wrong?"
    log.Fatal(err)  // yes — stop the program with the error message
}
// only here is cfg safe to use — err was nil
```

This pattern appears on virtually every I/O function in Go. It is the language's answer
to exceptions — no `try/catch`, just a value you inspect. The compiler cannot force you
to check errors, but Go convention (and `go vet`) will flag ignored errors.

### Why not just panic?

Go has `panic` (like throwing an exception) but it is reserved for truly unrecoverable
situations — bugs like nil pointer dereferences, out-of-bounds access. For expected
failure modes (missing env var, file not found, network error), returning an `error`
value keeps the caller in control of how to handle it: log and exit, retry, return a
different error, or wrap it with more context.

---

## Stack vs Heap

**First seen in:** `internal/health/checker.go` — `return &Registry{}`

### Two memory regions

Every running program has two regions of memory:

**Stack** — fast, automatic, short-lived.
Each function call gets a "frame" on the stack. Local variables live there. When the
function returns, the frame is destroyed — all its variables vanish instantly.
The OS manages this; no garbage collector involved.

**Heap** — slower, longer-lived, garbage collected.
Memory that needs to outlive the function that created it goes here. Go's garbage
collector tracks heap allocations and frees them when nothing points to them anymore.

### How Go decides — escape analysis

You never call `malloc`/`free` in Go. The compiler runs **escape analysis**: if it detects
that a value's address will outlive the function (e.g. you return `&r`), it automatically
places the value on the heap. Otherwise it stays on the stack.

```go
func newRegistry() Registry {
    r := Registry{}  // r lives on the STACK — dies when this function returns
    return r         // a COPY of r is returned to the caller
}

func NewRegistry() *Registry {
    r := &Registry{} // & signals: "this must outlive me" — Go moves r to the HEAP
    return r         // pointer to heap allocation — caller can use it indefinitely
}
```

You can inspect what Go decided:

```bash
go build -gcflags="-m" ./...
# output includes: "r escapes to heap"
```

### Stack vs Heap comparison

| | Stack | Heap |
|---|---|---|
| Speed | Very fast | Slower (GC overhead) |
| Lifetime | Dies with the function | Lives until GC collects it |
| Who frees it | OS automatically | Go garbage collector |
| When to use | Short-lived values, small structs | Anything returned as a pointer, shared state |

### The practical rule in this codebase

Dependency structs (`Registry`, `HealthHandler`, DB connections) are always heap-allocated
and passed as pointers — because the whole application must share **one** instance, and it
must outlive `main`'s inner scopes.

Small computation values (`CheckResult`, `int`, `string`) are fine as stack copies — they
are cheaper and the GC never touches them.

```go
// Heap — one shared instance, pointer passed everywhere
func NewRegistry() *Registry       { return &Registry{} }
func NewHealthHandler(...) *HealthHandler { return &HealthHandler{...} }

// Stack — cheap copy, short-lived, no pointer needed
func buildCheckResult(err error) CheckResult {
    return CheckResult{Status: StatusUnhealthy, Error: err.Error()}
}
```

---

## How these concepts connect to Functional Programming

The service layer (`internal/service/`) follows FP discipline. The three concepts above
each have a specific relationship with FP principles.

### `context.Context` — explicit dependencies (FP-friendly)

FP requires that a function's dependencies are visible in its signature — no hidden
globals, no implicit state. Passing `ctx context.Context` explicitly is exactly this:
instead of reading from a global "current request" variable (a hidden side effect),
every dependency travels through the function signature.

```go
// FP-friendly: all inputs visible in the signature.
// This function cannot access anything that isn't in ctx, input, or its arguments.
func validateAndEnrich(ctx context.Context, input SubmitEventInput, now time.Time) (model.Event, error)
```

### Pointer receivers — the tension with FP

Pointer receivers *allow* mutation. Mutation is the opposite of FP immutability.
The project resolves this tension with a clear layer rule:

| Layer | Pointer receiver purpose | FP rule |
|---|---|---|
| `handler/` | Holds dependencies (logger, service). OK to use pointer receivers. | No FP rule — handler is infrastructure. |
| `repository/` | Holds `*gorm.DB`. OK to use pointer receivers. | No FP rule — repository is infrastructure. |
| `service/` | Holds repository interfaces. Pointer receiver on the struct is OK for DI. | **Business logic functions must not mutate inputs.** Return new values instead. |

```go
// The service struct uses a pointer receiver — that's fine, it's dependency injection.
// But the logic inside is pure: input is not mutated, a new value is returned.
//
// FP: immutability
// Instead of modifying the event we received, we construct and return a new one.
// The caller's original value is untouched. This prevents bugs where shared state
// is modified unexpectedly across function boundaries.
func (s *eventService) Approve(e model.Event, by model.User, at time.Time) model.Event {
    return model.Event{
        ID:              e.ID,
        Status:          model.EventStatusApproved,
        LastUpdatedByID: by.ID,
        UpdatedAt:       at,
        // All other fields copied from e — original e is never touched
    }
}
```

The non-FP version (what to avoid) would look like:
```go
// BAD: mutates the input struct — side effect, breaks immutability
func (s *eventService) Approve(e *model.Event, by model.User) {
    e.Status = model.EventStatusApproved // caller's struct is now changed — surprising
}
```

### Interfaces — pure functions and testability (FP-friendly)

Interfaces make FP testability possible in Go. A service function that accepts a
`repository.EventRepository` interface (rather than a concrete `*gorm.DB`) has no
hidden dependency on a database. You can call it with a mock that returns deterministic
values — which is exactly what a pure function requires: same input, same output,
no observable side effects.

```go
// Because EventRepository is an interface, this function's behaviour is fully
// determined by its arguments — no hidden DB state, no global connections.
// That makes it testable like a pure function.
func (s *eventService) Submit(ctx context.Context, input SubmitEventInput) (model.Event, error) {
    existing, err := s.repo.FindUserByEmail(ctx, input.Submitter.Email)
    // s.repo is the interface — in tests it is a mock with controlled return values
}
```

### The layered FP strategy

```
cmd/api/main.go       ← wires everything together (side effects: start server, connect DB)
internal/handler/     ← handles HTTP (side effects: read request, write response)
internal/repository/  ← handles DB (side effects: SQL queries)
internal/service/     ← pure business logic (no I/O, no mutation, FP rules enforced)
```

Side effects are pushed to the edges (handler, repository). The service layer stays
pure. This is the architecture of most FP programs: a pure core surrounded by an
impure shell that handles I/O.
