package utils

import (
	"fmt"
	"penguin-backend/internal/models"
	"regexp"
	"strings"
	"time"
)

// Formats with timezone information (try these first)
var DefaultFormatsWithTZ = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05.999999999Z07:00",
	"2006-01-02T15:04:05Z07:00",
}

// Formats without timezone information (use local timezone)
var DefaultFormatsWithoutTZ = []string{
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
	"20060102",
	"2006/01/02",
	"2006/1/2",
	"2006.01.02",
	"2006.1.2",
	"2006-01-02",
	"2006-1-2",
}

// 日付パース基本的な置換ルール
var FormatDateRules = map[string]string{
	"2006":      `\d{4}`,                   // 年
	"01":        `(?:0[1-9]|1[0-2])`,       // 月
	"02":        `(?:0[1-9]|[12]\d|3[01])`, // 日
	"15":        `(?:[01]\d|2[0-3])`,       // 時（24時間）
	"04":        `[0-5]\d`,                 // 分
	"05":        `[0-5]\d`,                 // 秒
	"999999999": `\d{9}`,                   // ナノ秒
	"Z07:00":    `(?:Z|[+-]\d{2}:\d{2})`,   // タイムゾーン
	"Z0700":     `(?:Z|[+-]\d{4})`,         // タイムゾーン（コロンなし）
	"Z07":       `(?:Z|[+-]\d{2})`,         // タイムゾーン（時間のみ）
}

var DefaultRegexpsWithTZ = make(map[string]regexp.Regexp, len(DefaultFormatsWithTZ))
var DefaultRegexpsWithoutTZ = make(map[string]regexp.Regexp, len(DefaultFormatsWithoutTZ))

// init initializes the default patterns
func init() {
	// Initialize patterns for formats with timezone
	for _, format := range DefaultFormatsWithTZ {
		regexString := timeFormatToRegex(format)
		re := regexp.MustCompile(regexString)
		DefaultRegexpsWithTZ[format] = *re
	}

	// Initialize patterns for formats without timezone
	for _, format := range DefaultFormatsWithoutTZ {
		regexString := timeFormatToRegex(format)
		re := regexp.MustCompile(regexString)
		DefaultRegexpsWithoutTZ[format] = *re
	}
}

// timeFormatToRegex converts a time format string to a regex pattern
func timeFormatToRegex(format string) string {
	// 特殊文字をエスケープ
	pattern := regexp.QuoteMeta(format)

	// 置換を実行
	for old, new := range FormatDateRules {
		pattern = strings.Replace(pattern, regexp.QuoteMeta(old), new, -1)
	}

	return pattern
}

// ParseDateTime parses various date/time string formats and returns a time.Time
// When no timezone is specified, it uses the server's local timezone
func ParseDateTime(s string) (time.Time, error) {
	t, _, err := parseDateTimeWithRest(s, false)
	return t, err
}

// 文字列をパースし、戻り値はtime.Timeと、日付文字列から日付文字列をどり除いた文字列を返す
func ParseDateStringAndRest(s string) (time.Time, string, error) {
	// タイムゾーン付きのフォーマットを試行
	for format, re := range DefaultRegexpsWithTZ {
		matches := re.FindStringIndex(s)
		if matches == nil {
			continue
		}

		// 日付部分を抽出
		dateStr := s[matches[0]:matches[1]]
		// 日付の前の部分を取得
		prefix := strings.TrimSpace(s[:matches[0]])

		if t, err := time.Parse(format, dateStr); err == nil {
			return t, prefix, nil
		}
	}

	// タイムゾーンなしのフォーマットを試行
	for format, re := range DefaultRegexpsWithoutTZ {
		matches := re.FindStringIndex(s)
		if matches == nil {
			continue
		}

		// 日付部分を抽出
		dateStr := s[matches[0]:matches[1]]
		// 日付の前の部分を取得
		prefix := strings.TrimSpace(s[:matches[0]])

		if t, err := time.ParseInLocation(format, dateStr, time.Local); err == nil {
			return t, prefix, nil
		}
	}

	return time.Time{}, s, fmt.Errorf("unable to parse date/time in the string: %s", s)
}

// parseDateTimeWithRest is a common implementation for parsing date/time strings
func parseDateTimeWithRest(s string, returnRest bool) (time.Time, string, error) {
	// Try formats with timezone first
	for _, format := range DefaultFormatsWithTZ {
		if t, err := time.Parse(format, s); err == nil {
			if returnRest {
				return t, s[len(format):], nil
			}
			return t, "", nil
		}
	}

	// Try formats without timezone using local timezone
	for _, format := range DefaultFormatsWithoutTZ {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			if returnRest {
				return t, s[len(format):], nil
			}
			return t, "", nil
		}
	}

	// Handle special compact format like "1971-0618"
	if len(s) >= 8 && strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) == 2 && len(parts[0]) == 4 {
			// Try to parse as YYYY-MMDD
			compactDate := parts[0] + parts[1]
			if len(compactDate) == 8 {
				if t, err := time.ParseInLocation("20060102", compactDate, time.Local); err == nil {
					if returnRest {
						return t, s[8:], nil
					}
					return t, "", nil
				}
			}
		}
	}

	return time.Time{}, s, fmt.Errorf("unable to parse date/time: %s", s)
}

// RFC3339Nano で文字列をパースする
func ParseRFC3339Nano(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}

// Koujiデータ内の日付データの正規化
func NormalizeKoujiDate(kouji *models.Kouji) (time.Time, error) {
	modifiedTime := kouji.Folder.ModifiedTime

}
