package models

import "time"

// KoujiFolder represents a construction project folder with additional metadata
// @Description Construction project folder information with extended attributes
type KoujiFolder struct {
	// Embed the base Folder struct
	Folder

	// Additional fields specific to Kouji folders
	ProjectID    string    `json:"project_id,omitempty" example:"PRJ-2024-001"`
	ProjectName  string    `json:"project_name,omitempty" example:"豊田築炉工事"`
	Status       string    `json:"status,omitempty" example:"進行中"`
	StartDate    time.Time `json:"start_date,omitempty" example:"2024-01-01T00:00:00Z"`
	EndDate      time.Time `json:"end_date,omitempty" example:"2024-12-31T00:00:00Z"`
	Description  string    `json:"description,omitempty" example:"工事関連の資料とドキュメント"`
	Tags         []string  `json:"tags,omitempty" example:"['工事', '豊田', '築炉']"`
	FileCount    int       `json:"file_count,omitempty" example:"42"`
	SubdirCount  int       `json:"subdir_count,omitempty" example:"5"`
}

// KoujiFolderListResponse represents the response for listing kouji folders
// @Description Response containing list of construction project folders
type KoujiFolderListResponse struct {
	Folders    []KoujiFolder `json:"folders" description:"List of kouji folders"`
	Count      int           `json:"count" example:"10" description:"Total number of folders returned"`
	Path       string        `json:"path" example:"~/penguin/豊田築炉/2-工事" description:"The path that was queried"`
	TotalSize  int64         `json:"total_size,omitempty" example:"1073741824" description:"Total size of all files in bytes"`
}