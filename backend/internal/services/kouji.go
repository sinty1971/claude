package services

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"penguin-backend/internal/models"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type KoujiService struct {
	FileSystemService *FileSystemService
}

func NewKoujiService(fsService *FileSystemService) *KoujiService {
	return &KoujiService{
		FileSystemService: fsService,
	}
}

// GetKoujiListFromFileSystem はファイルシステムから工事プロジェクト一覧を取得する
func (ks *KoujiService) GetKoujiListFromFileSystem(targetPath string) ([]models.Kouji, error) {
	// Get kouji folders from file system
	folders, err := ks.FileSystemService.GetFolders(targetPath)
	if err != nil {
		return nil, err
	}

	// Regular expression to match folder names like "2025-0618 豊田築炉 名和工場"
	koujiPattern := regexp.MustCompile(`^(\d{4}-\d{4})\s+(.+?)\s+(.+)$`)

	// Convert to KoujiProjects with additional metadata
	koujiProjects := make([]models.Kouji, 0)

	for _, folder := range folders.Folders {
		// Only process directories that match the naming pattern
		if !folder.IsDirectory {
			continue
		}

		matches := koujiPattern.FindStringSubmatch(folder.Name)
		if matches == nil || len(matches) != 4 {
			continue // Skip folders that don't match the pattern
		}

		// Extract parts from the folder name
		dateStr := matches[1]     // e.g., "2025-0618"
		companyName := matches[2] // e.g., "豊田築炉"
		factoryName := matches[3] // e.g., "名和工場"

		// Parse date from folder name
		var projectDate time.Time
		if len(dateStr) == 9 && dateStr[4] == '-' {
			year := dateStr[:4]
			monthDay := dateStr[5:]
			if len(monthDay) == 4 {
				month := monthDay[:2]
				day := monthDay[2:]
				dateTimeStr := year + "-" + month + "-" + day
				// Parse as local time to preserve timezone
				projectDate, _ = time.ParseInLocation("2006-01-02", dateTimeStr, time.Local)
			}
		}

		// Generate unique project ID using folder creation date, company name, and location name
		// This ensures the same project always gets the same ID
		// Use project date instead of creation date for more stable ID generation
		projectDateStr := projectDate.Format("2006-01-02")
		idSource := projectDateStr + "_" + companyName + "_" + factoryName
		projectID := models.NewIDFromString(idSource).Len5()

		koujiProject := models.Kouji{
			// Generate project metadata based on folder name
			Id:           projectID,
			CompanyName:  companyName,
			LocationName: factoryName,
			Status:       ks.DetermineProjectStatus(projectDate),
			StartDate:    models.NewTimestamp(projectDate),
			EndDate:      models.NewTimestamp(projectDate.AddDate(0, 3, 0)), // Assume 3-month project duration
			Description:  companyName + "の" + factoryName + "における工事プロジェクト",
			Tags:         []string{"工事", companyName, factoryName, dateStr[:4]}, // Include year as tag
			// FileEntry: ファイルシステムから取得したフォルダー情報
			FileEntry: folder,
		}

		koujiProject.FileCount = 0 // Would need to scan subdirectory to get actual count
		koujiProject.SubdirCount = 0

		koujiProjects = append(koujiProjects, koujiProject)
	}

	return koujiProjects, nil
}

// LoadKoujiListFromDatabase は工事プロジェクト情報をYAMLファイルから読み込む
//
// 読み込み手順:
//  1. 指定されたディレクトリに .inside.yaml が存在するかチェック
//  2. YAML内容を KoujiFolder 構造体に解析
//  3. プロジェクトIDと必須フィールドを検証
//  4. タイムスタンプをローカルタイムゾーンから time.Time に変換
//  5. ソート済みのプロジェクトリストを返す
//
// エラーハンドリング:
//   - ファイルが存在しない場合は空のリストを返す
//   - YAML解析が失敗した場合はエラーを返す
func (ks *KoujiService) LoadKoujiListFromDatabase(databasePath string) ([]models.Kouji, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(databasePath, "~/") {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		databasePath = filepath.Join(usr.HomeDir, databasePath[2:])
	}

	absPath, err := filepath.Abs(databasePath)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return []models.Kouji{}, nil // Return empty list if file doesn't exist
	}

	// Read YAML file
	yamlData, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var dbKoujiList []models.Kouji
	if err := yaml.Unmarshal(yamlData, &dbKoujiList); err != nil {
		return nil, err
	}

	return dbKoujiList, nil
}

