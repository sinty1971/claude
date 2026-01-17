<script lang="ts">
  import { onMount } from "svelte";
  import { create } from "@bufbuild/protobuf";
  import { timestampDate } from "@bufbuild/protobuf/wkt";
  import { createClient } from "@connectrpc/connect";
  import { createConnectTransport } from "@connectrpc/connect-web";
  import {
    GetKojiesRequestSchema,
    KojiService,
    type Koji,
  } from "../../gen/grpc/v1/toyotachikuro_pb";

  const baseUrl =
    import.meta.env.VITE_CONNECT_BASE_URL ?? "http://localhost:9090";
  const transport = createConnectTransport({
    baseUrl,
    useBinaryFormat: true,
  });
  const client = createClient(KojiService, transport);

  let kojies: Koji[] = [];
  let isLoading = true;
  let errorMessage: string | null = null;

  const displayDate = (value?: Koji["start"]): string => {
    if (!value) return "-";
    const date = timestampDate(value);
    const time = date.getTime();
    if (Number.isNaN(time)) return "-";
    return date.toLocaleDateString("ja-JP");
  };

  const loadKojies = async (): Promise<void> => {
    isLoading = true;
    errorMessage = null;
    try {
      const request = create(GetKojiesRequestSchema, {});
      const response = await client.getKojies(request);
      const list = Object.values(response.kojies ?? {});
      list.sort((a, b) => {
        const aTime = a.start ? timestampDate(a.start).getTime() : -Infinity;
        const bTime = b.start ? timestampDate(b.start).getTime() : -Infinity;
        if (aTime !== bTime) return bTime - aTime;
        return (a.companyName || "").localeCompare(b.companyName || "", "ja");
      });
      kojies = list;
    } catch (error) {
      errorMessage =
        error instanceof Error ? error.message : "工事一覧の取得に失敗しました";
    } finally {
      isLoading = false;
    }
  };

  onMount(() => {
    void loadKojies();
  });
</script>

<svelte:head>
  <title>工事一覧</title>
</svelte:head>

<section class="page">
  <header class="page-header">
    <div>
      <h1>工事一覧</h1>
      <p class="lead">gRPC サーバーから取得した工事情報を表示しています。</p>
    </div>
    <button class="reload" on:click={loadKojies} disabled={isLoading}>
      {isLoading ? "読み込み中..." : "再読み込み"}
    </button>
  </header>

  {#if errorMessage}
    <div class="state error">{errorMessage}</div>
  {:else if isLoading}
    <div class="state loading">工事情報を取得しています...</div>
  {:else if kojies.length === 0}
    <div class="state empty">工事情報が見つかりませんでした。</div>
  {:else}
    <p class="count">件数: {kojies.length}</p>
    <div class="table-wrap">
      <table class="koji-table">
        <thead>
          <tr>
            <th>会社名</th>
            <th>工事場所</th>
            <th>状態</th>
            <th>開始</th>
            <th>終了</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {#each kojies as koji (koji.id)}
            <tr>
              <td>
                <div class="name">{koji.companyName || "-"}</div>
                <div class="sub">ID: {koji.id}</div>
              </td>
              <td>{koji.locationName || "-"}</td>
              <td>{koji.status || "-"}</td>
              <td>{displayDate(koji.start)}</td>
              <td>{displayDate(koji.mfEnd)}</td>
              <td>
                <a class="action" href={`/kojies/${koji.id}`}>編集</a>
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

  .koji-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.95rem;
  }

  .koji-table thead,
  .koji-table tbody {
    display: table;
    width: 100%;
    table-layout: fixed;
  }

  .koji-table tbody {
    display: block;
    max-height: 520px;
    overflow-y: auto;
  }

  .koji-table thead tr,
  .koji-table tbody tr {
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
    td:nth-child(4),
    th:nth-child(5),
    td:nth-child(5) {
      display: none;
    }
  }
  </style>
