package core

import (
	"os"
	"strings"
)

// サーバーの環境設定を定義します。
var ServerConfiguration = map[string]string{
	"FileServiceTarget":          "{ROOT}",
	"CompanyListFolder":          "{ROOT}/1 会社",
	"CompanyPersistFilename":     "@company.yaml",
	"CompanyPollIntervalMillSec": "3000",
	"KojiListFolder":             "{ROOT}/2 工事",
	"KojiPersistFilename":        "@koji.yaml",
	"MemberPersistFilename":      "@member.yaml",
	"PersistDBPath":              "{USERPROFILE}/.persist/@persist.db",
}

// ワーカーの環境設定を定義します。
var WorkerConfiguration = map[string]int{
	"MinumWorkers":   2,
	"MaximumWorkers": 16,
	"CpuMultiplier":  2,
}

func init() {
	ParseConfiguration()
}

func ParseConfiguration() {
	// PCのホスト名の取得
	hostname, err := os.Hostname()
	if err != nil {
		return
	}
	// {ROOT}: ホスト名に基づく設定の上書き
	root := ""
	switch hostname {
	case "DESKTOP-HHR7FT6":
		root = "C:/SyncFolder/SynologyDrive/豊田築炉"
	case "SINTY-OMEN":
		root = "O:/"
	}
	for key, value := range ServerConfiguration {
		ServerConfiguration[key] = strings.ReplaceAll(value, "{ROOT}", root)
	}
	// {USERPROFILE}: ユーザープロファイルに基づく設定の上書き
	userProfile := os.Getenv("USERPROFILE")
	for key, value := range ServerConfiguration {
		ServerConfiguration[key] = strings.ReplaceAll(value, "{USERPROFILE}", userProfile)
	}
}
