package services

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
