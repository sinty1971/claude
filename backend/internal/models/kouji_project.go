package models

import "time"

// KoujiProject represents a construction project folder with additional metadata
// @Description Construction project folder information with extended attributes
type KoujiProject struct {
	// Embed the base Folder struct
	Folder

	// Additional fields specific to Kouji folders
	ProjectID    string    `json:"project_id,omitempty" yaml:"projectid" example:"A3K7M"`
	ProjectName  string    `json:"project_name,omitempty" yaml:"projectname" example:"豊田築炉 名和工場工事"`
	CompanyName  string    `json:"company_name,omitempty" yaml:"companyname" example:"豊田築炉"`
	LocationName string    `json:"location_name,omitempty" yaml:"locationname" example:"名和工場"`
	Status       string    `json:"status,omitempty" yaml:"status" example:"進行中"`
	StartDate    time.Time `json:"start_date,omitempty" yaml:"startdate" example:"2024-01-01T00:00:00Z"`
	EndDate      time.Time `json:"end_date,omitempty" yaml:"enddate" example:"2024-12-31T00:00:00Z"`
	Description  string    `json:"description,omitempty" yaml:"description" example:"工事関連の資料とドキュメント"`
	Tags         []string  `json:"tags,omitempty" yaml:"tags" example:"['工事', '豊田築炉', '名和工場']"`
	FileCount    int       `json:"file_count,omitempty" yaml:"filecount" example:"42"`
	SubdirCount  int       `json:"subdir_count,omitempty" yaml:"subdircount" example:"5"`
}

// KoujiProjectListResponse represents the response for listing kouji projects
// @Description Response containing list of construction project folders
type KoujiProjectListResponse struct {
	Projects  []KoujiProject `json:"projects" description:"List of kouji projects"`
	Count     int            `json:"count" example:"10" description:"Total number of projects returned"`
	TotalSize int64          `json:"total_size,omitempty" example:"1073741824" description:"Total size of all files in bytes"`
}
