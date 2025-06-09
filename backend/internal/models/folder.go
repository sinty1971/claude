package models

import "time"

type Folder struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	IsDirectory  bool      `json:"is_directory"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
}

type FolderListResponse struct {
	Folders []Folder `json:"folders"`
	Count   int      `json:"count"`
	Path    string   `json:"path"`
}