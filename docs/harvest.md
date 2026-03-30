# Harvest: urushi-chronicle

## 使えたもの
- [x] Makefile (make check / make quality)
- [x] lint設定 (golangci-lint + oxlint + biome)
- [x] CI YAML
- [x] CLAUDE.md (28行、50行以下達成)
- [x] ADR テンプレート (1件: core-data-model)
- [x] 品質チェックリスト (make quality)
- [x] E2Eテスト雛形 (test/api/api.hurl)
- [x] Hooks（PostToolUse golangci-lint + oxlint）
- [x] lefthook
- [x] startup.sh

## 使えなかったもの（理由付き）
- 特になし。全テンプレートが機能した

## テンプレート改善提案

| 対象ファイル | 変更内容 | 根拠 |
|-------------|---------|------|
| golangci-lint テンプレート | exportloopref→copyloopvar, tenv→usetesting に更新 | deprecated linter警告 |
| lefthook.yml テンプレート | archgate存在チェック済み（v5.8で対応） | archgate未インストールでフック失敗 |
| biome.json テンプレート | v2スキーマ更新済み（v5.8で対応） | npm最新版でv2が入る |

## メトリクス

| 項目 | 値 |
|------|-----|
| Issue (closed/total) | 5/5 |
| PR merged | 5 |
| テスト数 | 45+ |
| CI失敗数 | 0 |
| ADR数 | 1 |
| テンプレート実装率 | 100% |
| CLAUDE.md行数 | 28 |

## 次のPJへの申し送り
- Go+TS モノレポでは post-lint.sh のファイル拡張子分岐が重要
- MQTT テストはモック subscriber で十分（実MQTT broker不要）
