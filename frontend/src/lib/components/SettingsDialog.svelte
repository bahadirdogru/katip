<script lang="ts">
  import {
    GetConfig, UpdateConfig, StartLLMServer, StopLLMServer, GetLLMStatus,
    CheckLlamaServer, DownloadLlamaServer, GetDownloadProgress,
    GetModelCatalog, GetInstalledModels, DownloadModel, GetModelDownloadProgress,
    GetServerLog, ReextractLlamaServer
  } from '../../../bindings/katip/internal/service/katipservice.js';

  interface Props {
    open: boolean;
    onClose: () => void;
  }

  interface CatalogModel {
    id: string;
    name: string;
    description: string;
    sizeLabel: string;
    sizeBytes: number;
    filename: string;
    language: string;
    minRAM: string;
    isDefault: boolean;
  }

  let { open, onClose }: Props = $props();

  let modelPath = $state('');
  let serverBinary = $state('');
  let serverHost = $state('127.0.0.1');
  let serverPort = $state(8089);
  let ctxSize = $state(4096);
  let threads = $state(4);
  let systemPrompt = $state('');

  let serverStatus = $state<{ running: boolean; healthy: boolean; endpoint: string; modelPath: string; lastError: string } | null>(null);
  let saving = $state(false);
  let statusMessage = $state('');
  let serverLog = $state('');
  let showLog = $state(false);
  let statusPollTimer: ReturnType<typeof setInterval> | null = null;

  let llamaInstalled = $state(false);
  let llamaPath = $state('');
  let llamaZipExists = $state(false);
  let extracting = $state(false);
  let downloading = $state(false);
  let downloadStatus = $state('');
  let downloadPercent = $state(0);
  let pollTimer: ReturnType<typeof setInterval> | null = null;

  let catalog = $state<CatalogModel[]>([]);
  let installedModelIDs = $state<string[]>([]);
  let modelDownloading = $state(false);
  let modelDownloadID = $state('');
  let modelDownloadStatus = $state('');
  let modelDownloadPercent = $state(0);
  let modelPollTimer: ReturnType<typeof setInterval> | null = null;

  async function loadConfig() {
    try {
      const cfg = await GetConfig();
      if (cfg) {
        modelPath = cfg.modelPath || '';
        serverBinary = cfg.serverBinary || '';
        serverHost = cfg.serverHost || '127.0.0.1';
        serverPort = cfg.serverPort || 8089;
        ctxSize = cfg.ctxSize || 4096;
        threads = cfg.threads || 4;
        systemPrompt = cfg.systemPrompt || '';
      }
    } catch (e) {
      console.error('Ayarlar yüklenemedi:', e);
    }
  }

  async function checkLlama() {
    try {
      const info = await CheckLlamaServer();
      const data = info as any;
      llamaInstalled = data.installed ?? false;
      llamaPath = data.path ?? '';
      llamaZipExists = data.zipExists ?? false;
      if (llamaInstalled && !serverBinary) {
        serverBinary = llamaPath;
      }
    } catch {
      llamaInstalled = false;
    }
  }

  async function handleReextract() {
    extracting = true;
    statusMessage = '';
    try {
      await ReextractLlamaServer();
      statusMessage = 'DLL dosyaları başarıyla çıkarıldı!';
      await checkLlama();
      await loadConfig();
    } catch (e) {
      statusMessage = 'Çıkarma hatası: ' + e;
    } finally {
      extracting = false;
    }
  }

  async function loadCatalog() {
    try {
      const models = await GetModelCatalog() as any;
      catalog = models ?? [];
      const installed = await GetInstalledModels() as any;
      installedModelIDs = installed ?? [];
    } catch (e) {
      console.error('Katalog yüklenemedi:', e);
    }
  }

  async function refreshStatus() {
    try {
      const status = await GetLLMStatus();
      serverStatus = status as any;
    } catch {
      serverStatus = null;
    }
  }

  async function handleSave() {
    saving = true;
    statusMessage = '';
    try {
      await UpdateConfig({
        modelPath,
        serverBinary,
        serverHost,
        serverPort,
        ctxSize,
        threads,
        systemPrompt,
      } as any);
      statusMessage = 'Ayarlar kaydedildi.';
    } catch (e) {
      statusMessage = 'Hata: ' + e;
    } finally {
      saving = false;
    }
  }

  async function handleStart() {
    statusMessage = '';
    serverLog = '';
    errorType = null;
    errorDetail = '';
    try {
      await StartLLMServer();
      statusMessage = 'Sunucu başlatılıyor... Model yükleniyor, bu birkaç dakika sürebilir.';
      startStatusPoll();
    } catch (e) {
      statusMessage = 'Başlatma hatası: ' + e;
    }
  }

  async function handleStop() {
    statusMessage = '';
    stopStatusPoll();
    try {
      await StopLLMServer();
      statusMessage = 'Sunucu durduruldu.';
      setTimeout(refreshStatus, 1000);
    } catch (e) {
      statusMessage = 'Durdurma hatası: ' + e;
    }
  }

  let errorType = $state<'ram' | 'model' | 'port' | 'generic' | null>(null);
  let errorDetail = $state('');

  function parseServerError(lastError: string) {
    if (!lastError) {
      errorType = null;
      errorDetail = '';
      return;
    }
    if (lastError.startsWith('BELLEK_YETERSIZ:')) {
      errorType = 'ram';
      errorDetail = lastError.replace('BELLEK_YETERSIZ: ', '');
    } else if (lastError.startsWith('MODEL_BOZUK:') || lastError.startsWith('MODEL_YÜKLENEMEDI:') || lastError.startsWith('MODEL_BULUNAMADI:')) {
      errorType = 'model';
      errorDetail = lastError.replace(/^[A-ZÜĞŞÇÖİ_]+: /, '');
    } else if (lastError.startsWith('PORT_KULLANILIYOR:')) {
      errorType = 'port';
      errorDetail = lastError.replace('PORT_KULLANILIYOR: ', '');
    } else {
      errorType = 'generic';
      errorDetail = lastError;
    }
  }

  function startStatusPoll() {
    stopStatusPoll();
    statusPollTimer = setInterval(async () => {
      await refreshStatus();
      try {
        serverLog = await GetServerLog() as any ?? '';
      } catch { /* ignore */ }

      if (serverStatus) {
        if (serverStatus.healthy) {
          statusMessage = 'Sunucu hazır!';
          errorType = null;
          stopStatusPoll();
        } else if (!serverStatus.running) {
          parseServerError(serverStatus.lastError);
          statusMessage = serverStatus.lastError
            ? 'Sunucu kapandı.'
            : 'Sunucu kapandı.';
          stopStatusPoll();
        }
      }
    }, 3000);
  }

  function stopStatusPoll() {
    if (statusPollTimer) {
      clearInterval(statusPollTimer);
      statusPollTimer = null;
    }
  }

  async function toggleLog() {
    showLog = !showLog;
    if (showLog) {
      try {
        serverLog = await GetServerLog() as any ?? '';
      } catch { /* ignore */ }
    }
  }

  async function handleDownload() {
    downloading = true;
    downloadStatus = 'Başlatılıyor...';
    downloadPercent = 0;
    try {
      await DownloadLlamaServer();
      startProgressPoll();
    } catch (e) {
      downloadStatus = 'Hata: ' + e;
      downloading = false;
    }
  }

  async function handleModelDownload(id: string) {
    modelDownloading = true;
    modelDownloadID = id;
    modelDownloadStatus = 'Başlatılıyor...';
    modelDownloadPercent = 0;
    try {
      await DownloadModel(id);
      startModelProgressPoll();
    } catch (e) {
      modelDownloadStatus = 'Hata: ' + e;
      modelDownloading = false;
    }
  }

  function startProgressPoll() {
    if (pollTimer) clearInterval(pollTimer);
    pollTimer = setInterval(async () => {
      try {
        const p = await GetDownloadProgress() as any;
        downloadStatus = p.status || '';
        downloadPercent = p.percent ?? 0;
        if (p.error) {
          downloading = false;
          downloadStatus = 'Hata: ' + p.error;
          stopProgressPoll();
        } else if (p.percent >= 100) {
          downloading = false;
          stopProgressPoll();
          await checkLlama();
          await loadConfig();
        }
      } catch {
        downloading = false;
        stopProgressPoll();
      }
    }, 500);
  }

  function startModelProgressPoll() {
    if (modelPollTimer) clearInterval(modelPollTimer);
    modelPollTimer = setInterval(async () => {
      try {
        const p = await GetModelDownloadProgress() as any;
        modelDownloadStatus = p.status || '';
        modelDownloadPercent = p.percent ?? 0;
        if (p.error) {
          modelDownloading = false;
          modelDownloadStatus = 'Hata: ' + p.error;
          stopModelProgressPoll();
        } else if (p.percent >= 100) {
          modelDownloading = false;
          stopModelProgressPoll();
          await loadCatalog();
          await loadConfig();
        }
      } catch {
        modelDownloading = false;
        stopModelProgressPoll();
      }
    }, 800);
  }

  function stopProgressPoll() {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }

  function stopModelProgressPoll() {
    if (modelPollTimer) {
      clearInterval(modelPollTimer);
      modelPollTimer = null;
    }
  }

  function formatBytes(bytes: number): string {
    if (bytes >= 1_000_000_000) return (bytes / 1_073_741_824).toFixed(1) + ' GB';
    if (bytes >= 1_000_000) return (bytes / 1_048_576).toFixed(0) + ' MB';
    return (bytes / 1024).toFixed(0) + ' KB';
  }

  $effect(() => {
    if (open) {
      loadConfig();
      refreshStatus();
      checkLlama();
      loadCatalog();
    }
    return () => {
      stopProgressPoll();
      stopModelProgressPoll();
      stopStatusPoll();
    };
  });
