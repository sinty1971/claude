package core

import (
	"os"
	"reflect"
	"strings"
)

// 環境設定を定義します。
var Config = struct {
	FileServiceTarget          string
	CompanyTargetFolder        string
	CompanyPersistFilename     string
	CompanyWatcherMaxDepth     int
	CompanyPollIntervalMillSec int
	KojiTargetFolder           string
	KojiPersistFilename        string
	MemberPersistFilename      string
	PersistDBPath              string
	MinumWorkers               int
	MaximumWorkers             int
	CpuMultiplier              int
}{
	FileServiceTarget:          "{ROOT}",
	CompanyTargetFolder:        "{ROOT}/1 会社",
	CompanyPersistFilename:     "@company.yaml",
	CompanyWatcherMaxDepth:     2,
	CompanyPollIntervalMillSec: 3000,
	KojiTargetFolder:           "{ROOT}/2 工事",
	KojiPersistFilename:        "@koji.yaml",
	MemberPersistFilename:      "@member.yaml",
	PersistDBPath:              "{USERPROFILE}/.persist/@persist.db",
	MinumWorkers:               2,
	MaximumWorkers:             16,
	CpuMultiplier:              2,
}

func init() {
	ParseConfigValue()
}

func ParseConfigValue() {
	// PCのホスト名の取得
	hostname, err := os.Hostname()
	if err != nil {
		return
	}

	// {USERPROFILE}: ユーザープロファイルに基づく設定の上書き
	userProfile := os.Getenv("USERPROFILE")

	// {ROOT}: ホスト名に基づく設定の上書き
	root := ""
	switch hostname {
	case "DESKTOP-HHR7FT6":
		root = "C:/SyncFolder/SynologyDrive/豊田築炉"
	case "SINTY-OMEN":
		root = "O:/"
	}

	// リフレクションを使って全ての文字列フィールドを処理
	v := reflect.ValueOf(&Config).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String && field.CanSet() {
			str := field.String()
			str = strings.ReplaceAll(str, "{ROOT}", root)
			str = strings.ReplaceAll(str, "{USERPROFILE}", userProfile)
			field.SetString(str)
		}
	}
}
