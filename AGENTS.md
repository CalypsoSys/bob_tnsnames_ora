# Repository Guidelines

## Project Structure & Module Organization
This repository is a small single-package Go module named `tnsnames`. Core parsing logic lives in `parser.go`, package documentation is in `doc.go`, and tests are in `parser_test.go`. Top-level project context lives in `README.md`, while CI is defined in `.github/workflows/test.yml`.

Keep new production code in the repository root unless the package is intentionally being split. Add tests alongside the feature they cover, using the existing `*_test.go` pattern.

## Build, Test, and Development Commands
- `go test ./...`: run the full test suite locally and in CI.
- `go test -run TestParseBasicAlias ./...`: run a focused test while iterating.
- `go test -cover ./...`: check coverage before opening a PR.
- `gofmt -w *.go`: format all Go files in the module.

There is no separate build step today; this library is validated through tests.

## Coding Style & Naming Conventions
Follow standard Go formatting and idioms. Use tabs as produced by `gofmt`, keep imports grouped by `gofmt`, and prefer small exported types and methods with clear doc comments when they are part of the public API.

Use `CamelCase` for exported identifiers (`ParseFile`, `DescriptorDetails`) and lower camel case for unexported helpers. Preserve the package’s current style of concise error messages prefixed with `tnsnames:` for library-facing errors.

## Testing Guidelines
Tests use Go’s built-in `testing` package. Name tests with descriptive `Test...` functions that reflect behavior, such as `TestParseSyntaxError`. Cover both valid descriptors and failure cases, especially parser errors and partial descriptor extraction.

Before submitting changes, run `go test ./...` and add or update tests for any public behavior change.

## Commit & Pull Request Guidelines
Recent commits use short, imperative subjects such as `Update Go to 1.26.2` and `Add README badges`. Keep commit messages brief, specific, and focused on one change.

Pull requests should include a clear summary, note any API or behavior changes, and mention the test command you ran. Link related issues when applicable. Screenshots are not needed for this repository unless documentation output changes in a meaningful way.
