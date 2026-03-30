# urushi-chronicle — アーキテクチャ概要

## 主要な設計判断
- ADR は `docs/adr/` に記録する

## システム構成
- Backend (Go): IoTセンサーデータ収集 + REST API
- Frontend (React): タイムライン型制作記録UI
- Database: PostgreSQL + TimescaleDB（時系列データ）
- Storage: GCP Cloud Storage（画像）

## 外部サービス連携
- MQTT Broker: IoTセンサーからのデータ受信
- GCP Cloud Storage: マクロ撮影画像の保存
- GCP Cloud Run: デプロイ先
