package core

import (
	"errors"
	"fmt"
	"strings"
	grpcv1migration "web-api/gen/sqlite/migrations"
)

// BuildInsertSQLFromDefaultMigrations は DefaultMigrations から INSERT 文を生成します。
func BuildInsertSQLFromDefaultMigrations(tableName string) (string, error) {
	if tableName == "" {
		return "", errors.New("tableName が空です")
	}

	for _, query := range grpcv1migration.DefaultMigrations() {
		columns, ok := extractColumnsFromCreateTable(query, tableName)
		if !ok {
			continue
		}
		if len(columns) == 0 {
			return "", errors.New("カラム情報が取得できません")
		}

		placeholders := make([]string, 0, len(columns))
		for range columns {
			placeholders = append(placeholders, "?")
		}

		return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
		), nil
	}

	return "", errors.New("対象テーブルが DefaultMigrations に見つかりません")
}

// BuildSelectByIDSQLFromDefaultMigrations は DefaultMigrations から SELECT 文を生成します。
func BuildSelectByIDSQLFromDefaultMigrations(tableName string) (string, error) {
	if tableName == "" {
		return "", errors.New("tableName が空です")
	}

	for _, query := range grpcv1migration.DefaultMigrations() {
		columns, ok := extractColumnsFromCreateTable(query, tableName)
		if !ok {
			continue
		}
		if len(columns) == 0 {
			return "", errors.New("カラム情報が取得できません")
		}

		return fmt.Sprintf(
			"SELECT %s FROM %s WHERE id = ?",
			strings.Join(columns, ", "),
			tableName,
		), nil
	}

	return "", errors.New("対象テーブルが DefaultMigrations に見つかりません")
}

func extractColumnsFromCreateTable(query string, tableName string) ([]string, bool) {
	trimmed := strings.TrimSpace(query)
	openIdx := strings.Index(trimmed, "(")
	closeIdx := strings.LastIndex(trimmed, ")")
	if openIdx == -1 || closeIdx == -1 || closeIdx <= openIdx {
		return nil, false
	}

	head := strings.TrimSpace(trimmed[:openIdx])
	parts := strings.Fields(head)
	if len(parts) == 0 {
		return nil, false
	}
	targetName := parts[len(parts)-1]
	if targetName != tableName {
		return nil, false
	}

	body := trimmed[openIdx+1 : closeIdx]
	defs := strings.Split(body, ",")
	columns := make([]string, 0, len(defs))
	for _, def := range defs {
		token := strings.TrimSpace(def)
		if token == "" {
			continue
		}
		tokenUpper := strings.ToUpper(token)
		if strings.HasPrefix(tokenUpper, "PRIMARY ") ||
			strings.HasPrefix(tokenUpper, "FOREIGN ") ||
			strings.HasPrefix(tokenUpper, "UNIQUE ") ||
			strings.HasPrefix(tokenUpper, "CHECK ") {
			continue
		}
		fields := strings.Fields(token)
		if len(fields) == 0 {
			continue
		}
		columns = append(columns, fields[0])
	}

	return columns, true
}
