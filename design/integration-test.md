# 統合テスト設計

## 目的

BSR（Boy Scout Rule lint filter）のエンドツーエンド動作を検証する統合テストを整備する。

- ベースライン作成 → コード変更 → `bsr check` の一連の流れを再現
- 将来の新しいlintツール対応時のリグレッション防止
- コントリビューションの敷居を下げる（シナリオ1つ追加するだけでテスト追加可能）

---

## 設計方針

### テストフィクスチャ = ディレクトリ

テストケースは `testdata/scenarios/` 配下の **ディレクトリ** として定義する。  
Markdownはテストを **定義するものではなく、ドキュメントとして** 使う。

| 役割 | ファイル | 説明 |
|------|---------|------|
| ドキュメント | `README.md` | linter名・バージョン・再現手順・期待結果 |
| ソースファイル | `*.go`, `*.ts`, `*.js` など | lint対象のコード。baseline入力の `file:line:` で参照される |
| baselineフェーズ入力 | `baseline.txt` | `bsr init` にパイプするlint出力 |
| checkフェーズ入力 | `check.txt` | `bsr check` にパイプするlint出力 |
| 期待出力 | `expected.txt` | `bsr check` の期待stdout（空ファイル=全抑制でexit 0） |

### ディレクトリ構造

```
testdata/scenarios/
├── all_suppressed/                # 全抑制（exit 0）
│   ├── README.md
│   ├── main.go                    # 両フェーズ共通のソース
│   ├── baseline.txt
│   ├── check.txt
│   └── expected.txt               # 空 = 出力なし、exit 0
│
├── new_issue/                     # 新規エラー検出（exit 1）
│   ├── README.md
│   ├── main.go
│   ├── baseline.txt
│   ├── check.txt
│   └── expected.txt               # 非空 = 新規エラー出力、exit 1
│
├── code_changed/                  # コード変更でhash不一致（exit 1）
│   ├── README.md
│   ├── before/                    # initフェーズのソース
│   │   └── main.go
│   ├── after/                     # checkフェーズのソース（変更後）
│   │   └── main.go
│   ├── baseline.txt
│   ├── check.txt
│   └── expected.txt
│
└── ... (その他シナリオ)
```

### フェーズ別ソースのルール

| 条件 | init で使うソース | check で使うソース |
|------|------------------|-------------------|
| 基本（`before/` も `after/` も無し） | ルートのソースファイル | 同じソースファイル |
| `before/` のみ存在 | `before/` 内のファイル | ルートのソースファイル |
| `after/`  のみ存在 | ルートのソースファイル | `after/` 内のファイル |
| 両方存在 | `before/` 内のファイル | `after/` 内のファイル |

---

## README.md のテンプレート

各シナリオの `README.md` は、人間が読んで再現できるドキュメントとする。

```markdown
# <シナリオ名>

## 概要
<このテストが検証することの簡潔な説明>

## 再現に使ったLinter

| 項目 | 値 |
|------|-----|
| Linter | golangci-lint |
| Version | v2.1.6 |
| Command | `golangci-lint run ./...` |
| 備考 | staticcheck, errcheck, unused を含むデフォルト設定 |

## 再現手順

1. ベースライン作成
   ```bash
   golangci-lint run ./... | bsr init -b baseline.json
   ```

2. コードを変更（before/ → after/）

3. 変更後のlint出力をチェック
   ```bash
   golangci-lint run ./... | bsr check -b baseline.json
   ```

## 期待される動作
- <期待するstdoutとexit codeの説明>
```

---

## テストランナー

### refactor: cmdパッケージのテスト容易化

テストランナーから直接呼べるように、`cmd/check.go` と `cmd/init.go` のコアロジックを `io.Reader` / `io.Writer` を受け取る関数に抽出する。

```go
// cmd/check.go

// check は lint 出力をベースラインと照合し、新規issueのみを出力する。
// stdin からlint出力を読み、stdout に新規issueを書き出す。
// 戻り値の int は新規issue数。
func check(stdin io.Reader, stdout io.Writer, baselinePath string) (int, error) { ... }

// CLIラッパー
func runCheck(cmd *cobra.Command, args []string) error {
    n, err := check(os.Stdin, os.Stdout, baselineFile)
    if err != nil { return err }
    if n > 0 { os.Exit(1) }
    return nil
}
```

