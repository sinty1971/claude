# CLAUDE.md

このファイルは、このリポジトリでコードを扱う際のClaude Code (claude.ai/code) へのガイダンスを提供します。

## プロジェクト概要

フォルダー管理システムは、フォルダーの閲覧と管理のためのWebインターフェースを提供します。構成は以下の通りです：
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
- `internal/models/`: データモデル (folder.go, instant.go)

バックエンドは `http://localhost:8080/api` でREST APIを提供し、主要なエンドポイントは以下です：
- `GET /api/folders?path=<オプション-パス>` - フォルダーの内容を返す

### フロントエンド構造
- `src/App.tsx`: ルーティング機能を持つメインアプリコンポーネント
- `src/components/`: UIコンポーネント (FolderGrid, FolderModal)
- `src/services/api.ts`: バックエンドAPIクライアント
- `src/types/`: TypeScript型定義

### 主要な実装詳細

1. **デフォルトパス**: システムは `~/penguin` ディレクトリを標準で参照します
2. **CORS**: バックエンドは `AllowOrigins: "*"` で全てのオリジンを許可します
3. **ファイル種別検出**: フロントエンドはファイル拡張子に基づいて異なるアイコンを表示します：
   - フォルダー: 📁
   - PDF: 📄
   - 画像 (jpg, jpeg, png, gif): 🖼️
   - 動画 (mp4, avi, mov): 🎬
   - 音声 (mp3, wav): 🎵
   - その他: 📎

4. **APIレスポンス形式**: バックエンドは name、path、size、isDirectory などのプロパティを持つフォルダーアイテムの配列を返します