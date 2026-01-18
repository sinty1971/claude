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
		baseURL  = flag.String("base-url", "http://localhost:9090", "MemberService のベース URL (例: http://localhost:9090)")
		memberID = flag.String("member-id", "", "取得する作業員ID (未指定時は全作業員を表示)")
		jsonOut  = flag.Bool("json", false, "JSON 形式で出力します")
		timeout  = flag.Duration("timeout", 5*time.Second, "RPC 呼び出しのタイムアウト")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	client := grpcv1connect.NewMemberServiceClient(http.DefaultClient, *baseURL)

	// 特定の作業員ID指定時
	if *memberID != "" {
		showMember(ctx, client, *memberID, *jsonOut)
		return
	}

	// 全作業員表示
	showAllMembers(ctx, client, *jsonOut)
}

// showMember は指定されたIDの作業員情報を表示します
func showMember(ctx context.Context, client grpcv1connect.MemberServiceClient, targetId string, jsonOut bool) {
	req := grpcv1.GetMemberRequest_builder{
		TargetId: targetId,
	}.Build()

	res, err := client.GetMember(ctx, req)
	if err != nil {
		log.Fatalf("GetMember の呼び出しに失敗しました: %v", err)
	}

	if jsonOut {
		data, err := json.MarshalIndent(res.GetMember(), "", "  ")
		if err != nil {
			log.Fatalf("レスポンスの JSON 変換に失敗しました: %v", err)
		}
		fmt.Println(string(data))
		return
	}

	// ターミナル表示
	member := res.GetMember()
	fmt.Printf("Member Information (ID: %s)\n", targetId)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("ID: %s\n", member.GetId())
	fmt.Printf("Name: %s\n", member.GetName())
	fmt.Printf("Pathist Folder: %s\n", member.GetDirPath())
	fmt.Printf("Company Name: %s\n", member.GetCompanyName())
	fmt.Printf("Role: %s\n", member.GetMfRole())
	if member.GetMfEmail() != "" {
		fmt.Printf("Email: %s\n", member.GetMfEmail())
	}
	if member.GetMfMobile() != "" {
		fmt.Printf("Phone: %s\n", member.GetMfMobile())
	}
}

// showAllMembers は全作業員の一覧を表示します
func showAllMembers(ctx context.Context, client grpcv1connect.MemberServiceClient, jsonOut bool) {
	req := grpcv1.GetMembersRequest_builder{}.Build()

	res, err := client.GetMembers(ctx, req)
	if err != nil {
		log.Fatalf("GetMembers の呼び出しに失敗しました: %v", err)
	}

	memberMap := res.GetMembers()

	if jsonOut {
		data, err := json.MarshalIndent(memberMap, "", "  ")
		if err != nil {
			log.Fatalf("レスポンスの JSON 変換に失敗しました: %v", err)
		}
		fmt.Println(string(data))
		return
	}

	// ターミナル表示
	fmt.Printf("Total Members: %d\n", len(memberMap))
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println("ID\t\tName\t\t\tCompany ID\tRole")
	fmt.Println(strings.Repeat("-", 100))

	for id, member := range memberMap {
		name := member.GetName()
		if len(name) > 20 {
			name = name[:17] + "..."
		}
		companyName := member.GetCompanyName()
		if len(companyName) > 20 {
			companyName = companyName[:17] + "..."
		}
		role := member.GetMfRole()
		if len(role) > 15 {
			role = role[:12] + "..."
		}

		fmt.Printf("%-15s\t%-20s\t%-20s\t%-15s\n",
			id, name, companyName, role)
	}

	fmt.Println("\nUsage:")
	fmt.Println("  --member-id <ID>    : 特定の作業員の詳細情報を表示")
	fmt.Println("  --json              : JSON形式で出力")
}
