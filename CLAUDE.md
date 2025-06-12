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
- `internal/handlers/`: HTTPリクエストハンドラー (folder_handler.go)
- `internal/services/`: ビジネスロジック (folder_service.go)
- `internal/models/`: データモデル (folder.go, kouji_project.go, id.go)

バックエンドは `http://localhost:8080/api` でREST APIを提供し、主要なエンドポイントは以下です：
- `GET /api/folders?path=<オプション-パス>` - フォルダーの内容を返す
- `GET /api/kouji-projects?path=<オプション-パス>` - 工事プロジェクトの一覧を返す
- `POST /api/kouji-projects/save` - 工事プロジェクト情報をYAMLファイルに保存

### フロントエンド構造
- `src/App.tsx`: ルーティング機能を持つメインアプリコンポーネント
- `src/components/`: UIコンポーネント (FolderGrid, FolderModal, KoujiProjectGrid, KoujiProjectPage)
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

4. **工事プロジェクト管理**: 
   - フォルダー命名規則: `YYYY-MMDD 会社名 現場名` (例: `2025-0618 豊田築炉 名和工場`)
   - プロジェクトID: BLAKE2bハッシュを使用した5文字の一意ID生成
   - YAML永続化: `.inside.yaml` ファイルでプロジェクト情報を保存
   - タイムゾーン: JST（ローカルタイム）で日時を保持

5. **APIレスポンス形式**: 
   - 一般フォルダー: name、path、size、isDirectory などのプロパティを持つ配列
   - 工事プロジェクト: project_id、company_name、location_name などの拡張プロパティを含む配列