# =============================================================================
# urushi-chronicle API E2E テスト（Hurl）
#
# 使い方:
#   1. サーバー起動: cd backend && go run ./cmd/urushi-chronicle/
#   2. テスト実行: hurl --test --variable base_url=http://localhost:8080 test/api/api.hurl
#
# 参考: https://hurl.dev/docs/manual.html
# =============================================================================

# --- ヘルスチェック ---

GET {{base_url}}/health
HTTP 200
[Asserts]
jsonpath "$.status" == "ok"


# --- 工程テンプレート一覧取得 ---

GET {{base_url}}/api/v1/templates
HTTP 200
[Asserts]
jsonpath "$" count == 3
jsonpath "$[0].name" isString
jsonpath "$[0].steps" isCollection


# --- 工程テンプレート個別取得 ---

GET {{base_url}}/api/v1/templates/standard
HTTP 200
[Asserts]
jsonpath "$.name" == "standard"
jsonpath "$.steps" count == 5


GET {{base_url}}/api/v1/templates/makie
HTTP 200
[Asserts]
jsonpath "$.name" == "makie"
jsonpath "$.steps" count == 7


# --- 制作工程 CRUD ---
# サンプル作品ID: 00000000-0000-0000-0000-000000000001

# 空の工程一覧取得
GET {{base_url}}/api/v1/works/00000000-0000-0000-0000-000000000001/steps
HTTP 200
[Asserts]
jsonpath "$.total" == 0
jsonpath "$.items" isCollection


# Create — 下塗り工程
POST {{base_url}}/api/v1/works/00000000-0000-0000-0000-000000000001/steps
Content-Type: application/json
{
    "name": "下塗り一回目",
    "step_order": 1,
    "category": "shitanuri",
    "notes": "生漆を使用"
}
HTTP 201
[Captures]
step_id: jsonpath "$.id"
[Asserts]
jsonpath "$.name" == "下塗り一回目"
jsonpath "$.step_order" == 1
jsonpath "$.category" == "shitanuri"
jsonpath "$.notes" == "生漆を使用"
jsonpath "$.work_id" == "00000000-0000-0000-0000-000000000001"


# Read (single)
GET {{base_url}}/api/v1/works/00000000-0000-0000-0000-000000000001/steps/{{step_id}}
HTTP 200
[Asserts]
jsonpath "$.id" == {{step_id}}
jsonpath "$.name" == "下塗り一回目"


# Read (list)
GET {{base_url}}/api/v1/works/00000000-0000-0000-0000-000000000001/steps
HTTP 200
[Asserts]
jsonpath "$.items" isCollection
jsonpath "$.total" >= 1


# Update
PUT {{base_url}}/api/v1/works/00000000-0000-0000-0000-000000000001/steps/{{step_id}}
Content-Type: application/json
{
    "name": "下塗り一回目（修正）",
    "notes": "黒漆に変更"
}
HTTP 200
[Asserts]
jsonpath "$.name" == "下塗り一回目（修正）"
jsonpath "$.notes" == "黒漆に変更"


# 画像アップロードURL取得
POST {{base_url}}/api/v1/works/00000000-0000-0000-0000-000000000001/steps/{{step_id}}/upload
Content-Type: application/json
{
    "content_type": "image/jpeg"
}
HTTP 200
[Asserts]
jsonpath "$.upload_url" isString
jsonpath "$.file_path" isString
jsonpath "$.content_type" == "image/jpeg"


# Delete
DELETE {{base_url}}/api/v1/works/00000000-0000-0000-0000-000000000001/steps/{{step_id}}
HTTP 204


# 削除確認
GET {{base_url}}/api/v1/works/00000000-0000-0000-0000-000000000001/steps/{{step_id}}
HTTP 404


# --- バリデーションエラー ---

# 空の名前
POST {{base_url}}/api/v1/works/00000000-0000-0000-0000-000000000001/steps
Content-Type: application/json
{
    "name": "",
    "step_order": 1,
    "category": "shitanuri"
}
HTTP 400
[Asserts]
jsonpath "$.errors" isCollection


# 存在しない作品
GET {{base_url}}/api/v1/works/99999999-9999-9999-9999-999999999999/steps
HTTP 404
