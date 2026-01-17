# Tailwind CSS + shadcn-svelte 導入完了

## 導入内容

### インストールしたパッケージ

```json
{
  "shadcn-svelte": "^1.1.0",
  "clsx": "^2.1.1",
  "tailwind-merge": "^3.4.0",
  "tailwind-variants": "^3.2.2",
  "tailwindcss": "^4.1.18",
  "@tailwindcss/vite": "^4.1.18",
  "@tailwindcss/forms": "^0.5.11",
  "@tailwindcss/typography": "^0.5.19"
}
```

### 追加したコンポーネント

- **Button**: ボタンコンポーネント（primary, secondary, ghost などのバリアント）
- **Card**: カードコンポーネント（Root, Header, Title, Description, Content, Footer）
- **Input**: 入力フィールドコンポーネント
- **Label**: ラベルコンポーネント
- **Table**: テーブルコンポーネント（Root, Header, Body, Head, Row, Cell）
- **Badge**: バッジコンポーネント（default, secondary, destructive など）

## 設定ファイル

### components.json

shadcn-svelte の設定ファイル。コンポーネントのインストール先やエイリアスを定義。

```json
{
  "$schema": "https://shadcn-svelte.com/schema.json",
  "tailwind": {
    "css": "src\\routes\\layout.css",
    "baseColor": "slate"
  },
  "aliases": {
    "components": "$lib/components",
    "utils": "$lib/utils",
    "ui": "$lib/components/ui",
    "hooks": "$lib/hooks",
    "lib": "$lib"
  }
}
```

### src/routes/layout.css

Tailwind CSS v4 のカスタムテーマとアプリケーション固有のスタイルを定義。

- **カラーパレット**: slate ベースの shadcn-svelte カラースキーム
- **レスポンシブ対応**: ダークモード対応の CSS 変数定義
- **カスタムコンポーネント**: `.app-shell`, `.app-header`, `.app-nav` などのレイアウト用クラス

### src/lib/utils.ts

shadcn-svelte が使用するユーティリティ関数。

- **cn()**: clsx と tailwind-merge を組み合わせたクラス名マージ関数

## 移行したページ

### 1. トップページ (src/routes/+page.svelte)

- Hero セクション: グリッドレイアウト + Card コンポーネント
- クイックリンク: Card コンポーネントでホバー効果
- Button コンポーネントで CTA ボタン

### 2. 会社一覧 (src/routes/companies/+page.svelte)

- Table コンポーネントで一覧表示
- Badge コンポーネントでカテゴリ表示
- Card.Root でテーブルをラップ
- レスポンシブデザイン（hidden md:table-cell, hidden lg:table-cell）

### 3. 工事一覧 (src/routes/kojies/+page.svelte)

- Table コンポーネントで一覧表示
- Badge コンポーネントで状態表示（進行中=red, 完了=green, 予定=blue, 不明=gray）
- カスタム背景色クラス（bg-green-600, bg-blue-600）

## カスタマイズポイント

### 工事状態バッジのカラー

`src/routes/kojies/+page.svelte` で状態に応じた色分けを実装:

```svelte
{#if generateKojiStatus(koji) === "進行中"}
  <Badge variant="destructive" class="font-bold">進行中</Badge>
{:else if generateKojiStatus(koji) === "完了"}
  <Badge class="bg-green-600 hover:bg-green-700">完了</Badge>
{:else if generateKojiStatus(koji) === "予定"}
  <Badge class="bg-blue-600 hover:bg-blue-700">予定</Badge>
{:else}
  <Badge variant="secondary">不明</Badge>
{/if}
```

### アプリケーション固有のスタイル

`src/routes/layout.css` の `@layer components` に追加:

- `.app-shell`: 最小高さ 100vh のフレックスレイアウト
- `.app-header`: ダークブルー背景のヘッダー
- `.app-nav`: ライトグレー背景のナビゲーション
- `.nav-toggle`: ナビゲーション切り替えボタン

## 既存コードとの互換性

- **カスタム CSS は削除済み**: すべてのページで `<style>` ブロックを削除し、Tailwind クラスに置き換え
- **既存ユーティリティ関数は保持**: `koji-utils.ts`, `form-utils.ts`, `grpc-client.ts` はそのまま使用
- **型定義は変更なし**: gRPC 生成コードや PageData 型は影響なし

## 今後の拡張

### 追加推奨コンポーネント

プロジェクトの成長に応じて以下のコンポーネントを追加できます:

```bash
cd my-svelte-app
bunx shadcn-svelte@latest add dialog
bunx shadcn-svelte@latest add dropdown-menu
bunx shadcn-svelte@latest add popover
bunx shadcn-svelte@latest add select
bunx shadcn-svelte@latest add toast
bunx shadcn-svelte@latest add alert
```

### カスタムテーマの変更

`src/routes/layout.css` の `:root` セクションで色を変更できます:

```css
:root {
  --primary: oklch(0.208 0.042 265.755); /* プライマリカラー */
  --destructive: oklch(0.577 0.245 27.325); /* エラー色 */
  --radius: 0.625rem; /* ボーダー半径 */
}
```

## 開発コマンド

```bash
# 開発サーバー起動
cd my-svelte-app
bun run dev

# 型チェック
bun run check

# ビルド
bun run build

# コンポーネント追加
bunx shadcn-svelte@latest add [component-name]
```

## 参考リンク

- [shadcn-svelte 公式ドキュメント](https://www.shadcn-svelte.com/)
- [Tailwind CSS v4 ドキュメント](https://tailwindcss.com/docs)
- [Svelte 5 ドキュメント](https://svelte.dev/docs/svelte/overview)
