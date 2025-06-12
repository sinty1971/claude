# Backend Architecture

このドキュメントは、Penguin Backend の internal パッケージ内の handlers、models、services の関係性を示しています。

## アーキテクチャ図

```mermaid
graph TB
    subgraph "HTTP Layer"
        Client[HTTP Client/Frontend]
    end
    
    Client -->|HTTP Request| FH

    subgraph "Handler Layer (internal/handlers)"
        FH[FolderHandler]
        FH_GetFolders[GetFolders]
        FH_GetKouji[GetKoujiProjects]
        FH_SaveKouji[SaveKoujiProjectsToYAML]
        FH_UpdateDates[UpdateKoujiProjectDates]
        FH_Cleanup[CleanupInvalidTimeData]
        
        FH --> FH_GetFolders
        FH --> FH_GetKouji
        FH --> FH_SaveKouji
        FH --> FH_UpdateDates
        FH --> FH_Cleanup
    end

    FH_GetFolders -->|calls| FS_GetFolders
    FH_GetKouji -->|calls| FS_GetFolders
    FH_GetKouji -->|calls| FS_LoadYAML
    FH_SaveKouji -->|calls| FS_LoadYAML
    FH_SaveKouji -->|calls| FS_SaveYAML
    FH_UpdateDates -->|calls| FS_LoadYAML
    FH_UpdateDates -->|calls| FS_SaveYAML
    FH_Cleanup -->|calls| FS_Cleanup

    subgraph "Service Layer (internal/services)"
        FS[FolderService]
        FS_GetFolders[GetFolders]
        FS_LoadYAML[LoadKoujiProjectsFromYAML]
        FS_SaveYAML[SaveKoujiProjectsToYAML]
        FS_Cleanup[CleanupInvalidTimeData]
        
        FS --> FS_GetFolders
        FS --> FS_LoadYAML
        FS --> FS_SaveYAML
        FS --> FS_Cleanup
    end

    FS_GetFolders -->|reads| FileSystem
    FS_LoadYAML -->|reads| YAMLFile
    FS_SaveYAML -->|writes| YAMLFile
    FS_Cleanup -->|read/write| YAMLFile

    FS_GetFolders -.uses.-> Folder
    FS_GetFolders -.returns.-> FolderListResponse
    FS_LoadYAML -.returns.-> KoujiProject
    FS_SaveYAML -.uses.-> KoujiProject
    FH_GetKouji -.returns.-> KoujiProjectListResponse

    subgraph "Model Layer (internal/models)"
        Folder[Folder struct]
        KoujiProject[KoujiProject struct]
        FolderListResponse[FolderListResponse]
        KoujiProjectListResponse[KoujiProjectListResponse]
        
        KoujiProject -.embed.-> Folder
    end

    subgraph "Data Storage"
        FileSystem[File System<br/>~/penguin/]
        YAMLFile[.inside.yaml]
    end

    style Client fill:#f9f,stroke:#333,stroke-width:2px
    style FileSystem fill:#ff9,stroke:#333,stroke-width:2px
    style YAMLFile fill:#ff9,stroke:#333,stroke-width:2px
```

## レイヤーごとの責務

### Handler層 (internal/handlers/folder_handler.go)
- HTTPリクエスト/レスポンスの処理
- バリデーション
- Service層の呼び出し
- ビジネスロジック（プロジェクトのマージ、ステータス判定など）

### Service層 (internal/services/folder_service.go)
- ファイルシステムやYAMLファイルへのアクセス
- データの読み書き処理
- パスの展開（`~/`の処理）
- 時刻データの検証とフォーマット

### Model層 (internal/models/)
- データ構造の定義
- `Folder`: 基本的なファイル/フォルダ情報（CreatedDateを含む）
- `KoujiProject`: Folderを埋め込み、工事プロジェクト固有の情報を追加
- レスポンス用の構造体定義

## データフロー例

1. **フォルダ一覧取得**: 
   - Client → FolderHandler.GetFolders → FolderService.GetFolders → FileSystem

2. **工事プロジェクト取得**: 
   - Client → FolderHandler.GetKoujiProjects → FolderService.GetFolders + LoadKoujiProjectsFromYAML → FileSystem + YAMLFile

3. **工事プロジェクト保存**: 
   - Client → FolderHandler.SaveKoujiProjectsToYAML → FolderService.LoadKoujiProjectsFromYAML + SaveKoujiProjectsToYAML → YAMLFile

## 主要な処理の流れ

### GetKoujiProjects の処理
1. ファイルシステムから工事フォルダ一覧を取得
2. YAMLファイルから既存の工事プロジェクト情報を読み込み
3. project_id を基準に両者をマージ
4. 開始日の降順でソート
5. レスポンスとして返却

### SaveKoujiProjectsToYAML の処理
1. 既存のYAMLファイルを読み込み
2. ファイルシステムから最新の工事フォルダ情報を取得
3. project_id を基準にマージ（更新日時で新旧判定）
4. 異常な時刻データを持つプロジェクトを除外
5. YAMLファイルに保存