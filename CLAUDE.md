# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) へのガイダンスを提供します。
レポート等は原則として日本語で行います。

## プロジェクト概要

フォルダー管理システムは、フォルダーの閲覧と管理、および工事プロジェクトの管理のためのWebインターフェースを提供します。構成は以下の通りです：
- **バックエンド**: Go (1.21) と Fiber v2 フレームワーク
- **フロントエンド**: React (19.1.0) と TypeScript、Vite

## 開発コマンド

### フロントエンド開発
```bash
cd frontend
npm install          # 依存関係をインストール
npm run dev          # 開発サーバーを起動 (http://localhost:5173)
npm run build        # 本番用にビルド
npm run lint         # ESLintを実行
npm run preview      # 本番ビルドをプレビュー
```

### バックエンド開発
```bash
cd backend
go mod tidy          # 依存関係をインストール/更新
go run cmd/main.go   # サーバーを起動 (http://localhost:8080)
```

## アーキテクチャ

### バックエンド構造
- `cmd/main.go`: エントリーポイント、CORSを持つFiberサーバーをセットアップ
- `internal/handlers/`: HTTPリクエストハンドラー (filesystem.go, kouji.go, time.go)
- `internal/services/`: ビジネスロジック (filesystem.go, kouji.go)
- `internal/models/`: データモデル (fileentry.go, kouji.go, id.go, time.go, timestamp.go)

バックエンドは `http://localhost:8080/api` でREST APIを提供し、主要なエンドポイントは以下です：
- `GET /api/file-entries?path=<オプション-パス>` - フォルダーの内容を返す
- `GET /api/kouji-entries?path=<オプション-パス>` - 工事プロジェクトの一覧を返す
- `POST /api/kouji-entries/save` - 工事プロジェクト情報をYAMLファイルに保存

### フロントエンド構造
- `src/App.tsx`: ルーティング機能を持つメインアプリコンポーネント
- `src/components/`: UIコンポーネント (FileEntryGrid, FileEntryModal, KoujiEntryGrid, KoujiEntryPage)
- `src/api/client.ts`: バックエンドAPIクライアント
- `src/types/`: TypeScript型定義 (kouji.ts)

### 主要な実装詳細

1. **デフォルトパス**: 
   - 一般フォルダー: `~/penguin` ディレクトリを標準で参照
   - 工事プロジェクト: `~/penguin/豊田築炉/2-工事` ディレクトリを標準で参照

2. **CORS**: バックエンドは `AllowOrigins: "*"` で全てのオリジンを許可します

3. **ファイル種別検出**: フロントエンドはファイル拡張子に基づいて異なるアイコンを表示します：
   - フォルダー: 📁
   - PDF: 📄
   - 画像 (jpg, jpeg, png, gif): 🖼️
   - 動画 (mp4, avi, mov): 🎬
   - 音声 (mp3, wav): 🎵
   - その他: 📎

4. **フォルダーリストの管理
   - Id: UnixシステムのInoのuint64を使用
   - 日時データ: 更新日
       - バックエンドのデータベース保存はフォーマット形式RFC3339Nanoの文字列保存

4. **工事一覧の管理**: 
   - フォルダー命名規則: `YYYY-MMDD 会社名 現場名` (例: `2025-0618 豊田築炉 名和工場`)
   - 工事ID: フォルダーのID+元請け会社名+現場名から一意ID生成
   - 工事データー永続化: `/home/<user>/penguin/豊田築炉/2-工事/.inside.yaml` ファイルで工事情報を保存
   - 日時データ: 工事開始日、工事完了日、フォルダー更新日
       - バックエンドのデータベース保存はフォーマット形式RFC3339Nanoの文字列保存
   - タイムゾーン: JST（ローカルタイム）で日時を保持

5. **APIレスポンス形式**: 
   - 一般フォルダー: name、path、size、isDirectory などのプロパティを持つ配列
   - 工事プロジェクト: id、company_name、location_name などの拡張プロパティを含む配列

6. **バックエンド内のデータソース定義**:
   - **FileSystem (fs)**: ファイルシステムから取得した情報
   - **Database (db)**: データベース（`.inside.yaml`ファイル）から取得した情報
     - 工事プロジェクトデータベースの保存場所: `~/penguin/豊田築炉/2-工事/.inside.yaml`
   - **Merge (mg)**: FileSystemとDatabaseのデータマージ処理
     - この処理は工事プロジェクト管理において重要な役割を持つ
     - 原則このデータをフロントエンドに提供する