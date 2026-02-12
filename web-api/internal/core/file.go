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
		absPath = PathJoin(usr.HomeDir, absPath[2:])
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

// パスからディレクトリ名を取得します
func PathDir(pathname string) string {
	dirname := filepath.Dir(pathname)
	return strings.ReplaceAll(dirname, "\\", "/")
}

// パスからファイル名またはフォルダー名を取得します
func PathBase(pathname string) string {
	basename := filepath.Base(pathname)
	if basename == "." || basename == "/" || basename == "\\" {
		return ""
	}
	return basename
}

// PathJoin は filepath.Join の代わりに常に '/' を区切り文字として使う関数
func PathJoin(elem ...string) string {
	path := filepath.Join(elem...)
	return strings.ReplaceAll(path, "\\", "/")
}

// PathSplit はパスを '/' で分割する関数
func PathSplit(path string) []string {
	// まず \\ を / に統一
	path = strings.ReplaceAll(path, "\\", "/")
	return strings.Split(path, "/")
}

// PathIsExcel エクセルファイルかどうかをチェック
func PathIsExcel(path string) bool {
	excelSuffix := []string{".xlsx", ".xls"}
	nameLower := strings.ToLower(path)

	for _, suffix := range excelSuffix {
		if strings.HasSuffix(nameLower, suffix) {
			return true
		}
	}
	return false
}

// EntryIsExcelFile エクセルファイルかどうかをチェック
func EntryIsExcelFile(entry os.DirEntry) bool {
	return PathIsExcel(entry.Name())
}
