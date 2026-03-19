<script lang="ts">
  import {
    DownloadLlamaServer, ReextractLlamaServer, GetDownloadProgress,
    DownloadModel, GetModelDownloadProgress
  } from '../../../bindings/katip/internal/service/katipservice.js';

  interface SetupInfo {
    status: string;
    llamaInstalled: boolean;
    llamaPath: string;
    zipExists: boolean;
    modelInstalled: boolean;
    modelPath: string;
    modelPartialBytes: number;
    defaultModelName: string;
    defaultModelSize: string;
    defaultModelID: string;
  }

  interface Props {
    setupInfo: SetupInfo;
    onComplete: () => void;
    onSkip: () => void;
  }

  let { setupInfo, onComplete, onSkip }: Props = $props();

  type WizardStep = 'welcome' | 'llama' | 'model' | 'done';

  let currentStep = $state<WizardStep>('welcome');
  let error = $state('');

  let llamaProgress = $state(0);
  let llamaStatus = $state('');
  let llamaBusy = $state(false);
  let llamaPollTimer: ReturnType<typeof setInterval> | null = null;

  let modelProgress = $state(0);
  let modelStatus = $state('');
  let modelBusy = $state(false);
  let modelPollTimer: ReturnType<typeof setInterval> | null = null;

  let needsLlama = $derived(!setupInfo.llamaInstalled);
  let needsModel = $derived(!setupInfo.modelInstalled);

  let totalSteps = $derived(2 + (needsLlama ? 1 : 0) + (needsModel ? 1 : 0));

  function getStepNumber(step: WizardStep): number {
    let n = 1;
    if (step === 'welcome') return n;
    if (needsLlama) { n++; if (step === 'llama') return n; }
    if (needsModel) { n++; if (step === 'model') return n; }
    n++;
    return n;
  }

  function nextAfterWelcome() {
    if (needsLlama) { currentStep = 'llama'; }
    else if (needsModel) { currentStep = 'model'; }
    else { currentStep = 'done'; }
  }

  function nextAfterLlama() {
    if (needsModel) { currentStep = 'model'; }
    else { currentStep = 'done'; }
  }

  async function handleLlamaSetup() {
    llamaBusy = true;
    error = '';
    try {
      if (setupInfo.zipExists) {
        llamaStatus = 'Arşiv açılıyor...';
        llamaProgress = 50;
        await ReextractLlamaServer();
        llamaProgress = 100;
        llamaStatus = 'Tamamlandı!';
        setTimeout(nextAfterLlama, 600);
      } else {
        llamaStatus = 'İndirme başlatılıyor...';
        llamaProgress = 0;
        await DownloadLlamaServer();
        startLlamaPoll();
      }
    } catch (e) {
      error = 'llama-server kurulumu başarısız: ' + e;
      llamaBusy = false;
    }
  }

  function startLlamaPoll() {
    if (llamaPollTimer) clearInterval(llamaPollTimer);
    llamaPollTimer = setInterval(async () => {
      try {
        const p = await GetDownloadProgress() as any;
        llamaStatus = p.status || '';
        llamaProgress = p.percent ?? 0;
        if (p.error) {
          error = p.error;
          llamaBusy = false;
          stopLlamaPoll();
        } else if (p.percent >= 100) {
          llamaBusy = false;
          stopLlamaPoll();
          setTimeout(nextAfterLlama, 600);
        }
      } catch {
        llamaBusy = false;
        stopLlamaPoll();
      }
    }, 500);
  }

  function stopLlamaPoll() {
    if (llamaPollTimer) { clearInterval(llamaPollTimer); llamaPollTimer = null; }
  }

  async function handleModelSetup() {
    modelBusy = true;
    error = '';
    modelStatus = 'İndirme başlatılıyor...';
    modelProgress = 0;
    try {
      await DownloadModel(setupInfo.defaultModelID);
      startModelPoll();
    } catch (e) {
      error = 'Model indirme başarısız: ' + e;
      modelBusy = false;
    }
  }

  function startModelPoll() {
    if (modelPollTimer) clearInterval(modelPollTimer);
    modelPollTimer = setInterval(async () => {
      try {
        const p = await GetModelDownloadProgress() as any;
        modelStatus = p.status || '';
        modelProgress = p.percent ?? 0;
        if (p.error) {
          error = p.error;
          modelBusy = false;
          stopModelPoll();
        } else if (p.percent >= 100) {
          modelBusy = false;
          stopModelPoll();
          setTimeout(() => { currentStep = 'done'; }, 600);
        }
      } catch {
        modelBusy = false;
        stopModelPoll();
      }
    }, 800);
  }

  function stopModelPoll() {
    if (modelPollTimer) { clearInterval(modelPollTimer); modelPollTimer = null; }
  }

  function formatBytes(bytes: number): string {
    if (bytes >= 1_000_000_000) return (bytes / 1_073_741_824).toFixed(1) + ' GB';
    if (bytes >= 1_000_000) return (bytes / 1_048_576).toFixed(0) + ' MB';
    return (bytes / 1024).toFixed(0) + ' KB';
  }
