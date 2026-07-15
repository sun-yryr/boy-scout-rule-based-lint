# Fingerprint & Match Strategy Design

## 2軸に分ける

**1. 同じエラーかどうかの判定方法**

`exact / smart / loose / scope` のような match strategy。

**2. 変更した場所では baseline を許すかどうか**

`allow / changed-file / changed-hunk / changed-scope` のような Boy Scout policy。

この2つを分けるのが大事。「同じエラーを追跡したい」のと「触った場所は掃除させたい」は似てるけど別問題。

---

## 多段 fingerprint（baseline v2 schema）

```json
{
  "version": 2,
  "entries": [
    {
      "file": "internal/context/extractor.go",
      "line": 42,
      "column": 8,
      "rule": "errcheck",
      "message": "Error return value of `file.Close` is not checked (errcheck)",
      "count": 1,
      "fingerprints": {
        "context_hash": "sha256:...",
        "line_hash": "sha256:...",
        "context_bag_hash": "sha256:...",
        "scope": "go:method:Extractor.Extract",
        "node_kind": "call_expression",
        "node_hash": "sha256:..."
      }
    }
  ]
}
```

| fingerprint           | 役割                 | 強さ            |
| --------------------- | ------------------ | ------------- |
| `context_hash`        | 今の完全一致。速い          | 厳密だが壊れやすい     |
| `line_hash`           | エラー行そのものの正規化 hash  | 行移動に強い        |
| `context_similarity`  | 周辺行の Jaccard 類似度など | 近傍編集に強い       |
| `scope`               | 関数/メソッド/class 等のパス | 行移動・周辺編集に強い   |
| `node_kind/node_hash` | エラー位置の AST ノード     | より細かい識別       |
| `rule`                | linter rule id     | message 揺れに強い |

- `Rule` は optional。汎用フォーマットから `--rule-regex` で抽出可能にする。
- 今の `context_hash` は fast path として残す。

---

## Match Strategy

### `exact`

現行互換。`file + message + context_hash` 完全一致。最も安全。

### `smart`（デフォルト候補）

```text
file が同じ
AND rule/message が同じ
AND (
  context_hash 完全一致
  OR line_hash 一致
  OR context_similarity >= 0.6
  OR scope 一致 + line_hash/context 類似
)
```

ポイント: **scope 一致だけでは通さない**。同じ関数内に同じ rule の新規エラーが増えた場合に誤って baseline 扱いするのを防ぐ。scope は「補助証拠」に留める。

### `scope`

Tree-sitter ありの強めモード。

```text
file + rule/message + scope + node/line/context のどれか
```

関数ごと移動、上に import 追加、周辺数行追加、くらいなら耐える。

### `loose`

PHPStan 寄り。`file + message/rule + count`。行番号を持たず、行移動に強い。ただし「古い1件が直って、新しい1件が同じ file/message で増えた」ケースを見逃しやすい。明示 opt-in が安全。

### スコアリング方式（`smart` / `scope` 向け）

boolean 連鎖より score のほうが拡張しやすい。

```text
required:
  same file
  same rule OR same normalized message

score:
  +100 exact context_hash
  +40  line_hash exact
  +30  context similarity >= 0.7
  +20  same scope
  +10  same node kind
  +0〜10 line distance small

accept:
  score >= 60
  ただし scope だけでは accept しない
```

---

## Count 消費（SessionMatcher）

`Count` は baseline 作成時には増やすが、filter 時の matcher は「候補があれば true」を返すだけで残数を減らしていない。baseline に `count: 1` しかなくても、同じ key の current issue が複数出た場合に全部 suppression される可能性がある。

matcher は stateless な `Match()` ではなく、1回の filter 実行中だけ state を持つ `SessionMatcher` にする。

