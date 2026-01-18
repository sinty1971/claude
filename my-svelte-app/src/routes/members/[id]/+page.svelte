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

  let member = $state<Member | null>(data.member);
  let form = $state({
    name: "",
    mfLastName: "",
    mfFirstName: "",
    mfMiddleName: "",
    mfKanaName: "",
    mfRole: "",
    mfPostalCode: "",
    mfAddress: "",
    mfMobile: "",
    mfTel: "",
    mfFax: "",
    mfEmail: "",
    mfWebsite: "",
  });
  let initialForm = $state<typeof form | null>(null);
  let isSaving = $state(false);
  let errorMessage: string | null = $state(null);
  let savedAt: Date | null = $state(null);

  // member が変更されたときにフォームと初期値を更新
  $effect(() => {
    if (member) {
      const newForm = {
        name: member.name ?? "",
        mfLastName: member.mfLastName ?? "",
        mfFirstName: member.mfFirstName ?? "",
        mfMiddleName: member.mfMiddleName ?? "",
        mfKanaName: member.mfKanaName ?? "",
        mfRole: member.mfRole ?? "",
        mfPostalCode: member.mfPostalCode ?? "",
        mfAddress: member.mfAddress ?? "",
        mfMobile: member.mfMobile ?? "",
        mfTel: member.mfTel ?? "",
        mfFax: member.mfFax ?? "",
        mfEmail: member.mfEmail ?? "",
        mfWebsite: member.mfWebsite ?? "",
      };
      form.name = newForm.name;
      form.mfLastName = newForm.mfLastName;
      form.mfFirstName = newForm.mfFirstName;
      form.mfMiddleName = newForm.mfMiddleName;
      form.mfKanaName = newForm.mfKanaName;
      form.mfRole = newForm.mfRole;
      form.mfPostalCode = newForm.mfPostalCode;
      form.mfAddress = newForm.mfAddress;
      form.mfMobile = newForm.mfMobile;
      form.mfTel = newForm.mfTel;
      form.mfFax = newForm.mfFax;
      form.mfEmail = newForm.mfEmail;
      form.mfWebsite = newForm.mfWebsite;
      initialForm = newForm;
    }
  });

  // 未保存の変更があるかを判定
  let hasUnsavedChanges = $derived(
    initialForm !== null &&
    (form.name !== initialForm.name ||
      form.mfLastName !== initialForm.mfLastName ||
      form.mfFirstName !== initialForm.mfFirstName ||
      form.mfMiddleName !== initialForm.mfMiddleName ||
      form.mfKanaName !== initialForm.mfKanaName ||
      form.mfRole !== initialForm.mfRole ||
      form.mfPostalCode !== initialForm.mfPostalCode ||
      form.mfAddress !== initialForm.mfAddress ||
      form.mfMobile !== initialForm.mfMobile ||
      form.mfTel !== initialForm.mfTel ||
      form.mfFax !== initialForm.mfFax ||
      form.mfEmail !== initialForm.mfEmail ||
      form.mfWebsite !== initialForm.mfWebsite)
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
          name: form.name,
          mfLastName: form.mfLastName,
          mfFirstName: form.mfFirstName,
          mfMiddleName: form.mfMiddleName,
          mfKanaName: form.mfKanaName,
          mfRole: form.mfRole,
          mfPostalCode: form.mfPostalCode,
          mfAddress: form.mfAddress,
          mfMobile: form.mfMobile,
          mfTel: form.mfTel,
          mfFax: form.mfFax,
          mfEmail: form.mfEmail,
          mfWebsite: form.mfWebsite,
        },
      });
      const response = await client.updateMember(request);
      // 更新後の状態をフォーム値で反映
      member = {
        ...member,
        name: form.name,
        mfLastName: form.mfLastName,
        mfFirstName: form.mfFirstName,
        mfMiddleName: form.mfMiddleName,
        mfKanaName: form.mfKanaName,
        mfRole: form.mfRole,
        mfPostalCode: form.mfPostalCode,
        mfAddress: form.mfAddress,
        mfMobile: form.mfMobile,
        mfTel: form.mfTel,
        mfFax: form.mfFax,
        mfEmail: form.mfEmail,
        mfWebsite: form.mfWebsite,
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
              class={hasUnsavedChanges ? "animate-pulse" : ""}
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
            <Label for="mfLastName" class="md:text-right">姓</Label>
            <Input 
              id="mfLastName" 
              type="text" 
              bind:value={form.mfLastName}
              placeholder="姓" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="mfFirstName" class="md:text-right">名</Label>
            <Input 
              id="mfFirstName" 
              type="text" 
              bind:value={form.mfFirstName}
              placeholder="名" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="mfMiddleName" class="md:text-right">ミドルネーム</Label>
            <Input 
              id="mfMiddleName" 
              type="text" 
              bind:value={form.mfMiddleName}
              placeholder="ミドルネーム" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="mfKanaName" class="md:text-right">カナ名</Label>
            <Input 
              id="mfKanaName" 
              type="text" 
              bind:value={form.mfKanaName}
              placeholder="カナ名" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="mfRole" class="md:text-right">役職</Label>
            <Input 
              id="mfRole" 
              type="text" 
              bind:value={form.mfRole}
              placeholder="役職" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="mfPostalCode" class="md:text-right">郵便番号</Label>
            <Input 
              id="mfPostalCode" 
              type="text" 
              bind:value={form.mfPostalCode}
              placeholder="郵便番号" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="mfAddress" class="md:text-right">住所</Label>
            <Input 
              id="mfAddress" 
              type="text" 
              bind:value={form.mfAddress}
              placeholder="住所" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="mfMobile" class="md:text-right">携帯</Label>
            <Input 
              id="mfMobile" 
              type="text" 
              bind:value={form.mfMobile}
              placeholder="携帯" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="mfTel" class="md:text-right">電話</Label>
            <Input 
              id="mfTel" 
              type="text" 
              bind:value={form.mfTel}
              placeholder="電話" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="mfFax" class="md:text-right">FAX</Label>
            <Input 
              id="mfFax" 
              type="text" 
              bind:value={form.mfFax}
              placeholder="FAX" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="mfEmail" class="md:text-right">メール</Label>
            <Input 
              id="mfEmail" 
              type="email" 
              bind:value={form.mfEmail}
              placeholder="メール" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>

          <div class="grid md:grid-cols-[140px_1fr] gap-4 items-center">
            <Label for="mfWebsite" class="md:text-right">ウェブサイト</Label>
            <Input 
              id="mfWebsite" 
              type="url" 
              bind:value={form.mfWebsite}
              placeholder="ウェブサイト" 
              onkeydown={handleEnterKeyNavigation} 
            />
          </div>
        </form>
      </Card.Content>
    </Card.Root>
  {/if}
</div>
