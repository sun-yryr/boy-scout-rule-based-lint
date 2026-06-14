# code_changed

## 概要

同じ file/message でもソース行が変わると hash が一致せず、新規 issue として出力される。

## 期待される動作

- stdout: hash 不一致になった issue
- 新規 issue 数: 1 以上（exit 1 相当）
