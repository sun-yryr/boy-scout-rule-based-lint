[日本語](README_ja.md)

# bsr (Boy Scout Rule)

A CLI tool that adds [PHPStan-like baseline](https://phpstan.org/user-guide/baseline) functionality to any lint tool.

Based on the Boy Scout Rule — "Leave the code better than you found it" — bsr allows existing errors while preventing new ones from being introduced.

## Installation

**mise**
```bash
mise use --pin github:sun-yryr/boy-scout-rule-based-lint
```

**go install**
```bash
go install github.com/sun-yryr/boy-scout-rule-based-lint/cmd/bsr@latest
```

**Manual Build**

```bash
git clone https://github.com/sun-yryr/boy-scout-rule-based-lint.git
cd boy-scout-rule-based-lint
go build -o bsr .
```

## Usage

### Initialize Baseline

Register all current lint output as the baseline:

```bash
golangci-lint run ./... | bsr init
```

### Filter New Errors

Output only new errors that are not in the baseline:

```bash
golangci-lint run ./... | bsr check
```

Returns exit code 1 if new errors are found, 0 otherwise.
```

## Options

```
-b, --baseline string   Path to the baseline file (default: ".bsr-baseline.json")
-c, --context int       Number of context lines used for matching (default: 2)
```

## Supported Formats

The following lint output formats are supported:

- `file:line:column: message` (golangci-lint, ESLint unix format, buf, etc.)
- `file:line: message`
- `file(line,column): message` (Visual Studio format)
- `file(line): message`
- ESLint stylish format (default ESLint output; multi-line with file path and indented issue lines)

## Verified Tools

### ESLint

**Default output (stylish)**

No extra configuration is required:

```sh
eslint . | bsr init
eslint . | bsr check
```

**Unix format** (one issue per line, `file:line:column: message`)

```sh
# ESLint 8 and earlier (unix formatter included in core)
eslint -f unix . | bsr check

# ESLint 9 and later (unix formatter removed from core)
npm install -D eslint-formatter-unix
eslint -f unix . | bsr check
```

Starting with ESLint v9.0.0, the `unix` formatter is no longer bundled with ESLint. Install the [`eslint-formatter-unix`](https://www.npmjs.com/package/eslint-formatter-unix) package separately. See the [ESLint v9 migration guide](https://eslint.org/docs/latest/use/migrate-to-9.0.0) for details.

### Buf

```sh
$ buf --version
1.63.0
$ buf lint | bsr check
```

### golangci-lint

```sh
$ golangci-lint --version
golangci-lint has version 2.8.0 built with go1.25.5 from e2e40021 on 2026-01-07T21:29:47Z
$ golangci-lint run ./... | bsr check
```

## Context-Based Matching

bsr uses the code context surrounding error lines for matching, not just line numbers. This allows it to correctly track the same errors even when line numbers shift due to code additions or deletions.

## CI/CD Example

```yaml
# GitHub Actions
- name: Run lint with baseline
  run: |
    golangci-lint run ./... | bsr check
```

## License

MIT
