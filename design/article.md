# bsr: ボーイスカウトルールをCIで検証可能にするlintベースラインツール

> **「来たときよりも綺麗に（Leave the code better than you found it.）」**
>
> —— Robert C. Martin, *Clean Code*

## はじめに: コード品質改善のジレンマ

あなたのチームはこう言われた経験はないだろうか。

「linterのルールを厳しくしよう」「でも既存コードに警告が2,000件出るんだよな」「じゃあ今度やろう」

あるいはこうだ。

「今回のPR、linter通った？」「いや、エラーあるけど前からあるやつだから…」「それ前も聞いた気がする」

linterはコード品質を保つ強力なツールだ。しかし**導入時点での既存エラー**という壁の前で、多くのチームが挫折する。全件修正するには時間が足りない。かといって導入を先送りにすれば状況は悪化する一方だ。

PHPStanのベースライン機能はこの問題に対して「既存エラーのリストを記録し、新規エラーだけを検出する」という解法を提供し、PHPコミュニティで広く支持されている[1]。では、この考え方を**あらゆるlinter**に適用できたらどうだろうか？

それが **bsr (Boy Scout Rule)** だ。

## bsr とは

bsr は任意のlinter出力にベースライン機能を追加するCLIツールである。仕組みはシンプルだ。

```
# 1. 現在のlintエラーをベースラインとして記録
golangci-lint run ./... | bsr init

# 2. 新規エラーだけを検出（CIに組み込む）
golangci-lint run ./... | bsr check
```

bsr init で生成された `.bsr-baseline.json` をリポジトリにコミットすれば、その後のlint実行では既存エラーを無視し、**新たに追加されたエラーだけ**を検出する。終了コードが1になるのは新規エラーがあるときだけなので、CIのチェックとしてそのまま使える。

### linter非依存

bsrの最大の特徴は**特定のlinterに依存しない**ことだ。golangci-lint、eslint、buf、Stylelint、RuboCop——標準的な `file:line: message` 形式の出力であれば、どんなツールでもパイプでつなぐだけでベースライン化できる。PHPStanのベースラインがPHPStan専用であるのに対し、bsrは「linterのためのベースライン基盤」として設計されている。

```bash
eslint src/ | bsr check
buf lint | bsr check
golangci-lint run ./... | bsr check
```

## 真のイノベーション: Boy Scout Policy

ここまで聞くと「単なるエラーフィルタか」と思うかもしれない。bsrの真価は **Boy Scout Policy** にある。

通常のベースラインは「既存エラーは全て許容する」という姿勢だ。しかしこれには問題がある。**変更したファイルも含めて**全ての既存エラーが無視されるため、「PRで触った関数のlintエラーを直さずにマージする」ことができてしまう。

ボーイスカウトルールの本質は「**触った場所は掃除して去る**」ことだ。bsrの Boy Scout Policy はこれをCI上で**ルールとして強制**する。

```bash
# 変更があったファイルではbaselineを無効化する（触ったファイルは必ず掃除）
bsr check --boy-scout-policy=file --base-ref=origin/main

# 変更があった行のエラーのみbaselineを無効化する
bsr check --boy-scout-policy=hunk --base-ref=origin/main

# 変更があった関数/メソッドscopeのエラーのみbaselineを無効化する（将来対応）
bsr check --boy-scout-policy=scope --base-ref=origin/main
```

### 3つのポリシーレベル

| ポリシー | 動作 | ユースケース |
|---------|------|-------------|
| `off`（デフォルト） | 変更箇所でもbaselineが有効 | 後方互換、従来のベースライン |
| `file` | 変更ファイル全体でbaseline無効 | 「触ったファイルくらいは綺麗にしろ」という思想 |
| `hunk` | 変更行のエラーのみbaseline無効 | 最小限の掃除義務。大きなファイルの部分修正に有効 |
| `scope`（計画中） | 変更された関数/メソッド単位でbaseline無効 | 理想形。「その関数を触ったならその関数くらいは整えろ」 |

### なぜこれが重要なのか

