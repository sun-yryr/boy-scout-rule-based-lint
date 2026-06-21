# Changelog

## [0.2.0](https://github.com/sun-yryr/boy-scout-rule-based-lint/compare/v0.1.7...v0.2.0) (2026-06-21)


### ⚠ BREAKING CHANGES

* Remove update command and related documentation from README files
* rename cmd check instead of filter

### delete

* Remove update command and related documentation from README files ([f024756](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/f024756aeac7f024e5e6a48d475b60cee49ca437))


### Features

* add Version ([bdd4dc8](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/bdd4dc8f624d63f87242010305e61b4b8d8b44d5))
* iinitをv2に。最大限の情報を残す方向に変更 ([91587b8](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/91587b8a05a7da2a138de2ae587cd7e841429a4a))
* Implement Boy Scout Policy in check command ([c874dbb](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/c874dbb20ef25e893f367ae7e501af328269917d))
* include staged changes in Boy Scout diff detection ([80f043e](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/80f043e814895214a8cec4c7620ef88ea1579a01))
* normalize baseline file paths to git repo root ([f3da37d](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/f3da37df08b4553d553422d8178d8d247774d9b3))
* publish baseline v2 JSON Schema ([ad91e61](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/ad91e6189340b9a44b68a775ad03a48f468e597d))
* rename cmd check instead of filter ([9d5e974](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/9d5e974a22e5fa65c221c6adf4ca88125a8d1006))
* store Boy Scout policy defaults in baseline config ([e997275](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/e9972752fb0449248be2699a4dde57164765c832))
* Update matcher and extractor to improve context handling and matching strategy ([ef59c18](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/ef59c18df3c97e09960921772df283c4d06e8f43))


### Bug Fixes

* increase bufio.Scanner buffer size to avoid token-too-long errors ([70aa0d9](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/70aa0d9d1a19457add6ff8c32333dd4957215ff3))
* preserve readable source_line in baseline JSON ([477bb3c](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/477bb3c051546961901aaffa97a29a2d91e1abc0))
* propagate Extract errors instead of creating empty-hash entries ([1720b04](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/1720b04dabf46c7fafb735ceaecf27a3195e761d))
* validate lineNum in Extract and return error on invalid/out-of-range lines ([406e64f](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/406e64f50fa26142338c676a01d18d9ae244a600))
* スキーマの修正 ([8f022a4](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/8f022a4c33f91d397e877d871d17395216d38e3d))
* プロジェクトルート外のファイル読み取りを防ぐ ([b507c56](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/b507c56dd156c90863158081e2dfd7a0613689ae))
* 無制限のfallback matchを削除 ([8ad3e77](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/8ad3e777e080e5f758248b2a9749681512059bf6))

## [0.1.7](https://github.com/sun-yryr/boy-scout-rule-based-lint/compare/v0.1.6...v0.1.7) (2026-02-15)


### Performance Improvements

* Extractでファイルをメモリ全展開しない ([a362928](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/a36292887493fb8eafad4eb0c46e984fb13638c2))
* Extractでファイルをメモリ全展開しない ([8c8badc](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/8c8badc7c85cf9cf6c7a581205167fdf9280d5fd))

## [0.1.6](https://github.com/sun-yryr/boy-scout-rule-based-lint/compare/v0.1.5...v0.1.6) (2026-01-14)


### Bug Fixes

* go installでbsrがインストールできるようにする ([9f34471](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/9f3447126bd11503edc366275da7fc591c970903))

## [0.1.5](https://github.com/sun-yryr/boy-scout-rule-based-lint/compare/v0.1.4...v0.1.5) (2026-01-14)


### Bug Fixes

* リリースの調整中 ([00d4fbf](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/00d4fbf651898ddb07e32a8a299ba85e4a14de55))

## [0.1.4](https://github.com/sun-yryr/boy-scout-rule-based-lint/compare/v0.1.3...v0.1.4) (2026-01-14)


### wip

* release-please のoutputチェック ([5e834a8](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/5e834a8f30a864aa89c6b5df08addb00cd048725))


### Features

* bsr - lint baseline filter tool ([e27869a](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/e27869ac6c68333613e4c2b8d5f65956436719e7))


### Bug Fixes

* お試しコミット ([7d80443](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/7d8044384c75070149cbec8b90bd100eaf507949))
* リリースフローの調整中 ([6d1347f](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/6d1347f3d7f70db5f25e5692090f9b77b9397214))
* リリースフロー調整中 ([8497d3c](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/8497d3c7d885ba9f46c43b8ea211294d8cd67e41))
* リリースフロー調整中 mainの最新にタグを打つ ([98199d4](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/98199d4181d6b61ad0b6c8b5853979dca63587c1))


### Miscellaneous Chores

* release 0.1.3 ([ea95eca](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/ea95ecaed1e1e696b2e6f596e166aadab9c73b2e))
* release 0.1.4 ([9401e39](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/9401e396f412042a6fddcfda1f27953184625917))

## [0.1.3](https://github.com/sun-yryr/boy-scout-rule-based-lint/compare/v0.1.2...v0.1.3) (2026-01-14)


### wip

* release-please のoutputチェック ([5e834a8](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/5e834a8f30a864aa89c6b5df08addb00cd048725))


### Features

* bsr - lint baseline filter tool ([e27869a](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/e27869ac6c68333613e4c2b8d5f65956436719e7))


### Bug Fixes

* お試しコミット ([7d80443](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/7d8044384c75070149cbec8b90bd100eaf507949))
* リリースフローの調整中 ([6d1347f](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/6d1347f3d7f70db5f25e5692090f9b77b9397214))


### Miscellaneous Chores

* release 0.1.3 ([ea95eca](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/ea95ecaed1e1e696b2e6f596e166aadab9c73b2e))

## [0.1.2](https://github.com/sun-yryr/boy-scout-rule-based-lint/compare/v0.1.1...v0.1.2) (2026-01-14)


### Bug Fixes

* お試しコミット ([7d80443](https://github.com/sun-yryr/boy-scout-rule-based-lint/commit/7d8044384c75070149cbec8b90bd100eaf507949))
