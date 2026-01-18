package core

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// ResolveAbsPath は絶対パスを最短パスに変換します。
// absPath は絶対パスです。
// '~' はホームディレクトリを展開して絶対パスに変換します。
// ディレクトリディスクリプタは'/'に変換されます。
// シンボリックリンクは解決されます。
// 絶対パスでない場合はエラーを返します。
func ResolveAbsPath(absPath string) (string, error) {
	// ディレクトリディスクリプタを'/'に変換

	// ホームディレクトリに展開
	if strings.HasPrefix(absPath, "~/") || strings.HasPrefix(absPath, "~\\") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		absPath = filepath.Join(usr.HomeDir, absPath[2:])
	}

	// シンボリックリンクを解決して絶対パスチェック
	absPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", err
	}

	// 絶対パスチェック
	if !filepath.IsAbs(absPath) {
		return "", errors.New("絶対パスの条件を満たしていません。")
	}

	absPath = strings.ReplaceAll(absPath, "\\", "/")
	return absPath, nil
}

// パスからファイル名またはフォルダー名を取得します
func GetBaseName(pathname string) string {
	basename := filepath.Base(pathname)
	if basename == "." || basename == "/" || basename == "\\" {
		return ""
	}
	return basename
}

// IsExcel エクセルファイルかどうかをチェック
func FilenameIsExcel(filename string) bool {
	excelSuffix := []string{".xlsx", ".xls"}
	nameLower := strings.ToLower(filename)

	for _, suffix := range excelSuffix {
		if strings.HasSuffix(nameLower, suffix) {
			return true
		}
	}
	return false
}

// EntryIsExcelFile エクセルファイルかどうかをチェック
func EntryIsExcelFile(entry os.DirEntry) bool {
	return FilenameIsExcel(entry.Name())
}
