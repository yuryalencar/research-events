# Go Package Management

Go has a built-in module system. No separate tool to install — it ships with the Go binary.

---

## The files

| Go | pnpm equivalent | Purpose |
|---|---|---|
| `go.mod` | `package.json` | declares the module name + direct + indirect deps |
| `go.sum` | `pnpm-lock.yaml` | cryptographic checksums of every downloaded module — never edit by hand |

---

## The commands

| Go command | pnpm equivalent | What it does |
|---|---|---|
| `go get github.com/foo/bar@latest` | `pnpm add foo` | adds a dependency (writes to `go.mod`) |
| `go get github.com/foo/bar@v1.2.3` | `pnpm add foo@1.2.3` | pins a specific version |
| `go mod tidy` | `pnpm install` (closest) | removes unused deps, adds missing ones, syncs `go.sum` |
| `go mod download` | `pnpm fetch` | downloads all deps to local cache |
| `go install tool@latest` | `pnpm dlx` / global install | installs a CLI tool (e.g. `mockgen`, `goose`, `air`) |

---

## Key differences from pnpm

### 1. No `node_modules` folder
Go downloads packages to a global cache (`$GOPATH/pkg/mod`, usually `~/go/pkg/mod`).
Every project on your machine shares the same cache. Nothing to `.gitignore` locally.

### 2. Import paths are URLs, not names
In pnpm you do `import { x } from "react"`. In Go the import path is the full module path:

```go
import "github.com/golang-jwt/jwt/v5"
```

This is also how `go get` knows where to fetch from — it's literally a URL.

### 3. `// indirect` comments
You'll see this in `go.mod` when a package was added but no Go file actually `import`s it yet.
Once you write the import and run `go mod tidy`, it becomes a direct dependency (comment disappears).

### 4. Major versions are part of the import path
Major versions ≥ 2 are embedded in the path: `github.com/pressly/goose/v3`.
This means v2 and v3 can coexist in the same project without conflicts — each is a different import path.

---

## Typical workflow

```bash
# Add a new library
cd backend && go get github.com/some/package@latest

# Remove unused deps / sync after deleting imports
cd backend && go mod tidy

# Install a CLI tool (used in Makefile/scripts, not imported as a library)
go install go.uber.org/mock/mockgen@latest
```

**Short version: `go get` = `pnpm add`, `go mod tidy` = cleanup pass, no `node_modules` ever.**
