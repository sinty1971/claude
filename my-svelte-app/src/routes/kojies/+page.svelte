<script lang="ts">
  import type { PageData } from "./$types";
  import { timestampDate } from "@bufbuild/protobuf/wkt";
  import { generateKojiStatus } from "$lib/koji-utils";
  import type { Koji } from "../../gen/grpc/v1/toyotachikuro_pb";
  import { Badge } from "$lib/components/ui/badge";
  import { Button } from "$lib/components/ui/button";
  import * as Table from "$lib/components/ui/table";
  import * as Card from "$lib/components/ui/card";

  let { data } = $props<{ data: PageData }>();

  let kojies = $derived(data.kojies);

  const displayDate = (value?: Koji["start"]): string => {
    if (!value) return "-";
    const date = timestampDate(value);
    const time = date.getTime();
    if (Number.isNaN(time)) return "-";
    return date.toLocaleDateString("ja-JP");
  };
</script>

<svelte:head>
  <title>工事一覧</title>
</svelte:head>

<div class="max-w-6xl mx-auto px-5 py-12">
  <div class="mb-6">
    <h1 class="text-3xl font-bold mb-2">工事一覧</h1>
    <p class="text-muted-foreground">gRPC サーバーから取得した工事情報を表示しています。</p>
  </div>

  {#if kojies.length === 0}
    <Card.Root>
      <Card.Content class="py-8 text-center text-muted-foreground">
        工事情報が見つかりませんでした。
      </Card.Content>
    </Card.Root>
  {:else}
    <p class="text-sm text-muted-foreground mb-3">件数: {kojies.length}</p>
    <Card.Root>
      <Table.Root>
        <Table.Header>
          <Table.Row>
            <Table.Head>会社名</Table.Head>
            <Table.Head class="hidden md:table-cell">工事場所</Table.Head>
            <Table.Head class="w-25">状態</Table.Head>
            <Table.Head class="hidden lg:table-cell">開始</Table.Head>
            <Table.Head class="hidden lg:table-cell">終了</Table.Head>
            <Table.Head class="w-25">操作</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {#each kojies as koji (koji.id)}
            <Table.Row>
              <Table.Cell>
                <div class="font-semibold">{koji.companyName || "-"}</div>
                <div class="text-xs text-muted-foreground mt-0.5">ID: {koji.id}</div>
              </Table.Cell>
              <Table.Cell class="hidden md:table-cell">{koji.locationName || "-"}</Table.Cell>
              <Table.Cell>
                {#if generateKojiStatus(koji) === "進行中"}
                  <Badge variant="destructive" class="font-bold">進行中</Badge>
                {:else if generateKojiStatus(koji) === "完了"}
                  <Badge class="bg-green-600 hover:bg-green-700">完了</Badge>
                {:else if generateKojiStatus(koji) === "予定"}
                  <Badge class="bg-blue-600 hover:bg-blue-700">予定</Badge>
                {:else}
                  <Badge variant="secondary">不明</Badge>
                {/if}
              </Table.Cell>
              <Table.Cell class="hidden lg:table-cell">{displayDate(koji.start)}</Table.Cell>
              <Table.Cell class="hidden lg:table-cell">{displayDate(koji.mfEnd)}</Table.Cell>
              <Table.Cell>
                <Button variant="ghost" size="sm" href={`/kojies/${koji.id}`}>編集</Button>
              </Table.Cell>
            </Table.Row>
          {/each}
        </Table.Body>
      </Table.Root>
    </Card.Root>
  {/if}
</div>
