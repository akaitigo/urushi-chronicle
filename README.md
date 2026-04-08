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

> **注意**: `docker-entrypoint-initdb.d` にマウントされたマイグレーションSQLは、PostgreSQLの初回起動時のみ実行されます。スキーマを変更した場合は、既存のボリュームを削除してから再起動してください:
> ```bash
> docker compose down -v
> docker compose up -d
> ```

## 開発コマンド

```bash
make check     # lint → test → build
make quality   # 品質チェック
make format    # フォーマット
```

## ライセンス

MIT License - Copyright (c) 2026 Ryusei
