<script lang="ts">
  import { onMount } from "svelte";
  import { create } from "@bufbuild/protobuf";
  import { createClient } from "@connectrpc/connect";
  import { createConnectTransport } from "@connectrpc/connect-web";
  import {
    CompanyService,
    GetCompaniesRequestSchema,
    GetCompanyCategoriesRequestSchema,
    type CompanyCategory,
    type Company,
  } from "../../gen/grpc/v1/toyotachikuro_pb";

  const baseUrl =
    import.meta.env.VITE_CONNECT_BASE_URL ?? "http://localhost:9090";
  const transport = createConnectTransport({
    baseUrl,
    useBinaryFormat: true,
  });
  const client = createClient(CompanyService, transport);

  let companies: Company[] = [];
  let categories: Map<number, string> = new Map();
  let isLoading = true;
  let errorMessage: string | null = null;

  const displayName = (company: Company): string =>
    company.shortName || company.mfLongName || "名称未設定";

  const displayTitle = (company: Company): string => {
    const name = company.shortName || company.mfLongName || "名称未設定";
    return `${name} (${company.id})`;
  };

  const loadCategories = async (): Promise<Map<number, string>> => {
    const request = create(GetCompanyCategoriesRequestSchema, {});
    const response = await client.getCompanyCategories(request);
    const map = new Map<number, string>();
    for (const category of response.categories ?? []) {
      map.set(category.index, category.label || "業種未設定");
    }
    return map;
  };

  const categoryLabel = (company: Company): string =>
    categories.get(company.categoryIndex) ?? "業種未設定";

  const loadCompanies = async (): Promise<void> => {
    isLoading = true;
    errorMessage = null;
    try {
      const [categoryMap, companiesResponse] = await Promise.all([
        loadCategories(),
        client.getCompanies(
          create(GetCompaniesRequestSchema, {
            forceReload: false,
          })
        ),
      ]);
      categories = categoryMap;
      const response = companiesResponse;
      const list = Object.values(response.companies ?? {});
      list.sort((a, b) => {
        const categoryCompare = (a.categoryIndex ?? 0) - (b.categoryIndex ?? 0);
        if (categoryCompare !== 0) return categoryCompare;
        return displayName(a).localeCompare(displayName(b), "ja");
      });
      companies = list;
    } catch (error) {
      errorMessage =
        error instanceof Error ? error.message : "会社一覧の取得に失敗しました";
    } finally {
      isLoading = false;
    }
  };

  onMount(() => {
    void loadCompanies();
  });
</script>

<svelte:head>
  <title>会社一覧</title>
</svelte:head>

<section class="page">
  <header class="page-header">
    <div>
      <h1>会社一覧</h1>
      <p class="lead">gRPC サーバーから取得した会社情報を表示しています。</p>
    </div>
    <button class="reload" on:click={loadCompanies} disabled={isLoading}>
      {isLoading ? "読み込み中..." : "再読み込み"}
    </button>
  </header>

  {#if errorMessage}
    <div class="state error">{errorMessage}</div>
  {:else if isLoading}
    <div class="state loading">会社情報を取得しています...</div>
  {:else if companies.length === 0}
    <div class="state empty">会社情報が見つかりませんでした。</div>
  {:else}
    <p class="count">件数: {companies.length}</p>
    <div class="table-wrap">
      <table class="company-table">
        <thead>
          <tr>
            <th>カテゴリ</th>
            <th>会社名</th>
            <th>住所</th>
            <th>電話</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {#each companies as company (company.id)}
            <tr>
              <td>{categoryLabel(company)}</td>
              <td>
                <div class="name">{displayTitle(company)}</div>
                {#if company.mfLongName && company.shortName}
                  <div class="sub">{company.mfLongName}</div>
                {/if}
              </td>
              <td>{company.mfAddress || "-"}</td>
              <td>{company.mfTel || "-"}</td>
              <td>
                <a class="action" href={`/companies/${company.id}`}>編集</a>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</section>

<style>
  :global(body) {
    font-family: "Zen Kaku Gothic New", system-ui, sans-serif;
    background: linear-gradient(120deg, #f9fafb, #eef2f7);
    color: #111827;
  }

  .page {
    max-width: 960px;
    margin: 0 auto;
    padding: 48px 20px 80px;
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 24px;
  }

  h1 {
    font-size: 2rem;
    margin: 0;
  }

  .lead {
    margin: 8px 0 0;
    color: #4b5563;
  }

  .reload {
    background: #0f172a;
    color: #ffffff;
    border: none;
    border-radius: 999px;
    padding: 10px 18px;
    font-size: 0.95rem;
    cursor: pointer;
  }

  .reload:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .state {
    padding: 18px;
    border-radius: 12px;
    background: #ffffff;
    box-shadow: 0 12px 30px rgba(15, 23, 42, 0.08);
  }

  .state.error {
    border: 1px solid #fca5a5;
    color: #991b1b;
  }

  .state.loading {
    border: 1px solid #e5e7eb;
  }

  .state.empty {
    border: 1px dashed #cbd5f5;
    color: #475569;
  }

  .count {
    margin: 0 0 12px;
    color: #4b5563;
  }

  .table-wrap {
    background: #ffffff;
    border-radius: 16px;
    box-shadow: 0 12px 30px rgba(15, 23, 42, 0.08);
    overflow: hidden;
  }

  .company-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.95rem;
  }

  .company-table thead,
  .company-table tbody {
    display: table;
    width: 100%;
    table-layout: fixed;
  }

  .company-table tbody {
    display: block;
    max-height: 520px;
    overflow-y: auto;
  }

  .company-table thead tr,
  .company-table tbody tr {
    display: table;
    width: 100%;
    table-layout: fixed;
  }

  th,
  td {
    text-align: left;
    padding: 14px 16px;
    border-bottom: 1px solid #e5e7eb;
    word-break: break-word;
  }

  th {
    background: #f8fafc;
    font-weight: 600;
  }

  tbody tr:hover {
    background: #f1f5f9;
  }

  .name {
    font-weight: 600;
  }

  .sub {
    font-size: 0.85rem;
    color: #6b7280;
    margin-top: 2px;
  }

  .mono {
    font-family: "JetBrains Mono", "SFMono-Regular", ui-monospace, monospace;
    font-size: 0.85rem;
    color: #475569;
  }

  .action {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border-radius: 999px;
    background: #e2e8f0;
    color: #0f172a;
    font-weight: 600;
    text-decoration: none;
    font-size: 0.85rem;
  }

  .action:hover {
    background: #cbd5f5;
  }

  @media (max-width: 720px) {
    .page-header {
      flex-direction: column;
      align-items: flex-start;
    }

    th:nth-child(4),
    td:nth-child(4) {
      display: none;
    }
  }
  </style>
