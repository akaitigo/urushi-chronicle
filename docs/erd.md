# Entity-Relationship Diagram

## 概要

urushi-chronicle のコアデータモデル。蒔絵・螺鈿の制作工程と IoT 環境データを紐付けて管理する。

## ER 図

```
┌─────────────────────────┐
│         works            │
├─────────────────────────┤
│ id          UUID    [PK] │
│ title       VARCHAR(200) │
│ description TEXT         │
│ technique   VARCHAR(50)  │  ← makie | raden | makie_raden | other
│ material    VARCHAR(100) │
│ status      VARCHAR(20)  │  ← in_progress | completed | archived
│ started_at  TIMESTAMPTZ  │
│ completed_at TIMESTAMPTZ │
│ created_at  TIMESTAMPTZ  │
│ updated_at  TIMESTAMPTZ  │
└──────────┬──────────────┘
           │ 1
           │
           │ *
┌──────────┴──────────────┐        ┌──────────────────────────────┐
│     process_steps        │        │     environment_readings      │
├─────────────────────────┤        ├──────────────────────────────┤
│ id          UUID    [PK] │        │ time          TIMESTAMPTZ    │  ← hypertable key
│ work_id     UUID    [FK] │───┐    │ sensor_id     VARCHAR(50)    │
│ name        VARCHAR(100) │   │    │ location      VARCHAR(100)   │
│ description TEXT         │   │    │ temperature   DOUBLE         │  ← -10.0 ~ 100.0
│ step_order  INT          │   │    │ humidity      DOUBLE         │  ← 0.0 ~ 100.0
│ category    VARCHAR(50)  │   ├───>│ work_id       UUID    [FK]   │
│ materials_used JSONB     │   │    │ process_step_id UUID  [FK]   │──┐
│ notes       TEXT         │   │    └──────────────────────────────┘  │
│ started_at  TIMESTAMPTZ  │   │                                      │
│ completed_at TIMESTAMPTZ │<─────────────────────────────────────────┘
│ created_at  TIMESTAMPTZ  │
│ updated_at  TIMESTAMPTZ  │
└──────────┬──────────────┘
           │
           │ *                    ┌──────────────────────────────┐
┌──────────┴──────────────┐      │     alert_thresholds          │
│        images            │      ├──────────────────────────────┤
├─────────────────────────┤      │ id              UUID    [PK] │
│ id              UUID [PK]│      │ sensor_id       VARCHAR(50)  │
│ work_id         UUID [FK]│      │ temperature_min DOUBLE       │
│ process_step_id UUID [FK]│      │ temperature_max DOUBLE       │
│ file_path       TEXT     │      │ humidity_min    DOUBLE       │
│ file_size_bytes BIGINT   │      │ humidity_max    DOUBLE       │
│ content_type    VARCHAR  │      │ enabled         BOOLEAN      │
│ image_type      VARCHAR  │      │ created_at      TIMESTAMPTZ  │
│ caption         VARCHAR  │      │ updated_at      TIMESTAMPTZ  │
│ taken_at        TIMESTAMPTZ│     └──────────────────────────────┘
│ created_at      TIMESTAMPTZ│
└─────────────────────────┘
```

## リレーション

| 関係 | カーディナリティ | 削除時の振る舞い |
|------|-----------------|-----------------|
| works → process_steps | 1:N | CASCADE（作品削除で工程も削除） |
| works → images | 1:N | CASCADE（作品削除で画像も削除） |
| process_steps → images | 1:N | SET NULL（工程削除でも画像は保持） |
| works → environment_readings | 1:N | SET NULL（作品削除でもデータは保持） |
| process_steps → environment_readings | 1:N | SET NULL（工程削除でもデータは保持） |

## 工程カテゴリ

| カテゴリ | 日本語 | 説明 |
|---------|--------|------|
| shitanuri | 下塗り | 漆の下地塗り |
| nakanuri | 中塗り | 中間層の塗り |
| uwanuri | 上塗り | 最終の塗り層 |
| makie | 蒔絵 | 金粉・銀粉による装飾 |
| raden | 螺鈿 | 貝片による装飾 |
| togidashi | 研ぎ出し | 漆を研いで下の装飾を出す |
| roiro | 呂色仕上げ | 最終光沢仕上げ |
| other | その他 | 上記以外の工程 |

## TimescaleDB 設計

- `environment_readings` テーブルは `time` カラムをパーティションキーとする hypertable
- 5 分間隔のセンサーデータを効率的に格納・クエリするために最適化
- `sensor_id` + `time DESC` の複合インデックスでセンサー別の最新データ取得を高速化
