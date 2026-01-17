<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
</script>

<div class="container mx-auto px-5 py-12">
  <section class="grid md:grid-cols-2 gap-6 mb-12">
    <div class="flex flex-col gap-4">
      <p class="text-xs uppercase tracking-widest text-muted-foreground">Penguin Console</p>
      <h1 class="text-4xl md:text-5xl font-bold">会社と工事を、すばやく見渡す。</h1>
      <p class="text-muted-foreground leading-relaxed">
        会社一覧と工事一覧を中心に、日常の確認と更新を最短で行うための
        シンプルな管理ビューです。
      </p>
      <div class="flex gap-3 flex-wrap mt-2">
        <Button href="/companies" size="lg">会社一覧</Button>
        <Button href="/kojies" variant="secondary" size="lg">工事一覧</Button>
      </div>
    </div>
    <Card.Root>
      <Card.Content class="pt-6 space-y-4">
        <div class="flex justify-between items-center">
          <span class="text-muted-foreground">現在の業務</span>
          <strong>会社・工事の把握</strong>
        </div>
        <div class="flex justify-between items-center">
          <span class="text-muted-foreground">更新フロー</span>
          <strong>一覧 → 編集 → 保存</strong>
        </div>
      </Card.Content>
    </Card.Root>
  </section>

  <section class="grid md:grid-cols-2 gap-4">
    <Card.Root class="hover:bg-accent transition-colors cursor-pointer">
      <a href="/companies" class="block">
        <Card.Content class="pt-6">
          <span class="text-sm text-muted-foreground">Company</span>
          <strong class="block text-lg mt-1">会社一覧を開く</strong>
        </Card.Content>
      </a>
    </Card.Root>
    <Card.Root class="hover:bg-accent transition-colors cursor-pointer">
      <a href="/kojies" class="block">
        <Card.Content class="pt-6">
          <span class="text-sm text-muted-foreground">Koji</span>
          <strong class="block text-lg mt-1">工事一覧を開く</strong>
        </Card.Content>
      </a>
    </Card.Root>
  </section>
</div>
