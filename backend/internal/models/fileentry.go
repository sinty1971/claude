package models

// FileEntry represents a file or directory
// @Description File or directory information
type FileEntry struct {
	Id uint64 `json:"id" yaml:"id" example:"123456"`
	// Name of the file or folder
	Name string `json:"name" yaml:"name" example:"documents"`
	// Full path to the file or folder
	Path string `json:"path" yaml:"path" example:"/home/user/documents"`
	// Whether this item is a directory
	IsDirectory bool `json:"is_directory" yaml:"is_directory" example:"true"`
	// Size of the file in bytes
	Size int64 `json:"size" yaml:"size" example:"4096"`
	// Last modification time
	ModifiedTime Timestamp `json:"modified_time" yaml:"modified_time"`
}

// FolderListResponse represents the response for folder listing
// @Description Response containing list of folders
type FolderListResponse struct {
	// List of folders
	Folders []FileEntry `json:"folders"`
	// Total number of folders returned
	Count int `json:"count" example:"10"`
}