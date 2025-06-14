package utils

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"

	"penguin-backend/internal/models"
)

// Formats and regexps with timezone information (try these first)
var DefaultRegexpsWithTZ = make(map[string]regexp.Regexp, len(DefaultFormatsWithTZ))
var DefaultFormatsWithTZ = []string{
	time.RFC3339Nano,
	time.RFC3339,
}

// Formats and regexps without timezone information (use local timezone)
var DefaultRegexpsWithoutTZ = make(map[string]regexp.Regexp, len(DefaultFormatsWithoutTZ))
var DefaultFormatsWithoutTZ = []string{
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-0102",
	"2006-01-02",
	"20060102",
	"2006/01/02",
	"2006.01.02",
	"2006/1/2",
	"2006.1.2",
}

// 置換ルール定義
var DateParseReplaceRule = map[string]string{
	"2006":      `\d{4}`,                 // 年
	"01":        `\d{2}`,                 // 月
	"02":        `\d{2}`,                 // 日
	"15":        `\d{2}`,                 // 時（24時間）
	"04":        `\d{2}`,                 // 分
	"05":        `\d{2}`,                 // 秒
	"999999999": `\d{9}`,                 // ナノ秒
	"Z07:00":    `(?:Z|[+-]\d{2}:\d{2})`, // タイムゾーン
	"Z0700":     `(?:Z|[+-]\d{4})`,       // タイムゾーン（コロンなし）
	"Z07":       `(?:Z|[+-]\d{2})`,       // タイムゾーン（時間のみ）
}

// init initializes the default patterns
func init() {
	formatToRegex := func(format string) regexp.Regexp {
		// 順序を考慮して置換を実行
		pattern := format

		// DateParseReplaceRuleからキーを取得し、文字数順でソート
		keys := slices.Collect(maps.Keys(DateParseReplaceRule))
		slices.SortFunc(keys, func(a, b string) int {
			return len(b) - len(a) // 文字数の大きい順
		})

		// 順序に従って置換
		for _, key := range keys {
			pattern = strings.Replace(pattern, key, DateParseReplaceRule[key], -1)
		}

		return *regexp.MustCompile(pattern)
	}

	// Initialize patterns for formats with timezone
	for _, format := range DefaultFormatsWithTZ {
		DefaultRegexpsWithTZ[format] = formatToRegex(format)
	}

	// Initialize patterns for formats without timezone
	for _, format := range DefaultFormatsWithoutTZ {
		DefaultRegexpsWithoutTZ[format] = formatToRegex(format)
	}
}

// Parse parses various date/time string formats and returns a time.Time
// When no timezone is specified, it uses the server's local timezone
func Parse(s string) (time.Time, error) {
	t, _, err := ParseDateAndRest(s)
	return t, err
}

// 日時文字列の抽出と、日時文字列から日付文字列をどり除いた文字列を返す
func findAndRest(re *regexp.Regexp, s string) (*string, *string) {
	matches := re.FindStringIndex(s)
	if matches == nil {
		return nil, nil
	}
	// 日時部分を抽出
	dateStr := s[matches[0]:matches[1]]
	// 日時の前の部分を取得
	prefix := strings.TrimSpace(s[:matches[0]])
	// 日時の後の部分を取得
	suffix := strings.TrimSpace(s[matches[1]:])
	// 日時の前の部分と後の部分を結合（間にスペースを追加）
	var restStr string
	if prefix != "" && suffix != "" {
		restStr = prefix + " " + suffix
	} else {
		restStr = prefix + suffix
	}

	return &dateStr, &restStr
}

