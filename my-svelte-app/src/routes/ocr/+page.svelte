<script lang="ts">
  import { writable } from 'svelte/store';
  import { Button } from '$lib/components/ui/button';
  import * as Card from '$lib/components/ui/card';

  let fileInput: HTMLInputElement;
  let selectedFile = writable<File | null>(null);
  let ocrResult = writable<any>(null);
  let isLoading = writable(false);
  let errorMessage = writable<string | null>(null);
  let previewUrl: string | null = null;
  let previewType: 'image' | 'pdf' | null = null;

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
      const f = target.files[0];
      selectedFile.set(f);
      // revoke old preview
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
        previewUrl = null;
        previewType = null;
      }

      const mime = f.type;
      if (mime === 'application/pdf') {
        previewType = 'pdf';
        previewUrl = URL.createObjectURL(f);
      } else if (mime.startsWith('image/')) {
        previewType = 'image';
        previewUrl = URL.createObjectURL(f);
      } else {
        previewType = null;
        previewUrl = null;
      }
    }
  }
</script>

<svelte:head>
  <title>OCR</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
  <h1 class="text-2xl font-bold mb-6">OCR</h1>

  <div class="grid md:grid-cols-2 gap-6">
    <div>
      <Card.Root>
        <Card.Content class="space-y-4 p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-muted-foreground">ファイルを選択</p>
              {#if $selectedFile}
                <div class="mt-1 text-sm">{$selectedFile.name}</div>
              {:else}
                <div class="mt-1 text-sm text-muted-foreground">jpg / png / pdf をサポート</div>
              {/if}
            </div>
            <div class="flex gap-2">
              <label class="inline-block">
                <input class="sr-only" type="file" accept="image/*,application/pdf" on:change={handleFileSelect} bind:this={fileInput} />
                <Button variant="outline" onclick={() => fileInput.click()}>ファイル選択</Button>
              </label>
              <Button onclick={handleOcr} disabled={$isLoading}>
                {#if $isLoading}
                  Processing...
                {:else}
                  Run OCR
                {/if}
              </Button>
            </div>
          </div>

          {#if previewUrl}
            <div class="mt-4">
              {#if previewType === 'image'}
                <img src={previewUrl} alt="preview" class="w-48 h-auto rounded shadow-sm object-contain" />
              {:else if previewType === 'pdf'}
                <embed src={previewUrl} type="application/pdf" width="220" height="300" class="rounded shadow-sm" />
              {/if}
            </div>
          {/if}

          {#if $errorMessage}
            <div class="text-destructive">{$errorMessage}</div>
          {/if}
        </Card.Content>
      </Card.Root>
    </div>

    <div>
      {#if $ocrResult}
        <Card.Root>
          <Card.Content class="p-6">
            <h2 class="text-lg font-semibold mb-2">OCR Result</h2>
            <pre class="whitespace-pre-wrap">{JSON.stringify($ocrResult, null, 2)}</pre>
          </Card.Content>
        </Card.Root>
      {/if}
    </div>
  </div>

  {#if $errorMessage}
    <div class="error-message">
      <p>{$errorMessage}</p>
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
