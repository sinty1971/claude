package core

import (
	"os"
	"reflect"
	"runtime"
	"strings"
)

// コンフィグレーションの型定義します。
type Configuration struct {
	DirectoryBaseDirPath           string
	CompanyBaseDirPath             string
	CompanyWatcherMaxDepth         int
	CompanyCategoryBaseDirPath     string
	CompanyCategoryWatcherMaxDepth int
	KojiBaseDirPath                string
	KojiWatcherMaxDepth            int
	MemberBaseDirPath              string
	MemberWatcherMaxDepth          int
	MinimumWorkers                 int
	MaximumWorkers                 int
	CpuMultiplier                  int

	// OCRServerExecutablePath は OCR サービスの実行可能ファイルのパス
	OcrServerExecutablePath string

	// PdfToImageExecutablePath は PDF を画像に変換する実行可能ファイルのパス
	PdfToImageExecutablePath string
	// OcrLanguage は OCR 実行時に渡す言語指定（例: "japan", "ch", "en"）
	OcrLanguage string
}

// Config はアプリケーション全体で使用される設定オブジェクトです。
var Config *Configuration

// EnvironmentMap はアプリケーション全体で使用される環境変数マップ
var EnvironmentMap map[string]string

func init() {
	Config = &Configuration{}

	// すべての環境変数を map に格納
	EnvironmentMap = make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			EnvironmentMap[pair[0]] = pair[1]
		}
	}

	// カスタム変数を追加: ホスト名に基づくルートパス
	root := ""

	// PCのホスト名の取得
	hostname, err := os.Hostname()
	if err != nil {
		return
	}

	switch hostname {
	case "DESKTOP-HHR7FT6":
		root = "C:/SyncFolder/SynologyDrive/豊田築炉"
	case "SINTY-OMEN":
		root = "O:"
	}
	EnvironmentMap["ROOT"] = root

	Config.ExpandPlaceholders()
}

// ExpandPlaceholders は設定内の {KEY} 形式のプレースホルダーを環境変数や設定値で展開します
func (c *Configuration) ExpandPlaceholders() {
	// リフレクションを使って全ての文字列フィールドを処理
	v := reflect.ValueOf(c).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String && field.CanSet() {
			str := field.String()
			// {KEY} 形式のプレースホルダーを展開
			for key, value := range EnvironmentMap {
				str = strings.ReplaceAll(str, "{"+key+"}", value)
			}
			field.SetString(str)
		}
	}
}

// CalculateWorkerCount は要素数とシステムリソースに基づいて最適なワーカー数を決定する
// itemCount: 処理する要素数
func (c *Configuration) CalculateWorkerCount(itemCount int) int {

	// 要素数が0の場合は0を返す
	if itemCount <= 0 {
		return 0
	}

	// CPU数ベースでワーカー数を計算
	numCPU := runtime.NumCPU()
	idealWorkers := numCPU * c.CpuMultiplier

	// 最小値と最大値の範囲内に収める
	numWorkers := max(
		c.MinimumWorkers,
		min(c.MaximumWorkers, idealWorkers))

	// 要素数が少ない場合はワーカー数を調整
	if itemCount < numWorkers {
		// 要素数の半分程度のワーカー数にする（最小1）
		numWorkers = max(1, itemCount/2)
	}

	return numWorkers
}
