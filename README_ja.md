[English](README.md)

# bsr

bsr は、ボーイスカウトルール（「来たときよりも綺麗に」）をlintツールに持ち込むためのCLIです。
任意のlintツールの出力を受け取り、ベースラインに記録済みのエラーを抑制したうえで新規エラーのみを報告します。これにより、既存の違反が大量にあっても厳しいルールを導入でき、必要に応じて「触ったコードは綺麗にしてからマージしなければならない」という運用を可能にします。

このツールは [PHPStanのベースライン機能](https://phpstan.org/user-guide/baseline) から着想を得ています。

## 特徴

- **lintツール非依存** — golangci-lint、ESLint、buf など、`file:line:...` 形式の出力を持つツールならそのまま使えます
- **行内容ベースのマッチング** — 行番号ではなくソース行内容のSHA-256ハッシュでエラーを識別します。該当行を編集するとハッシュが変わり、エラーは新規として再検出されます
- **カウントによる抑制上限** — ベースラインに記録された件数を超えて抑制しません。1件のエントリが多数の同一エラーを覆い隠すことはありません
- **Boy Scout Policy** — どの範囲を「触った」と判定してベースラインを無効にするか調節できます
- **Gitネイティブ** — `--base-ref` を指定するだけで、ベースリファレンスからの差分（`base...HEAD`）で変更箇所を判定します
- **CI向き** — 新規エラーがあるときだけ終了コード1を返します

## インストール

GitHub Release からダウンロードするか、以下の方法でインストールしてください。

**mise**（推奨）

```bash
mise use --pin github:sun-yryr/boy-scout-rule-based-lint
```

**go install**

```bash
go install github.com/sun-yryr/boy-scout-rule-based-lint/cmd/bsr@latest
```

**手動ビルド**

```bash
git clone https://github.com/sun-yryr/boy-scout-rule-based-lint.git
cd boy-scout-rule-based-lint
go build -o bsr .
```

## 使い方

### 1. ベースラインの初期化

```bash
golangci-lint run ./... | bsr init
```

現在のlint違反をすべて記録した `.bsr-baseline.json` が生成されます。これをリポジトリにコミットしてください。

対話可能な端末で実行した場合、`bsr init` は Boy Scout Policy の設定を質問し、baseline の `config` に保存します。非対話環境（stdin がパイプの場合など）では質問をスキップし、`config` は書き込みません。

### 2. 新規違反の検出

```bash
golangci-lint run ./... | bsr check
```

- ベースラインに存在しない違反のみが出力されます
- 新規違反がある場合は終了コード1、ない場合は0を返します
- ベースラインで抑制された違反は出力されません

### 3. Boy Scout Policy

デフォルトの設定（ベースラインだけの運用）では、汚いファイルに変更を入れても既存のエラーを直さずにマージできてしまいます。Boy Scout Policy を有効にすると、このブランチで変更したファイルや行についてはベースラインを無視し、そこにあるエラーをすべて報告します。

デフォルトのポリシーは `.bsr-baseline.json` に保存でき、毎回フラグを指定しなくても `bsr check` で使えます。

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

CLI フラグは明示的に指定した場合、baseline の `config` より優先されます。

```bash
# .bsr-baseline.json の config を使う
golangci-lint run ./... | bsr check

# 一時的にポリシーを上書き
golangci-lint run ./... | bsr check --boy-scout-policy file
```

すべてをコマンドラインで指定することもできます。

```bash
# 変更したファイル内のエラーは、ベースラインに関係なくすべて報告
golangci-lint run ./... | bsr check --boy-scout-policy file --base-ref origin/main

# 変更した行（hunk内）のエラーのみ報告
golangci-lint run ./... | bsr check --boy-scout-policy hunk --base-ref origin/main
```

| ポリシー | 動作 | 想定ユースケース |
|----------|------|------------------|
| `off`（デフォルト） | ベースラインがコードベース全体で有効 | 通常のベースライン運用 |
| `file` | 変更したファイル内のエラーはベースラインを無視 | ファイルに触ったらファイル全体を綺麗にする |
| `hunk` | 変更行（hunk内）のエラーのみベースラインを無視 | 大きなファイルの部分修正で触った行だけを直す |
| `scope` | 計画中 — 関数／メソッド単位での無効化 | 触ったコードと同じスコープは綺麗にしてからマージする |

ポリシーを `off` 以外で使う場合は `--base-ref`（通常は `origin/main`）を指定するか、baseline の `config` に保存してください。

`bsr check` の解決順序:

1. 明示的に指定した CLI フラグ（`--boy-scout-policy`, `--base-ref`）
2. baseline の `config`
3. デフォルト（`off`、空の `base_ref`）

## 仕組み

### マッチング戦略

bsrは2種類のマッチング戦略を内部に持っていますが、現時点で利用されるのは Exact のみです。CLIから切り替える手段は提供していません。

#### Exact マッチング（現在のデフォルト）

`bsr init` はエラーが発生したソース行を読み取り、空白を正規化したうえでSHA-256ハッシュを計算してエントリに記録します。

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

`bsr check` 時も同じ処理を行い、ハッシュが一致すれば該当エントリで抑制します（後述の `count` の範囲内で）。

行内容で識別するため、無関係な行の追加・削除では追跡が壊れません。逆に該当行を編集するとハッシュが変わり、エラーは新規として浮上します。これがbsrの「触ったコードから綺麗になっていく」挙動を支えています。

#### Loose マッチング

`file:message` をキーにする緩いマッチング戦略も実装してあります。フォーマット変更などで行内容が変わっても抑制が維持される反面、同じファイル・同じメッセージのエラーであれば別の場所で発生したものも抑制してしまうため、CLIからは利用できません。

### カウントによる抑制の上限

各ベースラインエントリは、そのエラーの発生回数（`count`）を保持しています。1回の `check` 実行では、ベースラインに記録された回数までしか抑制されません。たとえば `count: 1` のエントリに対して同じキーのエラーが2件発生した場合、2件目は新規エラーとして報告されます。

### Boy Scout Policy の実装

`git diff --unified=0 <base-ref>...HEAD` と `git diff --cached --unified=0` を実行し、結果を合成して変更されたファイル・行範囲を計算します。ブランチ上のコミット済み変更と staged（index）の変更の両方を対象に、指定されたポリシーに応じて該当箇所のベースライン抑制を無効化します。

## オプション

### 全サブコマンド共通

```text
-b, --baseline string   ベースラインファイルのパス（デフォルト: .bsr-baseline.json）
```

### bsr check

```text
--boy-scout-policy string   ボーイスカウトポリシー: off, file, hunk（デフォルト: off）
                            指定時は baseline config を上書き
--base-ref string           Gitのベースリファレンス（例: origin/main）
                            ポリシーが off 以外の場合は必須
                            指定時は baseline config を上書き
```

## 対応フォーマット

bsrは以下のlint出力形式に対応しています。

| フォーマット | 例 | 主な対応ツール |
|--------------|----|----------------|
| `file:line:column: message` | `main.go:42:3: unused variable x` | golangci-lint、ESLint unix形式、buf など |
| `file:line: message` | `main.go:42: unused variable x` | 一部のlintツール |
| `file(line,column): message` | `main.go(42,3): unused variable x` | Visual Studio / MSBuild 形式 |
| `file(line): message` | `main.go(42): unused variable x` | 一部のツール |
| ESLint stylish形式 | ファイルパス行＋インデントされたissue行 | ESLintのデフォルト出力 |

パースできない行はそのまま標準出力にパススルーされます（サマリー行など）。

## 動作確認済みのlintツール

### golangci-lint

```bash
golangci-lint run ./... | bsr init
golangci-lint run ./... | bsr check
```

### ESLint

**デフォルト出力（stylish）** — 追加設定なしで動作します。

```bash
eslint . | bsr init
eslint . | bsr check
```

**unix形式**（1行1件）

```bash
# ESLint 8 以前: unix フォーマッターはコアに含まれる
eslint -f unix . | bsr check

# ESLint 9 以降: unix フォーマッターはコアから分離
npm install -D eslint-formatter-unix
eslint -f unix . | bsr check
```

ESLint v9.0.0 以降は `unix` フォーマッターがコアから外れているため、[`eslint-formatter-unix`](https://www.npmjs.com/package/eslint-formatter-unix) パッケージを別途インストールしてください（詳細は [ESLint v9 移行ガイド](https://eslint.org/docs/latest/use/migrate-to-9.0.0) を参照）。

### Buf

```bash
buf lint | bsr check
```

## CI/CDでの使用例

### 基本形（新規エラーのみ検出）

```yaml
- name: Lint with baseline
  run: |
    golangci-lint run ./... | bsr check
```

### Boy Scout Policy を有効にする場合

```yaml
- name: Checkout with full history
  uses: actions/checkout@v6
  with:
    fetch-depth: 0   # git diff のために履歴が必要

- name: Run linter with Boy Scout policy
  run: |
    golangci-lint run ./... | bsr check \
      --boy-scout-policy hunk \
      --base-ref origin/main
```

## ロードマップ

**実装済み**

- `bsr init` / `bsr check`
- ソース行ハッシュによる Exact マッチング
- カウントによる抑制の上限
- Boy Scout Policy: `file` と `hunk`
- ESLint stylish 形式の対応

**設計中・計画中**

- `scope` ポリシー（tree-sitter による関数／メソッド単位の判定）
- より賢いマッチング戦略（周辺コンテキストの類似度、スコープ考慮など）
- `prune` コマンド（使われなくなったベースラインエントリの整理）
- `--format=github-actions`（GitHub Actions アノテーションへのネイティブ対応）

## コントリビュート

- バグ報告・機能要望は GitHub Issues へお願いします
- 新しいlintツールのフォーマット対応も歓迎です

## ライセンス

MIT
