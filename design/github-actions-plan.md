# GitHub Actions 対応 実装計画

## 課題

通常のlinterは GitHub Actions の `::error` / `::warning` workflow command を出力することで、PRのdiff上にインラインのアノテーション（コードコメント）を表示できる。

```
::error file=internal/foo.go,line=10,col=5::unused variable 'x' (unused)
```

しかし `bsr check` にパイプすると、baseline で抑制されるべきエラーまでアノテーションとして表示されてしまう（現状 `check` はパースできない行をそのまま passthrough するため）。また、linterによっては GitHub Actions 形式で出力するが `bsr` の `LineParser` がその形式をパースできない。

## 方針

`bsr check` に `--format=github-actions` を追加し、新規エラーだけを `::error` 形式で出力する。加えて、`bsr` をセットアップする composite action を提供する。

---

## Phase 1: bsr CLI に `--format` フラグを追加

### 概要

```
bsr check --format=default          # 現行動作（元のlint行をそのまま出力）
bsr check --format=github-actions   # ::error file=...,line=...,col=...::message 形式で出力
```

### `--format` 仕様

| 値 | 動作 |
|---|------|
| `default` | 現行互換。元のlint行をそのまま stdout に出力 |
| `github-actions` | パース済み Issue から `::error file=...,line=...,col=...::message` を生成。パース不可な行は passthrough |

### 実装方針

`check()` 関数の中の出力箇所（現在 `fmt.Fprintln(stdout, line)` が4箇所）を、format に応じて出力を切り替える。

```go
func formatOutput(format string, issue *parser.Issue, rawLine string) string {
    switch format {
    case "github-actions":
        return fmt.Sprintf("::error file=%s,line=%d,col=%d::%s",
            issue.File, issue.Line, issue.Column, issue.Message)
    default:
        return rawLine
    }
}
```

課題: Boy Scout policy の分岐やパース失敗時は `issue` がなく `rawLine` だけの場合がある。
→ `parser.Issue` がある場合のみフォーマット変換し、それ以外は常に rawLine を passthrough。

### 変更ファイル

| ファイル | 変更内容 |
|---------|---------|
| `cmd/check.go` | `--format` フラグ追加。出力箇所で format に応じた出力切り替え |

### CLI 例

```bash
# GitHub Actions で使う
golangci-lint run ./... | bsr check --format=github-actions --boy-scout-policy=hunk --base-ref=origin/main
```

---

## Phase 2: GitHub Action の提供

### 2-a. `bsr-check` composite action

`.github/actions/bsr-check/action.yml` として配置（リポジトリ内蔵）。

```yaml
name: 'bsr check'
description: 'Run lint with bsr baseline filter and output GitHub Actions annotations'
inputs:
  linter-command:
    description: 'Linter command to run'
    required: true
  baseline:
    description: 'Path to the baseline file'
    required: false
    default: '.bsr-baseline.json'
  boy-scout-policy:
    description: 'Boy Scout policy (off, file, hunk)'
    required: false
    default: 'off'
  base-ref:
    description: 'Git base ref for Boy Scout policy'
    required: false
    default: ''
  bsr-version:
    description: 'Version of bsr to use'
    required: false
    default: 'latest'
runs:
  using: 'composite'
  steps:
    - name: Setup bsr
      uses: jdx/mise-action@v3
      with:
        tools: |
          bsr
    - name: Run bsr check
      shell: bash
      run: |
        ${{ inputs.linter-command }} | bsr check \
          --baseline ${{ inputs.baseline }} \
          --format=github-actions \
          --boy-scout-policy ${{ inputs.boy-scout-policy }} \
          ${{ inputs.base-ref && format('--base-ref {0}', inputs.base-ref) || '' }}
```

### 2-b. 利用者側のワークフロー例

```yaml
# .github/workflows/lint.yaml
name: Lint with Boy Scout Rule

on:
  pull_request:

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0  # Boy Scout Policy の git diff に必要

      - name: Setup Go
        uses: actions/setup-go@v6
        with:
          go-version: '1.25'

      - name: bsr lint check
        uses: sun-yryr/boy-scout-rule-based-lint/.github/actions/bsr-check@main
        with:
          linter-command: 'golangci-lint run ./...'
          boy-scout-policy: 'hunk'
          base-ref: 'origin/main'
```

### 2-c. 別リポジトリ向けの配布方法（marketplace action）

複合 action は同一リポジトリ内でしか `uses: ./` で参照できない。
他リポジトリから使う場合、以下の方法がある:

| 方法 | Pros | Cons |
|------|------|------|
| README に YAML スニペットを記載（推奨） | シンプル。ユーザがコピペで使える | action 自体のバージョン管理が手動 |
| `sun-yryr/bsr-action` リポジトリを別途作成 | 専用actionとして配布・バージョニング可能 | 管理コスト増 |
| GoReleaser で配布するバイナリを使う composite action | bsrバイナリのダウンロードだけで動く | 初期セットアップの記述が増える |

**最初は README の YAML スニペット方式**で十分。ユーザは go install / mise / 直接バイナリDL で bsr をインストールし、以下のスニペットを `.github/workflows/lint.yaml` に貼れば動く:

```yaml
- name: Run lint with Boy Scout baseline
  run: |
    golangci-lint run ./... | bsr check --format=github-actions --boy-scout-policy=hunk --base-ref=origin/main
```

---

## 実装順序

### Step 1: `--format` フラグの実装（`cmd/check.go`）

- `outputFormat`変数と `--format` フラグを追加
- `check()` に `format string` を渡す
- 出力箇所で `format` に応じた文字列を生成

### Step 2: テスト

- `check_test.go` を作成（cmd以下にテストファイルはないため新規）
- `default` フォーマット（passthrough）のテスト
- `github-actions` フォーマットのテスト
- パース不可行の passthrough テスト

### Step 3: composite action（`.github/actions/bsr-check/action.yml`）

- リポジトリ内に composite action を作成
- bsr プロジェクト自身の CI にも適用して dogfooding

### Step 4: README に利用ガイド追加

- 「GitHub Actions Annotations」セクションを追加
- YAML スニペットを掲載
- `--format=github-actions` の説明