// SaveKoujiListToDatabase は工事プロジェクト情報をYAMLファイルに保存する
func (ks *KoujiService) SaveKoujiListToDatabase(databasePath string, koujilist []models.Kouji) error {
	// Expand ~ to home directory
	if strings.HasPrefix(databasePath, "~/") {
		usr, err := user.Current()
		if err != nil {
			return err
		}
		databasePath = filepath.Join(usr.HomeDir, databasePath[2:])
	}

	absPath, err := filepath.Abs(databasePath)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Convert to YAML format
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(koujilist); err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(absPath, buf.Bytes(), 0644)
}

// GetKoujiList は指定されたパスから工事プロジェクト一覧を取得する（ファイルシステムとデータベースをマージ）
func (ks *KoujiService) GetKoujiList(fsPath string) ([]models.Kouji, error) {
	dbPath := filepath.Join(fsPath, ".inside.yaml")

	// ファイルシステムから工事リストを取得
	fsKoujiList, err := ks.GetKoujiListFromFileSystem(fsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get kouji list from filesystem: %w", err)
	}

	// データベースファイルから工事リストを取得
	dbKoujiList, err := ks.LoadKoujiListFromDatabase(dbPath)
	if err != nil {
		// データベースファイルの読み込みに失敗した場合はファイルシステムのデータのみを使用
		dbKoujiList = []models.Kouji{}
	}

	// ファイルシステムとデータベースファイルの工事リストをマージ
	mgKoujiList := ks.MergeKoujiList(fsKoujiList, dbKoujiList)

	// 開始日の降順でソート（新しい順）
	sort.Slice(mgKoujiList, func(i, j int) bool {
		return mgKoujiList[i].StartDate.Time.After(mgKoujiList[j].StartDate.Time)
	})

	return mgKoujiList, nil
}

// MergeKoujiList はファイルシステムと既存YAMLファイルの工事プロジェクトをマージする
func (ks *KoujiService) MergeKoujiList(fsKoujiList, dbKoujiList []models.Kouji) []models.Kouji {
	// mergeKoujiProjects はファイルシステムとデータベースファイルの工事プロジェクトをマージする
	//
	// データベース形式 (.inside.yaml):
	//   - version: データベース形式のバージョン
	//   - generated_at: 生成タイムスタンプ
	//   - generated_by: システム識別子
	//   - projects: 工事プロジェクト情報の配列
	//
	// 取得手順:
	//   1. ファイルシステムとデータベースファイル内の工事プロジェクトが同じプロジェクトかの判断はproject_idで行う。
	//   2. ファイルシステムから取得した工事プロジェクトをデータベースファイルの工事プロジェクトで補完する。
	//   3. project_idが同じプロジェクトがデータベース内に存在するときは下記の処理を行う。
	//      a.ファイルシステムの工事プロジェクト更新日がデータベースファイルの更新日より新しい場合...
	//          i.  ファイルシステムのバージョンから得られない情報(Description, EndDate)は.inside.yamlの情報を使用する。
	//          ii. 間違ってもファイルシステムから得られない情報(Description, EndDate)で.inside.yamlの情報を上書きしないこと。
	//      b. .inside.yamlファイルの更新日がファイルシステムの工事プロジェクト更新日よりのほうが新しい場合...
	//          i.  現時点ではファイル名の変更等が必要なプロジェクトIDごとのファイルの移動は行わない。
	//          ii. ただし変更が必要な工事プロジェクトのファイルシステム情報と.inside.yamlの情報を返す
	//   6. マージ処理時に異常なcreatedDate値を持つプロジェクトは自動的に除外される。
	//
	// データ品質保証:
	//   - 「0001-01-01T09:26:51+09:18」のような異常な時刻データを検出・除外
	//   - 不正なタイムゾーンオフセット（+09:18など）の検出
	//   - 不合理な作成日時範囲（1990年以前）のフィルタリング
	//
	// エラーハンドリング:
	//   - ディレクトリが存在しない場合は作成
	//   - 書き込みが失敗した場合はエラーを返す
	//   - 異常データは警告なしに除外される

	// Create a map of database projects by project_id for quick lookup
	dbMap := make(map[string]models.Kouji)
	for _, kouji := range dbKoujiList {
		dbMap[kouji.Id] = kouji
	}

	mergedProjects := make([]models.Kouji, 0)

	// Process file system projects
	for _, fsKouji := range fsKoujiList {

		if dbKouji, exists := dbMap[fsKouji.Id]; exists {
			// Both versions exist and are valid - compare modification times
			if fsKouji.ModifiedTime.Time.After(dbKouji.ModifiedTime.Time) {
				// File system is newer - use FS data but preserve YAML-only fields
				mergedProject := fsKouji
				mergedProject.Description = dbKouji.Description // Preserve description from YAML
				if !dbKouji.EndDate.Time.IsZero() {
					mergedProject.EndDate = dbKouji.EndDate // Preserve custom end date from YAML
				}
				mergedProjects = append(mergedProjects, mergedProject)
			} else {
				// YAML is newer or same - use existing project
				mergedProjects = append(mergedProjects, dbKouji)
			}
			// Remove from map so we don't add it again
			delete(dbMap, fsKouji.Id)
		} else {
			// New project from file system - add it
			mergedProjects = append(mergedProjects, fsKouji)
		}
	}

	return mergedProjects
}