</script>

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center"
    onkeydown={(e) => e.key === 'Escape' && onClose()}>
    <div class="bg-surface rounded-xl shadow-2xl w-[600px] max-h-[80vh] overflow-y-auto">
      <div class="flex items-center justify-between p-4 border-b border-border">
        <h2 class="text-lg font-semibold">Ayarlar</h2>
        <button class="text-text-secondary hover:text-text-primary text-xl" onclick={onClose}>×</button>
      </div>

      <div class="p-4 space-y-4">
        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">LLM Sunucu Durumu</h3>
          <div class="flex items-center gap-3 p-3 bg-surface-secondary rounded-lg">
            {#if serverStatus}
              <div class="w-2.5 h-2.5 rounded-full flex-shrink-0 {serverStatus.healthy ? 'bg-green-500' : serverStatus.running ? 'bg-amber-500 animate-pulse' : 'bg-red-400'}"></div>
              <span class="text-sm">
                {serverStatus.healthy ? 'Çalışıyor' : serverStatus.running ? 'Model yükleniyor...' : 'Durdurulmuş'}
              </span>
              <span class="text-xs text-text-secondary ml-auto">{serverStatus.endpoint}</span>
            {:else}
              <div class="w-2.5 h-2.5 rounded-full bg-gray-300"></div>
              <span class="text-sm text-text-secondary">Durum bilinmiyor</span>
            {/if}
          </div>
          {#if errorType}
            <div class="mt-2 p-3 rounded-lg border {errorType === 'ram' ? 'bg-red-50 border-red-200' : errorType === 'model' ? 'bg-amber-50 border-amber-200' : errorType === 'port' ? 'bg-blue-50 border-blue-200' : 'bg-gray-50 border-gray-200'}">
              <div class="flex items-start gap-2">
                <span class="text-base flex-shrink-0 mt-0.5">
                  {errorType === 'ram' ? '⚠️' : errorType === 'model' ? '📁' : errorType === 'port' ? '🔌' : '❌'}
                </span>
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium {errorType === 'ram' ? 'text-red-800' : errorType === 'model' ? 'text-amber-800' : errorType === 'port' ? 'text-blue-800' : 'text-gray-800'}">
                    {errorType === 'ram' ? 'Yetersiz Bellek (RAM)' : errorType === 'model' ? 'Model Hatası' : errorType === 'port' ? 'Port Çakışması' : 'Sunucu Hatası'}
                  </p>
                  <p class="text-xs mt-1 {errorType === 'ram' ? 'text-red-700' : errorType === 'model' ? 'text-amber-700' : errorType === 'port' ? 'text-blue-700' : 'text-gray-700'}">
                    {errorDetail}
                  </p>
                  {#if errorType === 'ram'}
                    <div class="mt-2 pt-2 border-t border-red-200">
                      <p class="text-xs text-red-700 font-medium">Önerilen çözümler:</p>
                      <ul class="text-xs text-red-600 mt-1 space-y-0.5 list-disc list-inside">
                        <li>Aşağıdaki katalogdan daha küçük bir model indirin</li>
                        <li>Qwen2.5-3B (~2 GB) veya BitNet-2B (~1 GB) önerilir</li>
                        <li>Diğer uygulamaları kapatarak RAM boşaltın</li>
                      </ul>
                    </div>
                  {/if}
                </div>
              </div>
            </div>
          {:else if serverStatus?.lastError}
            <p class="text-xs text-red-600 mt-1 px-1">{serverStatus.lastError}</p>
          {/if}
          <div class="flex gap-2 mt-2">
            <button class="text-xs px-3 py-1.5 rounded bg-green-600 text-white hover:bg-green-700 disabled:opacity-50"
              onclick={handleStart}
              disabled={serverStatus?.running}>Başlat</button>
            <button class="text-xs px-3 py-1.5 rounded bg-red-500 text-white hover:bg-red-600 disabled:opacity-50"
              onclick={handleStop}
              disabled={!serverStatus?.running}>Durdur</button>
            <button class="text-xs px-3 py-1.5 rounded bg-surface-secondary text-text-secondary hover:bg-border"
              onclick={refreshStatus}>Durumu Yenile</button>
            <button class="text-xs px-3 py-1.5 rounded bg-surface-secondary text-text-secondary hover:bg-border ml-auto"
              onclick={toggleLog}>{showLog ? 'Logu Gizle' : 'Sunucu Logu'}</button>
          </div>
          {#if showLog}
            <pre class="mt-2 p-2 bg-gray-900 text-gray-200 text-[11px] rounded-md max-h-[200px] overflow-y-auto whitespace-pre-wrap break-all font-mono">{serverLog || 'Henüz log yok...'}</pre>
          {/if}
        </section>

        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">llama-server Kurulumu</h3>
          <div class="p-3 bg-surface-secondary rounded-lg space-y-2">
            <div class="flex items-center gap-2">
              <div class="w-2.5 h-2.5 rounded-full {llamaInstalled ? 'bg-green-500' : 'bg-red-400'}"></div>
              <span class="text-sm">
                {llamaInstalled ? 'llama-server kurulu' : 'llama-server bulunamadı'}
              </span>
            </div>
            {#if llamaInstalled}
              <p class="text-xs text-text-secondary break-all">{llamaPath}</p>
            {:else}
              <p class="text-xs text-text-secondary">
                llama-server, AI metin iyileştirme özelliği için gereklidir.
                Aşağıdaki butonla en son sürümü GitHub'dan otomatik indirebilirsiniz (~37 MB).
              </p>
            {/if}
            <div class="flex gap-2 flex-wrap">
              {#if llamaZipExists}
                <button
                  class="text-xs px-3 py-1.5 rounded bg-green-600 text-white hover:bg-green-700 disabled:opacity-50"
                  onclick={handleReextract}
                  disabled={extracting || downloading}>
                  {extracting ? 'Çıkarılıyor...' : 'Arşivi Tekrar Aç (DLL dahil)'}
                </button>
              {/if}
              <button
                class="text-xs px-3 py-1.5 rounded {llamaInstalled && llamaZipExists ? 'bg-surface-secondary text-text-secondary hover:bg-border' : 'bg-blue-600 text-white hover:bg-blue-700'} disabled:opacity-50 disabled:cursor-not-allowed"
                onclick={handleDownload}
                disabled={downloading || extracting}>
                {downloading ? 'İndiriliyor...' : llamaInstalled ? 'Yeniden İndir (DLL dahil)' : 'llama-server İndir'}
              </button>
            </div>
            {#if downloading || downloadStatus}
              <div class="space-y-1">
                {#if downloading}
                  <div class="w-full bg-gray-200 rounded-full h-2 overflow-hidden">
                    <div class="bg-blue-600 h-2 rounded-full transition-all duration-300"
                      style="width: {Math.max(0, Math.min(100, downloadPercent))}%"></div>
                  </div>
                {/if}
                <p class="text-xs {downloadStatus.startsWith('Hata') ? 'text-red-600' : 'text-text-secondary'}">
                  {downloadStatus}
                  {#if downloading && downloadPercent > 0 && downloadPercent < 100}
                    ({downloadPercent}%)
                  {/if}
                </p>
              </div>
            {/if}
          </div>
        </section>

        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">GGUF Model Katalogu</h3>
          <p class="text-xs text-text-secondary mb-3">
            AI metin iyileştirme için bir GGUF model seçip indirebilirsiniz. İndirilen model otomatik olarak ayarlanır.
          </p>
          <div class="space-y-2">
            {#each catalog as model (model.id)}
              {@const isInstalled = installedModelIDs.includes(model.id)}
              {@const isThisDownloading = modelDownloading && modelDownloadID === model.id}
              <div class="p-3 border rounded-lg {isInstalled ? 'border-green-300 bg-green-50' : 'border-border bg-surface-secondary'}">
                <div class="flex items-start justify-between gap-2">
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2">
                      <span class="text-sm font-medium">{model.name}</span>
                      {#if model.isDefault}
                        <span class="text-[10px] px-1.5 py-0.5 rounded bg-primary/10 text-primary font-medium">Varsayılan</span>
                      {/if}
                      <span class="text-[10px] px-1.5 py-0.5 rounded bg-gray-200 text-text-secondary">{model.sizeLabel}</span>
                      <span class="text-[10px] px-1.5 py-0.5 rounded bg-orange-100 text-orange-700">RAM: {model.minRAM}+</span>
                      <span class="text-[10px] px-1.5 py-0.5 rounded {model.language === 'Türkçe' ? 'bg-blue-100 text-blue-700' : model.language === 'Çok dilli' ? 'bg-purple-100 text-purple-700' : 'bg-gray-100 text-gray-600'}">{model.language}</span>
                    </div>
                    <p class="text-xs text-text-secondary mt-0.5">{model.description}</p>
                  </div>
                  <div class="flex-shrink-0">
                    {#if isInstalled}
                      <span class="text-xs text-green-600 font-medium">Kurulu</span>
                    {:else if isThisDownloading}
                      <span class="text-xs text-blue-600">%{modelDownloadPercent}</span>
                    {:else}
                      <button
                        class="text-xs px-3 py-1.5 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap"
                        onclick={() => handleModelDownload(model.id)}
                        disabled={modelDownloading || downloading}>
                        İndir
                      </button>
                    {/if}
                  </div>
                </div>
                {#if isThisDownloading}
                  <div class="mt-2 space-y-1">
                    <div class="w-full bg-gray-200 rounded-full h-1.5 overflow-hidden">
                      <div class="bg-blue-600 h-1.5 rounded-full transition-all duration-300"
                        style="width: {Math.max(0, Math.min(100, modelDownloadPercent))}%"></div>
                    </div>
                    <p class="text-[11px] text-text-secondary">{modelDownloadStatus}</p>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
          {#if !modelDownloading && modelDownloadStatus && modelDownloadStatus.startsWith('Hata')}
            <p class="text-xs text-red-600 mt-2">{modelDownloadStatus}</p>
          {/if}
        </section>

        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">Model ve Sunucu Yolları</h3>
          <div class="space-y-3">
            <div>
              <label class="block text-xs text-text-secondary mb-1" for="serverBinary">llama-server Binary Yolu</label>
              <input id="serverBinary" type="text" bind:value={serverBinary}
                placeholder="Ör: C:\llama.cpp\build\bin\llama-server.exe"
                class="w-full px-3 py-2 text-sm border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary/30" />
            </div>
            <div>
              <label class="block text-xs text-text-secondary mb-1" for="modelPath">GGUF Model Dosyası Yolu</label>
              <input id="modelPath" type="text" bind:value={modelPath}
                placeholder="Katalogdan indirilirse otomatik ayarlanır"
                class="w-full px-3 py-2 text-sm border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary/30" />
            </div>
          </div>
        </section>

        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">Sunucu Ayarları</h3>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs text-text-secondary mb-1" for="serverHost">Host</label>
              <input id="serverHost" type="text" bind:value={serverHost}
                class="w-full px-3 py-2 text-sm border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary/30" />
            </div>
            <div>
              <label class="block text-xs text-text-secondary mb-1" for="serverPort">Port</label>
              <input id="serverPort" type="number" bind:value={serverPort}
                class="w-full px-3 py-2 text-sm border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary/30" />
            </div>
            <div>
              <label class="block text-xs text-text-secondary mb-1" for="ctxSize">Context Uzunluğu</label>
              <input id="ctxSize" type="number" bind:value={ctxSize}
                class="w-full px-3 py-2 text-sm border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary/30" />
            </div>
            <div>
              <label class="block text-xs text-text-secondary mb-1" for="threads">Thread Sayısı</label>
              <input id="threads" type="number" bind:value={threads}
                class="w-full px-3 py-2 text-sm border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary/30" />
            </div>
          </div>
        </section>

        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">Sistem Prompt'u</h3>
          <textarea bind:value={systemPrompt} rows="5"
            class="w-full px-3 py-2 text-sm border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary/30 resize-y"
            placeholder="AI'ın metin düzeltme davranışını belirleyen prompt..."></textarea>
        </section>

        {#if statusMessage}
          <p class="text-sm {statusMessage.startsWith('Hata') ? 'text-red-600' : 'text-green-600'}">
            {statusMessage}
          </p>
        {/if}
      </div>

      <div class="flex justify-end gap-2 p-4 border-t border-border">
        <button class="px-4 py-2 text-sm rounded-md bg-surface-secondary text-text-secondary hover:bg-border"
          onclick={onClose}>Kapat</button>
        <button class="px-4 py-2 text-sm rounded-md bg-primary text-white hover:bg-primary-dark disabled:opacity-50"
          onclick={handleSave} disabled={saving}>
          {saving ? 'Kaydediliyor...' : 'Kaydet'}
        </button>
      </div>
    </div>
  </div>
{/if}
