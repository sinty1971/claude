package core

import (
	"errors"
	"strconv"
	"strings"
	grpcv1 "web-api/gen/grpc/v1"
)

// CompanyCategoryCacheMap は会社業種カテゴリーのキャッシュマップです
var CompanyCategoryCacheMap map[string]*grpcv1.CompanyCategory

var CompanyCategoryMin int32 = 0
var CompanyCategoryMax int32 = 0

func init() {
	CompanyCategoryCacheMap = make(map[string]*grpcv1.CompanyCategory)
}

// ErrorCompanyCategoryIndex 引数: idx が有効な範囲内かをチェックします
func ErrorCompanyCategoryIndex(idx int32) error {
	if CompanyCategoryMin <= idx && idx <= CompanyCategoryMax {
		return nil
	}
	return errors.New("invalid CompanyCategoryIndex")
}

// ParseCompanyCategoryFromDirName は業種カテゴリーインスタンス文字列からインデックスを取得します
// 引数: dirName 数字から始まるときは業種インデックス、それ以外は業種名として判断します
func ParseCompanyCategoryFromDirName(dirName string) (*grpcv1.CompanyCategory, error) {
	if len(dirName) < 3 {
		return nil, errors.New("会社ディレクトリ名が短すぎます: " + dirName)
	}

	// 文字列の先頭が数字かどうかを判定
	firstChar := dirName[0]
	idx, err := strconv.Atoi(string(firstChar))

	// 数字の場合はインデックスとして処理
	secondChar := dirName[1]
	if err == nil && secondChar == ' ' {
		for _, cat := range CompanyCategoryCacheMap {
			if cat.GetIndex() == int32(idx) {
				return cat, nil
			}
		}
	}

	// 数字でない場合は名前として処理
	for _, cc := range CompanyCategoryCacheMap {
		// name が含まれるかどうかを判定
		if strings.Contains(dirName, cc.GetName()) {
			return cc, nil
		}
	}
	// 見つからなかった場合はエラーを返す
	return nil, errors.New("ディレクトリ名から CompanyCategory を取得できませんでした: " + dirName)
}

// ContainCompanyCategoryName は業種カテゴリー名が text に含まれるかを判定します
func ContainCompanyCategoryName(cat *grpcv1.CompanyCategory, text string) bool {
	return strings.Contains(text, cat.GetName())
}
