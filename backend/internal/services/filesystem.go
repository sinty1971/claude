package services

import (
	"os"
	"os/user"
	"path/filepath"
	"penguin-backend/internal/models"
	"strings"
	"syscall"
)

type FileSystemService struct{}

func NewFileSystemService() *FileSystemService {
	return &FileSystemService{}
}

func (fs *FileSystemService) GetFolders(targetPath string) (*models.FolderListResponse, error) {
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

	var folders []models.FileEntry
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		stat := info.Sys().(*syscall.Stat_t)

		// Check if entry is a directory
		// For symlinks, check the target's type
		isDirectory := entry.IsDir()
		entryPath := filepath.Join(absPath, entry.Name())

		// If it's a symlink, check what it points to
		if info.Mode()&os.ModeSymlink != 0 {
			targetInfo, err := os.Stat(entryPath) // Follow the symlink
			if err == nil {
				isDirectory = targetInfo.IsDir()
			}
		}

		folder := models.FileEntry{
			Id:           stat.Ino,
			Name:         entry.Name(),
			Path:         entryPath,
			IsDirectory:  isDirectory,
			Size:         info.Size(),
			ModifiedTime: models.NewTimestamp(info.ModTime()),
		}
		folders = append(folders, folder)
	}

	return &models.FolderListResponse{
		Folders: folders,
		Count:   len(folders),
	}, nil
}