</script>

<div class="flex items-center justify-center h-screen w-screen bg-surface">
  <div class="w-full max-w-md px-8">
    <!-- Step indicator -->
    <div class="flex items-center justify-center gap-1.5 mb-10">
      {#each Array(totalSteps) as _, i}
        <div class="h-1 rounded-full transition-all duration-300 {i < getStepNumber(currentStep) ? 'bg-primary w-8' : 'bg-gray-200 w-6'}"></div>
      {/each}
    </div>

    <!-- Welcome -->
    {#if currentStep === 'welcome'}
      <div class="text-center">
        <h1 class="text-2xl font-bold text-text-primary mb-2">Katip'e Hoş Geldiniz</h1>
        <p class="text-sm text-text-secondary mb-8 leading-relaxed">
          Katip, AI destekli Türkçe metin düzeltme aracıdır.
          Başlamak için birkaç bileşenin kurulması gerekiyor.
        </p>

        <div class="text-left bg-surface-secondary rounded-lg p-4 mb-8 space-y-2.5">
          <div class="flex items-center gap-3">
            <div class="w-5 h-5 rounded-full flex items-center justify-center text-xs flex-shrink-0
              {setupInfo.llamaInstalled ? 'bg-green-100 text-green-600' : 'bg-blue-100 text-blue-600'}">
              {setupInfo.llamaInstalled ? '✓' : '1'}
            </div>
            <div class="flex-1">
              <span class="text-sm {setupInfo.llamaInstalled ? 'text-text-secondary line-through' : 'text-text-primary'}">
                llama-server {setupInfo.zipExists ? '(arşiv mevcut, açılacak)' : ''}
              </span>
              {#if setupInfo.llamaInstalled}
                <span class="text-xs text-green-600 ml-1">Kurulu</span>
              {/if}
            </div>
          </div>
          <div class="flex items-center gap-3">
            <div class="w-5 h-5 rounded-full flex items-center justify-center text-xs flex-shrink-0
              {setupInfo.modelInstalled ? 'bg-green-100 text-green-600' : 'bg-blue-100 text-blue-600'}">
              {setupInfo.modelInstalled ? '✓' : needsLlama ? '2' : '1'}
            </div>
            <div class="flex-1">
              <span class="text-sm {setupInfo.modelInstalled ? 'text-text-secondary line-through' : 'text-text-primary'}">
                {setupInfo.defaultModelName || 'Varsayılan Model'} ({setupInfo.defaultModelSize || '~4.5 GB'})
              </span>
              {#if setupInfo.modelInstalled}
                <span class="text-xs text-green-600 ml-1">Kurulu</span>
              {:else if setupInfo.modelPartialBytes > 0}
                <span class="text-xs text-blue-600 ml-1">{formatBytes(setupInfo.modelPartialBytes)} indirilmiş</span>
              {/if}
            </div>
          </div>
        </div>

        <button
          class="w-full py-2.5 px-4 rounded-lg bg-primary text-white text-sm font-medium hover:bg-primary-dark transition-colors"
          onclick={nextAfterWelcome}
        >
          Kuruluma Başla
        </button>
        <button
          class="w-full mt-3 py-2 px-4 text-sm text-text-secondary hover:text-text-primary transition-colors"
          onclick={onSkip}
        >
          Şimdi değil, daha sonra kur
        </button>
      </div>

    <!-- llama-server step -->
    {:else if currentStep === 'llama'}
      <div class="text-center">
        <h1 class="text-xl font-bold text-text-primary mb-2">llama-server Kurulumu</h1>
        <p class="text-sm text-text-secondary mb-6 leading-relaxed">
          {#if setupInfo.zipExists}
            Arşiv dosyası mevcut. Dosyalar çıkarılacak.
          {:else}
            llama-server GitHub'dan indirilecek (~37 MB).
          {/if}
        </p>

        {#if llamaBusy || llamaProgress > 0}
          <div class="mb-6">
            <div class="w-full bg-gray-100 rounded-full h-2 overflow-hidden mb-2">
              <div class="bg-primary h-2 rounded-full transition-all duration-300"
                style="width: {Math.max(0, Math.min(100, llamaProgress))}%"></div>
            </div>
            <p class="text-xs text-text-secondary">
              {llamaStatus}
              {#if llamaProgress > 0 && llamaProgress < 100}
                ({llamaProgress}%)
              {/if}
            </p>
          </div>
        {/if}

        {#if error}
          <div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
            <p class="text-xs text-red-700">{error}</p>
          </div>
        {/if}

        {#if !llamaBusy && llamaProgress < 100}
          <button
            class="w-full py-2.5 px-4 rounded-lg bg-primary text-white text-sm font-medium hover:bg-primary-dark transition-colors"
            onclick={handleLlamaSetup}
          >
            {setupInfo.zipExists ? 'Arşivi Aç' : 'İndir ve Kur'}
          </button>
          <button
            class="w-full mt-3 py-2 px-4 text-sm text-text-secondary hover:text-text-primary transition-colors"
            onclick={onSkip}
          >
            Sonra Yap
          </button>
        {/if}
      </div>

    <!-- Model step -->
    {:else if currentStep === 'model'}
      <div class="text-center">
        <h1 class="text-xl font-bold text-text-primary mb-2">Model İndirme</h1>
        <p class="text-sm text-text-secondary mb-6 leading-relaxed">
          {setupInfo.defaultModelName || 'Varsayılan model'} indiriliyor ({setupInfo.defaultModelSize || '~4.5 GB'}).
          {#if setupInfo.modelPartialBytes > 0}
            <br/><span class="text-blue-600">{formatBytes(setupInfo.modelPartialBytes)} daha önce indirilmiş, kaldığı yerden devam edilecek.</span>
          {/if}
        </p>

        {#if modelBusy || modelProgress > 0}
          <div class="mb-6">
            <div class="w-full bg-gray-100 rounded-full h-2.5 overflow-hidden mb-2">
              <div class="bg-primary h-2.5 rounded-full transition-all duration-300"
                style="width: {Math.max(0, Math.min(100, modelProgress))}%"></div>
            </div>
            <p class="text-xs text-text-secondary">
              {modelStatus}
              {#if modelProgress > 0 && modelProgress < 100}
                ({modelProgress}%)
              {/if}
            </p>
          </div>
        {/if}

        {#if error}
          <div class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
            <p class="text-xs text-red-700">{error}</p>
          </div>
        {/if}

        {#if !modelBusy && modelProgress < 100}
          <button
            class="w-full py-2.5 px-4 rounded-lg bg-primary text-white text-sm font-medium hover:bg-primary-dark transition-colors"
            onclick={handleModelSetup}
          >
            {setupInfo.modelPartialBytes > 0 ? 'Kaldığı Yerden Devam Et' : 'İndirmeye Başla'}
          </button>
          <button
            class="w-full mt-3 py-2 px-4 text-sm text-text-secondary hover:text-text-primary transition-colors"
            onclick={onSkip}
          >
            Sonra Yap
          </button>
        {/if}
      </div>

    <!-- Done -->
    {:else if currentStep === 'done'}
      <div class="text-center">
        <div class="w-14 h-14 rounded-full bg-green-100 flex items-center justify-center mx-auto mb-4">
          <span class="text-2xl text-green-600">✓</span>
        </div>
        <h1 class="text-xl font-bold text-text-primary mb-2">Kurulum Tamamlandı!</h1>
        <p class="text-sm text-text-secondary mb-8 leading-relaxed">
          Katip kullanıma hazır. Metin yazın ve AI ile düzeltme yapın.
        </p>
        <button
          class="w-full py-2.5 px-4 rounded-lg bg-primary text-white text-sm font-medium hover:bg-primary-dark transition-colors"
          onclick={onComplete}
        >
          Kullanmaya Başla
        </button>
      </div>
    {/if}
  </div>
</div>
