# =============================================================================
# Web API E2E テストテンプレート（Hurl）
#
# 使い方:
#   1. {base_url} をAPIのベースURLに置換（または --variable で渡す）
#   2. エンドポイントをプロジェクトに合わせて編集
#   3. `hurl --test test/api.hurl` で実行
#
# 参考: https://hurl.dev/docs/manual.html
# =============================================================================

# --- ヘルスチェック ---

GET {{base_url}}/health
HTTP 200
[Asserts]
jsonpath "$.status" == "ok"


# --- 認証なしでのアクセス拒否 ---

GET {{base_url}}/api/v1/resources
HTTP 401


# --- 認証フロー ---

# トークン取得（プロジェクトの認証方式に合わせて変更）
POST {{base_url}}/api/v1/auth/login
Content-Type: application/json
{
    "username": "test@example.com",
    "password": "test-password"
}
HTTP 200
[Captures]
auth_token: jsonpath "$.token"
[Asserts]
jsonpath "$.token" isString


# --- CRUD 操作 ---

# Create
POST {{base_url}}/api/v1/resources
Content-Type: application/json
Authorization: Bearer {{auth_token}}
{
    "name": "test-resource",
    "description": "E2E test resource"
}
HTTP 201
[Captures]
resource_id: jsonpath "$.id"
[Asserts]
jsonpath "$.name" == "test-resource"


# Read (single)
GET {{base_url}}/api/v1/resources/{{resource_id}}
Authorization: Bearer {{auth_token}}
HTTP 200
[Asserts]
jsonpath "$.id" == {{resource_id}}
jsonpath "$.name" == "test-resource"


# Read (list)
GET {{base_url}}/api/v1/resources
Authorization: Bearer {{auth_token}}
HTTP 200
[Asserts]
jsonpath "$.items" isCollection
jsonpath "$.total" >= 1


# Update
PUT {{base_url}}/api/v1/resources/{{resource_id}}
Content-Type: application/json
Authorization: Bearer {{auth_token}}
{
    "name": "updated-resource"
}
HTTP 200
[Asserts]
jsonpath "$.name" == "updated-resource"


# Delete
DELETE {{base_url}}/api/v1/resources/{{resource_id}}
Authorization: Bearer {{auth_token}}
HTTP 204


# 削除確認
GET {{base_url}}/api/v1/resources/{{resource_id}}
Authorization: Bearer {{auth_token}}
HTTP 404


# --- バリデーションエラー ---

POST {{base_url}}/api/v1/resources
Content-Type: application/json
Authorization: Bearer {{auth_token}}
{
    "name": ""
}
HTTP 400
[Asserts]
jsonpath "$.errors" isCollection
