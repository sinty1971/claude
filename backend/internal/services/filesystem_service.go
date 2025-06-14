package services

import (
	"bytes"
	"os"
	"os/user"
	"path/filepath"
	"penguin-backend/internal/models"
	"penguin-backend/internal/utils"
	"strings"
	"syscall"

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

		folder := models.Folder{
			Id:           stat.Ino,
			Name:         entry.Name(),
			Path:         entryPath,
			IsDirectory:  isDirectory,
			Size:         info.Size(),
			ModifiedTime: info.ModTime(),
		}
		folders = append(folders, folder)
	}

	return &models.FolderListResponse{
		Folders: folders,
		Count:   len(folders),
	}, nil
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

	var dbKoujiYamlList []models.KoujiYAML
	if err := yaml.Unmarshal(yamlData, &dbKoujiYamlList); err != nil {
		return nil, err
	}

	// Convert to KoujiProject structures
	dbKoujiList := make([]models.Kouji, len(dbKoujiYamlList))
	for i, yaml := range dbKoujiYamlList {
		kouji, err := utils.ConvertToKouji(&yaml)
		if err != nil {
			return nil, err
		}
		dbKoujiList[i] = kouji
	}

	return dbKoujiList, nil
}

// @TODO: この関数は内部で使用するだけで、外部に公開する必要はない
// @Description: 工事プロジェクト情報をYAMLファイルに保存する
// @param targetPath: 保存先パス
// @param koujiProjects: 保存する工事プロジェクト情報
// @return error: エラー
func (fs *FileSystemService) SaveKoujiListToYAML(targetPath string, koujilist []models.Kouji) error {
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

	koujiYamlList := make([]models.KoujiYAML, len(koujilist))
	for i, kf := range koujilist {
		koujiYamlList[i] = utils.ConvertToKoujiYAML(&kf)
	}

	// Convert to YAML format
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(koujiYamlList); err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(absPath, buf.Bytes(), 0644)
}
