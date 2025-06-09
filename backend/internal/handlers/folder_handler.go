package handlers

import (
	"penguin-backend/internal/services"

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

func (fh *FolderHandler) GetFolders(c *fiber.Ctx) error {
	targetPath := c.Query("path", "~/penguin/2-工事")
	
	folders, err := fh.folderService.GetFolders(targetPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to read directory",
			"message": err.Error(),
		})
	}

	return c.JSON(folders)
}