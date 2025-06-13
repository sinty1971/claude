package services

import (
	"bytes"
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

type FileSystemService struct{}

func NewFileSystemService() *FileSystemService {
	return &FileSystemService{}
}

func (fss *FileSystemService) GetFolders(targetPath string) (*models.FolderListResponse, error) {
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
			Id:           stat.Ino,
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
//  1. 指定されたディレクトリに .inside.yaml が存在するかチェック
//  2. YAML内容を KoujiFolder 構造体に解析
//  3. プロジェクトIDと必須フィールドを検証
//  4. タイムスタンプをローカルタイムゾーンから time.Time に変換
//  5. ソート済みのプロジェクトリストを返す
//
// エラーハンドリング:
//   - ファイルが存在しない場合は空のリストを返す
//   - YAML解析が失敗した場合はエラーを返す
func (fss *FileSystemService) LoadKoujiProjectsFromYAML(databasePath string) ([]models.Kouji, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(databasePath, "~/") {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		databasePath = filepath.Join(usr.HomeDir, databasePath[2:])
	}

	absPath, err := filepath.Abs(databasePath)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return []models.Kouji{}, nil // Return empty list if file doesn't exist
	}

	// Read YAML file
	yamlData, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	var dbProjects []models.Kouji
	if err := yaml.Unmarshal(yamlData, &dbProjects); err != nil {
		return nil, err
	}

	// Convert to KoujiProject structures
	fsProjects := make([]models.Kouji, len(dbProjects))
	for i, yf := range dbProjects {
		// 日付データを正規化
		startDate, err := parseTimeWithValidation(yf.StartDate)
		if err != nil {
			return nil, err
		}
		endDate, err := parseTimeWithValidation(yf.EndDate)
		if err != nil {
			return nil, err
		}

		fsProjects[i] = models.Kouji{
			Folder:       yf.Folder,
			Id:           yf.Id,
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

	return fsProjects, nil
}

// @TODO: この関数は内部で使用するだけで、外部に公開する必要はない
// @Description: 工事プロジェクト情報をYAMLファイルに保存する
// @param targetPath: 保存先パス
// @param koujiProjects: 保存する工事プロジェクト情報
// @return error: エラー
func (fs *FileSystemService) SaveKoujiProjectsToYAML(targetPath string, koujiProjects []models.Kouji) error {
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
	type DatabaseKoujiProject struct {
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
		Folder       models.Folder `yaml:"folder"`
	}

	insideProjects := make([]DatabaseKoujiProject, len(koujiProjects))
	for i, kf := range koujiProjects {
		insideProjects[i] = DatabaseKoujiProject{
			Folder:       kf.Folder,
			ProjectID:    kf.Id,
			ProjectName:  kf.KoujiName,
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
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(insideProjects); err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(absPath, buf.Bytes(), 0644)
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
