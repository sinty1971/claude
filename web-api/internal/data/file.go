package data

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	grpc "web-api/gen/grpc/v1"
	grpcConnect "web-api/gen/grpc/v1/grpcv1connect"
	"web-api/internal/core"
	"web-api/internal/models"
)

// FileStorage の実装

// FileStorage exposes FileStorage operations via Connect handlers.
type FileStorage struct {
	// Embed the unimplemented handler for forward compatibility
	grpcConnect.UnimplementedFileServiceHandler

	// manager はStorageManagerへの参照
	manager *StorageManager

	// StoragePath はファイルサービスの絶対パスフォルダー
	StoragePath string `json:"storagePath" yaml:"storage_path" example:"/penguin/豊田築炉"`
}

// Name はサービス名を返します
func (srv *FileStorage) Name() string {
	return "FileStorage"
}

func (srv *FileStorage) Start(services *StorageManager, options *map[string]string) error {
	// オプションの取得
	optTarget, exists := (*options)["FileServiceTarget"]
	if !exists {
		return errors.New("FileServiceTarget option is required")
	}

	// パスを正規化
	target, err := core.NormalizeAbsPath(optTarget)
	if err != nil {
		return err
	}

	srv.manager = services
	srv.StoragePath = target

	return nil
}

func (s *FileStorage) Cleanup() {
	// 現在はクリーンアップ処理は不要
}

// SyncToDB はファイルサービスの情報を SQLite へ同期する際に利用します。
func (s *FileStorage) SyncToDB(_ *sql.DB) error {
	return nil
}

func (s *FileStorage) GetFileBasePath(
	ctx context.Context, req *grpc.GetFilePathistFolderRequest) (
	*grpc.GetFilePathistFolderResponse, error) {
	// コンテキストを無視
	_ = ctx
	_ = req

	res := grpc.GetFilePathistFolderResponse_builder{}.Build()
	res.SetPathistFolder(s.StoragePath)
	return res, nil
}

// GetFiles は指定されたパスのファイル情報一覧を返す
func (s *FileStorage) GetFiles(
	ctx context.Context, req *grpc.GetFilesRequest) (
	*grpc.GetFilesResponse, error) {

	// 無視する引数
	_ = ctx

	// リクエスト情報の取得
	reqTarget := req.GetPathistFolder()

	// 絶対パスを取得
	absPath, err := s.GetAbsPathFrom(reqTarget)
	if err != nil {
		return nil, err
	}

	// ファイルエントリ配列を取得
	dirs, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	// ファイルエントリが0の場合は空配列を返す
	res := grpc.GetFilesResponse_builder{}.Build()
	files := make([]*grpc.File, 0)
	dirsNum := len(dirs)
	if dirsNum == 0 {
		res.SetFiles(files)
		return res, nil
	}

	// ワーカーグループとチャンネルを設定
	workerNum := core.DecideNumWorkers(dirsNum)
	var wg sync.WaitGroup
	channelIn := make(chan int, dirsNum)
	channelOut := make(chan *grpc.File, dirsNum)

	// ワーカーを起動
	for range workerNum {
		wg.Go(func() {
			for idx := range channelIn {
				dir := dirs[idx]
				fullpath := filepath.Join(absPath, dir.Name())

				fi := models.NewFile()
				if err := fi.ParseFrom(fullpath); err == nil {
					channelOut <- fi.File
				}
			}
		})
	}

	// ジョブを送信
	for idx := range dirs {
		channelIn <- idx
	}
	close(channelIn)

	// ワーカーの完了を待つ
	go func() {
		wg.Wait()
		close(channelOut)
	}()

	// 結果を収集
	files = make([]*grpc.File, 0, len(dirs))
	for fi := range channelOut {
		files = append(files, fi)
	}

	// レスポンスを更新して返す
	res.SetFiles(files)
	return res, nil
}

