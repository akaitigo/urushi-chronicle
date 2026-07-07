# ADR-003: 画像アップロード用presigned URLの生成戦略

**ステータス**: 承認
**日付**: 2026-07-07
**決定者**: Ryusei

## コンテキスト

制作工程のマクロ撮影画像は GCP Cloud Storage に保存する。クライアント（フロントエンド）は
バックエンドに presigned URL を要求し、その URL に対して直接 PUT でアップロードする。

従来の `storage.go` は presigned URL の署名部分をモック文字列
（`X-Goog-Signature=mock-presigned-token`）でハードコードしており、プロダクションコードに
モックが混在していた。この URL は実 GCS に対して無効であり、本番相当環境で画像アップロードが
できない不具合となっていた（Issue #23）。

## 決定

### 本番は実GCS V4署名、ローカル開発は明示的なモックに分離する

ストレージ実装を `ImageUploader` インターフェースの2実装に分離し、環境変数 `STORAGE_MODE` で
選択する（`internal/storage/storage.go` の `NewUploaderFromEnv`）。

- **本番（デフォルト）**: `GCSUploader` が `cloud.google.com/go/storage` の V4 署名
  （`SigningSchemeV4`）を用いて実際に有効な presigned URL を生成する。
  - 署名はサービスアカウント秘密鍵を用いて**ローカルで**行うため、URL発行時のネットワーク
    往復や IAM SignBlob 呼び出しは不要。
  - 認証情報は `GOOGLE_APPLICATION_CREDENTIALS`（サービスアカウントJSONキーのパス）から
    `client_email` と `private_key` を読み込む。未設定の場合は**起動時にエラーで停止**し、
    モックへサイレントフォールバックしない。
- **ローカル開発**: `STORAGE_MODE=mock` を明示設定した場合のみ `MockUploader` を使用する。
  返却する URL は `x-goog-signature=mock-development-only` を含む無効なプレースホルダで、
  GCP認証情報なしで動作確認するための開発専用実装。

### モックをプロダクションコードパスから除去する

`MockUploader` は `STORAGE_MODE=mock` が明示された場合にのみ生成される。デフォルト（本番）の
コードパスはモックへ到達しないため、モック署名が本番で返ることはない。

## 検討した代替案

### モックのまま維持し理由をADRに記載するのみ

- **却下理由**: 実アップロードが不可能なままとなり、機能的な不具合が解消されない。

### 秘密鍵を持たず `storage.Client` + IAM SignBlob API で署名する

- **却下理由**: URL発行のたびに IAM Credentials API への往復が発生しレイテンシが増加する。
  ローカル署名（秘密鍵）方式の方がシンプルで、Cloud Run でも秘密鍵をSecret Manager経由で
  マウントすれば同等に運用できる。

### V4署名を標準ライブラリで自前実装する

- **却下理由**: 署名の正準リクエスト整形は誤りが混入しやすく、公式SDKの利用が安全。

## 影響

- `go.mod` に `cloud.google.com/go/storage` が追加される（GCPプロジェクトとして正当な依存）。
- ローカル開発者は `.env` で `STORAGE_MODE=mock` を設定する（`.env.example` に既定値記載済み）。
- 本番デプロイでは `GOOGLE_APPLICATION_CREDENTIALS` の設定が必須。未設定なら起動失敗で検知できる。
- presigned URL の有効期限は 15 分（`uploadURLTTL`）。

## リスク

- サービスアカウント秘密鍵の管理が必要（中: Secret Manager等で管理し、コードやリポジトリに含めない）。
- 署名の正当性は本番のGCPバケットに対してのみ最終検証可能（低: 単体テストでは生成した
  テスト用RSA鍵でV4署名フォーマット `GOOG4-RSA-SHA256` を検証済み）。
