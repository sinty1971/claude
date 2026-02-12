# ディレクトリデリミタ '/' 統一実装

本プロジェクトでは、すべてのパス操作において OS に依存しないスラッシュ区切り (`/`) を使用するように統一しています。

## 実装概要

### ユーティリティ関数

[file.go](web-api/internal/core/file.go) に以下の関数を追加しました:

- **`PathJoin(...string) string`**: `filepath.Join` の代わりに使用。Windows でも常に `/` 区切りを返す
- **`PathSplit(string) []string`**: パスを `/` で分割。`\` と `/` の混在パスにも対応

### 置き換え箇所

以下のファイルで `filepath.Join` を `core.PathJoin` に置き換えました:

1. [services/member.go](web-api/internal/services/member.go)
   - 一人親方ディレクトリパス生成
   - 社員ディレクトリパス生成
   - 退職者ディレクトリパス生成

2. [services/company.go](web-api/internal/services/company.go)
   - 会社ディレクトリパス生成

3. [services/koji.go](web-api/internal/services/koji.go)
   - 工事ディレクトリパス生成

4. [services/directory.go](web-api/internal/services/directory.go)
   - ファイル一覧取得時のパス結合
   - 相対パスから絶対パス生成
   - ディレクトリコピー時のパス生成

5. [models/company.go](web-api/models/company.go)
   - 会社フォルダーパス更新

6. [models/koji.go](web-api/models/koji.go)
   - 工事フォルダーパス更新

7. [models/member.go](web-api/internal/models/member.go)
   - `strings.Split(dirPath, "\\")` を `core.PathSplit(dirPath)` に置き換え

8. [core/persist.go](web-api/internal/core/persist.go)
   - 永続化ファイルパス生成

9. [core/file.go](web-api/internal/core/file.go)
   - ホームディレクトリ展開時のパス結合

## テスト

[file_test.go](web-api/internal/core/file_test.go) に以下のテストを追加しました:

- `TestPathJoin`: 基本的なパス結合のテスト
- `TestPathJoinWindowsCompatibility`: Windows 環境での `/` 統一テスト
- `TestPathSplit`: パス分割のテスト（`/` と `\` の混在にも対応）
- `TestPathSplitWindowsPath`: Windows 形式パスの分割テスト
- `TestPathJoinAndSplitRoundTrip`: 往復変換テスト

### テスト実行結果

```bash
$ cd web-api
$ go test ./internal/core -v -run TestPath
=== RUN   TestPathJoin
=== RUN   TestPathJoin/基本的なパス結合
=== RUN   TestPathJoin/深いパス結合
=== RUN   TestPathJoin/単一要素
=== RUN   TestPathJoin/空文字列を含む
--- PASS: TestPathJoin (0.00s)
=== RUN   TestPathJoinWindowsCompatibility
--- PASS: TestPathJoinWindowsCompatibility (0.00s)
=== RUN   TestPathSplit
--- PASS: TestPathSplit (0.00s)
=== RUN   TestPathSplitWindowsPath
--- PASS: TestPathSplitWindowsPath (0.00s)
=== RUN   TestPathJoinAndSplitRoundTrip
--- PASS: TestPathJoinAndSplitRoundTrip (0.00s)
PASS
ok      web-api/internal/core   0.987s
```

## 使用例

### パス結合

```go
// 従来の方法（OS依存）
path := filepath.Join("home", "user", "file.txt")
// Windows: "home\\user\\file.txt"
// Unix:    "home/user/file.txt"

// 新しい方法（常に '/' 区切り）
path := core.PathJoin("home", "user", "file.txt")
// すべてのOS: "home/user/file.txt"
```

### パス分割

```go
// 従来の方法（区切り文字を明示的に指定）
parts := strings.Split(path, "\\")  // Windowsのみ
parts := strings.Split(path, "/")   // Unixのみ

// 新しい方法（すべての区切り文字に対応）
parts := core.PathSplit(path)
// "home\\user\\file.txt" → ["home", "user", "file.txt"]
// "home/user/file.txt"   → ["home", "user", "file.txt"]
// "home/user\\file.txt"  → ["home", "user", "file.txt"]
```

## 利点

1. **クロスプラットフォーム対応**: Windows でも Unix でも常に `/` 区切りで統一
2. **フロントエンドとの整合性**: JavaScript は常に `/` を使用するため、API レスポンスとの互換性が向上
3. **テスト容易性**: パスの形式が統一されているため、テストが簡単
4. **可読性**: パス区切りが常に `/` であることが保証されるため、コードが読みやすい

## 注意事項

- **ファイルシステム操作**: `os.Open`、`os.ReadDir` などの標準ライブラリ関数では、Windows でも `/` 区切りが正しく処理されます
- **絶対パス判定**: `filepath.IsAbs` は引き続き使用可能（内部で OS 判定を行うため）
- **ディレクトリ取得**: `filepath.Dir` や `filepath.Base` は引き続き使用可能（区切り文字に依存しない実装）
- **既存コード**: 新規コードでは `core.PathJoin` / `core.PathSplit` を使用してください

## 今後の展開

- [ ] 残りの `filepath.Join` 使用箇所の段階的な置き換え
- [ ] フロントエンド側でのパス検証強化
- [ ] CI/CD で Windows と Unix 両方のテストを実行
