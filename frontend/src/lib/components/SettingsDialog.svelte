<script lang="ts">
  import {
    GetConfig, UpdateConfig, StartLLMServer, StopLLMServer, GetLLMStatus,
    CheckLlamaServer, DownloadLlamaServer, DownloadLlamaServerForBackend, GetDownloadProgress,
    GetModelCatalog, GetInstalledModels, DownloadModel, GetModelDownloadProgress,
    GetServerLog, ReextractLlamaServer, GetHardwareProfile, RefreshHardwareProfile,
    SelectModel, DeleteModel, GetDiskUsage, ListHuggingFaceGGUF, DownloadHuggingFaceModel
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
    fitStatus?: string;
    recommendedQuant?: string;
    quantTier?: string;
  }

  interface HardwareProfile {
    totalRAMBytes: number;
    availableRAMBytes: number;
    cpuCores: number;
    gpus: { id: string; name: string; vramMB: number; backend: string }[];
    recommendedBackend: string;
    hasGPUAcceleration: boolean;
    platform: string;
    arch: string;
  }

  let { open, onClose }: Props = $props();

  let modelPath = $state('');
  let serverBinary = $state('');
  let serverHost = $state('127.0.0.1');
  let serverPort = $state(8089);
  let ctxSize = $state(4096);
  let threads = $state(4);
  let systemPrompt = $state('');
  let gpuLayers = $state(-1);
  let backend = $state('auto');
  let autoStartLLM = $state(true);
  let huggingFaceToken = $state('');

  let hardwareProfile = $state<HardwareProfile | null>(null);
  let diskUsage = $state<{ totalBytes: number; modelCount: number } | null>(null);
  let hfRepoID = $state('');
  let hfFiles = $state<{ filename: string; size: number; url: string }[]>([]);
  let hfLoading = $state(false);
  let hfError = $state('');
  let activeTab = $state<'models' | 'engine' | 'advanced'>('models');

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
        gpuLayers = cfg.gpuLayers ?? -1;
        backend = cfg.backend || 'auto';
        autoStartLLM = cfg.autoStartLLM ?? true;
        huggingFaceToken = cfg.huggingFaceToken || '';
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

  async function loadHardware() {
    try {
      hardwareProfile = await GetHardwareProfile() as HardwareProfile;
    } catch {
      hardwareProfile = null;
    }
  }

  async function loadDiskUsage() {
    try {
      diskUsage = await GetDiskUsage() as any;
    } catch {
      diskUsage = null;
    }
  }

  async function handleRefreshHardware() {
    try {
      hardwareProfile = await RefreshHardwareProfile() as HardwareProfile;
      await loadCatalog();
    } catch { /* ignore */ }
  }

  function fitLabel(status: string | undefined): { text: string; class: string } {
    switch (status) {
      case 'fits': return { text: 'Sığar', class: 'bg-green-100 text-green-700' };
      case 'may_be_slow': return { text: 'Yavaş olabilir', class: 'bg-amber-100 text-amber-700' };
      case 'wont_fit': return { text: 'Sığmaz', class: 'bg-red-100 text-red-700' };
      default: return { text: '', class: '' };
    }
  }

  function quantLabel(tier: string | undefined): string {
    switch (tier) {
      case 'small': return 'Küçük';
      case 'balanced': return 'Dengeli';
      case 'large': return 'Büyük';
      default: return '';
    }
  }

  function isActiveModel(id: string): boolean {
    const m = catalog.find(c => c.id === id);
    if (!m || !modelPath) return false;
    return modelPath.endsWith(m.filename);
  }

  async function handleSelectModel(id: string) {
    statusMessage = '';
    try {
      await SelectModel(id);
      await loadConfig();
      statusMessage = 'Aktif model değiştirildi.';
      await refreshStatus();
    } catch (e) {
      statusMessage = 'Hata: ' + e;
    }
  }

  async function handleDeleteModel(id: string) {
    if (!confirm('Bu modeli silmek istediğinize emin misiniz?')) return;
    statusMessage = '';
    try {
      await DeleteModel(id);
      await loadCatalog();
      await loadDiskUsage();
      await loadConfig();
      statusMessage = 'Model silindi.';
    } catch (e) {
      statusMessage = 'Hata: ' + e;
    }
  }

  async function handleHFSearch() {
    hfError = '';
    hfFiles = [];
    if (!hfRepoID.trim()) return;
    hfLoading = true;
    try {
      if (huggingFaceToken) {
        await UpdateConfig({
          modelPath, serverBinary, serverHost, serverPort, ctxSize, threads,
          systemPrompt, gpuLayers, backend, autoStartLLM, huggingFaceToken,
        } as any);
      }
      const files = await ListHuggingFaceGGUF(hfRepoID.trim()) as any;
      hfFiles = files ?? [];
    } catch (e) {
      hfError = String(e);
    } finally {
      hfLoading = false;
    }
  }

  async function handleHFDownload(filename: string) {
    modelDownloading = true;
    modelDownloadID = 'hf:' + filename;
    modelDownloadStatus = 'Başlatılıyor...';
    modelDownloadPercent = 0;
    try {
      await DownloadHuggingFaceModel(hfRepoID.trim(), filename);
      startModelProgressPoll();
    } catch (e) {
      modelDownloadStatus = 'Hata: ' + e;
      modelDownloading = false;
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
        gpuLayers,
        backend,
        autoStartLLM,
        huggingFaceToken,
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
      const resolvedBackend = backend === 'auto' ? '' : backend;
      if (resolvedBackend && resolvedBackend !== 'cpu') {
        await DownloadLlamaServerForBackend(resolvedBackend);
      } else {
        await DownloadLlamaServer();
      }
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
      loadHardware();
      loadDiskUsage();
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
    <div class="bg-surface rounded-xl shadow-2xl w-[680px] max-h-[85vh] overflow-y-auto">
      <div class="flex items-center justify-between p-4 border-b border-border">
        <h2 class="text-lg font-semibold">Ayarlar</h2>
        <button class="text-text-secondary hover:text-text-primary text-xl" onclick={onClose}>×</button>
      </div>

      <div class="flex border-b border-border px-4 gap-1">
        <button class="text-xs px-3 py-2 border-b-2 transition-colors {activeTab === 'models' ? 'border-primary text-primary font-medium' : 'border-transparent text-text-secondary'}"
          onclick={() => activeTab = 'models'}>Model Merkezi</button>
        <button class="text-xs px-3 py-2 border-b-2 transition-colors {activeTab === 'engine' ? 'border-primary text-primary font-medium' : 'border-transparent text-text-secondary'}"
          onclick={() => activeTab = 'engine'}>Motor & Sunucu</button>
        <button class="text-xs px-3 py-2 border-b-2 transition-colors {activeTab === 'advanced' ? 'border-primary text-primary font-medium' : 'border-transparent text-text-secondary'}"
          onclick={() => activeTab = 'advanced'}>Gelişmiş</button>
      </div>

      <div class="p-4 space-y-4">
        {#if hardwareProfile}
          <section class="p-3 bg-surface-secondary rounded-lg border border-border/50">
            <div class="flex items-center justify-between mb-2">
              <h3 class="text-xs font-medium text-text-primary uppercase tracking-wide">Donanım Profili</h3>
              <button class="text-[10px] text-primary hover:underline" onclick={handleRefreshHardware}>Yenile</button>
            </div>
            <div class="grid grid-cols-2 gap-2 text-xs text-text-secondary">
              <span>RAM: {formatBytes(hardwareProfile.totalRAMBytes)} ({formatBytes(hardwareProfile.availableRAMBytes)} boş)</span>
              <span>CPU: {hardwareProfile.cpuCores} çekirdek</span>
              <span>Backend: {hardwareProfile.recommendedBackend.toUpperCase()}</span>
              {#if hardwareProfile.gpus?.length}
                <span>GPU: {hardwareProfile.gpus[0].name} ({hardwareProfile.gpus[0].vramMB} MB)</span>
              {:else if hardwareProfile.recommendedBackend === 'metal'}
                <span>GPU: Apple Silicon (Metal)</span>
              {:else}
                <span>GPU: Yok (CPU modu)</span>
              {/if}
            </div>
            {#if diskUsage}
              <p class="text-[10px] text-text-secondary mt-2">Disk: {diskUsage.modelCount} model, {formatBytes(diskUsage.totalBytes)}</p>
            {/if}
          </section>
        {/if}

        {#if activeTab === 'models'}
        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">GGUF Model Kataloğu</h3>
          <p class="text-xs text-text-secondary mb-3">
            Uygunluk rozeti donanımınıza göre hesaplanır. İndirdikten sonra "Kullan" ile aktif modeli değiştirin.
          </p>
          <div class="space-y-2">
            {#each catalog as model (model.id)}
              {@const isInstalled = installedModelIDs.includes(model.id)}
              {@const isActive = isActiveModel(model.id)}
              {@const isThisDownloading = modelDownloading && (modelDownloadID === model.id || modelDownloadID === 'hf:' + model.filename)}
              {@const fit = fitLabel(model.fitStatus)}
              <div class="p-3 border rounded-lg {isActive ? 'border-primary bg-primary/5' : isInstalled ? 'border-green-300 bg-green-50/50' : 'border-border bg-surface-secondary'}">
                <div class="flex items-start justify-between gap-2">
                  <div class="flex-1 min-w-0">
                    <div class="flex flex-wrap items-center gap-1.5">
                      <span class="text-sm font-medium">{model.name}</span>
                      {#if isActive}
                        <span class="text-[10px] px-1.5 py-0.5 rounded bg-primary text-white font-medium">Aktif</span>
                      {/if}
                      {#if model.fitStatus}
                        <span class="text-[10px] px-1.5 py-0.5 rounded {fit.class}">{fit.text}</span>
                      {/if}
                      {#if model.quantTier}
                        <span class="text-[10px] px-1.5 py-0.5 rounded bg-gray-200 text-text-secondary">{quantLabel(model.quantTier)}</span>
                      {/if}
                      <span class="text-[10px] px-1.5 py-0.5 rounded bg-gray-200 text-text-secondary">{model.sizeLabel}</span>
                      <span class="text-[10px] px-1.5 py-0.5 rounded {model.language === 'Türkçe' ? 'bg-blue-100 text-blue-700' : 'bg-gray-100 text-gray-600'}">{model.language}</span>
                    </div>
                    <p class="text-xs text-text-secondary mt-0.5">{model.description}</p>
                  </div>
                  <div class="flex flex-col gap-1 flex-shrink-0">
                    {#if isInstalled}
                      {#if !isActive}
                        <button class="text-xs px-2 py-1 rounded bg-primary text-white hover:bg-primary-dark"
                          onclick={() => handleSelectModel(model.id)}>Kullan</button>
                      {/if}
                      <button class="text-xs px-2 py-1 rounded text-red-600 hover:bg-red-50"
                        onclick={() => handleDeleteModel(model.id)}>Sil</button>
                    {:else if isThisDownloading}
                      <span class="text-xs text-blue-600">%{modelDownloadPercent}</span>
                    {:else}
                      <button class="text-xs px-3 py-1.5 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
                        onclick={() => handleModelDownload(model.id)}
                        disabled={modelDownloading || downloading || model.fitStatus === 'wont_fit'}>
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
        </section>

        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">HuggingFace'den İçe Aktar</h3>
          <div class="flex gap-2 mb-2">
            <input type="text" bind:value={hfRepoID} placeholder="ör: TheBloke/Mistral-7B-GGUF"
              class="flex-1 px-3 py-2 text-sm border border-border rounded-md" />
            <button class="text-xs px-3 py-2 rounded bg-surface-secondary border border-border hover:bg-border disabled:opacity-50"
              onclick={handleHFSearch} disabled={hfLoading}>
              {hfLoading ? 'Aranıyor...' : 'Ara'}
            </button>
          </div>
          {#if hfError}
            <p class="text-xs text-red-600 mb-2">{hfError}</p>
          {/if}
          {#if hfFiles.length > 0}
            <div class="max-h-40 overflow-y-auto space-y-1 border border-border rounded-md p-2">
              {#each hfFiles as f}
                <div class="flex items-center justify-between gap-2 text-xs">
                  <span class="truncate flex-1">{f.filename} ({formatBytes(f.size)})</span>
                  <button class="px-2 py-1 rounded bg-blue-600 text-white hover:bg-blue-700 shrink-0"
                    onclick={() => handleHFDownload(f.filename)} disabled={modelDownloading}>İndir</button>
                </div>
              {/each}
            </div>
          {/if}
        </section>

        {:else if activeTab === 'engine'}
        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">LLM Sunucu Durumu</h3>
          <div class="flex items-center gap-3 p-3 bg-surface-secondary rounded-lg">
            {#if serverStatus}
              <div class="w-2.5 h-2.5 rounded-full flex-shrink-0 {serverStatus.healthy ? 'bg-green-500' : serverStatus.running ? 'bg-amber-500 animate-pulse' : 'bg-red-400'}"></div>
              <span class="text-sm">
                {serverStatus.healthy ? 'Çalışıyor' : serverStatus.running ? 'Model yükleniyor...' : 'Durdurulmuş'}
              </span>
              <span class="text-xs text-text-secondary ml-auto truncate">{serverStatus.endpoint}</span>
            {:else}
              <span class="text-sm text-text-secondary">Durum bilinmiyor</span>
            {/if}
          </div>
          {#if errorType}
            <div class="mt-2 p-3 rounded-lg border bg-red-50 border-red-200">
              <p class="text-xs text-red-700">{errorDetail}</p>
            </div>
          {/if}
          <div class="flex gap-2 mt-2 flex-wrap">
            <button class="text-xs px-3 py-1.5 rounded bg-green-600 text-white hover:bg-green-700 disabled:opacity-50"
              onclick={handleStart} disabled={serverStatus?.running}>Başlat</button>
            <button class="text-xs px-3 py-1.5 rounded bg-red-500 text-white hover:bg-red-600 disabled:opacity-50"
              onclick={handleStop} disabled={!serverStatus?.running}>Durdur</button>
            <button class="text-xs px-3 py-1.5 rounded bg-surface-secondary text-text-secondary hover:bg-border"
              onclick={toggleLog}>{showLog ? 'Logu Gizle' : 'Sunucu Logu'}</button>
          </div>
          {#if showLog}
            <pre class="mt-2 p-2 bg-gray-900 text-gray-200 text-[11px] rounded-md max-h-[160px] overflow-y-auto whitespace-pre-wrap font-mono">{serverLog || 'Henüz log yok...'}</pre>
          {/if}
        </section>

        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">llama-server Kurulumu</h3>
          <div class="p-3 bg-surface-secondary rounded-lg space-y-2">
            <div class="flex items-center gap-2">
              <div class="w-2.5 h-2.5 rounded-full {llamaInstalled ? 'bg-green-500' : 'bg-red-400'}"></div>
              <span class="text-sm">{llamaInstalled ? 'Kurulu' : 'Bulunamadı'}</span>
            </div>
            {#if llamaInstalled}
              <p class="text-xs text-text-secondary break-all">{llamaPath}</p>
            {/if}
            <div class="flex gap-2 flex-wrap">
              {#if llamaZipExists}
                <button class="text-xs px-3 py-1.5 rounded bg-green-600 text-white hover:bg-green-700 disabled:opacity-50"
                  onclick={handleReextract} disabled={extracting || downloading}>
                  {extracting ? 'Çıkarılıyor...' : 'Arşivi Aç'}
                </button>
              {/if}
              <button class="text-xs px-3 py-1.5 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
                onclick={handleDownload} disabled={downloading || extracting}>
                {downloading ? 'İndiriliyor...' : 'llama-server İndir'}
              </button>
            </div>
            {#if downloading}
              <div class="w-full bg-gray-200 rounded-full h-2 overflow-hidden">
                <div class="bg-blue-600 h-2 rounded-full transition-all" style="width: {downloadPercent}%"></div>
              </div>
              <p class="text-xs text-text-secondary">{downloadStatus}</p>
            {/if}
          </div>
        </section>

        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">Motor Ayarları</h3>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs text-text-secondary mb-1" for="backend">Backend</label>
              <select id="backend" bind:value={backend}
                class="w-full px-3 py-2 text-sm border border-border rounded-md bg-surface">
                <option value="auto">Otomatik</option>
                <option value="cuda">CUDA</option>
                <option value="vulkan">Vulkan</option>
                <option value="metal">Metal</option>
                <option value="cpu">CPU</option>
              </select>
            </div>
            <div>
              <label class="block text-xs text-text-secondary mb-1" for="gpuLayers">GPU Katmanları</label>
              <input id="gpuLayers" type="number" bind:value={gpuLayers} min={-1}
                class="w-full px-3 py-2 text-sm border border-border rounded-md" />
            </div>
          </div>
          <label class="flex items-center gap-2 text-sm mt-3">
            <input type="checkbox" bind:checked={autoStartLLM} />
            <span>Açılışta AI sunucusunu otomatik başlat</span>
          </label>
        </section>

        {:else if activeTab === 'advanced'}
        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">HuggingFace Token</h3>
          <input type="password" bind:value={huggingFaceToken} placeholder="Gated modeller için"
            class="w-full px-3 py-2 text-sm border border-border rounded-md" />
        </section>

        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">Yollar</h3>
          <div class="space-y-2">
            <input id="serverBinary" type="text" bind:value={serverBinary} placeholder="llama-server yolu"
              class="w-full px-3 py-2 text-sm border border-border rounded-md" />
            <input id="modelPath" type="text" bind:value={modelPath} placeholder="GGUF model yolu"
              class="w-full px-3 py-2 text-sm border border-border rounded-md" />
          </div>
        </section>

        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">Sunucu Parametreleri</h3>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs text-text-secondary mb-1" for="serverHost">Host</label>
              <input id="serverHost" type="text" bind:value={serverHost}
                class="w-full px-3 py-2 text-sm border border-border rounded-md" />
            </div>
            <div>
              <label class="block text-xs text-text-secondary mb-1" for="serverPort">Port</label>
              <input id="serverPort" type="number" bind:value={serverPort}
                class="w-full px-3 py-2 text-sm border border-border rounded-md" />
            </div>
            <div>
              <label class="block text-xs text-text-secondary mb-1" for="ctxSize">Context</label>
              <input id="ctxSize" type="number" bind:value={ctxSize}
                class="w-full px-3 py-2 text-sm border border-border rounded-md" />
            </div>
            <div>
              <label class="block text-xs text-text-secondary mb-1" for="threads">Thread</label>
              <input id="threads" type="number" bind:value={threads}
                class="w-full px-3 py-2 text-sm border border-border rounded-md" />
            </div>
          </div>
        </section>

        <section>
          <h3 class="text-sm font-medium text-text-primary mb-2">Sistem Prompt'u</h3>
          <textarea bind:value={systemPrompt} rows="5"
            class="w-full px-3 py-2 text-sm border border-border rounded-md resize-y"
            placeholder="AI düzeltme davranışı..."></textarea>
        </section>
        {/if}

        {#if statusMessage}
          <p class="text-sm {statusMessage.startsWith('Hata') ? 'text-red-600' : 'text-green-600'}">{statusMessage}</p>
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
