# AGENTS.md

## Cursor Cloud specific instructions

`bsr` is a single Go CLI (module `github.com/sun-yryr/boy-scout-rule-based-lint`); there are no servers, databases, or containers to run. Standard commands live in `README.md` and `.github/workflows/check.yaml`.

- **Go toolchain**: The system `go` is older than the version pinned in `go.mod` (`go 1.25.5`). Running any `go` command inside `/workspace` auto-downloads and uses the pinned toolchain, so `go build`/`go test` "just work" — do not be alarmed if `go version` outside the repo reports an older release.
- **Build / test / lint** (mirror CI):
  - `go build ./...`
  - `go test ./...`
  - `golangci-lint run ./...`
- **golangci-lint**: v2 is required (config `.golangci.yml` uses `version: "2"`). It is installed into `$(go env GOPATH)/bin`, which is already on `PATH`. `go install` is not the supported install method for golangci-lint v2; use the official `install.sh` script if it is missing.
- **Running the tool**: `go build -o bsr .` then pipe any linter output into it: `<lint-cmd> | ./bsr init` (writes `.bsr-baseline.json`), `<lint-cmd> | ./bsr check` (prints only new violations, exit 1 when new violations exist). `bsr init` skips the interactive Boy Scout Policy prompt when stdin is a pipe.
- **git dependency**: Boy Scout Policy modes (`--boy-scout-policy file|hunk` / `--base-ref`) shell out to `git diff`; run those from inside a git repo with the base ref available.
