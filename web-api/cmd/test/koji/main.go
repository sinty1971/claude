package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	grpcv1 "web-api/gen/grpc/v1"
	"web-api/gen/grpc/v1/grpcv1connect"
)

func main() {
	var (
		baseURL = flag.String("base-url", "http://localhost:9090", "KojiService のベース URL (例: http://localhost:9090)")
		kojiID  = flag.String("koji-id", "", "取得する工事ID (未指定時は全工事を表示)")
		jsonOut = flag.Bool("json", false, "JSON 形式で出力します")
		timeout = flag.Duration("timeout", 5*time.Second, "RPC 呼び出しのタイムアウト")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	client := grpcv1connect.NewKojiServiceClient(http.DefaultClient, *baseURL)

	// 特定の工事ID指定時
	if *kojiID != "" {
		showKoji(ctx, client, *kojiID, *jsonOut)
		return
	}

	// 全工事表示
	showAllKojies(ctx, client, *jsonOut)
}

// showKoji は指定されたIDの工事情報を表示します
func showKoji(ctx context.Context, client grpcv1connect.KojiServiceClient, targetId string, jsonOut bool) {
	req := grpcv1.GetKojiRequest_builder{
		TargetId: targetId,
	}.Build()

	res, err := client.GetKoji(ctx, req)
	if err != nil {
		log.Fatalf("GetKoji の呼び出しに失敗しました: %v", err)
	}

	if jsonOut {
		data, err := json.MarshalIndent(res.GetKoji(), "", "  ")
		if err != nil {
			log.Fatalf("レスポンスの JSON 変換に失敗しました: %v", err)
		}
		fmt.Println(string(data))
		return
	}

	// ターミナル表示
	koji := res.GetKoji()
	jst := time.FixedZone("JST", 9*60*60)
	fmt.Printf("Koji Information (ID: %s)\n", targetId)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("ID: %s\n", koji.GetId())
	fmt.Printf("Pathist Folder: %s\n", koji.GetDirPath())
	if koji.GetStart() != nil && koji.GetStart().IsValid() {
		fmt.Printf("Start: %s\n", koji.GetStart().AsTime().In(jst).Format("2006-01-02 15:04:05 MST"))
	} else {
		fmt.Printf("Start: N/A\n")
	}
	fmt.Printf("Company: %s\n", koji.GetCompanyName())
	fmt.Printf("Location: %s\n", koji.GetLocationName())
	if koji.GetPrEnd() != nil && koji.GetPrEnd().IsValid() {
		fmt.Printf("End Date: %s\n", koji.GetPrEnd().AsTime().In(jst).Format("2006-01-02 15:04:05 MST"))
	} else {
		fmt.Printf("End Date: N/A\n")
	}
}

// showAllKojies は全工事の一覧を表示します
func showAllKojies(ctx context.Context, client grpcv1connect.KojiServiceClient, jsonOut bool) {
	req := grpcv1.GetKojiesRequest_builder{}.Build()

	res, err := client.GetKojies(ctx, req)
	if err != nil {
		log.Fatalf("GetKojies の呼び出しに失敗しました: %v", err)
	}

	kojiMap := res.GetKojies()

	if jsonOut {
		data, err := json.MarshalIndent(kojiMap, "", "  ")
		if err != nil {
			log.Fatalf("レスポンスの JSON 変換に失敗しました: %v", err)
		}
		fmt.Println(string(data))
		return
	}

	// ターミナル表示
	fmt.Printf("Total Kojies: %d\n", len(kojiMap))
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println("ID\t\tLong Name\t\t\tCompany ID\tStart Date")
	fmt.Println(strings.Repeat("-", 100))

	jst := time.FixedZone("JST", 9*60*60)
	for id, koji := range kojiMap {
		start := koji.GetStart()
		startText := "N/A"
		if start != nil && start.IsValid() {
			startText = start.AsTime().In(jst).Format("2006-01-02")
		}
		companyName := koji.GetCompanyName()
		if len(companyName) > 30 {
			companyName = companyName[:27] + "..."
		}
		locationName := koji.GetLocationName()
		if len(locationName) > 30 {
			locationName = locationName[:27] + "..."
		}

		fmt.Printf("%-15s\t%-15s\t%-30s\t%-10s\n",
			id, startText, companyName, locationName)
	}

	fmt.Println("\nUsage:")
	fmt.Println("  --koji-id <ID>      : 特定の工事の詳細情報を表示")
	fmt.Println("  --json              : JSON形式で出力")
}
