# Contributor Guide

## Testing Instructions

- Find the CI pipeline in the `.github/workflows` directory, and ensure that all CI checks pass for you submitted PR.
- Fix any failing test before submitting a pull request. `go test ./...` should pass with no errors.
- Add or update godoc comments for any new or modified functions, types, and packages.
- Ensure that all new code is covered by tests. Use `go test -cover` to check coverage.
- Add or update tests for the code you modify or add.

## Code Style

- Use `gofmt` to format your code before committing.
- Follow Go's idiomatic style and conventions.
- Use meaningful variable and function names.
- Keep functions small and focused on a single task.
- Use comments to explain complex logic or decisions.
- Use `godoc` comments for public functions, types, and packages.
- Use `go doc` to check your documentation.
- Title format for commit messages: `Short description (50 characters or less)`.
- Use the imperative mood in commit messages (e.g., "Fix bug" instead of "Fixed bug").
- Use [semantic pull request](https://pulsar.apache.org/contribute/develop-semantic-title/ ) in the format `[type][scope] Summary` where:
  * `type` is one of `feat`, `improve`, `fix`, `cleanup`, `refactor`, or `revert`
  * `scope` is the affected area (e.g., `model`, `identifier`, `makefile`)
  * `Summary` is a present tense imperative sentence with a capitalized first letter and no period.

- Don't add a PR description. The maintainers will handle that.
- Don't add any labels to the PR. The maintainers will handle that.
- Add `godoc` comments for any new tests you write, explaining what the test does and why it's necessary.
- Update `godoc` comments for any existing functions, test, types, or packages that you modify.
- Update the `README.md` file with any new features or changes you make.

## Import Cycle Management

**CRITICAL RULE**: Never duplicate types, constants, or code to avoid import cycles. This is considered a serious architectural violation.

### The Problem
Import cycles occur when packages have circular dependencies (A imports B, B imports A). While Go's compiler prevents these, the solution is never to duplicate code.

### The Solution
When facing an import cycle, create a shared types package:

1. **Identify the root cause**: Find which shared types are causing the cycle
2. **Create a types package**: Make a new package (e.g., `core/types`) for shared primitive types
3. **Move shared types**: Place the types in the shared package
4. **Update imports**: Both original packages can now import the types package
5. **Verify**: Run `make build`, `make test`, and `make lint`

### Example
```go
// ❌ WRONG: Duplicating to avoid import cycle
type Level int64  // Duplicated from another package

// ✅ CORRECT: Using shared types package
import "github.com/thep2p/skipgraph-go/core/types"
func DoSomething(level types.Level) { ... }
```

### Detection
- Any comment saying "duplicated to avoid import cycle" should be flagged in code review
- Identical type definitions across packages indicate a problem
- Check for copy-pasted constants or enums

### Reference
See the `core/types` package for the established pattern in this codebase.