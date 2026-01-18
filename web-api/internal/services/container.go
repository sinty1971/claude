package services

import (
	"net/http"

	"connectrpc.com/grpcreflect"
)

// ContainerService はデータストレージサービスを管理します。
type ContainerService struct {
	Services map[string]*Service
}

// NewContainerService は ContainerService の新しいインスタンスを作成します。
func NewContainerService() *ContainerService {
	cs := &ContainerService{}
	cs.Services = make(map[string]*Service)
	return cs
}

// Service はファイルシステムとデータベースを橋渡しするインターフェースを定義します。
type Service interface {
	Name() string
	Start() error
	Cleanup()
	GenerateHandler() (servicePath string, handler http.Handler, serviceName string)
}

func (cs *ContainerService) GenerateMux() (*http.ServeMux, error) {

	mux := http.NewServeMux()

	handleMap := make(map[string]http.Handler, len(cs.Services))
	serviceNames := make([]string, 0, len(cs.Services))

	// ServicePath, ConnectHandlerの取得
	for _, service := range cs.Services {
		servicePath, connectHandler, serviceName := (*service).GenerateHandler()
		handleMap[servicePath] = connectHandler
		serviceNames = append(serviceNames, serviceName)
	}

	// ハンドラの登録
	for servicePath, connectHandler := range handleMap {
		mux.Handle(servicePath, connectHandler)
	}

	// gRPC ハンドラの登録
	reflector := grpcreflect.NewStaticReflector(serviceNames...)

	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	return mux, nil
}

// AddService はサービスを追加する
func (cs *ContainerService) AddService(service Service) {
	cs.Services[service.Name()] = &service
}

// Start はすべてのサービスを起動する
func (cs *ContainerService) Start() error {
	for _, p := range cs.Services {
		if err := (*p).Start(); err != nil {
			return err
		}
	}
	return nil
}

// CleanupAll はサービスをクリーンアップする
func (cs *ContainerService) CleanupAll() {
	for _, srv := range cs.Services {
		(*srv).Cleanup()
	}
}
