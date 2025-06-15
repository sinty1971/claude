package services

import (
	"os"
	"os/user"
	"path/filepath"
	"penguin-backend/internal/models"
	"strings"
	"syscall"
)

// FileSystemService is a service for managing the file system
type FileSystemService struct {
	// Root is the root directory of the file system
	Root string `json:"root" yaml:"root" example:"/home/<user>/penguin"`
}

// NewFileSystemService creates a new FileSystemService
func NewFileSystemService(root string) (*FileSystemService, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(root, "~/") {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		root = filepath.Join(usr.HomeDir, root[2:])
	}

	// Get absolute path
	absPath, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	return &FileSystemService{
		Root: absPath,
	}, nil
}

// GetFileEntries gets the file entries from the file system
func (s *FileSystemService) GetFileEntries(fsPath string) (*models.FileEntriesListResponse, error) {
	// Expand ~ to home directory
	fsPath = filepath.Join(s.Root, fsPath)

	absPath, err := filepath.Abs(fsPath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	var fileEntries []models.FileEntry
	var folderCount int
	var fileCount int
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

		fileEntries = append(fileEntries, models.FileEntry{
			Id:           stat.Ino,
			Name:         entry.Name(),
			Path:         entryPath,
			IsDirectory:  isDirectory,
			Size:         info.Size(),
			ModifiedTime: models.NewTimestamp(info.ModTime()),
		})

		if isDirectory {
			folderCount++
		} else {
			fileCount++
		}
	}

	return &models.FileEntriesListResponse{
		FileEntries: fileEntries,
		FolderCount: folderCount,
		FileCount:   fileCount,
	}, nil
}
