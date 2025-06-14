package utils

import (
	"reflect"
	"testing"
	"time"

	"penguin-backend/internal/models"
)

func TestParse(t *testing.T) {
	// Set location to JST for testing
	jst, _ := time.LoadLocation("Asia/Tokyo")
	time.Local = jst

	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		// RFC3339 formats with timezone
		{
			name:    "RFC3339Nano with timezone",
			input:   "2025-01-14T15:30:45.123456789+09:00",
			want:    time.Date(2025, 1, 14, 15, 30, 45, 123456789, time.FixedZone("", 9*3600)),
			wantErr: false,
		},
		{
			name:    "RFC3339 with timezone",
			input:   "2025-01-14T15:30:45+09:00",
			want:    time.Date(2025, 1, 14, 15, 30, 45, 0, time.FixedZone("", 9*3600)),
			wantErr: false,
		},
		// Date only formats (local timezone)
		{
			name:    "YYYYMMDD format",
			input:   "20250114",
			want:    time.Date(2025, 1, 14, 0, 0, 0, 0, jst),
			wantErr: false,
		},
		{
			name:    "YYYY-MMDD format",
			input:   "2025-0114",
			want:    time.Date(2025, 1, 14, 0, 0, 0, 0, jst),
			wantErr: false,
		},
		{
			name:    "YYYY-MM-DD format",
			input:   "2025-01-14",
			want:    time.Date(2025, 1, 14, 0, 0, 0, 0, jst),
			wantErr: false,
		},
		{
			name:    "YYYY/MM/DD format",
			input:   "2025/01/14",
			want:    time.Date(2025, 1, 14, 0, 0, 0, 0, jst),
			wantErr: false,
		},
		{
			name:    "YYYY.MM.DD format",
			input:   "2025.01.14",
			want:    time.Date(2025, 1, 14, 0, 0, 0, 0, jst),
			wantErr: false,
		},
		// Date and time formats (local timezone)
		{
			name:    "ISO format without timezone with nanoseconds",
			input:   "2025-01-14T15:30:45.123456789",
			want:    time.Date(2025, 1, 14, 15, 30, 45, 123456789, jst),
			wantErr: false,
		},
		{
			name:    "ISO format without timezone",
			input:   "2025-01-14T15:30:45",
			want:    time.Date(2025, 1, 14, 15, 30, 45, 0, jst),
			wantErr: false,
		},
		{
			name:    "DateTime with space separator",
			input:   "2025-01-14 15:30:45",
			want:    time.Date(2025, 1, 14, 15, 30, 45, 0, jst),
			wantErr: false,
		},
		// Single digit month/day
		{
			name:    "YYYY/M/D format",
			input:   "2025/1/2",
			want:    time.Date(2025, 1, 2, 0, 0, 0, 0, jst),
			wantErr: false,
		},
		{
			name:    "YYYY.M.D format",
			input:   "2025.1.2",
			want:    time.Date(2025, 1, 2, 0, 0, 0, 0, jst),
			wantErr: false,
		},
		// Error cases
		{
			name:    "Invalid format",
			input:   "not a date",
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDateAndRest(t *testing.T) {
	// Set location to JST for testing
	jst, _ := time.LoadLocation("Asia/Tokyo")
	time.Local = jst

	tests := []struct {
		name     string
		input    string
		wantTime time.Time
		wantRest string
		wantErr  bool
	}{
		{
			name:     "Date at beginning",
			input:    "2025-01-14 豊田築炉 名和工場",
			wantTime: time.Date(2025, 1, 14, 0, 0, 0, 0, jst),
			wantRest: "豊田築炉 名和工場",
			wantErr:  false,
		},
		{
			name:     "Date in middle",
			input:    "豊田築炉 2025-01-14 名和工場",
			wantTime: time.Date(2025, 1, 14, 0, 0, 0, 0, jst),
			wantRest: "豊田築炉 名和工場",
			wantErr:  false,
		},
		{
			name:     "Date at end",
			input:    "豊田築炉 名和工場 2025-01-14",
			wantTime: time.Date(2025, 1, 14, 0, 0, 0, 0, jst),
			wantRest: "豊田築炉 名和工場",
			wantErr:  false,
		},
		{
			name:     "DateTime with RFC3339",
			input:    "Project 2025-01-14T15:30:45+09:00 Description",
			wantTime: time.Date(2025, 1, 14, 15, 30, 45, 0, time.FixedZone("", 9*3600)),
			wantRest: "Project Description",
			wantErr:  false,
		},
		{
			name:     "Only date string",
			input:    "2025-01-14",
			wantTime: time.Date(2025, 1, 14, 0, 0, 0, 0, jst),
			wantRest: "",
			wantErr:  false,
		},
		{
			name:     "Multiple spaces",
			input:    "  2025-01-14   豊田築炉   名和工場  ",
			wantTime: time.Date(2025, 1, 14, 0, 0, 0, 0, jst),
			wantRest: "豊田築炉   名和工場",
			wantErr:  false,
		},
		{
			name:     "No date in string",
			input:    "豊田築炉 名和工場",
			wantTime: time.Time{},
			wantRest: "豊田築炉 名和工場",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTime, gotRest, err := ParseDateAndRest(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateAndRest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !gotTime.Equal(tt.wantTime) {
				t.Errorf("ParseDateAndRest() time = %v, want %v", gotTime, tt.wantTime)
			}
			if gotRest != tt.wantRest {
				t.Errorf("ParseDateAndRest() rest = %v, want %v", gotRest, tt.wantRest)
			}
		})
	}
}

func TestParseRFC3339Nano(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "Valid RFC3339Nano",
			input:   "2025-01-14T15:30:45.123456789+09:00",
			want:    time.Date(2025, 1, 14, 15, 30, 45, 123456789, time.FixedZone("", 9*3600)),
			wantErr: false,
		},
		{
			name:    "RFC3339Nano in text",
			input:   "Created at 2025-01-14T15:30:45.123456789Z by user",
			want:    time.Date(2025, 1, 14, 15, 30, 45, 123456789, time.UTC),
			wantErr: false,
		},
		{
			name:    "No RFC3339Nano format",
			input:   "2025-01-14",
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "Invalid string",
			input:   "not a date",
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRFC3339Nano(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRFC3339Nano() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("ParseRFC3339Nano() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToFolderYAML(t *testing.T) {
	modTime := time.Date(2025, 1, 14, 15, 30, 45, 123456789, time.UTC)

	folder := &models.Folder{
		Id:           12345,
		Name:         "test-folder",
		Path:         "/path/to/folder",
		IsDirectory:  true,
		Size:         1024,
		ModifiedTime: modTime,
	}

	got := ConvertToFolderYAML(folder)

	expected := models.FolderYAML{
		Id:           12345,
		Name:         "test-folder",
		Path:         "/path/to/folder",
		IsDirectory:  true,
		Size:         1024,
		ModifiedTime: "2025-01-14T15:30:45.123456789Z",
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ConvertToFolderYAML() = %v, want %v", got, expected)
	}
}

func TestConvertToFolder(t *testing.T) {
	tests := []struct {
		name    string
		input   *models.FolderYAML
		want    models.Folder
		wantErr bool
	}{
		{
			name: "Valid conversion",
			input: &models.FolderYAML{
				Id:           12345,
				Name:         "test-folder",
				Path:         "/path/to/folder",
				IsDirectory:  true,
				Size:         1024,
				ModifiedTime: "2025-01-14T15:30:45.123456789Z",
			},
			want: models.Folder{
				Id:           12345,
				Name:         "test-folder",
				Path:         "/path/to/folder",
				IsDirectory:  true,
				Size:         1024,
				ModifiedTime: time.Date(2025, 1, 14, 15, 30, 45, 123456789, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "Invalid time format",
			input: &models.FolderYAML{
				Id:           12345,
				Name:         "test-folder",
				Path:         "/path/to/folder",
				IsDirectory:  true,
				Size:         1024,
				ModifiedTime: "invalid-time",
			},
			want:    models.Folder{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertToFolder(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToFolder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToFolder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToKoujiYAML(t *testing.T) {
	startDate := time.Date(2025, 1, 14, 9, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 20, 17, 0, 0, 0, time.UTC)
	modTime := time.Date(2025, 1, 14, 15, 30, 45, 123456789, time.UTC)

	kouji := &models.Kouji{
		Id:           "kouji-id",
		CompanyName:  "豊田築炉",
		LocationName: "名和工場",
		Status:       "進行中",
		StartDate:    startDate,
		EndDate:      endDate,
		Description:  "Test description",
		Tags:         []string{"tag1", "tag2"},
		FileCount:    10,
		SubdirCount:  3,
		Folder: models.Folder{
			Id:           67890,
			Name:         "2025-0114 豊田築炉 名和工場",
			Path:         "/path/to/kouji",
			IsDirectory:  true,
			Size:         2048,
			ModifiedTime: modTime,
		},
	}

	got := ConvertToKoujiYAML(kouji)

	expected := models.KoujiYAML{
		Id:           "kouji-id",
		CompanyName:  "豊田築炉",
		LocationName: "名和工場",
		Status:       "進行中",
		StartDate:    "2025-01-14T09:00:00Z",
		EndDate:      "2025-01-20T17:00:00Z",
		Description:  "Test description",
		Tags:         []string{"tag1", "tag2"},
		FileCount:    10,
		SubdirCount:  3,
		Folder: models.FolderYAML{
			Id:           67890,
			Name:         "2025-0114 豊田築炉 名和工場",
			Path:         "/path/to/kouji",
			IsDirectory:  true,
			Size:         2048,
			ModifiedTime: "2025-01-14T15:30:45.123456789Z",
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ConvertToKoujiYAML() = %v, want %v", got, expected)
	}
}

func TestConvertToKouji(t *testing.T) {
	tests := []struct {
		name    string
		input   *models.KoujiYAML
		want    models.Kouji
		wantErr bool
	}{
		{
			name: "Valid conversion",
			input: &models.KoujiYAML{
				Id:           "kouji-id",
				CompanyName:  "豊田築炉",
				LocationName: "名和工場",
				Status:       "進行中",
				StartDate:    "2025-01-14T09:00:00Z",
				EndDate:      "2025-01-20T17:00:00Z",
				Description:  "Test description",
				Tags:         []string{"tag1", "tag2"},
				FileCount:    10,
				SubdirCount:  3,
				Folder: models.FolderYAML{
					Id:           67890,
					Name:         "2025-0114 豊田築炉 名和工場",
					Path:         "/path/to/kouji",
					IsDirectory:  true,
					Size:         2048,
					ModifiedTime: "2025-01-14T15:30:45.123456789Z",
				},
			},
			want: models.Kouji{
				Id:           "kouji-id",
				CompanyName:  "豊田築炉",
				LocationName: "名和工場",
				Status:       "進行中",
				StartDate:    time.Date(2025, 1, 14, 9, 0, 0, 0, time.UTC),
				EndDate:      time.Date(2025, 1, 20, 17, 0, 0, 0, time.UTC),
				Description:  "Test description",
				Tags:         []string{"tag1", "tag2"},
				FileCount:    10,
				SubdirCount:  3,
				Folder: models.Folder{
					Id:           67890,
					Name:         "2025-0114 豊田築炉 名和工場",
					Path:         "/path/to/kouji",
					IsDirectory:  true,
					Size:         2048,
					ModifiedTime: time.Date(2025, 1, 14, 15, 30, 45, 123456789, time.UTC),
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid start date",
			input: &models.KoujiYAML{
				Id:           "kouji-id",
				CompanyName:  "豊田築炉",
				LocationName: "名和工場",
				Status:       "進行中",
				StartDate:    "invalid-date",
				EndDate:      "2025-01-20T17:00:00Z",
				Description:  "Test description",
				Tags:         []string{"tag1", "tag2"},
				FileCount:    10,
				SubdirCount:  3,
				Folder: models.FolderYAML{
					Id:           12345,
					Name:         "2025-0114 豊田築炉 名和工場",
					Path:         "/path/to/kouji",
					IsDirectory:  true,
					Size:         2048,
					ModifiedTime: "2025-01-14T15:30:45.123456789Z",
				},
			},
			want:    models.Kouji{},
			wantErr: true,
		},
		{
			name: "Invalid end date",
			input: &models.KoujiYAML{
				Id:           "kouji-id",
				CompanyName:  "豊田築炉",
				LocationName: "名和工場",
				Status:       "進行中",
				StartDate:    "2025-01-14T09:00:00Z",
				EndDate:      "invalid-date",
				Description:  "Test description",
				Tags:         []string{"tag1", "tag2"},
				FileCount:    10,
				SubdirCount:  3,
				Folder: models.FolderYAML{
					Id:           12345,
					Name:         "2025-0114 豊田築炉 名和工場",
					Path:         "/path/to/kouji",
					IsDirectory:  true,
					Size:         2048,
					ModifiedTime: "2025-01-14T15:30:45.123456789Z",
				},
			},
			want:    models.Kouji{},
			wantErr: true,
		},
		{
			name: "Invalid folder time",
			input: &models.KoujiYAML{
				Id:           "kouji-id",
				CompanyName:  "豊田築炉",
				LocationName: "名和工場",
				Status:       "進行中",
				StartDate:    "2025-01-14T09:00:00Z",
				EndDate:      "2025-01-20T17:00:00Z",
				Description:  "Test description",
				Tags:         []string{"tag1", "tag2"},
				FileCount:    10,
				SubdirCount:  3,
				Folder: models.FolderYAML{
					Id:           12345,
					Name:         "2025-0114 豊田築炉 名和工場",
					Path:         "/path/to/kouji",
					IsDirectory:  true,
					Size:         2048,
					ModifiedTime: "invalid-time",
				},
			},
			want:    models.Kouji{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertToKouji(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToKouji() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToKouji() = %v, want %v", got, tt.want)
			}
		})
	}
}
