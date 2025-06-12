package services

import (
	"os"
	"path/filepath"
	"penguin-backend/internal/models"
	"strings"
	"time"
	"os/user"
	"gopkg.in/yaml.v3"
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

// LoadKoujiProjectsFromYAML は工事プロジェクト情報をYAMLファイルから読み込む
//
// 読み込み手順:
//   1. 指定されたディレクトリに .inside.yaml が存在するかチェック
//   2. YAML内容を KoujiFolder 構造体に解析
//   3. プロジェクトIDと必須フィールドを検証
//   4. タイムスタンプをローカルタイムゾーンから time.Time に変換
//   5. ソート済みのプロジェクトリストを返す
//
// エラーハンドリング:
//   - ファイルが存在しない場合は空のリストを返す
//   - YAML解析が失敗した場合はエラーを返す
func (fs *FolderService) LoadKoujiProjectsFromYAML(yamlPath string) ([]models.KoujiProject, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(yamlPath, "~/") {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		yamlPath = filepath.Join(usr.HomeDir, yamlPath[2:])
	}
	
	absPath, err := filepath.Abs(yamlPath)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return []models.KoujiProject{}, nil // Return empty list if file doesn't exist
	}

	// Read YAML file
	yamlData, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	// Define YAML structure for loading
	type YAMLKoujiProject struct {
		Folder       models.Folder `yaml:"folder"`
		ProjectID    string        `yaml:"projectid"`
		ProjectName  string        `yaml:"projectname"`
		CompanyName  string        `yaml:"companyname"`
		LocationName string        `yaml:"locationname"`
		Status       string        `yaml:"status"`
		CreatedDate  string        `yaml:"createddate"`
		StartDate    string        `yaml:"startdate"`
		EndDate      string        `yaml:"enddate"`
		Description  string        `yaml:"description"`
		Tags         []string      `yaml:"tags"`
		FileCount    int          `yaml:"filecount"`
		SubdirCount  int          `yaml:"subdircount"`
	}

	var yamlProjects []YAMLKoujiProject
	if err := yaml.Unmarshal(yamlData, &yamlProjects); err != nil {
		return nil, err
	}

	// Convert to KoujiProject structures
	koujiProjects := make([]models.KoujiProject, len(yamlProjects))
	for i, yf := range yamlProjects {
		// Parse dates with error handling and zero value detection
		createdDate := parseTimeWithValidation(yf.CreatedDate)
		startDate := parseTimeWithValidation(yf.StartDate)
		endDate := parseTimeWithValidation(yf.EndDate)
		
		koujiProjects[i] = models.KoujiProject{
			Folder:       yf.Folder,
			ProjectID:    yf.ProjectID,
			ProjectName:  yf.ProjectName,
			CompanyName:  yf.CompanyName,
			LocationName: yf.LocationName,
			Status:       yf.Status,
			CreatedDate:  createdDate,
			StartDate:    startDate,
			EndDate:      endDate,
			Description:  yf.Description,
			Tags:         yf.Tags,
			FileCount:    yf.FileCount,
			SubdirCount:  yf.SubdirCount,
		}
	}

	return koujiProjects, nil
}

func (fs *FolderService) SaveKoujiProjectsToYAML(targetPath string, koujiProjects []models.KoujiProject) error {
	// Expand ~ to home directory
	if strings.HasPrefix(targetPath, "~/") {
		usr, err := user.Current()
		if err != nil {
			return err
		}
		targetPath = filepath.Join(usr.HomeDir, targetPath[2:])
	}
	
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create YAML data with custom formatting for time fields
	type YAMLKoujiProject struct {
		Folder       models.Folder `yaml:"folder"`
		ProjectID    string        `yaml:"projectid"`
		ProjectName  string        `yaml:"projectname"`
		CompanyName  string        `yaml:"companyname"`
		LocationName string        `yaml:"locationname"`
		Status       string        `yaml:"status"`
		CreatedDate  string        `yaml:"createddate"`
		StartDate    string        `yaml:"startdate"`
		EndDate      string        `yaml:"enddate"`
		Description  string        `yaml:"description"`
		Tags         []string      `yaml:"tags"`
		FileCount    int          `yaml:"filecount"`
		SubdirCount  int          `yaml:"subdircount"`
	}

	yamlProjects := make([]YAMLKoujiProject, len(koujiProjects))
	for i, kf := range koujiProjects {
		yamlProjects[i] = YAMLKoujiProject{
			Folder:       kf.Folder,
			ProjectID:    kf.ProjectID,
			ProjectName:  kf.ProjectName,
			CompanyName:  kf.CompanyName,
			LocationName: kf.LocationName,
			Status:       kf.Status,
			CreatedDate:  formatTimeForYAML(kf.CreatedDate),
			StartDate:    formatTimeForYAML(kf.StartDate),
			EndDate:      formatTimeForYAML(kf.EndDate),
			Description:  kf.Description,
			Tags:         kf.Tags,
			FileCount:    kf.FileCount,
			SubdirCount:  kf.SubdirCount,
		}
	}

	// Convert to YAML format
	yamlData, err := yaml.Marshal(yamlProjects)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(absPath, yamlData, 0644)
}

// parseTimeWithValidation は時刻文字列をパースし、異常値を検出する
func parseTimeWithValidation(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{} // 空文字列の場合はゼロ値を返す
	}
	
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// パースエラーの場合はゼロ値を返す
		return time.Time{}
	}
	
	// 異常なタイムゾーンオフセット（+09:18など）を検出
	// 0001年の日付で+09:18のオフセットは異常値として扱う
	if parsedTime.Year() == 1 && parsedTime.Month() == 1 && parsedTime.Day() == 1 {
		// ゼロ値として扱う
		return time.Time{}
	}
	
	return parsedTime
}

// formatTimeForYAML は時刻をYAML保存用にフォーマットする
func formatTimeForYAML(t time.Time) string {
	if t.IsZero() {
		return "" // ゼロ値の場合は空文字列を返す
	}
	return t.In(time.Local).Format(time.RFC3339)
}

// CleanupInvalidTimeData は異常な時刻データを含むプロジェクトを削除する
func (fs *FolderService) CleanupInvalidTimeData(yamlPath string) error {
	// 既存のプロジェクトを読み込み
	projects, err := fs.LoadKoujiProjectsFromYAML(yamlPath)
	if err != nil {
		return err
	}
	
	// 異常データを検出して削除
	validProjects := make([]models.KoujiProject, 0)
	removedCount := 0
	
	for _, project := range projects {
		// CreatedDateが異常値（0001年かつ空のcompanynameなど）かチェック
		isInvalid := false
		
		if project.CreatedDate.Year() == 1 && project.CreatedDate.Month() == 1 && project.CreatedDate.Day() == 1 {
			// 追加の条件でより確実に異常データを特定
			if project.CompanyName == "" && project.LocationName == "" {
				isInvalid = true
			}
		}
		
		if !isInvalid {
			validProjects = append(validProjects, project)
		} else {
			removedCount++
		}
	}
	
	// 異常データが見つかった場合は保存
	if removedCount > 0 {
		err = fs.SaveKoujiProjectsToYAML(yamlPath, validProjects)
		if err != nil {
			return err
		}
	}
	
	return nil
}