// GetAbsPathFrom BasePathに引数の相対パスを追加した絶対パスを返す
func (s *FileStorage) GetAbsPathFrom(relPath string) (res string, err error) {
	// 絶対パスがある場合はエラーを返す
	if strings.HasPrefix(relPath, "~/") || filepath.IsAbs(relPath) {
		return "", errors.New("絶対パスは使用できません")
	}

	res = filepath.Join(s.StoragePath, relPath)

	return // naked return
}

// CopyFile はファイルまたはディレクトリをコピーする
func (s *FileStorage) CopyFile(relSrc, relDst string) (err error) {
	var absSrc, absDst string

	// relSrcがパスチェック及び絶対パス変換
	absSrc, err = s.GetAbsPathFrom(relSrc)
	if err != nil {
		return
	}

	// relDstのパスチェック及び絶対パス変換
	absDst, err = s.GetAbsPathFrom(relDst)
	if err != nil {
		return
	}

	// コピー元の存在確認
	srcOsFi, err := os.Stat(absSrc)
	if err != nil {
		return
	}

	// ディレクトリの場合
	if srcOsFi.IsDir() {
		err = s.absCopyDir(absSrc, absDst)
	} else {
		// ファイルの場合
		err = s.absCopyFile(absSrc, absDst)
	}

	return
}

// absCopyFile はファイルをコピーする内部関数
func (s *FileStorage) absCopyFile(absSrc, absDst string) (err error) {
	// コピー元ファイルを開く
	srcFile, err := os.Open(absSrc)
	if err != nil {
		return
	}
	defer srcFile.Close()

	// コピー先のディレクトリが存在しない場合は作成
	dstDir := filepath.Dir(absDst)
	if err = os.MkdirAll(dstDir, 0755); err != nil {
		return
	}

	// コピー先ファイルを作成
	dstFile, err := os.Create(absDst)
	if err != nil {
		return
	}
	defer dstFile.Close()

	// ファイル内容をコピー
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return
	}

	// ファイル権限をコピー
	if fi, err := os.Stat(absSrc); err != nil {
		return err
	} else {
		return os.Chmod(absDst, fi.Mode())
	}
}

// absCopyDir はディレクトリを再帰的にコピーする内部関数
func (s *FileStorage) absCopyDir(absSrc, absDst string) error {
	// コピー元ディレクトリの情報を取得
	srcInfo, err := os.Stat(absSrc)
	if err != nil {
		return err
	}

	// コピー先ディレクトリを作成
	if err := os.MkdirAll(absDst, srcInfo.Mode()); err != nil {
		return err
	}

	// ディレクトリ内のエントリを読み取り
	entries, err := os.ReadDir(absSrc)
	if err != nil {
		return err
	}

	// 各エントリを処理
	for _, entry := range entries {
		srcPath := filepath.Join(absSrc, entry.Name())
		dstPath := filepath.Join(absDst, entry.Name())

		if entry.IsDir() {
			// サブディレクトリの場合、再帰的にコピー
			if err := s.absCopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// ファイルの場合、ファイルをコピー
			if err := s.absCopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// MoveFile はファイルを移動する
func (s *FileStorage) MoveFile(relSrc, relDst string) error {
	absSrc, err := s.GetAbsPathFrom(relSrc)
	if err != nil {
		return err
	}
	absDst, err := s.GetAbsPathFrom(relDst)
	if err != nil {
		return err
	}

	// 移動先のディレクトリが存在するかチェック
	if _, err := os.Stat(absSrc); os.IsNotExist(err) {
		return errors.New("移動元のファイル/ディレクトリが存在しません: " + relSrc)
	}

	// 移動先の親ディレクトリを作成（必要に応じて）
	dstParent := filepath.Dir(absDst)
	if err := os.MkdirAll(dstParent, 0755); err != nil {
		return err
	}

	return os.Rename(absSrc, absDst)
}

// DeleteFile はファイルを削除する
func (s *FileStorage) DeleteFile(relPath string) error {
	absPath, err := s.GetAbsPathFrom(relPath)
	if err != nil {
		return err
	}

	return os.Remove(absPath)
}
