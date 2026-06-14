[日本語](README_ja.md)

# bsr

bsr is a CLI tool that brings the Boy Scout Rule ("leave things cleaner than you found them") to your linter. It takes the output of any lint tool, suppresses errors recorded in the baseline, and reports only new ones. This lets you introduce strict rules even when there are many existing violations, and optionally enforce that any code you touch must be cleaned up before it can be merged.

This tool is inspired by [PHPStan's baseline feature](https://phpstan.org/user-guide/baseline).

## Features

- **Linter-agnostic** — works with golangci-lint, ESLint, buf, and any tool that emits `file:line:...` style output
- **Content-based matching** — errors are identified by a SHA-256 hash of the source line content, not by line number. Edit the line and the hash changes, so the error resurfaces as new
- **Count-based suppression cap** — never suppresses more than the recorded count; a single entry cannot mask an explosion of identical errors
- **Boy Scout Policy** — control which scope of "touched" code disables the baseline
- **Git-native** — pass `--base-ref` and bsr detects changes via `base...HEAD` diff
- **CI-friendly** — exits with code 1 only when new errors exist

## Installation

Download from GitHub Releases or use one of the methods below.

**mise** (recommended)

```bash
mise use --pin github:sun-yryr/boy-scout-rule-based-lint
```

**go install**

```bash
go install github.com/sun-yryr/boy-scout-rule-based-lint/cmd/bsr@latest
```

**Manual build**

```bash
git clone https://github.com/sun-yryr/boy-scout-rule-based-lint.git
cd boy-scout-rule-based-lint
go build -o bsr .
```

## Usage

### 1. Initialize the baseline

```bash
golangci-lint run ./... | bsr init
```

This generates `.bsr-baseline.json` with every current lint violation. Commit it to your repository.

When run from an interactive terminal, `bsr init` prompts for optional Boy Scout Policy settings and stores them in the baseline `config`. In non-interactive environments (for example, when stdin is a pipe), the prompt is skipped and no `config` is written.

### 2. Detect new violations

```bash
golangci-lint run ./... | bsr check
```

- Only violations not present in the baseline are printed
- Exits with code 1 if there are new violations, 0 otherwise
- Suppressed violations are not printed

### 3. Boy Scout Policy

With the default configuration (baseline only), you can modify a dirty file and merge it without fixing any of its existing errors. Enable the Boy Scout Policy and bsr ignores the baseline for files or lines you changed on this branch. Every error in those locations is reported.

You can store the default policy in `.bsr-baseline.json` so `bsr check` does not need flags on every run:

```json
{
  "version": 2,
  "config": {
    "boy_scout_policy": "hunk",
    "base_ref": "origin/main"
  },
  "entries": []
}
```

CLI flags override baseline config when specified:

```bash
# Uses config from .bsr-baseline.json
golangci-lint run ./... | bsr check

# Temporarily override the stored policy
golangci-lint run ./... | bsr check --boy-scout-policy file
```

You can also pass everything on the command line:

```bash
# All errors in changed files are reported, regardless of the baseline
golangci-lint run ./... | bsr check --boy-scout-policy file --base-ref origin/main

# Only errors on changed lines (inside hunks) are reported
golangci-lint run ./... | bsr check --boy-scout-policy hunk --base-ref origin/main
```

| Policy | Behavior | When to use |
|--------|----------|-------------|
| `off` (default) | Baseline applies across the entire codebase | Standard baseline workflow |
| `file` | Baseline is ignored for errors in changed files | If you touch the file, clean the whole file |
| `hunk` | Baseline is ignored only for errors on changed lines (inside hunks) | Only the lines you touched need to be clean (useful for large files) |
| `scope` | Planned — baseline ignored at function/method granularity | If you touch the function, clean the function |

Pass `--base-ref` (typically `origin/main`) whenever the policy is not `off`, or store it in baseline `config`.

Resolution order for `bsr check`:

1. CLI flags (`--boy-scout-policy`, `--base-ref`) when explicitly provided
2. Baseline `config`
3. Defaults (`off`, empty `base_ref`)

## How it works

### Matching strategies

bsr has two matching strategies internally, but only Exact is used. There is no CLI flag to switch between them.

#### Exact matching (current default)

`bsr init` reads the source line where each error occurred, normalizes whitespace, computes a SHA-256 hash, and records it in the entry:

```json
{
  "file": "internal/foo.go",
  "message": "error return value not checked",
  "source_line": "if err := do(); err != nil {",
  "count": 1,
  "fingerprints": {
    "line_hash": "a1b2c3..."
  }
}
```

`bsr check` does the same and suppresses any error whose hash matches an entry (subject to the `count` cap described below).

Because matching is based on line content, unrelated additions or deletions elsewhere don't break tracking. Conversely, editing the matching line changes its hash, causing the error to reappear as new. This is what powers bsr's "touch it, clean it" behavior.

#### Loose matching

A looser strategy keyed on `file:message` is also implemented. It survives reformatting but cannot distinguish between two occurrences of the same message in the same file, so it is not exposed via the CLI.

### Count-based suppression cap

Each baseline entry records how many times the error occurred (`count`). A single `check` run will never suppress more than the recorded count. For example, if an entry has `count: 1` and the current lint output produces two errors with the same key, the second is reported as new.

### Boy Scout Policy implementation

bsr runs `git diff --unified=0 <base-ref>...HEAD` to compute changed files and line ranges, then selectively disables baseline suppression in those regions according to the chosen policy.

## Options

### Common to all subcommands

```text
-b, --baseline string   Path to the baseline file (default: .bsr-baseline.json)
```

### bsr check

```text
--boy-scout-policy string   Boy Scout policy: off, file, hunk (default: off)
                            Overrides baseline config when provided
--base-ref string           Git base ref (e.g. origin/main)
                            Required when the policy is not off
                            Overrides baseline config when provided
```

## Supported formats

bsr handles the following lint output formats:

| Format | Example | Supported tools |
|--------|---------|-----------------|
| `file:line:column: message` | `main.go:42:3: unused variable x` | golangci-lint, ESLint unix format, buf, and more |
| `file:line: message` | `main.go:42: unused variable x` | Some linters |
| `file(line,column): message` | `main.go(42,3): unused variable x` | Visual Studio / MSBuild style |
| `file(line): message` | `main.go(42): unused variable x` | Some tools |
| ESLint stylish format | File path line followed by indented issue lines | ESLint default output |

Unparseable lines are passed through to stdout unchanged (summary lines, etc.).

## Tested linters

### golangci-lint

```bash
golangci-lint run ./... | bsr init
golangci-lint run ./... | bsr check
```

### ESLint

**Default output (stylish)** — works without extra configuration:

```bash
eslint . | bsr init
eslint . | bsr check
```

**Unix format** (one issue per line)

```bash
# ESLint 8 and earlier: unix formatter is bundled
eslint -f unix . | bsr check

# ESLint 9 and later: unix formatter is not bundled
npm install -D eslint-formatter-unix
eslint -f unix . | bsr check
```

Starting with ESLint v9.0.0, the `unix` formatter is no longer bundled with ESLint. Install the [`eslint-formatter-unix`](https://www.npmjs.com/package/eslint-formatter-unix) package separately (see the [ESLint v9 migration guide](https://eslint.org/docs/latest/use/migrate-to-9.0.0) for details).

### Buf

```bash
buf lint | bsr check
```

## CI/CD examples

### Basic — detect new errors only

```yaml
- name: Lint with baseline
  run: |
    golangci-lint run ./... | bsr check
```

### With Boy Scout Policy enabled

```yaml
- name: Checkout with full history
  uses: actions/checkout@v6
  with:
    fetch-depth: 0   # full history is required for git diff

- name: Run linter with Boy Scout policy
  run: |
    golangci-lint run ./... | bsr check \
      --boy-scout-policy hunk \
      --base-ref origin/main
```

## Roadmap

**Implemented**

- `bsr init` / `bsr check`
- Exact matching via source line hash
- Count-based suppression cap
- Boy Scout Policy: `file` and `hunk`
- ESLint stylish format support

**Designed / planned**

- `scope` policy (function/method granularity via tree-sitter)
- Smarter matching strategies (surrounding-context similarity, scope-aware matching)
- `prune` command to clean up stale baseline entries
- `--format=github-actions` for native PR annotations

## Contributing

- Bug reports and feature requests via GitHub Issues
- PRs adding support for new linter output formats are welcome

## License

MIT
