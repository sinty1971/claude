package handlers

import (
	"fmt"
	"path/filepath"
	"penguin-backend/internal/models"
	"penguin-backend/internal/services"
	"time"

	"github.com/gofiber/fiber/v2"
)

// KoujiHandler handles kouji-related HTTP requests
type KoujiHandler struct {
	fileSystemService *services.FileSystemService
	koujiService      *services.KoujiService
}

// NewKoujiHandler creates a new KoujiHandler instance
func NewKoujiHandler(fileSystemService *services.FileSystemService) *KoujiHandler {
	return &KoujiHandler{
		fileSystemService: fileSystemService,
		koujiService:      services.NewKoujiService(fileSystemService),
	}
}

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
func (h *KoujiHandler) GetKoujiList(c *fiber.Ctx) error {
	// パラメータの取得
	fsPath := c.Query("path", "~/penguin/豊田築炉/2-工事")

	// KoujiServiceを使用して工事リストを取得
	koujiList, err := h.koujiService.GetKoujiList(fsPath)
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
func (h *KoujiHandler) SaveKoujiListToDatabase(c *fiber.Ctx) error {
	// Default paths
	targetPath := c.Query("path", "~/penguin/豊田築炉/2-工事")

	// KoujiServiceを使用して工事プロジェクトを保存
	count, err := h.koujiService.SaveKoujiProjects(targetPath)
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
func (h *KoujiHandler) UpdateKoujiProjectDates(c *fiber.Ctx) error {
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
	err = h.koujiService.UpdateProjectDates(projectID, startDate, endDate)
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
func (h *KoujiHandler) CleanupInvalidTimeData(c *fiber.Ctx) error {
	yamlPath := c.Query("yaml_path", "~/penguin/豊田築炉/2-工事/.inside.yaml")

	// Load existing projects to count before cleanup
	projectsBefore, err := h.koujiService.LoadKoujiListFromDatabase(yamlPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to load YAML file",
			"message": err.Error(),
		})
	}

	countBefore := len(projectsBefore)

	// Load projects again to count after cleanup
	projectsAfter, err := h.koujiService.LoadKoujiListFromDatabase(yamlPath)
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