<script lang="ts">
  import type { PageData } from "./$types";
  import { create } from "@bufbuild/protobuf";
  import { timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt";
  import { createGrpcClient } from "$lib/grpc-client";
  import { handleEnterKeyNavigation } from "$lib/form-utils";
  import { generateKojiStatus, getKojiStatusClass } from "$lib/koji-utils";
  import {
    KojiService,
    UpdateKojiRequestSchema,
    type Koji,
  } from "../../../gen/grpc/v1/toyotachikuro_pb";

  let { data } = $props<{ data: PageData }>();

  const client = createGrpcClient(KojiService);

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

  let koji = $state<Koji | null>(data.koji);
  let form = $state({
    companyName: "",
    locationName: "",
    start: "",
    mfEnd: "",
  });
  let initialForm = $state<typeof form | null>(null);
  let isSaving = $state(false);
  let errorMessage: string | null = $state(null);
  let savedAt: Date | null = $state(null);

  // koji が変更されたときにフォームと初期値を更新
  $effect(() => {
    if (koji) {
      const newForm = {
        companyName: koji.companyName ?? "",
        locationName: koji.locationName ?? "",
        start: toInputValue(koji.start),
        mfEnd: toInputValue(koji.mfEnd),
      };
      form.companyName = newForm.companyName;
      form.locationName = newForm.locationName;
      form.start = newForm.start;
      form.mfEnd = newForm.mfEnd;
      initialForm = newForm;
    }
  });

  // 未保存の変更があるかを判定
  let hasUnsavedChanges = $derived(
    initialForm !== null &&
    (form.companyName !== initialForm.companyName ||
      form.locationName !== initialForm.locationName ||
      form.start !== initialForm.start ||
      form.mfEnd !== initialForm.mfEnd)
  );

  const displayName = (source: Koji | null): string =>
    source?.companyName || "名称未設定";

  const saveKoji = async (): Promise<void> => {
    if (!koji) return;
    isSaving = true;
    errorMessage = null;
    try {
      const request = create(UpdateKojiRequestSchema, {
        targetId: koji.id,
        sourceKoji: {
          id: koji.id,
          companyName: form.companyName,
          locationName: form.locationName,
          dirPath: koji.dirPath ?? "",
          start: toTimestamp(form.start),
          mfEnd: toTimestamp(form.mfEnd),
        },
      });
      const response = await client.updateKoji(request);
      // 更新後の状態をフォーム値で反映
      koji = {
        ...koji,
        companyName: form.companyName,
        locationName: form.locationName,
        start: toTimestamp(form.start),
        mfEnd: toTimestamp(form.mfEnd),
      };
      // 初期値を更新
      initialForm = { ...form };
      savedAt = new Date();
    } catch (error) {
      errorMessage =
        error instanceof Error ? error.message : "工事情報の更新に失敗しました";
    } finally {
      isSaving = false;
    }
  };
</script>

<svelte:head>
  <title>工事編集</title>
</svelte:head>

<section class="page">
  <header class="page-header">
    <div>
      <h1>工事編集</h1>
      {#if savedAt}
        <p class="lead success-message">
          {savedAt.getHours()}:{savedAt.getMinutes().toString().padStart(2, '0')} に保存しました
        </p>
      {:else}
        <p class="lead">工事情報を編集して保存できます。</p>
      {/if}
    </div>
    <a class="back" href="/kojies">工事一覧に戻る</a>
  </header>

  {#if errorMessage}
    <div class="state error">{errorMessage}</div>
  {/if}

  {#if koji === null}
    <div class="state empty">工事情報が見つかりませんでした。</div>
  {:else}
    <form class="detail" onsubmit={(e) => { e.preventDefault(); void saveKoji(); }}>
      <div class="actions top">
        <button class="save" class:unsaved={hasUnsavedChanges} type="submit" disabled={isSaving || !hasUnsavedChanges}>
          {isSaving ? "保存中..." : "保存"}
        </button>
        <span class="hint">会社名: {displayName(koji)}</span>
      </div>
      <div class="detail-row">
        <span class="label">状態</span>
        <span class="value">
          <span class="status {getKojiStatusClass(generateKojiStatus(koji))}">{generateKojiStatus(koji)}</span>
        </span>
      </div>
      <div class="detail-row">
        <label class="label" for="companyName">会社名</label>
        <input id="companyName" type="text" bind:value={form.companyName} onkeydown={handleEnterKeyNavigation} />
      </div>
      <div class="detail-row">
        <label class="label" for="locationName">工事場所</label>
        <input id="locationName" type="text" bind:value={form.locationName} onkeydown={handleEnterKeyNavigation} />
      </div>
      <div class="detail-row">
        <label class="label" for="start">開始</label>
        <input id="start" type="date" bind:value={form.start} onkeydown={handleEnterKeyNavigation} />
      </div>
      <div class="detail-row">
        <label class="label" for="mfEnd">終了</label>
        <input id="mfEnd" type="date" bind:value={form.mfEnd} onkeydown={handleEnterKeyNavigation} />
      </div>
      <div class="detail-row">
        <span class="label">ID</span>
        <span class="value mono">{koji.id}</span>
      </div>
      <div class="detail-row">
        <span class="label">ディレクトリ</span>
        <span class="value mono">{koji.dirPath || "-"}</span>
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

  .lead.success-message {
    color: #166534;
    font-weight: 600;
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

  .actions.top {
    margin-top: 0;
    margin-bottom: 16px;
    padding-bottom: 16px;
    border-bottom: 1px solid #e2e8f0;
  }

  .save {
    background: #0f172a;
    color: #ffffff;
    border: none;
    border-radius: 999px;
    padding: 10px 18px;
    font-size: 0.95rem;
    cursor: pointer;
    transition: all 0.3s ease;
  }

  .save.unsaved {
    animation: pulse 2s ease-in-out infinite;
    box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.7);
  }

  @keyframes pulse {
    0%, 100% {
      box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.7);
    }
    50% {
      box-shadow: 0 0 0 8px rgba(59, 130, 246, 0);
    }
  }

  .save:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .hint {
    color: #64748b;
    font-size: 0.9rem;
  }

  .status {
    display: inline-block;
    padding: 4px 12px;
    border-radius: 6px;
    font-size: 0.85rem;
    font-weight: 600;
  }

  .status-active {
    color: #dc2626;
    font-weight: 700;
  }

  .status-completed {
    color: #16a34a;
  }

  .status-scheduled {
    color: #2563eb;
  }

  .status-unknown {
    color: #6b7280;
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
