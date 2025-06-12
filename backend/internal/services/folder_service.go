package services

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"penguin-backend/internal/models"
	"strings"
	"syscall"
	"time"

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
		stat := info.Sys().(*syscall.Stat_t)

		folder := models.Folder{
			Name:         entry.Name(),
			Path:         filepath.Join(absPath, entry.Name()),
			IsDirectory:  entry.IsDir(),
			Size:         info.Size(),
			CreatedDate:  time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec),
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
//  1. 指定されたディレクトリに .inside.yaml が存在するかチェック
//  2. YAML内容を KoujiFolder 構造体に解析
//  3. プロジェクトIDと必須フィールドを検証
//  4. タイムスタンプをローカルタイムゾーンから time.Time に変換
//  5. ソート済みのプロジェクトリストを返す
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
		StartDate    string        `yaml:"startdate"`
		EndDate      string        `yaml:"enddate"`
		Description  string        `yaml:"description"`
		Tags         []string      `yaml:"tags"`
		FileCount    int           `yaml:"filecount"`
		SubdirCount  int           `yaml:"subdircount"`
	}

	var yamlProjects []YAMLKoujiProject
	if err := yaml.Unmarshal(yamlData, &yamlProjects); err != nil {
		return nil, err
	}

	// Convert to KoujiProject structures
	koujiProjects := make([]models.KoujiProject, len(yamlProjects))
	for i, yf := range yamlProjects {
		// Parse dates with error handling and zero value detection
		startDate, err := parseTimeWithValidation(yf.StartDate)
		if err != nil {
			return nil, err
		}
		endDate, err := parseTimeWithValidation(yf.EndDate)
		if err != nil {
			return nil, err
		}

		koujiProjects[i] = models.KoujiProject{
			Folder:       yf.Folder,
			ProjectID:    yf.ProjectID,
			ProjectName:  yf.ProjectName,
			CompanyName:  yf.CompanyName,
			LocationName: yf.LocationName,
			Status:       yf.Status,
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

// @TODO: この関数は内部で使用するだけで、外部に公開する必要はない
// @Description: 工事プロジェクト情報をYAMLファイルに保存する
// @param targetPath: 保存先パス
// @param koujiProjects: 保存する工事プロジェクト情報
// @return error: エラー
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

	// Create Inside data with custom formatting for time fields
	type InsideKoujiProject struct {
		Folder       models.Folder `yaml:"folder"`
		ProjectID    string        `yaml:"projectid"`
		ProjectName  string        `yaml:"projectname"`
		CompanyName  string        `yaml:"companyname"`
		LocationName string        `yaml:"locationname"`
		Status       string        `yaml:"status"`
		StartDate    string        `yaml:"startdate"`
		EndDate      string        `yaml:"enddate"`
		Description  string        `yaml:"description"`
		Tags         []string      `yaml:"tags"`
		FileCount    int           `yaml:"filecount"`
		SubdirCount  int           `yaml:"subdircount"`
	}

	insideProjects := make([]InsideKoujiProject, len(koujiProjects))
	for i, kf := range koujiProjects {
		insideProjects[i] = InsideKoujiProject{
			Folder:       kf.Folder,
			ProjectID:    kf.ProjectID,
			ProjectName:  kf.ProjectName,
			CompanyName:  kf.CompanyName,
			LocationName: kf.LocationName,
			Status:       kf.Status,
			StartDate:    formatTimeForYAML(kf.StartDate),
			EndDate:      formatTimeForYAML(kf.EndDate),
			Description:  kf.Description,
			Tags:         kf.Tags,
			FileCount:    kf.FileCount,
			SubdirCount:  kf.SubdirCount,
		}
	}

	// Convert to YAML format
	yamlData, err := yaml.Marshal(insideProjects)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(absPath, yamlData, 0644)
}

// parseTimeWithValidation は時刻文字列をパースし、異常値を検出する
// @param timeStr: パースする時刻文字列
// @return time.Time: パースされた時刻
func parseTimeWithValidation(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, errors.New("timeStr is empty") // 空文字列の場合はエラーを返す
	}

	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// パースエラーの場合はゼロ値を返す
		return time.Time{}, err
	}

	// 異常値を検出する包括的なチェック
	if isInvalidParsedTime(parsedTime) {
		return time.Time{}, errors.New("invalid parsed time")
	}

	return parsedTime, nil
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
		// より包括的な異常値チェック
		isInvalid := isInvalidProject(project)

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

// isInvalidParsedTime は解析済み時刻が異常値かどうかをチェックする
func isInvalidParsedTime(t time.Time) bool {
	// ゼロ値
	if t.IsZero() {
		return true
	}

	// 0001年（Go言語のゼロ値に近い異常値）
	if t.Year() == 1 {
		return true
	}

	// 不合理な将来日付（1年以上先）
	if t.After(time.Now().AddDate(1, 0, 0)) {
		return true
	}

	// 不合理な過去日付（2000年より前）
	if t.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return true
	}

	return false
}

// isInvalidProject は工事プロジェクトが異常データかどうかをチェックする
func isInvalidProject(project models.KoujiProject) bool {
	// CreatedDateの異常値チェック
	if isInvalidParsedTime(project.CreatedDate) {
		return true
	}

	// 0001年の特定パターン（"0001-01-01T09:26:51+09:18"に近い値）
	if project.CreatedDate.Year() == 1 {
		// 会社名や場所名が空の場合は明らかに異常
		if project.CompanyName == "" && project.LocationName == "" {
			return true
		}
		// プロジェクトIDが空の場合も異常
		if project.ProjectID == "" {
			return true
		}
	}

	return false
}
