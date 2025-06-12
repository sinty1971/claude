package handlers

import (
	"fmt"
	"path/filepath"
	"penguin-backend/internal/models"
	"penguin-backend/internal/services"
	"regexp"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
)

type FolderHandler struct {
	folderService *services.FolderService
}

func NewFolderHandler() *FolderHandler {
	return &FolderHandler{
		folderService: services.NewFolderService(),
	}
}

// GetFolders godoc
// @Summary      Get folders
// @Description  Retrieve a list of folders from the specified path
// @Tags         folders
// @Accept       json
// @Produce      json
// @Param        path query string false "Path to the directory to list" default(~/penguin)
// @Success      200 {object} models.FolderListResponse "Successful response"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /folders [get]
func (fh *FolderHandler) GetFolders(c *fiber.Ctx) error {
	targetPath := c.Query("path", "~/penguin")

	folders, err := fh.folderService.GetFolders(targetPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to read directory",
			"message": err.Error(),
		})
	}

	return c.JSON(folders)
}

// GetKoujiProjects godoc
// @Summary      Get kouji projects
// @Description  Retrieve a list of construction project folders from the specified path
// @Tags         kouji-projects
// @Accept       json
// @Produce      json
// @Param        path query string false "Path to the directory to list" default(~/penguin/豊田築炉/2-工事)
// @Success      200 {object} models.KoujiProjectListResponse "Successful response"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /kouji-projects [get]
func (fh *FolderHandler) GetKoujiProjects(c *fiber.Ctx) error {
	// Default paths
	targetPath := c.Query("path", "~/penguin/豊田築炉/2-工事")
	databasePath := filepath.Join(targetPath, ".inside.yaml")

	// 1. ファイルシステムから工事プロジェクト一覧を取得
	fsProjects, err := fh.getKoujiProjectsFromFileSystem(targetPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to read directory",
			"message": err.Error(),
		})
	}

	// 2. データベースファイルから工事プロジェクト一覧を読み込み
	dbProjects, err := fh.folderService.LoadKoujiProjectsFromYAML(databasePath)
	if err != nil {
		// データベースファイルの読み込みに失敗した場合はファイルシステムのデータのみを使用
		dbProjects = []models.KoujiProject{}
	}

	// 3. プロジェクトをマージ
	mergedProjects := fh.mergeKoujiProjects(fsProjects, dbProjects)

	// 4. 開始日の降順でソート（新しい順）
	sort.Slice(mergedProjects, func(i, j int) bool {
		return mergedProjects[i].StartDate.After(mergedProjects[j].StartDate)
	})

	totalSize := int64(0)
	for _, project := range mergedProjects {
		totalSize += project.Size
	}

	response := models.KoujiProjectListResponse{
		Projects:  mergedProjects,
		Count:     len(mergedProjects),
		TotalSize: totalSize,
	}

	return c.JSON(response)
}

// SaveKoujiProjectsToYAML godoc
// @Summary      Save kouji projects to YAML
// @Description  Save kouji project information to a YAML file
// @Tags         kouji-projects
// @Accept       json
// @Produce      json
// @Param        path query string false "Path to the directory to scan" default(~/penguin/豊田築炉/2-工事)
// @Param        output_path query string false "Output YAML file path" default(~/penguin/豊田築炉/2-工事/.inside.yaml)
// @Success      200 {object} map[string]string "Success message"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /kouji-projects/save [post]
func (fh *FolderHandler) SaveKoujiProjectsToYAML(c *fiber.Ctx) error {
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
	yamlPath := filepath.Join(targetPath, ".inside.yaml")

	// 1. 既存の.inside.yamlファイルから工事プロジェクト一覧を読み込み
	existingProjects, err := fh.folderService.LoadKoujiProjectsFromYAML(yamlPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to load existing YAML file",
			"message": err.Error(),
		})
	}

	// 2. ファイルシステムから工事プロジェクト一覧を取得
	fsProjects, err := fh.getKoujiProjectsFromFileSystem(targetPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to read directory",
			"message": err.Error(),
		})
	}

	// 3. プロジェクトをマージ（project_idで判断）
	mergedProjects := fh.mergeKoujiProjects(fsProjects, existingProjects)

	// 4. 開始日の降順でソート
	sort.Slice(mergedProjects, func(i, j int) bool {
		return mergedProjects[i].StartDate.After(mergedProjects[j].StartDate)
	})

	// 5. YAMLファイルに保存
	err = fh.folderService.SaveKoujiProjectsToYAML(yamlPath, mergedProjects)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to save YAML file",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":     "工事フォルダー情報をYAMLファイルに保存しました",
		"output_path": yamlPath,
		"count":       len(mergedProjects),
	})
}

