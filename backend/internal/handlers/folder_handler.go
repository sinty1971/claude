package handlers

import (
	"penguin-backend/internal/models"
	"penguin-backend/internal/services"
	"regexp"
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
// @Param        path query string false "Path to the directory to list" default(~/penguin/豊田築炉/2-工事)
// @Success      200 {object} models.FolderListResponse "Successful response"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /folders [get]
func (fh *FolderHandler) GetFolders(c *fiber.Ctx) error {
	targetPath := c.Query("path", "~/penguin/豊田築炉/2-工事")

	folders, err := fh.folderService.GetFolders(targetPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to read directory",
			"message": err.Error(),
		})
	}

	return c.JSON(folders)
}

// GetKoujiFolders godoc
// @Summary      Get kouji folders
// @Description  Retrieve a list of construction project folders from the specified path
// @Tags         kouji-folders
// @Accept       json
// @Produce      json
// @Param        path query string false "Path to the directory to list" default(~/penguin/豊田築炉/2-工事)
// @Success      200 {object} models.KoujiFolderListResponse "Successful response"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /kouji-folders [get]
func (fh *FolderHandler) GetKoujiFolders(c *fiber.Ctx) error {
	// Fixed target path for kouji folders
	targetPath := "~/penguin/豊田築炉/2-工事"
	
	// Allow override via query parameter if needed
	if queryPath := c.Query("path"); queryPath != "" {
		targetPath = queryPath
	}

	folders, err := fh.folderService.GetFolders(targetPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to read directory",
			"message": err.Error(),
		})
	}

	// Regular expression to match folder names like "2025-0618 豊田築炉 名和工場"
	// Pattern: YYYY-MMDD Company Factory
	koujiPattern := regexp.MustCompile(`^(\d{4}-\d{4})\s+(.+?)\s+(.+)$`)
	
	// Convert to KoujiFolders with additional metadata
	koujiFolders := make([]models.KoujiFolder, 0)
	totalSize := int64(0)
	subdirCount := 0

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
		dateStr := matches[1]      // e.g., "2025-0618"
		companyName := matches[2]   // e.g., "豊田築炉"
		factoryName := matches[3]   // e.g., "名和工場"
		
		// Parse date from folder name
		var projectDate time.Time
		if len(dateStr) == 9 && dateStr[4] == '-' {
			year := dateStr[:4]
			monthDay := dateStr[5:]
			if len(monthDay) == 4 {
				month := monthDay[:2]
				day := monthDay[2:]
				dateTimeStr := year + "-" + month + "-" + day
				projectDate, _ = time.Parse("2006-01-02", dateTimeStr)
			}
		}
		
		koujiFolder := models.KoujiFolder{
			Folder: folder,
			// Generate project metadata based on folder name
			ProjectID:   "PRJ-" + dateStr,
			ProjectName: companyName + " " + factoryName + "工事",
			Status:      determineProjectStatus(projectDate),
			StartDate:   projectDate,
			EndDate:     projectDate.AddDate(0, 3, 0), // Assume 3-month project duration
			Description: companyName + "の" + factoryName + "における工事プロジェクト",
			Tags:        []string{"工事", companyName, factoryName, dateStr[:4]}, // Include year as tag
		}
		
		subdirCount++
		koujiFolder.FileCount = 0 // Would need to scan subdirectory to get actual count
		koujiFolder.SubdirCount = 0
		
		koujiFolders = append(koujiFolders, koujiFolder)
	}

	response := models.KoujiFolderListResponse{
		Folders:   koujiFolders,
		Count:     len(koujiFolders),
		Path:      targetPath,
		TotalSize: totalSize,
	}

	return c.JSON(response)
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