```go
type SessionMatcher struct {
    baseline  *Baseline
    remaining map[EntryID]int
}

func (m *SessionMatcher) Match(entry Entry) bool {
    candidate := m.bestCandidate(entry)
    if candidate == nil {
        return false
    }
    if m.remaining[candidate.ID] <= 0 {
        return false
    }
    m.remaining[candidate.ID]--
    return true
}
```

これを入れるだけでも `loose` や fuzzy match を安全寄りにできる。

---

## Tree-sitter: `scope fingerprint` と `changed-scope policy`

### 1. `scope` fingerprint

エラー行・column から最小 named node を探し、親を辿って function/method/class/type を見つける。

```text
go:method:Extractor.Extract
ts:class:UserService.method:createUser
php:class:App\Service\UserService.method:create
```

### 2. `changed-scope` policy

```text
変更された関数/メソッド/class 内の lint error は baseline で隠さない
変更されていない scope の lint error は baseline で隠す
```

`changed-file` だと大きいファイルを1行触っただけで既存エラー全部が出て厳しすぎる。`changed-hunk` だと「関数を触ったんだからその関数くらいはきれいにしよう」という思想が弱い。`changed-scope` はちょうどいい落としどころ。

### プロバイダインタフェース

```go
type FingerprintProvider interface {
    Supports(path string) bool
    Fingerprint(path string, loc Location) (Fingerprint, error)
}
```

最初は `.go`, `.js`, `.ts`, `.php` など人気どころだけ対応し、未対応拡張子は text fingerprint fallback。

---

## CLI オプション案

```bash
bsr filter \
  --match=smart \
  --changed-policy=allow
```

Boy Scout 強め:

```bash
bsr filter \
  --match=scope \
  --changed-policy=changed-scope \
  --base-ref=origin/main
```

### `--match` 候補

| value   | 説明 |
| ------- | ---- |
| `exact` | 現行互換。file + message + context_hash |
| `smart` | line_hash / fuzzy context / scope を組み合わせる |
| `scope` | Tree-sitter scope を重視する |
| `loose` | file + rule/message + count。PHPStan 風 |

### `--changed-policy` 候補

| value            | 説明 |
| ---------------- | ---- |
| `allow`          | 変更ファイルでも baseline を許す（現行寄り） |
| `changed-file`   | 変更されたファイルでは baseline を無効化 |
| `changed-hunk`   | 変更 hunk 周辺の diagnostic だけ baseline 無効化 |
| `changed-scope`  | 変更された関数/メソッド/class scope 内で baseline 無効化 |

`changed-policy` は `--base-ref` か `--diff` を受け取れると CI で使いやすい。

```bash
git diff --unified=0 origin/main...HEAD | bsr filter --diff - --changed-policy=changed-scope
```

---

## 推奨設定

| 用途 | match | changed-policy |
| ---- | ----- | -------------- |
| デフォルト（後方互換） | `exact` | `allow` |
| 推奨（現実的） | `smart` | `allow` |
| Boy Scout 強め | `smart` | `changed-hunk` |
| 理想形 | `scope` | `changed-scope` |

---

## 実装順

1. **`Count` 消費を入れる** — fuzzy 化する前に誤 suppression の上限を作る
2. **baseline v2 に `line`, `column`, `line_hash` を追加** — `Issue` には line/column があるので保存するだけ
3. **既存の `context_lines` で fuzzy match** — Jaccard 類似度を使い「周辺に1行追加された」くらいを救う
4. **`--match=smart` を追加** — Tree-sitter なしで始める
5. **Tree-sitter scope を optional provider として追加**
6. **`changed-policy` を追加** — まず `changed-file` / `changed-hunk`、その後に `changed-scope`

---

## 注意点

- Tree-sitter の `scope` だけで一致させるのは危ない。同じ関数内に同じ種類のエラーが複数あると混ざる。必ず `line_hash`、`node_hash`、`context_similarity`、`count` 消費とセットにする。
- `--report-stale-baseline` や `bsr prune` があると、未使用 entry を見つけて削除できる導線ができて良さそう（PHPStan 思想の踏襲）。
