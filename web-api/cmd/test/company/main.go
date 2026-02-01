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
		baseURL   = flag.String("base-url", "http://localhost:9090", "CompanyService のベース URL (例: http://localhost:9090)")
		companyID = flag.String("company-id", "", "取得する会社ID (未指定時は全会社を表示)")
		jsonOut   = flag.Bool("json", false, "JSON 形式で出力します")
		timeout   = flag.Duration("timeout", 5*time.Second, "RPC 呼び出しのタイムアウト")
		showCat   = flag.Bool("categories", false, "会社カテゴリー一覧を表示")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	companyServiceClient := grpcv1connect.NewCompanyServiceClient(http.DefaultClient, *baseURL)
	companyCategoryServiceClient := grpcv1connect.NewCompanyCategoryServiceClient(http.DefaultClient, *baseURL)

	// カテゴリー表示モード
	if *showCat {
		showCompanyCategories(ctx, companyCategoryServiceClient, *jsonOut)
		return
	}

	// 特定の会社ID指定時
	if *companyID != "" {
		showCompany(ctx, companyServiceClient, *companyID, *jsonOut)
		return
	}

	// 全会社表示
	showAllCompanies(ctx, companyServiceClient, *jsonOut)
}

// showCompanyCategories は会社カテゴリー一覧を表示します
func showCompanyCategories(ctx context.Context, client grpcv1connect.CompanyCategoryServiceClient, jsonOut bool) {
	req := grpcv1.GetCompanyCategoriesRequest_builder{}.Build()
	log.Printf("GetCompanyCategoriesRequest_builder, OK!\n")

	res, err := client.GetCompanyCategories(ctx, req)
	if err != nil {
		log.Fatalf("GetCompanyCategories の呼び出しに失敗しました: %v", err)
	}

	if jsonOut {
		data, err := json.MarshalIndent(res.GetCategories(), "", "  ")
		if err != nil {
			log.Fatalf("レスポンスの JSON 変換に失敗しました: %v", err)
		}
		fmt.Println(string(data))
		return
	}

	// ターミナル表示
	fmt.Println("Company Categories:")
	fmt.Println("Index\tLabel")
	fmt.Println("-----\t-----")
	for _, category := range res.GetCategories() {
		fmt.Printf("%d\t%s\n", category.GetIndex(), category.GetName())
	}
}

// showCompany は指定されたIDの会社情報を表示します
func showCompany(ctx context.Context, client grpcv1connect.CompanyServiceClient, targetId string, jsonOut bool) {
	req := grpcv1.GetCompanyRequest_builder{
		TargetId: targetId,
	}.Build()

	res, err := client.GetCompany(ctx, req)
	if err != nil {
		log.Fatalf("GetCompany の呼び出しに失敗しました: %v", err)
	}

	if jsonOut {
		data, err := json.MarshalIndent(res.GetCompany(), "", "  ")
		if err != nil {
			log.Fatalf("レスポンスの JSON 変換に失敗しました: %v", err)
		}
		fmt.Println(string(data))
		return
	}

	// ターミナル表示
	company := res.GetCompany()
	log.Println("Long Name:", company.GetPrLongName())
	fmt.Printf("Company Information (ID: %s)\n", targetId)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("ID: %s\n", company.GetId())
	fmt.Printf("Pathist Folder: %s\n", company.GetDirPath())
	fmt.Printf("Name: %s\n", company.GetName())
	fmt.Printf("Long Name: %s\n", company.GetPrLongName())
	fmt.Printf("Category Index: %d\n", company.GetCategoryIndex())
	fmt.Printf("Postal Code: %s\n", company.GetPrPostalCode())
	fmt.Printf("Address: %s\n", company.GetPrAddress())
	fmt.Printf("Tel: %s\n", company.GetPrTel())
	fmt.Printf("Fax: %s\n", company.GetPrFax())
	fmt.Printf("Email: %s\n", company.GetPrEmail())
	fmt.Printf("Website: %s\n", company.GetPrWebsite())
}

// showAllCompanies は全会社の一覧を表示します
func showAllCompanies(ctx context.Context, client grpcv1connect.CompanyServiceClient, jsonOut bool) {
	req := grpcv1.GetCompaniesRequest_builder{}.Build()

	res, err := client.GetCompanies(ctx, req)
	if err != nil {
		log.Fatalf("GetCompanies の呼び出しに失敗しました: %v", err)
	}

	companyMap := res.GetCompanies()

	if jsonOut {
		data, err := json.MarshalIndent(companyMap, "", "  ")
		if err != nil {
			log.Fatalf("レスポンスの JSON 変換に失敗しました: %v", err)
		}
		fmt.Println(string(data))
		return
	}

	// ターミナル表示
	fmt.Printf("Total Companies: %d\n", len(companyMap))
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("ID\t\tShort Name\t\tLegal Name\t\tCategory")
	fmt.Println(strings.Repeat("-", 80))

	for id, company := range companyMap {
		name := company.GetName()
		if len(name) > 15 {
			name = name[:12] + "..."
		}

		legalName := company.GetPrLongName()
		if len(legalName) > 30 {
			legalName = legalName[:27] + "..."
		}

		fmt.Printf("%-15s\t%-15s\t%-20s\t%d\n",
			id, name, legalName, company.GetCategoryIndex())
	}

	fmt.Println("\nUsage:")
	fmt.Println("  --company-id <ID>    : 特定の会社の詳細情報を表示")
	fmt.Println("  --categories         : 会社カテゴリー一覧を表示")
	fmt.Println("  --json              : JSON形式で出力")
}
