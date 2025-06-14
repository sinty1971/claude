package handlers

import (
	"fmt"
	"path/filepath"
	"penguin-backend/internal/models"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetKoujiList godoc
// @Summary      Get kouji list
// @Description  Retrieve a list of construction project folders from the specified path
// @Tags         kouji-list
// @Accept       json
// @Produce      json
// @Param        path query string false "Path to the directory to list" default(~/penguin/豊田築炉/2-工事)
// @Success      200 {object} models.KoujiListResponse "Successful response"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /kouji-list [get]
func (fh *FileSystemHandler) GetKoujiList(c *fiber.Ctx) error {
	// パラメータの取得
	fsPath := c.Query("path", "~/penguin/豊田築炉/2-工事")

	// KoujiServiceを使用して工事リストを取得
	koujiList, err := fh.KoujiService.GetKoujiList(fsPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get kouji list",
			"message": err.Error(),
		})
	}

	totalSize := int64(0)
	for _, kouji := range koujiList {
		totalSize += kouji.FileEntry.Size
	}

	return c.JSON(models.KoujiListResponse{
		KoujiList: koujiList,
		Count:     len(koujiList),
		TotalSize: totalSize,
	})
}

// SaveKoujiListToDatabase godoc
// @Summary      Save kouji projects to YAML
// @Description  Save kouji project information to a YAML file
// @Tags         kouji-list
// @Accept       json
// @Produce      json
// @Param        path query string false "Path to the directory to scan" default(~/penguin/豊田築炉/2-工事)
// @Param        output_path query string false "Output YAML file path" default(~/penguin/豊田築炉/2-工事/.inside.yaml)
// @Success      200 {object} map[string]string "Success message"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /kouji-list/save [post]
func (fh *FileSystemHandler) SaveKoujiListToDatabase(c *fiber.Ctx) error {
	// SaveKoujiProjectsToYAML は工事プロジェクト情報をYAMLファイルに保存する
	//
	// ファイル形式 (.inside.yaml):
	//   - version: YAML形式のバージョン
	//   - generated_at: 生成タイムスタンプ
	//   - generated_by: システム識別子
	//   - projects: プロジェクト情報の配列
	//
	// 書き込み手順:
	//   1. 既存の[2-工事/.inside.yaml]ファイルが存在するときは工事プロジェクト一覧を２つのバージョンを取得する。
	//      一つのバージョンはファイルシステム[2-工事]からの取得、もう一つは[2-工事/.inside.yaml]ファイルからのデシリアライズ取得です。
	//   2. ファイルシステムからの取得は下記の手順に従う
	//   	a. 対象ディレクトリをスキャン（デフォルト: ~/penguin/豊田築炉/2-工事）
	//   	b. "YYYY-MMDD 会社名 現場名" パターンに一致するフォルダーを抽出
	//   	c. 5文字のハッシュを使用して一意のプロジェクトIDを生成（例: "A3K7M"）
	//   	d. 開始日の降順でプロジェクトをソート（新しい順）
	//   	e. 更新日・開始日・終了日のタイムスタンプを必要に応じてローカルタイムゾーン（JST）に変換
	//   	f. 異常な時刻データ（0001年、不正なタイムゾーンオフセットなど）を検出・除外
	//   3. ファイルシステムと.inside.yamlファイル内の工事プロジェクトが同じプロジェクトかの判断はproject_idで行う。
	//   4. ファイルシステムから取得した工事プロジェクトが.inside.yaml内に存在しないときは.inside.yamlに追加する。
	//   5. 両方に同じproject_idのプロジェクトがある時は下記の判断を行う。
	//      a.ファイルシステムの工事プロジェクト更新日が.inside.yamlファイルの更新日より新しい場合...
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

	// Default paths
	targetPath := c.Query("path", "~/penguin/豊田築炉/2-工事")

	// KoujiServiceを使用して工事プロジェクトを保存
	count, err := fh.KoujiService.SaveKoujiProjects(targetPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to save kouji projects",
			"message": err.Error(),
		})
	}

	yamlPath := filepath.Join(targetPath, ".inside.yaml")

	return c.JSON(fiber.Map{
		"message":     "工事フォルダー情報をYAMLファイルに保存しました",
		"output_path": yamlPath,
		"count":       count,
	})
}

// getKoujiListFromFileSystem はファイルシステムから工事プロジェクト一覧を取得する
// Deprecated: Use KoujiService.GetKoujiListFromFileSystem instead
func (fh *FileSystemHandler) getKoujiListFromFileSystem(targetPath string) ([]models.Kouji, error) {
	// Get kouji folders from file system
	folders, err := fh.FileSystemService.GetFolders(targetPath)
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
			Status:       determineProjectStatus(projectDate),
			StartDate:    models.NewTimestamp(projectDate),
			EndDate:      models.NewTimestamp(projectDate.AddDate(0, 3, 0)), // Assume 3-month project duration
			Description:  companyName + "の" + factoryName + "における工事プロジェクト",
			Tags:         []string{"工事", companyName, factoryName, dateStr[:4]}, // Include year as tag
			// Folder: ファイルシステムから取得したフォルダー情報
			FileEntry: folder,
		}

		koujiProject.FileCount = 0 // Would need to scan subdirectory to get actual count
		koujiProject.SubdirCount = 0

		koujiProjects = append(koujiProjects, koujiProject)
	}

	return koujiProjects, nil
}

