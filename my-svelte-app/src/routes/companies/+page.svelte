<script lang="ts">
  import type { PageData } from "./$types";
  import type { Company } from "../../gen/grpc/v1/toyotachikuro_pb";
  import { Badge } from "$lib/components/ui/badge";
  import { Button } from "$lib/components/ui/button";
  import * as Table from "$lib/components/ui/table";
  import * as Card from "$lib/components/ui/card";

  let { data } = $props<{ data: PageData }>();

  let companies = $derived(data.companies);
  let categories = $derived(
    new Map<number, string>(Object.entries(data.categoryMap).map(([k, v]) => [Number(k), v as string]))
  );

  const displayTitle = (company: Company): string => {
    const name = company.name || company.prLongName || "名称未設定";
    return `${name} (${company.id})`;
  };

  const categoryLabel = (company: Company): string =>
    categories.get(company.categoryIndex) ?? "業種未設定";
</script>

<svelte:head>
  <title>会社一覧</title>
</svelte:head>

<div class="max-w-6xl mx-auto px-5 py-12">
  <div class="mb-6">
    <h1 class="text-3xl font-bold mb-2">会社一覧</h1>
    <p class="text-muted-foreground">gRPC サーバーから取得した会社情報を表示しています。</p>
  </div>

  {#if companies.length === 0}
    <Card.Root>
      <Card.Content class="py-8 text-center text-muted-foreground">
        会社情報が見つかりませんでした。
      </Card.Content>
    </Card.Root>
  {:else}
    <p class="text-sm text-muted-foreground mb-3">件数: {companies.length}</p>
    <Card.Root>
      <Table.Root>
        <Table.Header>
          <Table.Row>
            <Table.Head class="w-30">カテゴリ</Table.Head>
            <Table.Head>会社名</Table.Head>
            <Table.Head class="hidden md:table-cell">住所</Table.Head>
            <Table.Head class="hidden lg:table-cell">電話</Table.Head>
            <Table.Head class="w-30">操作</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {#each companies as company (company.id)}
            <Table.Row>
              <Table.Cell>
                <Badge variant="secondary">{categoryLabel(company)}</Badge>
              </Table.Cell>
              <Table.Cell>
                <div class="font-semibold">{displayTitle(company)}</div>
                {#if company.prLongName && company.name}
                  <div class="text-sm text-muted-foreground mt-0.5">{company.prLongName}</div>
                {/if}
              </Table.Cell>
              <Table.Cell class="hidden md:table-cell">{company.prAddress || "-"}</Table.Cell>
              <Table.Cell class="hidden lg:table-cell">{company.prTel || "-"}</Table.Cell>
              <Table.Cell>
                <Button variant="ghost" size="sm" href={`/companies/${company.id}`}>編集</Button>
              </Table.Cell>
            </Table.Row>
          {/each}
        </Table.Body>
      </Table.Root>
    </Card.Root>
  {/if}
</div>
