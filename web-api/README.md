# Pathist gRPC Server

Connect ベースの gRPC サーバーです。ディレクトリ・会社・工事ドメインの情報を公開し、フロントエンドからは Connect-Web 経由で利用します。

## 主な機能

### DirectoryService

ディレクトリ情報の管理を提供します。

- `GetPathList` : 指定パスのディレクトリ一覧を取得

### CompanyService

会社データの管理を提供します。

- `GetCompanies` : 全会社データの取得（キャッシュベース）
- `GetCompany` : 特定会社の詳細情報を取得
- `UpdateCompany` : 会社情報の更新（フォルダー名変更、マニフェスト更新）
- `GetCompanyCategories` : 会社カテゴリー一覧の取得

**マニフェストフィールド（`mf_`プレフィックス）:**
- 会社正式名、郵便番号、住所、電話、FAX、メール、Webサイト

### KojiService

工事データの管理を提供します。

- `GetKojies` : 全工事データの取得（キャッシュベース）
- `GetKoji` : 特定工事の詳細情報を取得
- `UpdateKoji` : 工事情報の更新（フォルダー名変更、マニフェスト更新）

**マニフェストフィールド:**
- `mf_end` : 工事終了日（未設定時は開始日と同じ値に自動補完）

API の定義は `proto/grpc/v1/toyotachikuro.proto` にまとまっており、`just generate-grpc` コマンドでサーバー側とフロントエンド側のスタブを再生成できます。

## マニフェストシステム

各ドメインモデル（Company、Koji）は、フォルダー名から解析できる情報に加え、追加の永続化データを `@manifest.yaml` ファイルで管理します。

### 特徴

- **`mf_`プレフィックス**: protobuf メッセージ内で `mf_` で始まるフィールドがマニフェストファイルに保存されます
- **日本時間対応**: Timestamp 型のフィールドは日本時間（JST、+09:00）で YAML に保存されます
- **自動補完**: Koji の `mf_end` が未設定の場合、`start` フィールドの値が自動的にコピーされます
- **埋め込みパターン**: `ManifestProvider` を埋め込むことで、各モデルが永続化機能を持ちます

### 保存場所

- 会社: `{会社フォルダ}/@manifest.yaml`
- 工事: `{工事フォルダ}/@manifest.yaml`

## 実行方法

```bash
# リポジトリルートから
just api
```

デフォルトでは HTTP/2 over h2c で `http://localhost:9090` を待ち受けます。TLS を有効化する場合は:

```bash
just api-tls
```

証明書が未作成の場合は自動的に自己署名証明書が生成されます。

## 動作確認

CLI テストクライアントで動作確認ができます。

```bash
# 会社サービスのテスト
just test-company                          # 全会社一覧を表示
just test-company --company-id <ID>        # 特定会社の詳細表示
just test-company --json                   # JSON 形式で出力

# 工事サービスのテスト
just test-koji                             # 全工事一覧を表示
just test-koji --koji-id <ID>              # 特定工事の詳細表示
just test-koji --json                      # JSON 形式で出力

# ファイルサービスのテスト
just test-file                             # ファイル一覧を表示
just test-file --json                      # JSON 形式で出力
```

## ディレクトリ構成

```text
web-api/
├── cmd/
│   ├── server/        # サーバーエントリポイント
│   ├── protodoc/      # プロト定義からドキュメント生成
│   └── test/          # テスト用 CLI ツール
│       ├── company/   # 会社 API テストクライアント
│       ├── file/      # ファイル API テストクライアント
│       └── koji/      # 工事 API テストクライアント
├── gen/               # プロト生成コード（buf generate で更新）
│   └── grpc/v1/       # Go/Connect-Go スタブ
├── internal/
│   ├── core/          # 共通機能
│   │   ├── config.go      # 設定管理
│   │   ├── file.go        # ファイル操作
│   │   ├── id.go          # ID 生成
│   │   ├── manifest.go    # マニフェスト永続化
│   │   └── watcher.go     # ファイルシステム監視
│   ├── models/        # ドメインモデル
│   │   ├── company.go     # 会社モデル
│   │   ├── koji.go        # 工事モデル
│   │   └── timestamp.go   # タイムスタンプ処理
│   └── services/      # ビジネスロジック（gRPC ハンドラー）
│       ├── container.go   # サービスコンテナ
│       ├── company.go     # 会社サービス
│       ├── koji.go        # 工事サービス
│       └── directory.go   # ディレクトリサービス
├── tools/             # go generate スクリプト
└── go.mod
```

## 主要コマンド

### サーバー起動

- `just api` : HTTP/2 over h2c で API サーバーを起動
- `just api-tls` : TLS 有効で API サーバーを起動

### コード生成

- `just generate-grpc` : Go/Connect-Go スタブを再生成
- `just generate-types` : TypeScript 型定義を生成（フロントエンド用）

### テスト

- `just test-company [OPTIONS]` : 会社サービスのテスト
- `just test-koji [OPTIONS]` : 工事サービスのテスト
- `just test-file [OPTIONS]` : ファイルサービスのテスト

詳細はリポジトリ直下の `justfile` を参照してください。

## 技術スタック

- **gRPC/Connect**: [connectrpc.com/connect](https://connectrpc.com/)
- **Protobuf**: Edition 2023、Opaque API モード
- **Buf**: スキーマ管理とコード生成
- **ファイルシステム監視**: fsnotify ベース
- **YAML 永続化**: gopkg.in/yaml.v3

## 設定

サーバー起動時のデフォルト設定（`cmd/server/main.go`）：

```go
core.Config.DirectoryBaseDirPath = "{ROOT}"
core.Config.CompanyBaseDirPath = "{ROOT}/1 会社"
core.Config.KojiBaseDirPath = "{ROOT}/2 工事"
core.Config.MaximumWorkers = 16
core.Config.MinimumWorkers = 2
```

`{ROOT}` は環境変数またはホームディレクトリ配下の `penguin` フォルダに展開されます。
