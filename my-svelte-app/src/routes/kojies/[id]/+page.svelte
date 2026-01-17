<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { create } from "@bufbuild/protobuf";
  import { timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt";
  import { createClient } from "@connectrpc/connect";
  import { createConnectTransport } from "@connectrpc/connect-web";
  import {
    GetKojiRequestSchema,
    KojiService,
    UpdateKojiRequestSchema,
    type Koji,
  } from "../../../gen/grpc/v1/toyotachikuro_pb";

  const baseUrl =
    import.meta.env.VITE_CONNECT_BASE_URL ?? "http://localhost:9090";
  const transport = createConnectTransport({
    baseUrl,
    useBinaryFormat: true,
  });
  const client = createClient(KojiService, transport);

  let koji: Koji | null = null;
  let form = {
    status: "",
    companyName: "",
    locationName: "",
    start: "",
    mfEnd: "",
  };
  let isLoading = true;
  let isSaving = false;
  let errorMessage: string | null = null;
  let successMessage: string | null = null;

  const displayName = (source: Koji | null): string =>
    source?.companyName || "名称未設定";

  const toInputValue = (value?: Koji["start"]): string => {
    if (!value) return "";
    const date = timestampDate(value);
    if (Number.isNaN(date.getTime())) return "";
    const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
    return local.toISOString().slice(0, 10);
  };

  const toTimestamp = (value: string): Koji["start"] | undefined => {
    if (!value) return undefined;
    const date = new Date(`${value}T00:00`);
    if (Number.isNaN(date.getTime())) return undefined;
    return timestampFromDate(date);
  };

  const loadKoji = async (id: string): Promise<void> => {
    isLoading = true;
    errorMessage = null;
    try {
      const request = create(GetKojiRequestSchema, { targetId: id });
      const response = await client.getKoji(request);
      koji = response.koji ?? null;
      if (koji) {
        form = {
          status: koji.status ?? "",
          companyName: koji.companyName ?? "",
          locationName: koji.locationName ?? "",
          start: toInputValue(koji.start),
          mfEnd: toInputValue(koji.mfEnd),
        };
      }
    } catch (error) {
      errorMessage =
        error instanceof Error ? error.message : "工事情報の取得に失敗しました";
    } finally {
      isLoading = false;
    }
  };

  const saveKoji = async (): Promise<void> => {
    if (!koji) return;
    isSaving = true;
    errorMessage = null;
    successMessage = null;
    try {
      const request = create(UpdateKojiRequestSchema, {
        targetId: koji.id,
        sourceKoji: {
          id: koji.id,
          status: form.status,
          companyName: form.companyName,
          locationName: form.locationName,
          dirPath: koji.dirPath ?? "",
          start: toTimestamp(form.start),
          mfEnd: toTimestamp(form.mfEnd),
        },
      });
      const response = await client.updateKoji(request);
      koji = response.prevKoji ?? koji;
      successMessage = "工事情報を更新しました。";
    } catch (error) {
      errorMessage =
        error instanceof Error ? error.message : "工事情報の更新に失敗しました";
    } finally {
      isSaving = false;
    }
  };

  onMount(() => {
    const id = page.params.id;
    if (id) {
      void loadKoji(id);
    } else {
      isLoading = false;
      errorMessage = "工事IDが指定されていません。";
    }
  });
</script>

<svelte:head>
  <title>工事編集</title>
</svelte:head>

<section class="page">
  <header class="page-header">
    <div>
      <h1>工事編集</h1>
      <p class="lead">工事情報を編集して保存できます。</p>
    </div>
    <a class="back" href="/kojies">工事一覧に戻る</a>
  </header>

  {#if errorMessage}
    <div class="state error">{errorMessage}</div>
  {:else if successMessage}
    <div class="state success">{successMessage}</div>
  {:else if isLoading}
    <div class="state loading">工事情報を取得しています...</div>
  {:else if koji === null}
    <div class="state empty">工事情報が見つかりませんでした。</div>
  {:else}
    <form class="detail" on:submit|preventDefault={saveKoji}>
      <div class="detail-row">
        <label class="label" for="status">状態</label>
        <input id="status" type="text" bind:value={form.status} />
      </div>
      <div class="detail-row">
        <label class="label" for="companyName">会社名</label>
        <input id="companyName" type="text" bind:value={form.companyName} />
      </div>
      <div class="detail-row">
        <label class="label" for="locationName">工事場所</label>
        <input id="locationName" type="text" bind:value={form.locationName} />
      </div>
      <div class="detail-row">
        <label class="label" for="start">開始</label>
        <input id="start" type="date" bind:value={form.start} />
      </div>
      <div class="detail-row">
        <label class="label" for="mfEnd">終了</label>
        <input id="mfEnd" type="date" bind:value={form.mfEnd} />
      </div>
      <div class="detail-row">
        <span class="label">ID</span>
        <span class="value mono">{koji.id}</span>
      </div>
      <div class="detail-row">
        <span class="label">ディレクトリ</span>
        <span class="value mono">{koji.dirPath || "-"}</span>
      </div>
      <div class="actions">
        <button class="save" type="submit" disabled={isSaving}>
          {isSaving ? "保存中..." : "保存"}
        </button>
        <span class="hint">会社名: {displayName(koji)}</span>
      </div>
    </form>
  {/if}
</section>

<style>
  :global(body) {
    font-family: "Zen Kaku Gothic New", system-ui, sans-serif;
    background: linear-gradient(120deg, #f9fafb, #eef2f7);
    color: #111827;
  }

  .page {
    max-width: 860px;
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

  .back {
    background: #0f172a;
    color: #ffffff;
    border: none;
    border-radius: 999px;
    padding: 10px 18px;
    font-size: 0.95rem;
    text-decoration: none;
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

  .state.success {
    border: 1px solid #86efac;
    color: #166534;
  }

  .state.loading {
    border: 1px solid #e5e7eb;
  }

  .state.empty {
    border: 1px dashed #cbd5f5;
    color: #475569;
  }

  .detail {
    background: #ffffff;
    border-radius: 16px;
    box-shadow: 0 12px 30px rgba(15, 23, 42, 0.08);
    padding: 20px;
    display: grid;
    gap: 12px;
  }

  .detail-row {
    display: grid;
    grid-template-columns: 120px 1fr;
    gap: 12px;
    align-items: center;
  }

  .label {
    font-weight: 600;
    color: #475569;
  }

  .value {
    color: #111827;
  }

  input {
    border: 1px solid #e2e8f0;
    border-radius: 10px;
    padding: 10px 12px;
    font-size: 0.95rem;
    background: #f8fafc;
  }

  input:focus {
    outline: 2px solid #94a3b8;
    outline-offset: 2px;
  }

  .actions {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 8px;
  }

  .save {
    background: #0f172a;
    color: #ffffff;
    border: none;
    border-radius: 999px;
    padding: 10px 18px;
    font-size: 0.95rem;
    cursor: pointer;
  }

  .save:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .hint {
    color: #64748b;
    font-size: 0.9rem;
  }

  .mono {
    font-family: "JetBrains Mono", "SFMono-Regular", ui-monospace, monospace;
    font-size: 0.9rem;
    color: #475569;
  }

  @media (max-width: 720px) {
    .page-header {
      flex-direction: column;
      align-items: flex-start;
    }

    .detail-row {
      grid-template-columns: 1fr;
    }
  }
  </style>
