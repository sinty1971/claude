package models

import (
	"errors"
	"os"

	grpcv1 "web-api/gen/grpc/v1"
	"web-api/internal/core"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// grpcv1.Folder 関連のヘルパー
// このファイルは proto ファイルから自動生成された gRPC メッセージ型を
// 補完するためのヘルパー関数や型を提供します。

// Folder は gRPC の Folder メッセージの拡張機能版です。
type Folder struct {
	*grpcv1.Folder

	// ManifestProvider は永続化設定を管理します。
	*core.ManifestProvider
}

// NewFolder Folder インスタンスを作成します。
func NewFolder() *Folder {
	return &Folder{
		Folder: grpcv1.Folder_builder{}.Build(),
	}
}

// ParseFrom は指定されたフルパスからファイル情報を解析して設定します
func (m *Folder) ParseFrom(folderPath string) error {
	var err error

	// 絶対パスの正規化
	normalized, err := core.NormalizeAbsPath(folderPath)
	if err != nil {
		return err
	}

	// ファイル情報の取得
	osFi, err := os.Stat(normalized)
	if err != nil {
		return err
	}

	// フォルダーであることを確認
	if !osFi.IsDir() {
		return errors.New("指定されたパスはフォルダーではありません: specified path is not a folder")
	}

	// 最終更新時刻の取得
	osModTime := osFi.ModTime()
	if osModTime.IsZero() {
		return errors.New("フォルダー最終更新日の取得に失敗しました: folder modification time is zero")
	}
	modifiedTime := timestamppb.New(osModTime)

	// フィールドの設定
	m.SetFolderPath(normalized)
	m.SetSize(osFi.Size())
	m.SetModifiedTime(modifiedTime)
	return nil
}
