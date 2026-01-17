<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { create } from "@bufbuild/protobuf";
  import { createClient } from "@connectrpc/connect";
  import { createConnectTransport } from "@connectrpc/connect-web";
  import {
    CompanyService,
    GetCompanyCategoriesRequestSchema,
    GetCompanyRequestSchema,
    UpdateCompanyRequestSchema,
    type Company,
    type CompanyCategory,
  } from "../../../gen/grpc/v1/toyotachikuro_pb";

  const baseUrl =
    import.meta.env.VITE_CONNECT_BASE_URL ?? "http://localhost:9090";
  const transport = createConnectTransport({
    baseUrl,
    useBinaryFormat: true,
  });
  const client = createClient(CompanyService, transport);

  let company: Company | null = null;
  let categories: CompanyCategory[] = [];
  let form = {
    shortName: "",
    categoryIndex: 0,
    mfLongName: "",
    mfPostalCode: "",
    mfAddress: "",
    mfTel: "",
    mfFax: "",
    mfEmail: "",
    mfWebsite: "",
  };
  let isLoading = true;
  let isSaving = false;
  let errorMessage: string | null = null;
  let successMessage: string | null = null;

  const displayName = (source: Company | null): string =>
    source?.shortName || source?.mfLongName || "名称未設定";

  const loadCompany = async (id: string): Promise<void> => {
    isLoading = true;
    errorMessage = null;
    try {
      const [companyResponse, categoryResponse] = await Promise.all([
        client.getCompany(create(GetCompanyRequestSchema, { targetId: id })),
        client.getCompanyCategories(create(GetCompanyCategoriesRequestSchema, {})),
      ]);
      company = companyResponse.company ?? null;
      categories = categoryResponse.categories ?? [];
      if (company) {
        form = {
          shortName: company.shortName ?? "",
          categoryIndex: company.categoryIndex ?? 0,
          mfLongName: company.mfLongName ?? "",
          mfPostalCode: company.mfPostalCode ?? "",
          mfAddress: company.mfAddress ?? "",
          mfTel: company.mfTel ?? "",
          mfFax: company.mfFax ?? "",
          mfEmail: company.mfEmail ?? "",
          mfWebsite: company.mfWebsite ?? "",
        };
      }
    } catch (error) {
      errorMessage =
        error instanceof Error ? error.message : "会社情報の取得に失敗しました";
    } finally {
      isLoading = false;
    }
  };

  const saveCompany = async (): Promise<void> => {
    if (!company) return;
    isSaving = true;
    errorMessage = null;
    successMessage = null;
    try {
      const request = create(UpdateCompanyRequestSchema, {
        targetId: company.id,
        sourceCompany: {
          id: company.id,
          shortName: form.shortName,
          categoryIndex: form.categoryIndex,
          dirPath: company.dirPath ?? "",
          mfLongName: form.mfLongName,
          mfPostalCode: form.mfPostalCode,
          mfAddress: form.mfAddress,
          mfTel: form.mfTel,
          mfFax: form.mfFax,
          mfEmail: form.mfEmail,
          mfWebsite: form.mfWebsite,
        },
      });
      const response = await client.updateCompany(request);
      company = response.prevCompany ?? company;
      successMessage = "会社情報を更新しました。";
    } catch (error) {
      errorMessage =
        error instanceof Error ? error.message : "会社情報の更新に失敗しました";
    } finally {
      isSaving = false;
    }
  };

  onMount(() => {
    const id = page.params.id;
    if (id) {
      void loadCompany(id);
    } else {
      isLoading = false;
      errorMessage = "会社IDが指定されていません。";
    }
  });
</script>

<svelte:head>
  <title>会社編集</title>
</svelte:head>

<section class="page">
  <header class="page-header">
    <div>
      <h1>会社編集</h1>
      <p class="lead">
        会社情報を編集して保存できます。
      </p>
    </div>
    <a class="back" href="/companies">会社一覧に戻る</a>
  </header>

  {#if errorMessage}
    <div class="state error">{errorMessage}</div>
  {:else if successMessage}
    <div class="state success">{successMessage}</div>
  {:else if isLoading}
    <div class="state loading">会社情報を取得しています...</div>
  {:else if company === null}
    <div class="state empty">会社情報が見つかりませんでした。</div>
  {:else}
    <form class="detail" on:submit|preventDefault={saveCompany}>
      <div class="detail-row">
        <label class="label" for="shortName">短縮名</label>
        <input
          id="shortName"
          type="text"
          bind:value={form.shortName}
          placeholder="短縮名"
        />
      </div>
      <div class="detail-row">
        <label class="label" for="category">カテゴリ</label>
        <select id="category" bind:value={form.categoryIndex}>
          {#each categories as category (category.index)}
            <option value={category.index}>
              {category.label || "業種未設定"}
            </option>
          {/each}
        </select>
      </div>
      <div class="detail-row">
        <label class="label" for="mfLongName">正式名称</label>
        <input
          id="mfLongName"
          type="text"
          bind:value={form.mfLongName}
          placeholder="正式名称"
        />
      </div>
      <div class="detail-row">
        <label class="label" for="mfPostalCode">郵便番号</label>
        <input
          id="mfPostalCode"
          type="text"
          bind:value={form.mfPostalCode}
          placeholder="郵便番号"
        />
      </div>
      <div class="detail-row">
        <label class="label" for="mfAddress">住所</label>
        <input
          id="mfAddress"
          type="text"
          bind:value={form.mfAddress}
          placeholder="住所"
        />
      </div>
      <div class="detail-row">
        <label class="label" for="mfTel">電話</label>
        <input
          id="mfTel"
          type="text"
          bind:value={form.mfTel}
          placeholder="電話番号"
        />
      </div>
      <div class="detail-row">
        <label class="label" for="mfFax">FAX</label>
        <input
          id="mfFax"
          type="text"
          bind:value={form.mfFax}
          placeholder="FAX"
        />
      </div>
      <div class="detail-row">
        <label class="label" for="mfEmail">メール</label>
        <input
          id="mfEmail"
          type="email"
          bind:value={form.mfEmail}
          placeholder="メールアドレス"
        />
      </div>
      <div class="detail-row">
        <label class="label" for="mfWebsite">Web</label>
        <input
          id="mfWebsite"
          type="url"
          bind:value={form.mfWebsite}
          placeholder="Webサイト"
        />
      </div>
      <div class="detail-row">
        <span class="label">ID</span>
        <span class="value mono">{company.id}</span>
      </div>
      <div class="detail-row">
        <span class="label">ディレクトリ</span>
        <span class="value mono">{company.dirPath || "-"}</span>
      </div>
      <div class="actions">
        <button class="save" type="submit" disabled={isSaving}>
          {isSaving ? "保存中..." : "保存"}
        </button>
        <span class="hint">会社名: {displayName(company)}</span>
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

  input,
  select {
    border: 1px solid #e2e8f0;
    border-radius: 10px;
    padding: 10px 12px;
    font-size: 0.95rem;
    background: #f8fafc;
  }

  input:focus,
  select:focus {
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
