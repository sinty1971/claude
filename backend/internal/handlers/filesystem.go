package handlers

import (
	"penguin-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

type FileSystemHandler struct {
	FileSystemService *services.FileSystemService
}

func NewFileSystemHandler(fsService *services.FileSystemService) *FileSystemHandler {
	return &FileSystemHandler{
		FileSystemService: fsService,
	}
}

// GetFileEntries godoc
// @Summary      Get folders
// @Description  Retrieve a list of folders from the specified path
// @Tags         file-entries
// @Accept       json
// @Produce      json
// @Param        path query string false "Path to the directory to list" default(~/penguin)
// @Success      200 {object} models.FolderListResponse "Successful response"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /file-entries [get]
func (h *FileSystemHandler) GetFileEntries(c *fiber.Ctx) error {
	fsPath := c.Query("path", "~/penguin")

	fileEntries, err := h.FileSystemService.GetFileEntries(fsPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to read directory",
			"message": err.Error(),
		})
	}

	return c.JSON(fileEntries)
}