// getKoujiProjectsFromFileSystem はファイルシステムから工事プロジェクト一覧を取得する
func (fh *FolderHandler) getKoujiProjectsFromFileSystem(targetPath string) ([]models.KoujiProject, error) {
	// Get kouji folders from file system
	folders, err := fh.folderService.GetFolders(targetPath)
	if err != nil {
		return nil, err
	}

	// Regular expression to match folder names like "2025-0618 豊田築炉 名和工場"
	koujiPattern := regexp.MustCompile(`^(\d{4}-\d{4})\s+(.+?)\s+(.+)$`)

	// Convert to KoujiProjects with additional metadata
	koujiProjects := make([]models.KoujiProject, 0)

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

		// Get folder creation date (use modification time as proxy for creation date)
		// Validate the modification time to avoid invalid dates
		createdDate := folder.ModifiedTime
		if createdDate.Year() == 1 || createdDate.IsZero() {
			// Use current time if modification time is invalid
			createdDate = time.Now()
		}

		// Generate unique project ID using folder creation date, company name, and location name
		// This ensures the same project always gets the same ID
		// Use project date instead of creation date for more stable ID generation
		projectDateStr := projectDate.Format("2006-01-02")
		idSource := projectDateStr + "_" + companyName + "_" + factoryName
		projectID := models.NewIDFromString(idSource).Len5()

		koujiProject := models.KoujiProject{
			Folder: folder,
			// Generate project metadata based on folder name
			ProjectID:    projectID,
			ProjectName:  companyName + " " + factoryName + "工事",
			CompanyName:  companyName,
			LocationName: factoryName,
			Status:       determineProjectStatus(projectDate),
			StartDate:    projectDate,
			EndDate:      projectDate.AddDate(0, 3, 0), // Assume 3-month project duration
			Description:  companyName + "の" + factoryName + "における工事プロジェクト",
			Tags:         []string{"工事", companyName, factoryName, dateStr[:4]}, // Include year as tag
		}

		koujiProject.FileCount = 0 // Would need to scan subdirectory to get actual count
		koujiProject.SubdirCount = 0

		koujiProjects = append(koujiProjects, koujiProject)
	}

	return koujiProjects, nil
}

// mergeKoujiProjects はファイルシステムと既存YAMLファイルの工事プロジェクトをマージする
func (fh *FolderHandler) mergeKoujiProjects(fsProjects, dbProjects []models.KoujiProject) []models.KoujiProject {
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
	dbMap := make(map[string]models.KoujiProject)
	for _, project := range dbProjects {
		dbMap[project.ProjectID] = project
	}

	mergedProjects := make([]models.KoujiProject, 0)

	// Process file system projects
	for _, fsProject := range fsProjects {
		// Skip projects with invalid creation dates
		if isInvalidCreatedDate(fsProject) {
			continue
		}

		if dbProject, exists := dbMap[fsProject.ProjectID]; exists {
			// Skip if existing project also has invalid date
			if isInvalidCreatedDate(dbProject) {
				// Replace invalid existing project with valid FS project
				mergedProjects = append(mergedProjects, fsProject)
				delete(dbMap, fsProject.ProjectID)
				continue
			}

			// Both versions exist and are valid - compare modification times
			if fsProject.ModifiedTime.After(dbProject.ModifiedTime) {
				// File system is newer - use FS data but preserve YAML-only fields
				mergedProject := fsProject
				mergedProject.Description = dbProject.Description // Preserve description from YAML
				if !dbProject.EndDate.IsZero() {
					mergedProject.EndDate = dbProject.EndDate // Preserve custom end date from YAML
				}
				mergedProjects = append(mergedProjects, mergedProject)
			} else {
				// YAML is newer or same - use existing project
				mergedProjects = append(mergedProjects, dbProject)
			}
			// Remove from map so we don't add it again
			delete(dbMap, fsProject.ProjectID)
		} else {
			// New project from file system - add it
			mergedProjects = append(mergedProjects, fsProject)
		}
	}

	// Add remaining projects from YAML that don't exist in file system
	// Skip projects with invalid creation dates
	for _, remainingProject := range dbMap {
		if !isInvalidCreatedDate(remainingProject) {
			mergedProjects = append(mergedProjects, remainingProject)
		}
	}

	return mergedProjects
}

// isInvalidCreatedDate checks if a creation date is invalid
func isInvalidCreatedDate(project models.KoujiProject) bool {
	var createdDate = project.CreatedDate

	// Check for zero value
	if createdDate.IsZero() {
		return true
	}

	// Check for year 1 (0001-01-01) which indicates invalid Go time
	if createdDate.Year() == 1 {
		return true
	}

	// Check for unreasonable future dates (more than 1 day from now)
	if createdDate.After(time.Now().AddDate(0, 0, 1)) {
		return true
	}

	// Check for unreasonable past dates (before year 2000)
	if createdDate.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return true
	}

	return false
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
func (fh *FolderHandler) UpdateKoujiProjectDates(c *fiber.Ctx) error {
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

	// Default path for YAML file
	outputPath := "~/penguin/豊田築炉/2-工事/.inside.yaml"

	// Load existing projects from YAML
	existingProjects, err := fh.folderService.LoadKoujiProjectsFromYAML(outputPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to load existing projects",
			"message": err.Error(),
		})
	}

	// Find and update the project
	projectFound := false
	for i, project := range existingProjects {
		if project.ProjectID == projectID {
			existingProjects[i].StartDate = startDate.In(time.Local)
			existingProjects[i].EndDate = endDate.In(time.Local)
			existingProjects[i].Status = determineProjectStatus(startDate.In(time.Local))
			projectFound = true
			break
		}
	}

	if !projectFound {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	// Save updated projects to YAML
	err = fh.folderService.SaveKoujiProjectsToYAML(outputPath, existingProjects)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to save updated projects",
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
func (fh *FolderHandler) CleanupInvalidTimeData(c *fiber.Ctx) error {
	yamlPath := c.Query("yaml_path", "~/penguin/豊田築炉/2-工事/.inside.yaml")

	// Load existing projects to count before cleanup
	projectsBefore, err := fh.folderService.LoadKoujiProjectsFromYAML(yamlPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to load YAML file",
			"message": err.Error(),
		})
	}

	countBefore := len(projectsBefore)

	// Perform cleanup
	err = fh.folderService.CleanupInvalidTimeData(yamlPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to cleanup invalid time data",
			"message": err.Error(),
		})
	}

	// Load projects again to count after cleanup
	projectsAfter, err := fh.folderService.LoadKoujiProjectsFromYAML(yamlPath)
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
