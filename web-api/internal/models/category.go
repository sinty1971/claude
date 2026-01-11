package models

import "errors"

// 業種カテゴリーの定義
const (
	CompanyCategoryMin      int32 = 0
	CompanyCategoryUnion    int32 = 0
	CompanyCategoryAgency   int32 = 1
	CompanyCategoryPeer     int32 = 2
	CompanyCategoryPersonal int32 = 3
	CompanyCategoryPrime    int32 = 4
	CompanyCategoryLease    int32 = 5
	CompanyCategorySales    int32 = 6
	CompanyCategorySales2   int32 = 7
	CompanyCategoryRecruit  int32 = 8
	CompanyCategoryOther    int32 = 9
	CompanyCategoryMax      int32 = 9
)

// CompanyCategoryMap 業種カテゴリーのラベルマップです
// 将来的にはyamlファイルから読み込む予定
var CompanyCategoryMap = map[int32]string{
	CompanyCategoryUnion:    "自社組合",
	CompanyCategoryAgency:   "下請会社",
	CompanyCategoryPeer:     "築炉会社",
	CompanyCategoryPersonal: "一人親方",
	CompanyCategoryPrime:    "元請会社",
	CompanyCategoryLease:    "リース会社",
	CompanyCategorySales:    "販売会社",
	CompanyCategorySales2:   "販売会社２",
	CompanyCategoryRecruit:  "求人会社",
	CompanyCategoryOther:    "一般会社",
}

// CompanyCategoryReverseMap は業種カテゴリーの逆引きマップです
var CompanyCategoryReverseMap = map[string]int{}

// init はパッケージ初期化時に呼び出され、逆引きマップを初期化します
func init() {
	// CompanyCategoryReverseMapを初期化
	for ci, name := range CompanyCategoryMap {
		CompanyCategoryReverseMap[name] = int(ci)
	}
}

// ErrorCompanyCategoryIndex 引数: idx が有効な範囲内かをチェックします
func ErrorCompanyCategoryIndex(ci int32) error {
	if CompanyCategoryMin <= ci && ci <= CompanyCategoryMax {
		return nil
	}
	return errors.New("invalid CompanyCategoryIndex")
}
