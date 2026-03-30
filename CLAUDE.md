# urushi-chronicle

蒔絵・螺鈿作品の制作工程を記録・共有するデジタルアーカイブ。IoTセンサーで漆風呂の温湿度を自動記録。

## 技術スタック
- Go: IoTデータ収集+REST API
- TypeScript/React: タイムライン型UI
- PostgreSQL+TimescaleDB / GCP Cloud Run+Storage

## ルール
- TypeScript: `~/.claude/rules/typescript.md`
- Go: golangci-lint + gofumpt

## コマンド
```
make check     # lint → test → build
make quality   # 品質ゲート
```

## 構造
```
backend/cmd/  internal/  pkg/  go.mod
frontend/src/ components/ hooks/ lib/ types/
test/api/     docs/adr/
```

## 環境変数
`.env.example` を参照