```go
// cmd/init.go

// init_ は lint 出力からベースラインを作成する。
func init_(stdin io.Reader, baselinePath string) (int, error) { ... }

func runInit(cmd *cobra.Command, args []string) error {
    n, err := init_(os.Stdin, baselineFile)
    if err != nil { return err }
    fmt.Fprintf(os.Stderr, "Baseline created with %d entries\n", n)
    return nil
}
```

### テスト本体: `cmd/integration_test.go`

```go
func TestIntegration(t *testing.T) {
    entries, _ := os.ReadDir("testdata/scenarios")
    for _, e := range entries {
        if !e.IsDir() { continue }
        t.Run(e.Name(), func(t *testing.T) {
            runScenario(t, filepath.Join("testdata/scenarios", e.Name()))
        })
    }
}

func runScenario(t *testing.T, dir string) {
    // 1. ソースディレクトリの準備
    srcDir := t.TempDir()
    copySourceFiles(dir, srcDir, "before")  // デフォルトは dir 直下

    // 2. baseline作成
    baselinePath := filepath.Join(t.TempDir(), "baseline.json")
    baselineInput := readFile(filepath.Join(dir, "baseline.txt"))
    n, err := init_(
        strings.NewReader(baselineInput),
        baselinePath,
    )
    require.NoError(t, err)

    // 3. コード変更がある場合、ソースをafterに差し替え
    if hasAfterDir(dir) {
        copySourceFiles(dir, srcDir, "after")
    }

    // 4. check実行
    checkInput := readFile(filepath.Join(dir, "check.txt"))
    var stdout bytes.Buffer
    newCount, err := check(
        strings.NewReader(checkInput),
        &stdout,
        baselinePath,
    )
    require.NoError(t, err)

    // 5. 期待出力と比較
    expected := readFile(filepath.Join(dir, "expected.txt"))
    assert.Equal(t, expected, stdout.String())

    // 6. exitコード相当の検証
    if expected == "" {
        assert.Equal(t, 0, newCount, "all suppressed → exit 0")
    } else {
        assert.Greater(t, newCount, 0, "new issues found → exit 1")
    }
}
```

---

## 初期シナリオ一覧

| # | シナリオ名 | テスト内容 | before/after |
|---|-----------|-----------|-------------|
| 1 | `all_suppressed` | baselineとcheckが同一 → 全抑制、exit 0 | なし |
| 2 | `new_issue` | checkに新規エラー1件 → その行が出力、exit 1 | なし |
| 3 | `multiple_new` | checkに新規エラー複数件 → 全件出力 | なし |
| 4 | `resolved_issue` | 一部修正済み → 修正分は出力されない | なし |
| 5 | `code_changed` | 同行だがソース変更 → hash不一致で新規扱い | before/after |
| 6 | `visual_studio` | MSBuild形式 `file(line,col): message` の入力 | なし |
| 7 | `empty_input` | 空のlint出力 → 出力なし、exit 0 | なし |
| 8 | `unparsable` | パース不能行 → そのままパススルー出力 | なし |
| 9 | `mixed_changes` | 一部修正＋一部新規 → 新規分のみ出力 | なし |

---

## 実装手順

1. **[refactor] `cmd/check.go` `cmd/init.go`** — `io.Reader`/`io.Writer` パラメータを受け取る関数に抽出
2. **[add] `cmd/integration_test.go`** — シナリオランナー実装
3. **[add] シナリオ1-9** — `testdata/scenarios/` にテストケースを作成
4. **[verify] `go test ./cmd/ -run Integration -v`** — 全シナリオがパスすることを確認
5. **[ci] `.github/workflows/check.yaml`** に `go test ./...` が含まれていることを確認（既存のまま）

---

## 新しいlintツール対応の追加手順（contributeガイド）

1. `testdata/scenarios/<linter名>_<内容>/` ディレクトリを作成
2. `README.md` にlinter名、バージョン、再現コマンドを記述
3. ソースファイル、`baseline.txt`、`check.txt`、`expected.txt` を配置
4. `go test ./cmd/ -run Integration -v` で検証
5. PR

---

## 補足: ESLintルールテストとの比較

ESLintの `RuleTester` は単一ルールの入出力をコード内で簡潔に書けるが、BSRのテストは本質的に **2フェーズ（baseline作成 → check）** を必要とし、さらに **実際のソースファイルのhash計算** が絡むため、同じようにはできない。

代わりにディレクトリベースのフィクスチャ + READMEドキュメントの組み合わせにより、再現性と可読性の両立を図る。
