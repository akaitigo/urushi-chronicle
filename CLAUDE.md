# urushi-chronicle

## 概要

蒔絵・螺鈿作品の制作工程を詳細に記録・共有するデジタルアーカイブ。漆の塗り重ね各層の乾燥条件をIoTセンサーで自動記録し、蒔絵の金粉蒔き・螺鈿の貝片配置をマクロ撮影で記録する。漆風呂の温湿度管理と経年変化トラッキングを提供する。

## 技術スタック

- Go: IoTセンサーデータ収集サービス（MQTT, goroutine）
- TypeScript/React: タイムライン型制作記録UI（SPA）
- PostgreSQL + TimescaleDB: 時系列環境データ
- GCP Cloud Run + Cloud Storage: ホスティング・画像ストレージ

## コーディングルール

- TypeScript: `~/.claude/rules/typescript.md` のルールに従うこと（`any`禁止、`as`最小限）
- Go: 標準の Go スタイルガイドに従うこと。golangci-lint の設定は `.golangci.yml` を参照

## ビルド & テスト

### フロントエンド（TypeScript/React）
```bash
npm install          # 依存インストール
npm run dev          # 開発サーバー起動
npm run build        # プロダクションビルド
npx vitest run       # テスト実行
npx oxlint .         # リント
npx biome format .   # フォーマット
```

### バックエンド（Go）
```bash
go build ./...                    # ビルド
go test -v -race ./...            # テスト
golangci-lint run ./...           # リント
gofumpt -w .                      # フォーマット
```

### 共通
```bash
make check           # lint → test → build 一括
make quality         # 品質ゲート
```

## ディレクトリ構造

```
.
├── backend/
│   ├── cmd/urushi-chronicle/main.go  # Go エントリポイント
│   ├── internal/                     # Go 内部パッケージ
│   ├── pkg/                          # Go 公開パッケージ
│   └── go.mod
├── frontend/
│   ├── src/                          # React SPA
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── lib/
│   │   └── types/
│   ├── public/
│   ├── package.json
│   └── tsconfig.json
├── test/api/                         # Hurl API テスト
├── docs/                             # ADR, quality-override
├── Makefile
└── CLAUDE.md
```

## 環境変数

```bash
DATABASE_URL=          # PostgreSQL + TimescaleDB 接続文字列
MQTT_BROKER_URL=       # MQTTブローカーURL
GCS_BUCKET=            # Cloud Storage バケット名
JWT_SECRET=            # JWT署名キー
ALERT_WEBHOOK_URL=     # アラート通知先（Slack等）
```
