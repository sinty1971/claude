package models

import (
	"errors"
	"os"

	grpcv1 "web-api/gen/grpc/v1"
	"web-api/internal/core"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

// grpcv1.Folder 関連のヘルパー
// このファイルは proto ファイルから自動生成された gRPC メッセージ型を
// 補完するためのヘルパー関数や型を提供します。

// Directory は gRPC の Directory メッセージの拡張機能版です。
type Directory struct{}

// ParseFrom は指定されたフルパスからファイル情報を解析して設定します
func (m *Directory) GenerateMessage(message protoreflect.Message) error {
	// message の型アサーション
	mes, ok := message.Interface().(*grpcv1.Directory)
	if !ok {
		return errors.New("invalid message type")
	}

	// 絶対パスの正規化
	absPath, err := core.ResolveAbsPath(mes.GetDirPath())
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

	// メッセージフィールドの設定
	mes.SetId(core.BytesToId([]byte(absPath)))
	mes.SetDirPath(absPath)
	return nil
}
