<script lang="ts">
  import type { PageData } from "./$types";
  import { create } from "@bufbuild/protobuf";
  import { createGrpcClient } from "$lib/grpc-client";
  import { handleEnterKeyNavigation } from "$lib/form-utils";
  import {
    MemberService,
    UpdateMemberRequestSchema,
    type Member,
  } from "../../../gen/grpc/v1/toyotachikuro_pb";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import * as Card from "$lib/components/ui/card";
  import * as Alert from "$lib/components/ui/alert";

  let { data } = $props<{ data: PageData }>();

  const client = createGrpcClient(MemberService);

  let member = $state<Member | null>(null);
  let form = $state({
    name: "",
    prLastName: "",
    prFirstName: "",
    prMiddleName: "",
    prKanaName: "",
    prRole: "",
    prBirthdate: "",
    prBloodType: "",
    prPostalCode: "",
    prAddress: "",
    prMobile: "",
    prTel: "",
    prFax: "",
    prEmail: "",
    prWebsite: "",
  });
  let initialForm = $state<typeof form | null>(null);

  // 初期化は$effect内でのみ行う（Svelteのリアクティブ仕様に準拠）
  let isSaving = $state(false);
  let errorMessage: string | null = $state(null);
  let savedAt: Date | null = $state(null);

  $effect(() => {
    if (data.member && (!member || data.member.id !== member.id)) {
      member = data.member;
      form.name = data.member.name ?? "";
      form.prLastName = data.member.prLastName ?? "";
      form.prFirstName = data.member.prFirstName ?? "";
      form.prMiddleName = data.member.prMiddleName ?? "";
      form.prKanaName = data.member.prKanaName ?? "";
      form.prRole = data.member.prRole ?? "";
      form.prBirthdate = data.member.prBirthdate ?? "";
      form.prBloodType = data.member.prBloodType ?? "";
      form.prPostalCode = data.member.prPostalCode ?? "";
      form.prAddress = data.member.prAddress ?? "";
      form.prMobile = data.member.prMobile ?? "";
      form.prTel = data.member.prTel ?? "";
      form.prFax = data.member.prFax ?? "";
      form.prEmail = data.member.prEmail ?? "";
      form.prWebsite = data.member.prWebsite ?? "";
      initialForm = { ...form };
    }
  });

  // 未保存の変更があるかを判定
  let hasUnsavedChanges = $derived(
    initialForm !== null && JSON.stringify(form) !== JSON.stringify(initialForm)
  );

  const displayName = (source: Member | null): string =>
    source?.name || "名称未設定";

  const saveMember = async (): Promise<void> => {
    if (!member) return;
    isSaving = true;
    errorMessage = null;
    try {
      const request = create(UpdateMemberRequestSchema, {
        targetId: member.id,
        sourceMember: {
          id: member.id,
          ...form,
        },
      });
      const response = await client.updateMember(request);
      // 更新後の状態をフォーム値で反映
      member = {
        ...member,
        ...form,
      };
      // 初期値を更新
      initialForm = { ...form };
      savedAt = new Date();
    } catch (error) {
      errorMessage =
        error instanceof Error ? error.message : "メンバー情報の更新に失敗しました";
    } finally {
      isSaving = false;
    }
  };
</script>

<svelte:head>
  <title>メンバー編集</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-5 py-12">
  <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
    <div>
      <h1 class="text-3xl font-bold mb-2">メンバー編集</h1>
      {#if savedAt}
        <p class="text-green-700 font-semibold">
          {savedAt.getHours()}:{savedAt.getMinutes().toString().padStart(2, '0')} に保存しました
        </p>
      {:else}
        <p class="text-muted-foreground">メンバー情報を編集して保存できます。</p>
      {/if}
    </div>
    <Button href="/members" variant="outline">メンバー一覧に戻る</Button>
  </div>

  {#if errorMessage}
    <Alert.Root variant="destructive" class="mb-6">
      <Alert.Description>{errorMessage}</Alert.Description>
    </Alert.Root>
  {/if}

  {#if member === null}
    <Card.Root>
      <Card.Content class="py-8 text-center text-muted-foreground">
        メンバー情報が見つかりませんでした。
      </Card.Content>
    </Card.Root>
  {:else}
    <Card.Root>
      <Card.Content class="pt-6">
        <form onsubmit={(e) => { e.preventDefault(); void saveMember(); }} class="space-y-6">
          <div class="flex items-center justify-between pb-4 border-b">
            <Button 
              type="submit" 
              disabled={isSaving || !hasUnsavedChanges}
              class={hasUnsavedChanges ? "animate-pulse bg-orange-400 hover:bg-orange-500 text-white" : ""}
            >
              {isSaving ? "保存中..." : "保存"}
            </Button>
            <span class="text-sm text-muted-foreground">メンバー名: {displayName(member)}</span>
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="name" class="md:text-right">名前</Label>
            <Input 
              id="name" 
              type="text" 
              bind:value={form.name}
              placeholder="名前" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prLastName" class="md:text-right">姓</Label>
            <Input 
              id="prLastName" 
              type="text" 
              bind:value={form.prLastName}
              placeholder="姓" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prFirstName" class="md:text-right">名</Label>
            <Input 
              id="prFirstName" 
              type="text" 
              bind:value={form.prFirstName}
              placeholder="名" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prMiddleName" class="md:text-right">ミドルネーム</Label>
            <Input 
              id="prMiddleName" 
              type="text" 
              bind:value={form.prMiddleName}
              placeholder="ミドルネーム" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prKanaName" class="md:text-right">カナ名</Label>
            <Input 
              id="prKanaName" 
              type="text" 
              bind:value={form.prKanaName}
              placeholder="カナ名" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prRole" class="md:text-right">役職</Label>
            <Input 
              id="prRole" 
              type="text" 
              bind:value={form.prRole}
              placeholder="役職" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prBirthdate" class="md:text-right">生年月日</Label>
            <Input 
              id="prBirthdate" 
              type="date" 
              bind:value={form.prBirthdate}
              placeholder="生年月日" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prBloodType" class="md:text-right">血液型</Label>
            <Input 
              id="prBloodType" 
              type="text" 
              bind:value={form.prBloodType}
              placeholder="血液型" 
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
            <Label for="prMobile" class="md:text-right">携帯</Label>
            <Input 
              id="prMobile" 
              type="text" 
              bind:value={form.prMobile}
              placeholder="携帯" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prTel" class="md:text-right">電話</Label>
            <Input 
              id="prTel" 
              type="text" 
              bind:value={form.prTel}
              placeholder="電話" 
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
              placeholder="メール" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="prWebsite" class="md:text-right">ウェブサイト</Label>
            <Input 
              id="prWebsite" 
              type="url" 
              bind:value={form.prWebsite}
              placeholder="ウェブサイト" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>
        </form>
      </Card.Content>
    </Card.Root>
  {/if}
</div>
