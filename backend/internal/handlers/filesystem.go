package handlers

import (
	"penguin-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

type FileSystemHandler struct {
	FileSystemService *services.FileSystemService
}

func NewFileSystemHandler() *FileSystemHandler {
	return &FileSystemHandler{
		FileSystemService: services.NewFileSystemService(),
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
func (h *FileSystemHandler) GetFolders(c *fiber.Ctx) error {
	fsPath := c.Query("path", "~/penguin")

	folders, err := h.FileSystemService.GetFolders(fsPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to read directory",
			"message": err.Error(),
		})
	}

	return c.JSON(folders)
}