// 文字列をパースし、戻り値はtime.Timeと、日付文字列から日付文字列をどり除いた文字列を返す
func ParseDateAndRest(s string) (time.Time, string, error) {
	// タイムゾーン付きのフォーマットを試行（配列順序で）
	for _, format := range DefaultFormatsWithTZ {
		re := DefaultRegexpsWithTZ[format]
		dateStr, restStr := findAndRest(&re, s)
		if dateStr == nil {
			continue
		}
		if t, err := time.Parse(format, *dateStr); err == nil {
			return t, *restStr, nil
		}
	}

	// タイムゾーンなしのフォーマットを試行（配列順序で）
	for _, format := range DefaultFormatsWithoutTZ {
		re := DefaultRegexpsWithoutTZ[format]
		dateStr, restStr := findAndRest(&re, s)
		if dateStr == nil {
			continue
		}

		if t, err := time.ParseInLocation(format, *dateStr, time.Local); err == nil {
			return t, *restStr, nil
		}
	}

	return time.Time{}, s, fmt.Errorf("unable to parse date/time in the string: %s", s)
}

// RFC3339Nano で文字列をパースする
func ParseRFC3339Nano(s string) (time.Time, error) {
	re := DefaultRegexpsWithTZ[time.RFC3339Nano]
	dateStr, _ := findAndRest(&re, s)
	if dateStr == nil {
		return time.Time{}, fmt.Errorf("unable to parse date/time in the string: %s", s)
	}
	return time.Parse(time.RFC3339Nano, *dateStr)
}

// FolderデータをFolderYAMLに変換する
func ConvertToFolderYAML(folder *models.Folder) models.FolderYAML {
	modifiedTime := folder.ModifiedTime.Format(time.RFC3339Nano)

	return models.FolderYAML{
		Id:           folder.Id,
		Name:         folder.Name,
		Path:         folder.Path,
		IsDirectory:  folder.IsDirectory,
		Size:         folder.Size,
		ModifiedTime: modifiedTime,
	}
}

// FolderYAMLデータをFolderに変換する
func ConvertToFolder(folderYAML *models.FolderYAML) (models.Folder, error) {
	modifiedTime, err := time.Parse(time.RFC3339Nano, folderYAML.ModifiedTime)
	if err != nil {
		return models.Folder{}, err
	}

	return models.Folder{
		Id:           folderYAML.Id,
		Name:         folderYAML.Name,
		Path:         folderYAML.Path,
		IsDirectory:  folderYAML.IsDirectory,
		Size:         folderYAML.Size,
		ModifiedTime: modifiedTime,
	}, nil
}

// KoujiデータをKoujiYAMLに変換する
func ConvertToKoujiYAML(kouji *models.Kouji) models.KoujiYAML {
	folderYAML := ConvertToFolderYAML(&kouji.Folder)
	startDate := kouji.StartDate.Format(time.RFC3339Nano)
	endDate := kouji.EndDate.Format(time.RFC3339Nano)

	return models.KoujiYAML{
		Id:           kouji.Id,
		CompanyName:  kouji.CompanyName,
		LocationName: kouji.LocationName,
		Status:       kouji.Status,
		StartDate:    startDate,
		EndDate:      endDate,
		Description:  kouji.Description,
		Tags:         kouji.Tags,
		FileCount:    kouji.FileCount,
		SubdirCount:  kouji.SubdirCount,
		Folder:       folderYAML,
	}
}

// KoujiYAMLデータをKoujiに変換する
func ConvertToKouji(koujiYAML *models.KoujiYAML) (models.Kouji, error) {
	startDate, err := time.Parse(time.RFC3339Nano, koujiYAML.StartDate)
	if err != nil {
		return models.Kouji{}, err
	}
	endDate, err := time.Parse(time.RFC3339Nano, koujiYAML.EndDate)
	if err != nil {
		return models.Kouji{}, err
	}
	folder, err := ConvertToFolder(&koujiYAML.Folder)
	if err != nil {
		return models.Kouji{}, err
	}

	return models.Kouji{
		Id:           koujiYAML.Id,
		CompanyName:  koujiYAML.CompanyName,
		LocationName: koujiYAML.LocationName,
		Status:       koujiYAML.Status,
		StartDate:    startDate,
		EndDate:      endDate,
		Description:  koujiYAML.Description,
		Tags:         koujiYAML.Tags,
		FileCount:    koujiYAML.FileCount,
		SubdirCount:  koujiYAML.SubdirCount,
		Folder:       folder,
	}, nil
}
