# Penguin Backend API

Go + Fiber製のフォルダー管理API

## 機能
- `~/penguin/2-工事`内のフォルダー一覧取得
- CORS対応でフロントエンドとの連携

## エンドポイント

### GET /api/folders
フォルダー一覧を取得

**クエリパラメータ:**
- `path` (optional): 対象パス (デフォルト: `~/penguin/2-工事`)

**レスポンス例:**
```json
{
  "folders": [
    {
      "name": "プロジェクトA",
      "path": "/home/user/penguin/2-工事/プロジェクトA",
      "is_directory": true,
      "size": 4096,
      "modified_time": "2024-01-01T12:00:00Z"
    }
  ],
  "count": 1,
  "path": "/home/user/penguin/2-工事"
}
```

## 実行方法

```bash
cd backend
go mod tidy
go run cmd/main.go
```

サーバーは http://localhost:8080 で起動します。

## API使用例

### 基本的な使用方法

1. **デフォルトパスのフォルダー取得**:
   ```bash
   curl "http://localhost:8080/api/folders"
   ```

2. **カスタムパスの指定**:
   ```bash
   curl "http://localhost:8080/api/folders?path=~/Documents"
   ```

3. **ブラウザでアクセス**:
   - `http://localhost:8080/api/folders` をブラウザで開く

### レスポンスの説明
- `folders`: フォルダー/ファイルの配列
- `count`: 取得された項目数
- `path`: 実際に読み取られたパス
- 各フォルダーオブジェクト:
  - `name`: ファイル/フォルダー名
  - `path`: フルパス
  - `is_directory`: ディレクトリかどうかのフラグ
  - `size`: ファイルサイズ（バイト）
  - `modified_time`: 最終更新時刻

## ディレクトリ構造

```
backend/
├── cmd/
│   └── main.go              # エントリーポイント
├── internal/
│   ├── handlers/            # HTTPハンドラー
│   ├── models/              # データモデル
│   └── services/            # ビジネスロジック
├── pkg/                     # 外部公開パッケージ
└── go.mod                   # Go modules
```