<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';

	let { children } = $props();
	let isNavOpen = $state(true);
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<div class="app-shell">
	<header class="app-header">
		<div class="brand">Penguin</div>
		<button
			class="nav-toggle"
			type="button"
			aria-expanded={isNavOpen}
			on:click={() => (isNavOpen = !isNavOpen)}
		>
			{isNavOpen ? 'ナビを隠す' : 'ナビを表示'}
		</button>
	</header>

	{#if isNavOpen}
		<nav class="app-nav">
			<a href="/">トップ</a>
			<a href="/companies">会社一覧</a>
			<a href="/kojies">工事一覧</a>
		</nav>
	{/if}

	<main class="app-main">
		{@render children()}
	</main>
</div>
