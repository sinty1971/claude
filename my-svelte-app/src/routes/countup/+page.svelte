<script>
	let count = 0;

	function increment() {
		count += 1;
	}
</script>

<main>
	<h1>Count Demo</h1>
	<p>Count: {count}</p>
	<button type="button" on:click={increment}>Count</button>
</main>
