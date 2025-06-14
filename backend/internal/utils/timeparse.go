package utils

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"
)

// Formats and regexps with timezone information (try these first)
var DefaultTimestampRegexpsWithTZ = make(map[string]regexp.Regexp, len(DefaultTimestampFormatsWithTZ))
var DefaultTimestampFormatsWithTZ = []string{
	time.RFC3339Nano,
	time.RFC3339,
}

// Formats and regexps without timezone information (use local timezone)
var DefaultTimestampRegexpsWithoutTZ = make(map[string]regexp.Regexp, len(DefaultTimestampFormatsWithoutTZ))
var DefaultTimestampFormatsWithoutTZ = []string{
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
var TimestampParseReplaceRule = map[string]string{
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
		keys := slices.Collect(maps.Keys(TimestampParseReplaceRule))
		slices.SortFunc(keys, func(a, b string) int {
			return len(b) - len(a) // 文字数の大きい順
		})

		// 順序に従って置換
		for _, key := range keys {
			pattern = strings.Replace(pattern, key, TimestampParseReplaceRule[key], -1)
		}

		return *regexp.MustCompile(pattern)
	}

	// Initialize patterns for formats with timezone
	for _, format := range DefaultTimestampFormatsWithTZ {
		DefaultTimestampRegexpsWithTZ[format] = formatToRegex(format)
	}

	// Initialize patterns for formats without timezone
	for _, format := range DefaultTimestampFormatsWithoutTZ {
		DefaultTimestampRegexpsWithoutTZ[format] = formatToRegex(format)
	}
}

// ParseTime parses various date/time string formats and returns a time.Time
// When no timezone is specified, it uses the server's local timezone
func ParseTime(s string) (time.Time, error) {
	ts, _, err := ParseTimeAndRest(s)
	return ts, err
}

// 日時文字列の抽出と、日時文字列から日付文字列をどり除いた文字列を返す
func findTimeStringAndRest(re *regexp.Regexp, s string) (*string, *string) {
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
func ParseTimeAndRest(s string) (time.Time, string, error) {
	// タイムゾーン付きのフォーマットを試行（配列順序で）
	for _, format := range DefaultTimestampFormatsWithTZ {
		re := DefaultTimestampRegexpsWithTZ[format]
		dateStr, restStr := findTimeStringAndRest(&re, s)
		if dateStr == nil {
			continue
		}
		if t, err := time.Parse(format, *dateStr); err == nil {
			return t, *restStr, nil
		}
	}

	// タイムゾーンなしのフォーマットを試行（配列順序で）
	for _, format := range DefaultTimestampFormatsWithoutTZ {
		re := DefaultTimestampRegexpsWithoutTZ[format]
		dateStr, restStr := findTimeStringAndRest(&re, s)
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
	re := DefaultTimestampRegexpsWithTZ[time.RFC3339Nano]
	dateStr, _ := findTimeStringAndRest(&re, s)
	if dateStr == nil {
		return time.Time{}, fmt.Errorf("unable to parse date/time in the string: %s", s)
	}
	t, err := time.Parse(time.RFC3339Nano, *dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse date/time in the string: %s", s)
	}
	return t, nil
}
