# all_suppressed

## 概要

baseline と check の lint 入力が同一のとき、既知 issue はすべて抑制され stdout は空になる。

## 期待される動作

- stdout: 空
- 新規 issue 数: 0（exit 0 相当）
