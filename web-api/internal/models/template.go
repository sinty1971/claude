package models

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"

	"web-api/internal/core"
)

// Template は gRPC grpc.v1.XXXX メッセージの拡張版です。
type Template struct{}

// GenerateMessage はモデルの protobuf メッセージを取得します。
func (m *Template) GenerateMessage(request protoreflect.Message) (protoreflect.ProtoMessage, error) {
	// mes := grpcv1.XXXX_builder{}.Build()
	// return mes, nil
	return nil, nil
}

// GenerateId は message からIDを生成します
func (m *Template) GenerateId(message protoreflect.Message) string {
	// idString := "some_field_value"
	return core.BytesToId([]byte(""))
}

// UpdateMessage は情報を更新します
func (m *Template) UpdateMessage(target protoreflect.Message, source protoreflect.Message) error {
	// 更新処理をここに実装
	return nil
}

func NewPersistModelTemplate(arg string) (*core.PersistModel[*Template], error) {
	// PersistModel を作成
	pm, err := core.NewPersistModel(&Template{}, "template.yaml")
	if err != nil {
		return nil, err
	}

	// 初期化
	// request := grpcv1.Template_builder{}.Build()
	// request.SetDirPath(dirPath)
	// err = pm.Initialize(request.ProtoReflect())
	// if err != nil {
	// 	return nil, err
	// }

	//
	return pm, nil
}
