<script lang="ts">
  import type { PageData } from "./$types";
  import type { Member } from "../../gen/grpc/v1/toyotachikuro_pb";
  import { Button } from "$lib/components/ui/button";
  import * as Table from "$lib/components/ui/table";
  import * as Card from "$lib/components/ui/card";

  let { data } = $props<{ data: PageData }>();

  let selectedCompany = $state<string | null>("豊田築炉");

  // カテゴリー名からインデックスへのマップ（会社カテゴリーの定義順）
  const categoryIndexMap: Record<string, number> = {
    "自社組合": 0,
    "下請会社": 1,
    "築炉会社": 2,
    "一人親方": 3,
    "元請会社": 4,
    "リース会社": 5,
    "販売会社": 6,
    "販売会社２": 7,
    "求人会社": 8,
    "一般会社": 9,
    "未設定": 999,
  };

  // 会社名のユニークな一覧を取得（カテゴリーインデックス順→会社名順）
  let companyNames = $derived(
    (Array.from(new Set(data.members.map((m: Member) => m.companyName || "未設定"))) as string[]).sort(
      (a, b) => {
        // 各会社名に対応するメンバーを探してカテゴリーを取得
        const memberA = data.members.find((m: Member) => (m.companyName || "未設定") === a);
        const memberB = data.members.find((m: Member) => (m.companyName || "未設定") === b);
        
        const categoryA = memberA?.companyCategoryName || "未設定";
        const categoryB = memberB?.companyCategoryName || "未設定";
        
        const indexA = categoryIndexMap[categoryA] ?? 999;
        const indexB = categoryIndexMap[categoryB] ?? 999;
        
        // カテゴリーインデックスが同じ場合は会社名でソート
        if (indexA === indexB) {
          return a.localeCompare(b, "ja");
        }
        return indexA - indexB;
      }
    )
  );

  // フィルタリングとソートされたメンバー一覧
  let members = $derived(
    [...data.members]
      .filter((m: Member) => {
        if (selectedCompany === null) return true;
        return (m.companyName || "未設定") === selectedCompany;
      })
      .sort((a: Member, b: Member) => {
        const companyA = a.companyName || "";
        const companyB = b.companyName || "";
        return companyA.localeCompare(companyB, "ja");
      })
  );

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

  {#if data.members.length === 0}
    <Card.Root>
      <Card.Content class="py-8 text-center text-muted-foreground">
        メンバー情報が見つかりませんでした。
      </Card.Content>
    </Card.Root>
  {:else}
    <!-- 会社フィルターボタン -->
    <div class="mb-6">
      <div class="flex flex-wrap gap-2">
        <Button
          variant={selectedCompany === null ? "default" : "outline"}
          size="sm"
          onclick={() => (selectedCompany = null)}
        >
          全て ({data.members.length})
        </Button>
        {#each companyNames as companyName}
          <Button
            variant={selectedCompany === companyName ? "default" : "outline"}
            size="sm"
            onclick={() => (selectedCompany = companyName as string)}
          >
            {companyName} ({data.members.filter((m: Member) => (m.companyName || "未設定") === companyName).length})
          </Button>
        {/each}
      </div>
    </div>

    <p class="text-sm text-muted-foreground mb-3">件数: {members.length}</p>
    <Card.Root>
      <Table.Root>
        <Table.Header>
          <Table.Row>
            <Table.Head>名前</Table.Head>
            <Table.Head class="hidden sm:table-cell">会社名</Table.Head>
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
                {#if member.prKanaName}
                  <div class="text-sm text-muted-foreground mt-0.5">{member.prKanaName}</div>
                {/if}
              </Table.Cell>
              <Table.Cell class="hidden sm:table-cell">{member.companyName || "-"}</Table.Cell>
              <Table.Cell class="hidden md:table-cell">{member.prRole || "-"}</Table.Cell>
              <Table.Cell class="hidden md:table-cell">{member.prEmail || "-"}</Table.Cell>
              <Table.Cell class="hidden lg:table-cell">{member.prMobile || member.prTel || "-"}</Table.Cell>
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
