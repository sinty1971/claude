# Migrate Manifest Scripts

このディレクトリには、`@company.yaml` から `@manifest.yaml` へキーをマッピングして移行するスクリプトが複数の言語で提供されています。

## スクリプト一覧

- `migrate_manifest.py` - Python版
- `migrate_manifest.go` - Go版
- `migrate_manifest.ts` - TypeScript版
- `migrate_manifest.ps1` - PowerShell版
- `migrate_manifest.sh` - Bash版
- `migrate_manifest.rb` - Ruby版
- `migrate_manifest.pl` - Perl版
- `migrate_manifest.php` - PHP版

## 機能

全てのスクリプトは以下の共通機能を持っています:

- 指定ディレクトリ配下の全ての `@company.yaml` を再帰的に検索
- 各 `@company.yaml` から以下のキーを読み取り:
  - `persist_long_name` → `mf_long_name`
  - `persist_postal_code` → `mf_postal_code`
  - `persist_address` → `mf_address`
  - `persist_tel` → `mf_tel`
  - `persist_fax` → `mf_fax`
  - `persist_email` → `mf_email`
  - `persist_website` → `mf_website`
- 同じディレクトリに `@manifest.yaml` として保存

## 使用方法

### Python版

**必要なパッケージ:**

```bash
pip install pyyaml
```

**実行:**

```bash
python migrate_manifest.py "C:\SyncFolder\SynologyDrive\豊田築炉\1 会社"
```

### Go版

**必要なパッケージ:**

```bash
go get gopkg.in/yaml.v3
```

**実行:**

```bash
go run migrate_manifest.go "C:\SyncFolder\SynologyDrive\豊田築炉\1 会社"
```

または、ビルドして実行:

```bash
go build -o migrate_manifest.exe migrate_manifest.go
.\migrate_manifest.exe "C:\SyncFolder\SynologyDrive\豊田築炉\1 会社"
```

### TypeScript版

**必要なパッケージ:**

```bash
npm install yaml glob
npm install -D @types/node tsx
```

**実行:**

```bash
# TypeScriptを直接実行
npx tsx migrate_manifest.ts "C:\SyncFolder\SynologyDrive\豊田築炉\1 会社"

# または、JavaScriptにコンパイルして実行
npx tsc migrate_manifest.ts
node migrate_manifest.js "C:\SyncFolder\SynologyDrive\豊田築炉\1 会社"
```

### PowerShell版

**必要なモジュール:**

```powershell
Install-Module -Name powershell-yaml -Scope CurrentUser
```

**実行:**

```powershell
.\migrate_manifest.ps1 "C:\SyncFolder\SynologyDrive\豊田築炉\1 会社"
```

### Bash版 (Linux/macOS)

**必要なツール:**

- [yq](https://github.com/mikefarah/yq) - YAML プロセッサ

```bash
# macOS
brew install yq

# Linux
sudo snap install yq
```

**実行:**

```bash
chmod +x migrate_manifest.sh
./migrate_manifest.sh "/path/to/1 会社"
```

### Ruby版

**必要なライブラリ:**

Ruby標準ライブラリのみ（追加インストール不要）

**実行:**

```bash
| Ruby       | Rails開発者         | Pythonに近い読みやすさ、標準YAML対応 |
| Perl       | Unix/Linuxシステム管理者 | テキスト処理に最適、正規表現が強力 |
| PHP        | Web開発者           | Web環境で標準、ファイル処理が得意 |
ruby migrate_manifest.rb "C:\SyncFolder\SynologyDrive\豊田築炉\1 会社"
```

### Perl版

**必要なモジュール:**

```bash
cpan YAML::XS
```

**実行:**

```bash
perl migrate_manifest.pl "C:\SyncFolder\SynologyDrive\豊田築炉\1 会社"
```

### PHP版

**必要な拡張:**

```bash
# YAML拡張のインストール
pecl install yaml

# php.iniに追加
extension=yaml.so  # Linux/macOS
extension=yaml.dll # Windows
```

**実行:**

```bash
php migrate_manifest.php "C:\SyncFolder\SynologyDrive\豊田築炉\1 会社"
```

## 選択ガイド

| 言語       | 推奨環境              | メリット                     |
|------------|---------------------|------------------------------|
| Python     | 汎用                | シンプル、可読性が高い         |
| Go         | バックエンド開発者     | 高速、スタンドアロンバイナリ   |
| TypeScript | フロントエンド開発者   | 型安全、Node.js環境との統合   |
| PowerShell | Windows管理者        | Windows標準、スクリプト統合   |
| Bash       | Linux/macOS         | Unix系OS標準、自動化に最適    |

## トラブルシューティング

### Python: ModuleNotFoundError: No module named 'yaml'

```bash
pip install pyyaml
```

### Go: package gopkg.in/yaml.v3 is not in GOROOT

```bash
go mod init scripts
go get gopkg.in/yaml.v3
```

### TypeScript: Cannot find module 'yaml'

```bash
npm install yaml glob
```

### PowerShell: powershell-yaml モジュールがインストールされていません

```powershell
Install-Module -Name powershell-yaml -Scope CurrentUser
```

### Bash: yq: command not found

```bash
# macOS
brew install yq


### Perl: Can't locate YAML/XS.pm

```bash
cpan YAML::XS
```

### PHP: Call to undefined function yaml_parse()

```bash
# YAML拡張をインストール
pecl install yaml

# php.iniに追加
extension=yaml.so  # Linux/macOS
extension=yaml.dll # Windows
```
# Linux (snap)
sudo snap install yq
```
