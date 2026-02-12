<script lang="ts">
  import type { PageData } from "./$types";
  import { create } from "@bufbuild/protobuf";
  import { createGrpcClient } from "$lib/grpc-client";
  import { handleEnterKeyNavigation } from "$lib/form-utils";
  import {
    CompanyService,
    UpdateCompanyRequestSchema,
    type Company,
    type CompanyCategory,
  } from "../../../gen/grpc/v1/toyotachikuro_pb";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import * as Card from "$lib/components/ui/card";
  import * as Alert from "$lib/components/ui/alert";

  // propsからPageDataを取得
  let { data } = $props<{ data: PageData }>();

  const client = createGrpcClient(CompanyService);

  let company = $state<Company | null>(null);
  let form = $state({
    name: "",
    categoryIndex: 0,
    prLongName: "",
    prPostalCode: "",
    prAddress: "",
    prTel: "",
    prFax: "",
    prEmail: "",
    prWebsite: "",
  });
  let initialForm = $state<typeof form | null>(null);
  let isSaving = $state(false);
  let errorMessage: string | null = $state(null);
  let savedAt: Date | null = $state(null);

  // data.companyの変更を監視してフォームを更新
  $effect(() => {
    // data.companyが存在し、変更があった場合のみ実行
    if (data.company && (!company || data.company.id !== company.id)) {
      company = data.company;
      form.name = data.company.name ?? "";
      form.categoryIndex = data.company.categoryIndex ?? 0;
      form.prLongName = data.company.prLongName ?? "";
      form.prPostalCode = data.company.prPostalCode ?? "";
      form.prAddress = data.company.prAddress ?? "";
      form.prTel = data.company.prTel ?? "";
      form.prFax = data.company.prFax ?? "";
      form.prEmail = data.company.prEmail ?? "";
      form.prWebsite = data.company.prWebsite ?? "";
      initialForm = { ...form };
    }
  });

  // 未保存の変更があるかを判定
  let hasUnsavedChanges = $derived(
    initialForm !== null && JSON.stringify(form) !== JSON.stringify(initialForm)
  );

  const displayName = (source: Company | null): string =>
    source?.name || source?.prLongName || "名称未設定";

  const saveCompany = async (): Promise<void> => {
    if (!company) return;
    isSaving = true;
    errorMessage = null;
    try {
      const request = create(UpdateCompanyRequestSchema, {
        targetId: company.id,
        sourceCompany: {
          id: company.id,
          dirPath: company.dirPath ?? "",
          ...form,
        },
      });
      const response = await client.updateCompany(request);
      // 更新後の状態をフォーム値で反映
      company = {
        ...company,
        ...form,
      };
      // 初期値を更新
      initialForm = { ...form };
      savedAt = new Date();
    } catch (error) {
      errorMessage =
        error instanceof Error ? error.message : "会社情報の更新に失敗しました";
    } finally {
      isSaving = false;
    }
  };
</script>

<svelte:head>
  <title>会社編集</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-5 py-12">
  <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
    <div>
      <h1 class="text-3xl font-bold mb-2">会社編集</h1>
      {#if savedAt}
        <p class="text-green-700 font-semibold">
          {savedAt.getHours()}:{savedAt.getMinutes().toString().padStart(2, '0')} に保存しました
        </p>
      {:else}
        <p class="text-muted-foreground">会社情報を編集して保存できます。</p>
      {/if}
    </div>
    <Button href="/companies" variant="outline">会社一覧に戻る</Button>
  </div>

  {#if errorMessage}
    <Alert.Root variant="destructive" class="mb-6">
      <Alert.Description>{errorMessage}</Alert.Description>
    </Alert.Root>
  {/if}

  {#if company === null}
    <Card.Root>
      <Card.Content class="py-8 text-center text-muted-foreground">
        会社情報が見つかりませんでした。
      </Card.Content>
    </Card.Root>
  {:else}
    <Card.Root>
      <Card.Content class="pt-6">
        <form onsubmit={(e) => { e.preventDefault(); void saveCompany(); }} class="space-y-6">
          <div class="flex items-center justify-between pb-4 border-b">
            <Button 
              type="submit" 
              disabled={isSaving || !hasUnsavedChanges}
              class={hasUnsavedChanges ? "animate-pulse bg-orange-400 hover:bg-orange-500 text-white" : ""}
            >
              {isSaving ? "保存中..." : "保存"}
            </Button>
            <span class="text-sm text-muted-foreground">会社名: {displayName(company)}</span>
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="name" class="md:text-right">会社名</Label>
            <Input 
              id="name" 
              type="text" 
              bind:value={form.name}
              placeholder="会社名" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="category" class="md:text-right">カテゴリ</Label>
            <select 
              id="category" 
              bind:value={form.categoryIndex} 
              onkeydown={handleEnterKeyNavigation}
              class="flex h-9 w-full items-center justify-between whitespace-nowrap rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
            >
              {#each data.categories as category (category.index)}
                <option value={category.index}>
                  {category.name || "業種未設定"}
                </option>
              {/each}
            </select>
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prLongName" class="md:text-right">正式名称</Label>
            <Input 
              id="prLongName" 
              type="text" 
              bind:value={form.prLongName}
              placeholder="正式名称" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prPostalCode" class="md:text-right">郵便番号</Label>
            <Input 
              id="prPostalCode" 
              type="text" 
              bind:value={form.prPostalCode}
              placeholder="郵便番号" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prAddress" class="md:text-right">住所</Label>
            <Input 
              id="prAddress" 
              type="text" 
              bind:value={form.prAddress}
              placeholder="住所" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prTel" class="md:text-right">電話</Label>
            <Input 
              id="prTel" 
              type="text" 
              bind:value={form.prTel}
              placeholder="電話番号" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prFax" class="md:text-right">FAX</Label>
            <Input 
              id="prFax" 
              type="text" 
              bind:value={form.prFax}
              placeholder="FAX" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prEmail" class="md:text-right">メール</Label>
            <Input 
              id="prEmail" 
              type="email" 
              bind:value={form.prEmail}
              placeholder="メールアドレス" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prWebsite" class="md:text-right">Web</Label>
            <Input 
              id="prWebsite" 
              type="url" 
              bind:value={form.prWebsite}
              placeholder="Webサイト" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label class="md:text-right">ID</Label>
            <span class="text-sm font-mono text-muted-foreground">{company.id}</span>
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label class="md:text-right">ディレクトリ</Label>
            <span class="text-sm font-mono text-muted-foreground">{company.dirPath || "-"}</span>
          </div>
        </form>
      </Card.Content>
    </Card.Root>
  {/if}
</div>
