package services

import (
	"os"
	"path/filepath"
	"penguin-backend/internal/models"
	"strings"
	"os/user"
)

type FolderService struct{}

func NewFolderService() *FolderService {
	return &FolderService{}
}

func (fs *FolderService) GetFolders(targetPath string) (*models.FolderListResponse, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(targetPath, "~/") {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		targetPath = filepath.Join(usr.HomeDir, targetPath[2:])
	}
	
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	var folders []models.Folder
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		folder := models.Folder{
			Name:         entry.Name(),
			Path:         filepath.Join(absPath, entry.Name()),
			IsDirectory:  entry.IsDir(),
			Size:         info.Size(),
			ModifiedTime: info.ModTime(),
		}
		folders = append(folders, folder)
	}

	response := &models.FolderListResponse{
		Folders: folders,
		Count:   len(folders),
		Path:    absPath,
	}

	return response, nil
}