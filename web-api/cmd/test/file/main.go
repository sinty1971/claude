package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	grpcv1 "web-api/gen/grpc/v1"
	"web-api/gen/grpc/v1/grpcv1connect"
)

func main() {
	var (
		baseURL = flag.String("base-url", "http://localhost:9090", "FileService のベース URL (例: http://localhost:9090)")
		target  = flag.String("target", "", "一覧を取得する相対パス (未指定時はルート)")
		jsonOut = flag.Bool("json", false, "JSON 形式で出力します")
		timeout = flag.Duration("timeout", 5*time.Second, "RPC 呼び出しのタイムアウト")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	client := grpcv1connect.NewDirectoryServiceClient(http.DefaultClient, *baseURL)

	req := grpcv1.GetPathListRequest_builder{
		RelativePath: *target,
	}.Build()

	res, err := client.GetPathList(ctx, req)
	if err != nil {
		log.Fatalf("ListFiles の呼び出しに失敗しました: %v", err)
	}

	if *jsonOut {
		output := struct {
			ManagedFolder string   `json:"managedFolder"`
			Target        string   `json:"target"`
			FilePaths     []string `json:"filePaths"`
		}{
			Target:    req.GetRelativePath(),
			FilePaths: res.GetPathList(),
		}

		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			log.Fatalf("レスポンスの JSON 変換に失敗しました: %v", err)
		}
		fmt.Println(string(data))
		return
	}

	// ターミナルで読みやすいように簡易フォーマットで出力する。
	fmt.Printf("PathistFolder: %s\n", req.GetRelativePath())
	fmt.Println("IsDir\tSize\tModified\tPathistFolder\tIdealPath")
	for _, filePath := range res.GetPathList() {
		fmt.Printf("%s\n", filePath)
	}
}
