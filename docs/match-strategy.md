# Match Strategy Design

## インタフェース

```go
type MatchStrategy interface {
    // GroupKey は remaining 消費と prune のためのエントリ識別子を返す。
    // exact/loose ではマッチ判定にも使うが、smart/scope ではマッチ判定には使わない。
    GroupKey(e Entry) string

    // Match は baseline entry と current lint issue が「同じエラー」かを判定する。
    // exact/loose では GroupKey の比較で済むが、smart ではスコアリングを行う。
    // マッチした場合は base 側のエントリの GroupKey を返す。
    // false の場合は key は空文字。
    Match(base Entry, current Entry) (matched bool, key string)

    // MatchAny は prune 用。baseline エントリが現行 lint 出力（candidates）の
    // いずれかとマッチするかを判定する。スマート戦略では全候補からスコアリングする。
    MatchAny(base Entry, candidates []Entry) (matched bool, key string)
}

type SessionMatcher struct {
    baseline  *Baseline
    strategy  MatchStrategy
    remaining map[string]int // GroupKey -> remaining count
}
```

## SessionMatcher

```go
func NewSessionMatcher(bl *Baseline, s MatchStrategy) *SessionMatcher {
    sm := &SessionMatcher{
        baseline:  bl,
        strategy:  s,
        remaining: make(map[string]int, len(bl.Entries)),
    }
    for _, e := range bl.Entries {
        sm.remaining[s.GroupKey(e)] += e.Count
    }
    return sm
}

func (sm *SessionMatcher) Match(current Entry) bool {
    for i := range sm.baseline.Entries {
        base := sm.baseline.Entries[i]
        matched, key := sm.strategy.Match(base, current)
        if !matched {
            continue
        }
        if sm.remaining[key] <= 0 {
            continue
        }
        sm.remaining[key]--
        return true
    }
    return false
}
```

## Strategy ごとの GroupKey と Match の関係

### exact

```go
func (s *ExactMatcher) GroupKey(e Entry) string {
    return e.File + ":" + e.Fingerprints.LineHash
}

func (s *ExactMatcher) Match(base, current Entry) (bool, string) {
    key := s.GroupKey(base)
    if s.GroupKey(current) == key {
        return true, key
    }
    return false, ""
}
```

GroupKey がそのままマッチ判定。1:1。

### loose

```go
func (s *LooseMatcher) GroupKey(e Entry) string {
    return e.File + ":" + e.Message
}

func (s *LooseMatcher) Match(base, current Entry) (bool, string) {
    key := s.GroupKey(base)
    if s.GroupKey(current) == key {
        return true, key
    }
    return false, ""
}
```

GroupKey がそのままマッチ判定。唯一 **N:1** 集約が起こる戦略。`remaining` に複数件が加算される。

### smart

```go
func (s *SmartMatcher) GroupKey(e Entry) string {
    // エントリを一意に識別するユニークID。マッチ判定には使わない。
    return fmt.Sprintf("%s:%d:%d:%s", e.File, e.Line, e.Column, e.Rule)
}

func (s *SmartMatcher) Match(base, current Entry) (bool, string) {
    score := 0
    if base.File != current.File || base.Rule != current.Rule {
        return false, ""
    }
    if base.ContextHash == current.ContextHash {
        score += 100
    }
    if base.LineHash == current.LineHash {
        score += 40
    }
    if jaccard(base.ContextLines, current.ContextLines) >= 0.7 {
        score += 30
    }
    if base.Scope == current.Scope {
        score += 20
    }
    if base.NodeKind == current.NodeKind {
        score += 10
    }
    if score >= 60 {
        return true, s.GroupKey(base)
    }
    return false, ""
}
```

- **マッチ判定**: スコアリング（GroupKey 非依存）
- **remaining 消費**: マッチした `base` エントリの GroupKey を使用
- **GroupKey の役割**: 純粋なユニーク識別子（file:line:column:rule）

### scope

```go
type ScopeMatcher = SmartMatcher // 基本同じ。scope の重みを大きくする程度。
```

## prune の設計

```go
func Prune(bl *Baseline, currentIssues []Entry, s MatchStrategy) *Baseline {
    var kept []Entry
    for _, base := range bl.Entries {
        matched, key := s.MatchAny(base, currentIssues)
        if !matched {
            continue // 現行 lint に一致なし → 削除
        }
        // マッチしたら残す（currentIssues 側のカウント消費は不要）
        kept = append(kept, base)
    }
    return &Baseline{Version: bl.Version, Entries: kept}
}
```

`exact`/`loose` の `MatchAny` は GroupKey の単純比較、`smart`/`scope` は全候補とのスコアリング。

## 責務の分離まとめ

| 戦略 | マッチ判定 | GroupKey | remainingキー | prune判定 |
|------|-----------|----------|--------------|-----------|
| `exact` | line_hash 比較 | `file:line_hash` | 同左 | line_hash 比較 |
| `loose` | file+message 比較 | `file:message` | 同左 | file+message 比較 |
| `smart` | **スコアリング** | `file:line:col:rule` | 同左 | **スコアリング** |
| `scope` | **スコアリング** | `file:line:col:rule` | 同左 | **スコアリング** |

- `exact`/`loose` は GroupKey と Match が一体化している
- `smart`/`scope` は GroupKey と Match が分離している
- prune は `Match`/`MatchAny` のメソッドをそのまま使うだけ
- `NormalizedBaseline` のような中間型は不要。`remaining map[string]int` で十分
