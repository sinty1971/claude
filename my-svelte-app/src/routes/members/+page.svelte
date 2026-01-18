<script lang="ts">
  import type { PageData } from "./$types";
  import type { Member } from "../../gen/grpc/v1/toyotachikuro_pb";
  import { Button } from "$lib/components/ui/button";
  import * as Table from "$lib/components/ui/table";
  import * as Card from "$lib/components/ui/card";

  let { data } = $props<{ data: PageData }>();

  let members = $derived(data.members);

  const displayName = (member: Member): string =>
    member.name || "名称未設定";

  const displayTitle = (member: Member): string => {
    const name = member.name || "名称未設定";
    return `${name} (${member.id})`;
  };
</script>

<svelte:head>
  <title>メンバー一覧</title>
</svelte:head>

<div class="max-w-6xl mx-auto px-5 py-12">
  <div class="mb-6">
    <h1 class="text-3xl font-bold mb-2">メンバー一覧</h1>
    <p class="text-muted-foreground">gRPC サーバーから取得したメンバー情報を表示しています。</p>
  </div>

  {#if members.length === 0}
    <Card.Root>
      <Card.Content class="py-8 text-center text-muted-foreground">
        メンバー情報が見つかりませんでした。
      </Card.Content>
    </Card.Root>
  {:else}
    <p class="text-sm text-muted-foreground mb-3">件数: {members.length}</p>
    <Card.Root>
      <Table.Root>
        <Table.Header>
          <Table.Row>
            <Table.Head>名前</Table.Head>
            <Table.Head class="hidden md:table-cell">役職</Table.Head>
            <Table.Head class="hidden md:table-cell">メール</Table.Head>
            <Table.Head class="hidden lg:table-cell">電話</Table.Head>
            <Table.Head class="w-30">操作</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {#each members as member (member.id)}
            <Table.Row>
              <Table.Cell>
                <div class="font-semibold">{displayTitle(member)}</div>
                {#if member.mfKanaName}
                  <div class="text-sm text-muted-foreground mt-0.5">{member.mfKanaName}</div>
                {/if}
              </Table.Cell>
              <Table.Cell class="hidden md:table-cell">{member.mfRole || "-"}</Table.Cell>
              <Table.Cell class="hidden md:table-cell">{member.mfEmail || "-"}</Table.Cell>
              <Table.Cell class="hidden lg:table-cell">{member.mfMobile || member.mfTel || "-"}</Table.Cell>
              <Table.Cell>
                <Button variant="ghost" size="sm" href={`/members/${member.id}`}>編集</Button>
              </Table.Cell>
            </Table.Row>
          {/each}
        </Table.Body>
      </Table.Root>
    </Card.Root>
  {/if}
</div>
