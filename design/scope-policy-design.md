# Scope Policy 実装設計

## 概要

`--boy-scout-policy=scope` は「変更された関数/メソッド/class 内の lint error は baseline で隠さない」ポリシー。tree-sitter を使って diff の変更行が属する scope を解決し、同じ scope 内の lint error をレポートする。

## データ構造

### `ChangeSet` 拡張

```go
type ChangeSet struct {
    Files         map[string]bool
    ChangedLines  map[string]map[int]bool
    ChangedScopes map[string]map[string]bool // file -> scope -> true （scope 用）
}

func (cs *ChangeSet) HasFile(path string) bool {
    return cs.Files[path]
}

func (cs *ChangeSet) HasLine(path string, line int) bool { ... }

func (cs *ChangeSet) HasScope(path, scope string) bool {
    return cs.ChangedScopes[path][scope]
}
```

### `ScopeResolver` インタフェース

```go
// diff パッケージに配置
type ScopeResolver interface {
    ResolveScope(file string, line int) (string, error)
}
```

tree-sitter への依存はインタフェースの背後に隠蔽し、`diff` パッケージは tree-sitter を import しない。

## `ResolveScopes`: diff 行から scope への解決

```go
func (cs *ChangeSet) ResolveScopes(resolver ScopeResolver) error {
    cs.ChangedScopes = make(map[string]map[string]bool)
    for file, lines := range cs.ChangedLines {
        scopes := make(map[string]bool)
        for line := range lines {
            scope, err := resolver.ResolveScope(file, line)
            if err != nil {
                continue // 未対応言語などはスキップ
            }
            scopes[scope] = true
        }
        if len(scopes) > 0 {
            cs.ChangedScopes[file] = scopes
        }
    }
    return nil
}
```

- `ChangedLines` の各 (file, line) について scope を問い合わせる
- 変更行数分だけの呼び出しなので現実的な計算量（100行のdiffなら100回のtree-sitterパース）

## `check()` での利用フロー

### `check()` シグネチャに `ScopeResolver` を追加

```go
func check(
    stdin io.Reader, stdout io.Writer,
    baselinePath string,
    policy string,
    changeSet *diff.ChangeSet,
    scopeResolver diff.ScopeResolver, // policy=="scope" のときのみ non-nil
) (int, error)
```

### `runCheck` での分岐

```go
if boyScoutPolicy == "scope" {
    resolver := treesitter.NewScopeResolver() // 言語プロバイダを束ねた実装
    changeSet.ResolveScopes(resolver)
}
newIssues, err := check(os.Stdin, os.Stdout, baselineFile, boyScoutPolicy, changeSet, resolver)
```

### ループ内でのマッチ

```go
if changeSet != nil {
    switch policy {
    case "file":
        if changeSet.HasFile(issue.File) { ... }
    case "hunk":
        if changeSet.HasLine(issue.File, issue.Line) { ... }
    case "scope":
        scope := extractScopeForIssue(issue, scopeResolver) // issue の行から scope を解決
        if changeSet.HasScope(issue.File, scope) { ... }
    }
}
```

**ポイント**: `scope` ポリシーの場合、lint error（issue）側でも scope を解決する必要がある。変更行の scope 集合と、issue の scope を比較して一致したらレポート。

## `treesitter.ScopeResolver` 実装

```go
// internal/treesitter/scope.go
package treesitter

type ScopeResolver struct {
    providers []ScopeProvider // 言語ごとのプロバイダ
}

type ScopeProvider interface {
    Supports(path string) bool
    ResolveScope(path string, line int) (string, error)
}

func (r *ScopeResolver) ResolveScope(file string, line int) (string, error) {
    for _, p := range r.providers {
        if p.Supports(file) {
            return p.ResolveScope(file, line)
        }
    }
    return "", fmt.Errorf("no scope provider for %s", file)
}
```

### scope 文字列のフォーマット

拡張子から言語を判別し、tree-sitter の named node を辿って scope を構築:

```text
go:method:Extractor.Extract
go:func:parseHunkNewStart
ts:class:UserService.method:createUser
php:class:App\Service\UserService.method:create
```

### Scope 解決アルゴリズム

1. 指定行・列から最小の named node を取得
2. 親を辿り、最初に見つかった function / method / class / type declaration を取り出す
3. その name を取得
4. `lang:kind:name` 形式で返す

見つからない場合は親方向の探索を続け、トップレベルでも見つからなければ空文字（またはエラー）。

## 実装順序

### Step 1: `ScopeResolver` インタフェース + `ChangeSet` 拡張

1. `internal/diff/parser.go` に `ScopeResolver` インタフェースを追加
2. `ChangeSet` に `ChangedScopes` フィールドと `HasScope`、`ResolveScopes` を追加
3. テストは mock `ScopeResolver` でカバー

### Step 2: `treesitter.ScopeResolver` 実装

1. `internal/treesitter/` パッケージを作成
2. `ScopeResolver` 本体（プロバイダのレジストリ）
3. Go の `ScopeProvider` 実装（tree-sitter-go 使用）
4. 対応言語を増やすごとに `ScopeProvider` を追加

### Step 3: `check()` 統合

1. `check()` のポリシー switch に `"scope"` ケースを追加
2. issue 側の scope 解決ロジック
3. `runCheck()` での ResolveScopes 呼び出し
4. `boyScoutPolicy == "scope"` のバリデーションエラーを解除

### Step 4: その他の `diff` 取得パス対応（任意）

1. `--diff` フラグ（ファイルからdiff読み込み）の追加
2. ワーキングツリー変更の `git diff` サポート

## 注意点

- **scope 解決は issue 側と diff 側の両方で必要**。同じ関数内に同じルールの別エラーがあると scope 一致だけでは区別できないため、scope は Boy Scout policy（変更領域の判定）にだけ使い、match strategy では `line_hash` や `node_hash` と組み合わせる
- **変更行が0のファイル**（削除のみ）では `ChangedScopes[file]` は空になる → `HasScope` は常に false、baseline が有効のまま。これは正しい動作
- **未対応言語**のファイルは `ResolveScope` がエラーを返す → `ResolveScopes` で skip、そのファイルの lint error は baseline が有効のまま