従来のベースラインは「今あるエラーを追跡する」だけだった。Boy Scout Policyはこれに「**触った場所は掃除させる**」という**行動規範**を加える。この2つを組み合わせることで、初めて「来たときよりも綺麗に」をCIで自動検証できるようになる。

```yaml
# GitHub Actionsでの使用例
- name: Boy Scout Lint Check
  run: |
    golangci-lint run ./... | bsr check \
      --boy-scout-policy=hunk \
      --base-ref=origin/main
```

この一行で、PRの変更行に新たにlintエラーを追加していないこと**かつ**触った行の既存lintエラーが解消されていることを検証できる。

## ベースラインの徐々に減らす文化

bsrが作りたいのは「ベースラインを少しずつ減らしていく」という開発文化だ。

- 新機能開発のついでに、触ったファイルのlintエラーを1つ直す
- Boy Scout PolicyがそれをCIで検証する
- チーム全体でベースラインが減っていくのが可視化される
- いつの間にかベースラインが空になり、linterのルールレベルを上げられる

これは『Clean Code』で語られるボーイスカウトルールの実践そのものだ。

> 「コードに散らかしを残す行為は、ゴミのポイ捨てと同じくらい社会的に許容されないようにすべきだ」
>
> —— Robert C. Martin

bsr はこの思想を、**願望からCIで検証可能なルールへ**と昇華する。

## マッチ戦略: エラーを賢く追跡する

ベースラインの実用上の課題は「コードが移動・変更されても同じエラーを正しく追跡できるか」だ。bsr は多段のfingerprint（指紋）を用いたマッチ戦略を提供する。

| 戦略 | 説明 | 耐性 |
|------|------|------|
| `exact` | ファイル+メッセージ+行ハッシュの完全一致 | 最も安全。行移動に弱い |
| `smart` | 行ハッシュ+コンテキスト類似度+scopeの複合 | 周辺編集・行移動に強い |
| `loose` | ファイル+メッセージのみ（PHPStan方式） | 行移動に強いが誤検出のリスクあり |
| `scope`（計画中） | tree-sitterで関数scopeを考慮 | 関数移動に強い |

さらに、誤った抑制を防ぐために**カウント消費**の仕組みも備えている。ベースラインに「このエラーは1件」と記録されていれば、同じ種類の新規エラーが2件発生しても1件しか抑制されない。

## なぜlinter本体ではなく外部ツールなのか

「ベースライン機能はlinter本体に実装すべきでは？」という疑問は当然だ。実際、PHPStanは本体に組み込まれている。

bsrを外部ツールにした理由は3つある。

1. **普遍性**: golangci-lint にも ESLint にも RuboCop にも、あらゆるlinterに同じ思想を適用できる
2. **アップデート不要**: linter本体のバージョンや機能追加を待たずに、bsr単体で改善を続けられる
3. **関心の分離**: lintはlint、ベースライン管理はbsr。それぞれが独立して進化できる

これは grep / sort / uniq のUNIX哲学そのものだ。単機能のツールをパイプで組み合わせて、全体として強力なシステムを作る。

## 今後の展望

現在はα版として開発中で、以下の機能を実装中だ。

- [x] ベースライン初期化 (`init`)
- [x] 新規エラーフィルタリング (`check`)
- [x] カウント消費（過剰抑制の防止）
- [x] Boy Scout Policy: `file`, `hunk`
- [ ] `--match=smart`（複合マッチ戦略）
- [ ] Boy Scout Policy: `scope`（tree-sitterによる関数単位の判定）
- [ ] `prune` コマンド（未使用ベースラインエントリの自動削除）
- [ ] `report-stale-baseline`（消失したエラーを通知）

## 参考

1. [PHPStan - The Baseline](https://phpstan.org/user-guide/baseline) — ベースライン機能の先駆者であり、bsrの設計思想の原点
2. [DevIQ - Boy Scout Rule](https://deviq.com/principles/boy-scout-rule) — ボーイスカウトルールの原則解説
3. Robert C. Martin, *Clean Code: A Handbook of Agile Software Craftsmanship* (Prentice Hall, 2008) — 第1章にボーイスカウトルールの記述

---

**bsr** は [GitHub](https://github.com/sun-yryr/boy-scout-rule-based-lint) で公開中（MIT License）。
