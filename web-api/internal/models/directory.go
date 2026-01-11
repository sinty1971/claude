package models

import (
	"errors"
	"os"

	grpcv1 "web-api/gen/grpc/v1"
	"web-api/internal/core"
)

// grpcv1.Folder 関連のヘルパー
// このファイルは proto ファイルから自動生成された gRPC メッセージ型を
// 補完するためのヘルパー関数や型を提供します。

// Directory は gRPC の Directory メッセージの拡張機能版です。
type Directory struct {
	*grpcv1.Directory

	// ManifestProvider は永続化設定を管理します。
	*core.ManifestProvider
}

// NewFolder Folder インスタンスを作成します。
func NewDirectory() *Directory {
	return &Directory{
		Directory: grpcv1.Directory_builder{}.Build(),
	}
}

// ParseFrom は指定されたフルパスからファイル情報を解析して設定します
func (m *Directory) ParseFromDirPath(dirPath string) error {
	var err error

	// 絶対パスの正規化
	absPath, err := core.NormalizeAbsPath(dirPath)
	if err != nil {
		return err
	}

	// ファイル情報の取得
	fi, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	// フォルダーであることを確認
	if !fi.IsDir() {
		return errors.New("指定されたパスはフォルダーではありません: specified path is not a folder")
	}

	// フィールドの設定
	m.SetDirPath(absPath)
	return nil
}
