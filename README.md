# bsr (Boy Scout Rule)

任意のlintツールにPHPStanのベースライン機能を追加するCLIツールです。

「来たときよりも綺麗に」の精神で、既存のエラーは許容しつつ新規エラーの追加を防ぎます。

## インストール

```bash
go install github.com/sun-yryr/boy-scout-rule-based-lint@latest
```

または、リポジトリをクローンしてビルド:

```bash
git clone https://github.com/sun-yryr/boy-scout-rule-based-lint.git
cd boy-scout-rule-based-lint
go build -o bsr .
```

## 使い方

### ベースラインの初期化

現在のlint出力をすべてベースラインとして登録します:

```bash
golangci-lint run ./... 2>&1 | bsr init
```

### 新規エラーのフィルタリング

ベースラインに登録されていない新規エラーのみを出力します:

```bash
golangci-lint run ./... 2>&1 | bsr filter
```

新規エラーがある場合は終了コード1、ない場合は0を返します。

### ベースラインの更新

現在のlint出力でベースラインを上書きします:

```bash
golangci-lint run ./... 2>&1 | bsr update
```

## オプション

```
-b, --baseline string   ベースラインファイルのパス (デフォルト: ".bsr-baseline.json")
-c, --context int       マッチングに使用するコンテキスト行数 (デフォルト: 2)
```

## 対応フォーマット

以下の形式のlint出力に対応しています:

- `file:line:column: message` (golangci-lint, eslint, buf など)
- `file:line: message`
- `file(line,column): message` (Visual Studio形式)
- `file(line): message`

## コンテキストベースマッチング

bsrは行番号だけでなく、エラー行の前後のコードコンテキストを使用してマッチングを行います。
これにより、コードの追加・削除で行番号がずれても、同じエラーを正しく追跡できます。

## CI/CDでの使用例

```yaml
# GitHub Actions
- name: Run lint with baseline
  run: |
    golangci-lint run ./... 2>&1 | bsr filter
```

## ライセンス

MIT
