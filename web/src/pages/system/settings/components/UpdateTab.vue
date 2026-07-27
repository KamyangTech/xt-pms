<template>
  <div class="update-tab">
    <!-- 部署信息 -->
    <t-card title="部署环境" :bordered="false" class="setting-card">
      <div class="deploy-info">
        <div class="deploy-row">
          <span class="deploy-label">部署方式</span>
          <t-tag v-if="installInfo?.update_mode === 'systemd'" theme="success" variant="light">systemd 服务</t-tag>
          <t-tag v-else-if="installInfo?.update_mode === 'docker'" theme="primary" variant="light">Docker 部署</t-tag>
          <t-tag v-else-if="installInfo?.update_mode === 'supervisor'" theme="primary" variant="light">supervisor 托管</t-tag>
          <t-tag v-else theme="warning" variant="light">手动运行（仅支持手动更新）</t-tag>
        </div>
        <div class="deploy-row" v-if="installInfo?.install_dir">
          <span class="deploy-label">安装目录</span>
          <span class="deploy-value">{{ installInfo?.install_dir }}</span>
        </div>
        <div class="deploy-row">
          <span class="deploy-label">当前版本</span>
          <span class="deploy-value version-value">v{{ appVersion }}</span>
        </div>
        <t-button theme="primary" :loading="checking" @click="handleCheckUpdate">
          <template #icon><t-icon name="refresh" /></template>
          检查更新
        </t-button>
        <span v-if="!installInfo?.can_auto_update && installInfo?.update_mode !== 'docker'" class="deploy-hint">
          当前环境无法自动安装，检查更新仅用于查看是否有新版本
        </span>
        <span v-else-if="installInfo?.update_mode === 'docker'" class="deploy-hint">
          Docker 部署通过镜像升级，检查更新仅用于查看是否有新版本
        </span>
      </div>
    </t-card>

    <!-- 更新可用 -->
    <t-card v-if="updateInfo?.has_update" title="发现新版本" :bordered="false" class="setting-card update-available-card">
      <div class="update-header">
        <div class="update-versions">
          <t-tag theme="default" variant="light">v{{ updateInfo.current_version }}</t-tag>
          <t-icon name="arrow-right" />
          <t-tag theme="success" variant="light">v{{ updateInfo.latest_version }}</t-tag>
        </div>
        <span class="update-pub-time" v-if="updateInfo.published_at">
          发布于 {{ formatDate(updateInfo.published_at) }}
        </span>
      </div>

      <!-- Docker 升级指引 -->
      <div v-if="installInfo?.update_mode === 'docker'" class="docker-upgrade-section">
        <t-alert theme="info" :close="false">
          <template #message>
            <p>Docker 部署不支持容器内在线更新，请通过镜像方式升级。</p>
          </template>
        </t-alert>
        <t-button
          variant="outline"
          theme="primary"
          class="docker-upgrade-btn"
          @click="openDockerUpdateDoc"
        >
          <template #icon><t-icon name="link" /></template>
          查看 Docker 升级指引
        </t-button>
      </div>

      <!-- 下载进度 -->
      <div v-if="downloading && installInfo?.update_mode !== 'docker'" class="download-progress">
        <t-progress :percentage="downloadPercent" :label="true" />
        <span class="progress-text">{{ formatSize(downloadedBytes) }} / {{ formatSize(updateInfo.file_size) }}</span>
      </div>

      <div class="update-actions" v-if="!downloading && !applying">
        <t-button v-if="installInfo?.update_mode !== 'docker'" theme="primary" @click="handleDownload" :loading="downloadStarting" :disabled="!installInfo?.can_auto_update">
          下载更新
        </t-button>
        <t-button
          v-if="packagePath && installInfo?.update_mode !== 'docker'"
          theme="success"
          @click="handleApply"
          :loading="applying"
          :disabled="!installInfo?.can_auto_update"
        >
          立即更新
        </t-button>
        <span v-if="!installInfo?.can_auto_update && installInfo?.update_mode !== 'docker'" class="deploy-hint">
          当前为手动运行环境，请按下方手动步骤升级
        </span>
      </div>

      <div v-if="applying" class="apply-status">
        <t-loading text="正在执行更新，服务即将重启..." />
      </div>

      <!-- Changelog -->
      <t-divider />
      <div class="changelog-section">
        <div class="changelog-title">更新日志</div>
        <!-- eslint-disable-next-line vue/no-v-html -- DOMPurify 已净化 -->
        <div class="changelog-body" v-html="renderedChangelog"></div>
      </div>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import DOMPurify from 'dompurify';
import {
  checkUpdate,
  downloadUpdateStream,
  applyUpdate,
  getInstallInfo,
  type CheckUpdateResult,
  type InstallInfo,
} from '@/api/system/update';

const appVersion = __APP_VERSION__;

const installInfo = ref<InstallInfo | null>(null);
const updateInfo = ref<CheckUpdateResult | null>(null);
const checking = ref(false);
const updateChecked = ref(false);

// 下载状态
const downloading = ref(false);
const downloadStarting = ref(false);
const downloadPercent = ref(0);
const downloadedBytes = ref(0);
const packagePath = ref('');
const packageSHA256 = ref('');
const downloadAbort = ref<AbortController | null>(null);

// 应用更新
const applying = ref(false);

