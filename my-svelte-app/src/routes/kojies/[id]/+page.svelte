<script lang="ts">
  import type { PageData } from "./$types";
  import { create } from "@bufbuild/protobuf";
  import { timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt";
  import { createGrpcClient } from "$lib/grpc-client";
  import { handleEnterKeyNavigation } from "$lib/form-utils";
  import { generateKojiStatus } from "$lib/koji-utils";
  import {
    KojiService,
    UpdateKojiRequestSchema,
    type Koji,
  } from "../../../gen/grpc/v1/toyotachikuro_pb";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Badge } from "$lib/components/ui/badge";
  import * as Card from "$lib/components/ui/card";
  import * as Alert from "$lib/components/ui/alert";

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

  let koji = $state<Koji | null>(null);
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
    koji = data.koji;
    if (koji) {
      form.companyName = koji.companyName ?? "";
      form.locationName = koji.locationName ?? "";
      form.start = toInputValue(koji.start);
      form.mfEnd = toInputValue(koji.mfEnd);
      initialForm = { ...form };
    }
  });

  // 未保存の変更があるかを判定
  let hasUnsavedChanges = $derived(
    initialForm !== null && JSON.stringify(form) !== JSON.stringify(initialForm)
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
          dirPath: koji.dirPath ?? "",
          companyName: form.companyName,
          locationName: form.locationName,
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

<div class="max-w-4xl mx-auto px-5 py-12">
  <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
    <div>
      <h1 class="text-3xl font-bold mb-2">工事編集</h1>
      {#if savedAt}
        <p class="text-green-700 font-semibold">
          {savedAt.getHours()}:{savedAt.getMinutes().toString().padStart(2, '0')} に保存しました
        </p>
      {:else}
        <p class="text-muted-foreground">工事情報を編集して保存できます。</p>
      {/if}
    </div>
    <Button href="/kojies" variant="outline">工事一覧に戻る</Button>
  </div>

  {#if errorMessage}
    <Alert.Root variant="destructive" class="mb-6">
      <Alert.Description>{errorMessage}</Alert.Description>
    </Alert.Root>
  {/if}

  {#if koji === null}
    <Card.Root>
      <Card.Content class="py-8 text-center text-muted-foreground">
        工事情報が見つかりませんでした。
      </Card.Content>
    </Card.Root>
  {:else}
    <Card.Root>
      <Card.Content class="pt-6">
        <form onsubmit={(e) => { e.preventDefault(); void saveKoji(); }} class="space-y-6">
          <div class="flex items-center justify-between pb-4 border-b">
            <Button 
              type="submit" 
              disabled={isSaving || !hasUnsavedChanges}
              class={hasUnsavedChanges ? "animate-pulse" : ""}
            >
              {isSaving ? "保存中..." : "保存"}
            </Button>
            <span class="text-sm text-muted-foreground">会社名: {displayName(koji)}</span>
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label class="md:text-right">状態</Label>
            <div>
              {#if generateKojiStatus(koji) === "進行中"}
                <Badge variant="destructive" class="font-bold">進行中</Badge>
              {:else if generateKojiStatus(koji) === "完了"}
                <Badge class="bg-green-600 hover:bg-green-700">完了</Badge>
              {:else if generateKojiStatus(koji) === "予定"}
                <Badge class="bg-blue-600 hover:bg-blue-700">予定</Badge>
              {:else}
                <Badge variant="secondary">不明</Badge>
              {/if}
            </div>
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="companyName" class="md:text-right">会社名</Label>
            <Input 
              id="companyName" 
              type="text" 
              bind:value={form.companyName} 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="locationName" class="md:text-right">工事場所</Label>
            <Input 
              id="locationName" 
              type="text" 
              bind:value={form.locationName} 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="start" class="md:text-right">開始</Label>
            <Input 
              id="start" 
              type="date" 
              bind:value={form.start} 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="mfEnd" class="md:text-right">終了</Label>
            <Input 
              id="mfEnd" 
              type="date" 
              bind:value={form.mfEnd} 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label class="md:text-right">ID</Label>
            <span class="text-sm font-mono text-muted-foreground">{koji.id}</span>
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label class="md:text-right">ディレクトリ</Label>
            <span class="text-sm font-mono text-muted-foreground">{koji.dirPath || "-"}</span>
          </div>
        </form>
      </Card.Content>
    </Card.Root>
  {/if}
</div>
