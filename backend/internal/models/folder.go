package models

import "time"

// Folder represents a file or directory
// @Description File or directory information
type Folder struct {
	Id uint64 `json:"id" yaml:"id" example:"123456"`
	// Name of the file or folder
	Name string `json:"name" yaml:"name" example:"documents"`
	// Full path to the file or folder
	Path string `json:"path" yaml:"path" example:"/home/user/documents"`
	// Whether this item is a directory
	IsDirectory bool `json:"is_directory" yaml:"isdirectory" example:"true"`
	// Size of the file in bytes
	Size int64 `json:"size" yaml:"size" example:"4096"`
	// Last modification time
	ModifiedTime time.Time `json:"modified_time" yaml:"modifiedtime" example:"2024-01-15T10:30:00Z"`
}

type FolderYAML struct {
	Id uint64 `json:"id" yaml:"id" example:"123456"`
	// Name of the file or folder
	Name string `json:"name" yaml:"name" example:"documents"`
	// Full path to the file or folder
	Path string `json:"path" yaml:"path" example:"/home/user/documents"`
	// Whether this item is a directory
	IsDirectory bool `json:"is_directory" yaml:"isdirectory" example:"true"`
	// Size of the file in bytes
	Size int64 `json:"size" yaml:"size" example:"4096"`
	// Last modification time
	ModifiedTime string `json:"modified_time" yaml:"modifiedtime" example:"2024-01-15T10:30:00Z"`
}

// FolderListResponse represents the response for folder listing
// @Description Response containing list of folders
type FolderListResponse struct {
	// List of folders
	Folders []Folder `json:"folders"`
	// Total number of folders returned
	Count int `json:"count" example:"10"`
}

type FolderListResponseYAML struct {
	// List of folders
	Folders []FolderYAML `json:"folders"`
	// Total number of folders returned
	Count int `json:"count" example:"10"`
}