const renderedChangelog = computed(() => {
  if (!updateInfo.value?.changelog) return '<p>暂无更新日志</p>';
  // 简单的 markdown 转 HTML（处理 # 标题和 - 列表）
  const html = updateInfo.value.changelog
    .replace(/^### (.+)$/gm, '<h4>$1</h4>')
    .replace(/^## (.+)$/gm, '<h3>$1</h3>')
    .replace(/^- (.+)$/gm, '<li>$1</li>')
    .replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>')
    .replace(/\n\n/g, '<br/>');
  // DOMPurify 净化，防止上游 API 被劫持后注入恶意脚本
  return DOMPurify.sanitize(html);
});

function formatDate(dateStr: string) {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatSize(bytes: number) {
  if (!bytes) return '0 B';
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / 1024 / 1024).toFixed(1) + ' MB';
}

async function loadInstallInfo() {
  try {
    installInfo.value = await getInstallInfo();
  } catch {
    // 降级：假设为 systemd
  }
}

async function handleCheckUpdate() {
  checking.value = true;
  updateChecked.value = false;
  try {
    updateInfo.value = await checkUpdate(appVersion);
    updateChecked.value = true;
    if (updateInfo.value?.has_update) {
      MessagePlugin.success(`发现新版本 v${updateInfo.value.latest_version}`);
    } else {
      MessagePlugin.info('当前已是最新版本');
    }
  } catch (e: any) {
    MessagePlugin.error(e?.message || '检查更新失败');
  } finally {
    checking.value = false;
  }
}

function handleDownload() {
  if (!updateInfo.value?.download_url) return;
  
  downloadStarting.value = true;
  downloading.value = true;
  downloadPercent.value = 0;
  downloadedBytes.value = 0;
  packagePath.value = '';
  packageSHA256.value = '';

  downloadAbort.value = downloadUpdateStream(
    updateInfo.value.download_url,
    (data) => {
      downloadPercent.value = data.percent;
      downloadedBytes.value = data.downloaded;
    },
    (data) => {
      // 下载完成
      const msg = data.message || '';
      const parts = msg.split('|');
      packagePath.value = parts[0] || '';
      if (parts.length > 1) {
        for (let i = 1; i < parts.length; i++) {
          if (parts[i].startsWith('sha256:')) {
            packageSHA256.value = parts[i].slice(7);
          }
        }
      }
      downloading.value = false;
      downloadStarting.value = false;
      downloadPercent.value = 100;
      MessagePlugin.success('下载完成');
    },
    (err) => {
      downloading.value = false;
      downloadStarting.value = false;
      MessagePlugin.error(err || '下载失败');
    }
  );
}

async function handleApply() {
  if (!packagePath.value || !updateInfo.value) return;
  
  applying.value = true;
  try {
    const result = await applyUpdate({
      from_version: updateInfo.value.current_version,
      to_version: updateInfo.value.latest_version,
      package_path: packagePath.value,
      sha256: packageSHA256.value || updateInfo.value.sha256,
      changelog: updateInfo.value.changelog,
      file_size: updateInfo.value.file_size,
    });
    MessagePlugin.success(result.message || '更新已启动');
    // 等待几秒后刷新
    setTimeout(() => {
      window.location.reload();
    }, 8000);
  } catch (e: any) {
    applying.value = false;
    MessagePlugin.error(e?.message || '更新失败');
  }
}

onMounted(() => {
  loadInstallInfo();
});

function openDockerUpdateDoc() {
  window.open('https://doc.pms.xintest.cn/guide/update.html#docker-update', '_blank');
}
</script>

<style scoped lang="less">
.update-tab {
  padding-top: 8px;
}

.setting-card {
  margin-bottom: var(--td-comp-margin-l);
}

.deploy-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.deploy-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.deploy-label {
  color: var(--td-text-color-secondary);
  font-size: 14px;
  min-width: 80px;
}

.deploy-value {
  color: var(--td-text-color-primary);
  font-size: 14px;
}

.deploy-hint {
  color: var(--td-text-color-secondary);
  font-size: 12px;
  margin-left: 8px;
}

.deploy-info :deep(.t-button) {
  align-self: flex-start;
}

.version-value {
  font-weight: 600;
  font-family: monospace;
}

.update-available-card {
  border-left: 3px solid var(--td-success-color);
}

.update-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.update-versions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.update-pub-time {
  color: var(--td-text-color-secondary);
  font-size: 13px;
}

.docker-upgrade-section {
  margin-bottom: 16px;

  .docker-upgrade-btn {
    margin-top: 12px;
  }
}

.download-progress {
  margin-bottom: 16px;

  .progress-text {
    display: block;
    margin-top: 4px;
    font-size: 12px;
    color: var(--td-text-color-secondary);
  }
}

.update-actions {
  display: flex;
  gap: 12px;
  margin-bottom: 8px;
}

.apply-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px 0;
}

.changelog-section {
  .changelog-title {
    font-weight: 500;
    margin-bottom: 8px;
    color: var(--td-text-color-primary);
  }

  .changelog-body {
    font-size: 13px;
    color: var(--td-text-color-secondary);
    line-height: 1.6;

    :deep(h3) {
      font-size: 14px;
      margin: 12px 0 6px;
    }

    :deep(h4) {
      font-size: 13px;
      margin: 8px 0 4px;
    }

    :deep(ul) {
      padding-left: 20px;
      margin: 4px 0;
    }

    :deep(li) {
      margin: 2px 0;
    }
  }
}
</style>
