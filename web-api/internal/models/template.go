package models

import (
	"google.golang.org/protobuf/proto"

	"web-api/internal/core"
)

// Template は gRPC grpc.v1.XXXX メッセージの拡張版です。
//
// Deprecated: 新規実装では TypedTemplate を使用してください。
type Template struct{}

// InitializeFromMessage はモデルの protobuf メッセージを取得します。
func (m *Template) InitializeFromMessage(message proto.Message) (proto.Message, error) {
	// mes := grpcv1.XXXX_builder{}.Build()
	// return mes, nil
	return nil, nil
}

// GenerateId は message からIDを生成します
func (m *Template) generateId(message proto.Message) string {
	// idString := "some_field_value"
	return core.BytesToId([]byte(""))
}

// UpdateMessage は情報を更新します
func (m *Template) UpdateMessage(target proto.Message, source proto.Message) error {
	// 更新処理をここに実装
	return nil
}

func NewPersistModelTemplate(arg string) (*core.PersistModel[proto.Message, *Template], error) {
	// PersistModel を作成
	pm, err := core.NewPersistModel(&Template{}, "template.yaml")
	if err != nil {
		return nil, err
	}

	// 初期化
	// request := grpcv1.Template_builder{}.Build()
	// request.SetDirPath(dirPath)
	// err = pm.Initialize(request)
	// if err != nil {
	// 	return nil, err
	// }

	//
	return pm, nil
}
