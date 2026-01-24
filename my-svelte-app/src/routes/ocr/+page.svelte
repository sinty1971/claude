<script lang="ts">
  import { writable } from 'svelte/store';

  let fileInput: HTMLInputElement;
  let selectedFile = writable<File | null>(null);
  let ocrResult = writable<any>(null);
  let isLoading = writable(false);
  let errorMessage = writable<string | null>(null);

  async function handleOcr() {
    if (!$selectedFile) {
      errorMessage.set('Please select a file first.');
      return;
    }

    isLoading.set(true);
    errorMessage.set(null);
    ocrResult.set(null);

    const formData = new FormData();
    formData.append('image', $selectedFile);

    try {
      const response = await fetch('http://localhost:9090/api/ocr', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`OCR request failed: ${errorText}`);
      }

      const result = await response.json();
      ocrResult.set(result);
    } catch (error) {
      if (error instanceof Error) {
        errorMessage.set(error.message);
      } else {
        errorMessage.set('An unknown error occurred.');
      }
    } finally {
      isLoading.set(false);
    }
  }

  function handleFileSelect(event: Event) {
    const target = event.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
      selectedFile.set(target.files[0]);
    }
  }
</script>

<svelte:head>
  <title>OCR</title>
</svelte:head>

<div class="container">
  <h1>OCR</h1>

  <div class="file-input-container">
    <input type="file" accept=".pdf,application/pdf" on:change={handleFileSelect} bind:this={fileInput} />
    <button on:click={handleOcr} disabled={$isLoading}>
      {#if $isLoading}
        Processing...
      {:else}
        Run OCR
      {/if}
    </button>
  </div>

  {#if $selectedFile}
    <p>Selected file: {$selectedFile.name}</p>
  {/if}

  {#if $errorMessage}
    <div class="error-message">
      <p>{$errorMessage}</p>
    </div>
  {/if}

  {#if $ocrResult}
    <div class="result-container">
      <h2>OCR Result</h2>
      <pre>{JSON.stringify($ocrResult, null, 2)}</pre>
    </div>
  {/if}
</div>

<style>
  .container {
    padding: 2rem;
  }

  .file-input-container {
    margin-bottom: 1rem;
  }

  .error-message {
    color: red;
    margin-bottom: 1rem;
  }

  .result-container {
    margin-top: 2rem;
    background-color: #f5f5f5;
    padding: 1rem;
    border-radius: 5px;
  }

  pre {
    white-space: pre-wrap;
    word-wrap: break-word;
  }
</style>