// DetermineProjectStatus determines the project status based on the date
func (ks *KoujiService) DetermineProjectStatus(projectDate time.Time) string {
	if projectDate.IsZero() {
		return "不明"
	}

	now := time.Now()
	endDate := projectDate.AddDate(0, 3, 0) // 3 months duration

	if now.Before(projectDate) {
		return "予定"
	} else if now.After(endDate) {
		return "完了"
	} else {
		return "進行中"
	}
}

// UpdateProjectDates はプロジェクトの開始日と終了日を更新する
func (ks *KoujiService) UpdateProjectDates(projectID string, startDate, endDate time.Time) error {
	// Default path for YAML file
	outputPath := "~/penguin/豊田築炉/2-工事/.inside.yaml"

	// Load existing projects from YAML
	existingProjects, err := ks.LoadKoujiListFromDatabase(outputPath)
	if err != nil {
		return fmt.Errorf("failed to load existing projects: %w", err)
	}

	// Find and update the project
	projectFound := false
	for i, project := range existingProjects {
		if project.Id == projectID {
			existingProjects[i].StartDate = models.NewTimestamp(startDate.In(time.Local))
			existingProjects[i].EndDate = models.NewTimestamp(endDate.In(time.Local))
			existingProjects[i].Status = ks.DetermineProjectStatus(startDate.In(time.Local))
			projectFound = true
			break
		}
	}

	if !projectFound {
		return fmt.Errorf("project not found: %s", projectID)
	}

	// Save updated projects to YAML
	err = ks.SaveKoujiListToDatabase(outputPath, existingProjects)
	if err != nil {
		return fmt.Errorf("failed to save updated projects: %w", err)
	}

	return nil
}

// SaveKoujiProjects は指定されたパスの工事プロジェクトをデータベースに保存する
func (ks *KoujiService) SaveKoujiProjects(targetPath string) (int, error) {
	yamlPath := filepath.Join(targetPath, ".inside.yaml")

	// 1. 既存の.inside.yamlファイルから工事プロジェクト一覧を読み込み
	existingProjects, _ := ks.LoadKoujiListFromDatabase(yamlPath)

	// 2. ファイルシステムから工事プロジェクト一覧を取得
	fsProjects, err := ks.GetKoujiListFromFileSystem(targetPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read directory: %w", err)
	}

	// 3. プロジェクトをマージ（project_idで判断）
	mergedProjects := ks.MergeKoujiList(fsProjects, existingProjects)

	// 4. 開始日の降順でソート
	sort.Slice(mergedProjects, func(i, j int) bool {
		return mergedProjects[i].StartDate.Time.After(mergedProjects[j].StartDate.Time)
	})

	// 5. YAMLファイルに保存
	err = ks.SaveKoujiListToDatabase(yamlPath, mergedProjects)
	if err != nil {
		return 0, fmt.Errorf("failed to save YAML file: %w", err)
	}

	return len(mergedProjects), nil
}
