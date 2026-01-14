package services

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	grpc "web-api/gen/grpc/v1"
	grpcConnect "web-api/gen/grpc/v1/grpcv1connect"
	"web-api/internal/core"
)

// FolderManager の実装

// DirectoryService exposes DirectoryService operations via Connect handlers.
type DirectoryService struct {
	// cs はContainerServiceへの参照
	cs *ContainerService

	// BaseDirPath はファイルサービスの絶対パスフォルダー
	BaseDirPath string `json:"dir_path" yaml:"dir_path" example:"/penguin/豊田築炉"`

	// Embed the unimplemented handler for forward compatibility
	grpcConnect.UnimplementedDirectoryServiceHandler
}

func NewDirectoryService(cs *ContainerService) *DirectoryService {

	// パスを正規化
	baseDirPath, err := core.ResolveAbsPath(core.Config.DirectoryBaseDirPath)
	if err != nil {
		panic(err)
	}

	return &DirectoryService{
		cs:          cs,
		BaseDirPath: baseDirPath,
	}
}

// Name はサービス名を返します
func (srv *DirectoryService) Name() string {
	return "DirectoryService"
}

func (srv *DirectoryService) Start() error {
	return nil
}

func (srv *DirectoryService) Cleanup() {
	// 現在はクリーンアップ処理は不要
}

// GetFiles は指定されたパスのファイル情報一覧を返す
func (srv *DirectoryService) GetFiles(
	// args
	ctx context.Context,
	req *grpc.GetPathListRequest) (

	// returns
	res *grpc.GetPathListResponse,
	err error) {

	// 無視する引数
	_ = ctx

	// リクエスト情報の取得
	relPath := req.GetRelativePath()

	// 絶対パスを取得
	absPath, err := srv.GetAbsPathFrom(relPath)
	if err != nil {
		return nil, err
	}

	// ファイルエントリ配列を取得
	dirs, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	// ファイルエントリが0の場合は空配列を返す
	res = grpc.GetPathListResponse_builder{}.Build()
	pathList := make([]string, 0)
	dirsNum := len(dirs)
	if dirsNum == 0 {
		res.SetPathList(pathList)
		return res, nil
	}

	// ワーカーグループとチャンネルを設定
	WorkerCount := core.Config.CalculateWorkerCount(dirsNum)
	var wg sync.WaitGroup
	channelIn := make(chan int, dirsNum)
	channelOut := make(chan string, dirsNum)

	// ワーカーを起動
	for range WorkerCount {
		wg.Go(func() {
			for idx := range channelIn {
				dir := dirs[idx]
				fullpath := filepath.Join(absPath, dir.Name())
				channelOut <- fullpath
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
	pathList = make([]string, 0, len(dirs))
	for fi := range channelOut {
		pathList = append(pathList, fi)
	}

	// レスポンスを更新して返す
	res.SetPathList(pathList)
	return res, nil
}

// GetAbsPathFrom BasePathに引数の相対パスを追加した絶対パスを返す
func (srv *DirectoryService) GetAbsPathFrom(relPath string) (res string, err error) {
	// 絶対パスがある場合はエラーを返す
	if strings.HasPrefix(relPath, "~/") || filepath.IsAbs(relPath) {
		return "", errors.New("絶対パスは使用できません")
	}

	res = filepath.Join(srv.BaseDirPath, relPath)

	return // naked return
}

// CopyFile はファイルまたはディレクトリをコピーする
func (srv *DirectoryService) CopyFile(relSrc, relDst string) (err error) {
	var absSrc, absDst string

	// relSrcがパスチェック及び絶対パス変換
	absSrc, err = srv.GetAbsPathFrom(relSrc)
	if err != nil {
		return
	}

	// relDstのパスチェック及び絶対パス変換
	absDst, err = srv.GetAbsPathFrom(relDst)
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
		err = srv.absCopyDir(absSrc, absDst)
	} else {
		// ファイルの場合
		err = srv.absCopyFile(absSrc, absDst)
	}

	return
}

// absCopyFile はファイルをコピーする内部関数
func (srv *DirectoryService) absCopyFile(absSrc, absDst string) (err error) {
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
func (srv *DirectoryService) absCopyDir(absSrc, absDst string) error {
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
			if err := srv.absCopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// ファイルの場合、ファイルをコピー
			if err := srv.absCopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// MoveFile はファイルを移動する
func (srv *DirectoryService) MoveFile(relSrc, relDst string) error {
	absSrc, err := srv.GetAbsPathFrom(relSrc)
	if err != nil {
		return err
	}
	absDst, err := srv.GetAbsPathFrom(relDst)
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
func (srv *DirectoryService) DeleteFile(relPath string) error {
	absPath, err := srv.GetAbsPathFrom(relPath)
	if err != nil {
		return err
	}

	return os.Remove(absPath)
}