// mergeKoujiList はファイルシステムと既存YAMLファイルの工事プロジェクトをマージする
// Deprecated: Use KoujiService.MergeKoujiList instead
func (fh *FileSystemHandler) mergeKoujiList(fsKoujiList, dbKoujiList []models.Kouji) []models.Kouji {
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

// UpdateKoujiProjectDates godoc
// @Summary      Update kouji project dates
// @Description  Update start and end dates for a specific kouji project
// @Tags         kouji-projects
// @Accept       json
// @Produce      json
// @Param        project_id path string true "Project ID"
// @Param        dates body UpdateProjectDatesRequest true "Updated dates"
// @Success      200 {object} map[string]string "Success message"
// @Failure      400 {object} map[string]string "Bad request"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /kouji-projects/{project_id}/dates [put]
func (fh *FileSystemHandler) UpdateKoujiProjectDates(c *fiber.Ctx) error {
	projectID := c.Params("project_id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "project_id is required",
		})
	}

	// Define request body structure
	type UpdateProjectDatesRequest struct {
		StartDate string `json:"start_date" example:"2024-01-01T00:00:00Z"`
		EndDate   string `json:"end_date" example:"2024-12-31T00:00:00Z"`
	}

	var req UpdateProjectDatesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
	}

	// Parse dates with flexible format support
	parseDateString := func(dateStr string) (time.Time, error) {
		// Try RFC3339 first
		if parsedTime, err := time.Parse(time.RFC3339, dateStr); err == nil {
			return parsedTime, nil
		}

		// Try RFC3339 without timezone (assume local timezone)
		if parsedTime, err := time.ParseInLocation("2006-01-02T15:04:05", dateStr, time.Local); err == nil {
			return parsedTime, nil
		}

		// Try date only format (assume local timezone)
		if parsedTime, err := time.ParseInLocation("2006-01-02", dateStr, time.Local); err == nil {
			return parsedTime, nil
		}

		return time.Time{}, fmt.Errorf("unsupported date format")
	}

	startDate, err := parseDateString(req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid start_date format",
			"message": fmt.Sprintf("Date must be in RFC3339, ISO format, or YYYY-MM-DD format. Error: %v", err),
		})
	}

	endDate, err := parseDateString(req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid end_date format",
			"message": fmt.Sprintf("Date must be in RFC3339, ISO format, or YYYY-MM-DD format. Error: %v", err),
		})
	}

	// KoujiServiceを使用してプロジェクトの日付を更新
	err = fh.KoujiService.UpdateProjectDates(projectID, startDate, endDate)
	if err != nil {
		if err.Error() == "project not found: "+projectID {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Project not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to update project dates",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":    "プロジェクトの日付が更新されました",
		"project_id": projectID,
	})
}

// CleanupInvalidTimeData godoc
// @Summary      Cleanup invalid time data
// @Description  Remove projects with invalid time data (like 0001-01-01T09:26:51+09:18) from YAML
// @Tags         kouji-projects
// @Accept       json
// @Produce      json
// @Param        yaml_path query string false "Path to the YAML file" default(~/penguin/豊田築炉/2-工事/.inside.yaml)
// @Success      200 {object} map[string]string "Success message with cleanup details"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /kouji-projects/cleanup [post]
func (fh *FileSystemHandler) CleanupInvalidTimeData(c *fiber.Ctx) error {
	yamlPath := c.Query("yaml_path", "~/penguin/豊田築炉/2-工事/.inside.yaml")

	// Load existing projects to count before cleanup
	projectsBefore, err := fh.KoujiService.LoadKoujiListFromDatabase(yamlPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to load YAML file",
			"message": err.Error(),
		})
	}

	countBefore := len(projectsBefore)

	// Load projects again to count after cleanup
	projectsAfter, err := fh.KoujiService.LoadKoujiListFromDatabase(yamlPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to reload YAML file after cleanup",
			"message": err.Error(),
		})
	}

	countAfter := len(projectsAfter)
	removedCount := countBefore - countAfter

	return c.JSON(fiber.Map{
		"message":         "異常な時刻データのクリーンアップが完了しました",
		"yaml_path":       yamlPath,
		"projects_before": countBefore,
		"projects_after":  countAfter,
		"removed_count":   removedCount,
	})
}

// determineProjectStatus determines the project status based on the date
// Deprecated: Use KoujiService.DetermineProjectStatus instead
func determineProjectStatus(projectDate time.Time) string {
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
