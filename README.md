# urushi-chronicle

蒔絵・螺鈿作品の制作工程を詳細に記録・共有するデジタルアーカイブ。

漆の塗り重ね各層の乾燥条件（温度・湿度・時間）をIoTセンサーで自動記録し、蒔絵の金粉蒔き・螺鈿の貝片配置をマクロ撮影で記録する。作品の経年変化トラッキング機能付き。

## 技術スタック

- **Backend**: Go (IoTデータ収集 + API)
- **Frontend**: TypeScript / React
- **Database**: PostgreSQL + TimescaleDB
- **Infrastructure**: GCP Cloud Run

> **ストア切り替え**: `STORE_TYPE` 環境変数で `postgres`（PostgreSQL+TimescaleDB）と `memory`（インメモリ）を切り替えられます。未設定の場合は `DATABASE_URL` の有無で自動判定します。

## セットアップ

```bash
# PostgreSQL+TimescaleDB を起動
docker compose up -d

# Backend
cd backend && go mod download && go build ./...

# Frontend
cd frontend && npm install && npm run dev
```

## データベースマイグレーション

スキーマは `backend/migrations/*.up.sql`（`NNN_説明.up.sql` 命名）で管理し、付属の
マイグレーションランナーで適用する。`schema_migrations` テーブルで適用済みバージョンを
追跡するため、`migrate` は冪等（未適用分のみ実行）。

```bash
# DATABASE_URL を設定して未適用のマイグレーションを適用する
export DATABASE_URL=postgres://urushi:urushi@localhost:5432/urushi_chronicle?sslmode=disable
make migrate
# または直接:
cd backend && go run ./cmd/migrate
```

環境変数:

- `DATABASE_URL`（必須）: PostgreSQL接続文字列
- `MIGRATIONS_DIR`（任意）: `*.up.sql` の配置ディレクトリ。デフォルト `migrations`

> **注意**: `docker-entrypoint-initdb.d` にマウントされたマイグレーションSQLは、PostgreSQLの初回起動時のみ実行されます。以降のスキーマ変更は `make migrate` で適用してください。既存ボリュームを作り直す場合:
> ```bash
> docker compose down -v && docker compose up -d
> ```

## 開発コマンド

```bash
make check     # lint → test → build
make quality   # 品質チェック
make format    # フォーマット
```

## ライセンス

MIT License - Copyright (c) 2026 Ryusei
