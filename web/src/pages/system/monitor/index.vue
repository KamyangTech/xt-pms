<template>
  <div class="monitor-container">
    <!-- 顶栏：标题 + 刷新控制 -->
    <div class="monitor-header">
      <div class="monitor-title">
        <t-icon name="dashboard" size="24px" />
        <span>系统性能监控</span>
        <t-tag :theme="healthTagTheme" variant="light" size="medium">
          {{ fullHealth?.status === 'ok' ? '正常' : fullHealth?.status === 'degraded' ? '降级' : '异常' }}
        </t-tag>
      </div>
      <div class="monitor-actions">
        <span class="refresh-label">自动刷新</span>
        <t-switch v-model="autoRefresh" size="small" />
        <t-button variant="outline" @click="refreshAll">
          <template #icon><t-icon name="refresh" /></template>
        </t-button>
      </div>
    </div>

    <!-- 第一行：健康状态 + 运行时长 -->
      <t-row :gutter="16" class="row-gap">
        <t-col :span="4">
          <t-card :bordered="false" class="stat-card health-card" :class="healthClass">
            <div class="stat-label">系统状态</div>
            <div class="stat-value health-value">
              <t-check-circle-filled-icon v-if="fullHealth?.status === 'ok'" size="36px" />
              <t-error-circle-filled-icon v-else size="36px" />
              <span>{{ fullHealth?.status === 'ok' ? '正常' : fullHealth?.status === 'degraded' ? '降级' : '异常' }}</span>
            </div>
            <div class="stat-desc">{{ fullHealth?.database?.error || '数据库连接正常' }}</div>
          </t-card>
        </t-col>
        <t-col :span="2">
          <t-card :bordered="false" class="stat-card">
            <div class="stat-label">运行时长</div>
            <div class="stat-value">{{ uptimeDisplay }}</div>
          </t-card>
        </t-col>
        <t-col :span="2">
          <t-card :bordered="false" class="stat-card">
            <div class="stat-label">请求总数</div>
            <div class="stat-value primary">{{ metrics?.total_requests?.toLocaleString() ?? '-' }}</div>
          </t-card>
        </t-col>
        <t-col :span="2">
          <t-card :bordered="false" class="stat-card">
            <div class="stat-label">处理中</div>
            <div class="stat-value warning">{{ metrics?.in_flight ?? '-' }}</div>
          </t-card>
        </t-col>
        <t-col :span="2">
          <t-card :bordered="false" class="stat-card">
            <div class="stat-label">慢请求</div>
            <div class="stat-value" :class="{ danger: (metrics?.slow_requests_total ?? 0) > 10 }">
              {{ metrics?.slow_requests_total ?? '-' }}
            </div>
            <div class="stat-desc">&gt;500ms</div>
          </t-card>
        </t-col>
      </t-row>

      <!-- 第二行：DB 连接池 -->
      <t-row :gutter="16" class="row-gap">
        <t-col :span="6">
          <t-card :bordered="false" title="数据库连接池" class="section-card">
            <template #actions>
              <t-tag :theme="poolUsageTheme" variant="light">{{ poolUsageText }}</t-tag>
            </template>
            <t-row :gutter="24">
              <t-col :span="4">
                <div class="pool-stat">
                  <t-progress
                    theme="circle"
                    :percentage="poolUsagePercent"
                    :color="poolProgressColors"
                    :stroke-width="8"
                    size="120"
                  />
                  <div class="pool-stat-label">使用率</div>
                </div>
              </t-col>
              <t-col :span="8">
                <div class="pool-grid">
                  <div class="pool-item">
                    <span class="pool-label">使用中</span>
                    <span class="pool-value accent">{{ poolStats?.in_use_connections ?? '-' }}</span>
                  </div>
                  <div class="pool-item">
                    <span class="pool-label">空闲</span>
                    <span class="pool-value success">{{ poolStats?.idle_connections ?? '-' }}</span>
                  </div>
                  <div class="pool-item">
                    <span class="pool-label">最大连接数</span>
                    <span class="pool-value">{{ poolStats?.max_open_connections ?? '-' }}</span>
                  </div>
                  <div class="pool-item">
                    <span class="pool-label">等待队列</span>
                    <span class="pool-value" :class="{ danger: (poolStats?.waiting_count ?? 0) > 0 }">
                      {{ poolStats?.waiting_count ?? '-' }}
                    </span>
                  </div>
                  <div class="pool-item">
                    <span class="pool-label">累计等待</span>
                    <span class="pool-value">{{ poolStats?.wait_duration ?? '-' }}</span>
                  </div>
                  <div class="pool-item">
                    <span class="pool-label">空闲超时关闭</span>
                    <span class="pool-value">{{ poolStats?.max_idle_closed?.toLocaleString() ?? '-' }}</span>
                  </div>
                </div>
              </t-col>
            </t-row>
          </t-card>
        </t-col>

        <!-- 请求状态分布 -->
        <t-col :span="6">
          <t-card :bordered="false" title="请求状态分布" class="section-card">
            <div class="status-grid">
              <div class="status-item">
                <span class="status-label">1xx (信息)</span>
                <span class="status-value">{{ metrics?.requests_by_status?.['1xx'] ?? 0 }}</span>
              </div>
              <div class="status-item">
                <span class="status-label">2xx (成功)</span>
                <span class="status-value success">{{ metrics?.requests_by_status?.['2xx'] ?? 0 }}</span>
              </div>
              <div class="status-item">
                <span class="status-label">3xx (重定向)</span>
                <span class="status-value">{{ metrics?.requests_by_status?.['3xx'] ?? 0 }}</span>
              </div>
              <div class="status-item">
                <span class="status-label">4xx (客户端错误)</span>
                <span class="status-value warning">{{ metrics?.requests_by_status?.['4xx'] ?? 0 }}</span>
              </div>
              <div class="status-item">
                <span class="status-label">5xx (服务端错误)</span>
                <span class="status-value danger">{{ metrics?.requests_by_status?.['5xx'] ?? 0 }}</span>
              </div>
              <div class="status-item">
                <span class="status-label">总请求数</span>
                <span class="status-value primary">{{ metrics?.total_requests?.toLocaleString() ?? 0 }}</span>
              </div>
            </div>
          </t-card>
        </t-col>
      </t-row>

      <!-- 第三行：延迟 + 运行时 -->
      <t-row :gutter="16" class="row-gap">
        <t-col :span="6">
          <t-card :bordered="false" title="请求延迟" class="section-card">
            <template #actions>
              <span class="card-hint">最近 {{ metrics?.latency_ms?.count ?? 0 }} 个样本</span>
            </template>
            <div class="latency-grid">
              <div class="latency-item">
                <span class="latency-label">平均</span>
                <span class="latency-value">{{ metrics?.latency_ms?.avg ?? '-' }} ms</span>
              </div>
              <div class="latency-item">
                <span class="latency-label">P50</span>
                <span class="latency-value success">{{ metrics?.latency_ms?.p50 ?? '-' }} ms</span>
              </div>
              <div class="latency-item">
                <span class="latency-label">P95</span>
                <span class="latency-value" :class="{ warning: (metrics?.latency_ms?.p95 ?? 0) > 500 }">
                  {{ metrics?.latency_ms?.p95 ?? '-' }} ms
                </span>
              </div>
              <div class="latency-item">
                <span class="latency-label">P99</span>
                <span class="latency-value" :class="{ danger: (metrics?.latency_ms?.p99 ?? 0) > 1000 }">
                  {{ metrics?.latency_ms?.p99 ?? '-' }} ms
                </span>
              </div>
              <div class="latency-item">
                <span class="latency-label">最大</span>
                <span class="latency-value accent">{{ metrics?.latency_ms?.max ?? '-' }} ms</span>
              </div>
            </div>
          </t-card>
        </t-col>

        <t-col :span="6">
          <t-card :bordered="false" title="运行时" class="section-card">
            <template #actions>
              <t-tag size="small" variant="light">{{ fullHealth?.runtime?.go_version }}</t-tag>
            </template>
            <div class="runtime-grid">
              <div class="runtime-item">
                <span class="runtime-label">Goroutines</span>
                <span class="runtime-value" :class="{ warning: (fullHealth?.runtime?.goroutines ?? 0) > 500 }">
                  {{ fullHealth?.runtime?.goroutines?.toLocaleString() ?? '-' }}
                </span>
              </div>
              <div class="runtime-item">
                <span class="runtime-label">堆内存</span>
                <span class="runtime-value">{{ fullHealth?.runtime?.heap_alloc_mb ?? '-' }} MB</span>
              </div>
              <div class="runtime-item">
                <span class="runtime-label">系统内存</span>
                <span class="runtime-value">{{ fullHealth?.runtime?.heap_sys_mb ?? '-' }} MB</span>
              </div>
              <div class="runtime-item">
                <span class="runtime-label">GC 次数</span>
                <span class="runtime-value">{{ fullHealth?.runtime?.num_gc?.toLocaleString() ?? '-' }}</span>
              </div>
              <div class="runtime-item">
                <span class="runtime-label">CPU 核心</span>
                <span class="runtime-value">{{ fullHealth?.runtime?.num_cpu ?? '-' }}</span>
              </div>
            </div>
          </t-card>
        </t-col>
      </t-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { ErrorCircleFilledIcon, CheckCircleFilledIcon } from 'tdesign-icons-vue-next';
import {
  getFullHealth,
  getMetrics,
  getPoolStats,
  type FullHealth,
  type RequestMetrics,
  type PoolStats,
} from '@/api/monitor';

defineOptions({ name: 'SystemMonitor' });

// ------ 状态 ------
const autoRefresh = ref(true);
const fullHealth = ref<FullHealth | null>(null);
const metrics = ref<RequestMetrics | null>(null);
const poolStats = ref<PoolStats | null>(null);

let timer: ReturnType<typeof setInterval> | null = null;

// ------ 计算属性 ------
const healthTagTheme = computed(() => {
  if (!fullHealth.value) return 'default';
  return fullHealth.value.status === 'ok' ? 'success' : fullHealth.value.status === 'degraded' ? 'warning' : 'danger';
});

const healthClass = computed(() => fullHealth.value?.status ?? '');

const uptimeDisplay = computed(() => {
  const sec = parseInt(metrics.value?.uptime_seconds ?? '0');
  if (sec < 60) return `${sec}s`;
  if (sec < 3600) return `${Math.floor(sec / 60)}m ${sec % 60}s`;
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  return `${h}h ${m}m`;
});

const poolUsagePercent = computed(() => {
  if (!poolStats.value || poolStats.value.max_open_connections <= 0) return 0;
  return Math.round((poolStats.value.in_use_connections / poolStats.value.max_open_connections) * 100);
});

const poolUsageTheme = computed(() => {
  if (poolUsagePercent.value > 90) return 'danger';
  if (poolUsagePercent.value > 75) return 'warning';
  return 'success';
});

const poolUsageText = computed(() => `${poolUsagePercent.value}%`);

const poolProgressColors = computed(() => {
  if (poolUsagePercent.value > 90) return { '0%': '#e34d59', '100%': '#e34d59' };
  if (poolUsagePercent.value > 75) return { '0%': '#ed7b2f', '100%': '#ed7b2f' };
  return { '0%': '#00a870', '100%': '#00a870' };
});

// ------ 数据加载 ------
async function loadHealth() {
  try {
    fullHealth.value = await getFullHealth();
  } catch {
    fullHealth.value = null;
  }
}

async function loadMetrics() {
  try {
    metrics.value = await getMetrics();
  } catch {
    metrics.value = null;
  }
}

async function loadPool() {
  try {
    poolStats.value = await getPoolStats();
  } catch {
    poolStats.value = null;
  }
}

async function refreshAll() {
  await Promise.allSettled([loadHealth(), loadMetrics(), loadPool()]);
}

function startAutoRefresh() {
  timer = setInterval(refreshAll, 10_000);
}

function stopAutoRefresh() {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
}

watch(autoRefresh, (val) => {
  if (val) {
    refreshAll().then(startAutoRefresh);
  } else {
    stopAutoRefresh();
  }
});

onMounted(() => {
  if (autoRefresh.value) {
    refreshAll().then(startAutoRefresh);
  } else {
    refreshAll();
  }
});

onUnmounted(stopAutoRefresh);
</script>

<style scoped lang="less">
.monitor-container {
  padding: 16px;
}

.monitor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  .monitor-title {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 18px;
    font-weight: 600;
    color: var(--td-text-color-primary);
  }

  .monitor-actions {
    display: flex;
    align-items: center;
    gap: 12px;

    .refresh-label {
      font-size: 14px;
      color: var(--td-text-color-secondary);
    }
  }
}

.row-gap {
  margin-bottom: 16px;
}

// ------ 通用统计卡片 ------
.stat-card {
  height: 140px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  text-align: center;

  .stat-label {
    font-size: 13px;
    color: var(--td-text-color-secondary);
    margin-bottom: 8px;
  }

  .stat-value {
    font-size: 28px;
    font-weight: 700;

    &.primary { color: var(--td-brand-color); }
    &.success { color: var(--td-success-color); }
    &.warning { color: var(--td-warning-color); }
    &.danger  { color: var(--td-error-color); }
  }

  .stat-desc {
    font-size: 12px;
    color: var(--td-text-color-placeholder);
    margin-top: 4px;
  }
}

// 健康卡片
.health-card {
  &.ok       { border-bottom: 3px solid var(--td-success-color); }
  &.degraded { border-bottom: 3px solid var(--td-warning-color); }
  &.down     { border-bottom: 3px solid var(--td-error-color); }

  .health-value {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    font-size: 22px;
    text-transform: uppercase;
    letter-spacing: 2px;
  }
}

// ------ 分区卡片 ------
.section-card {
  .card-hint {
    font-size: 12px;
    color: var(--td-text-color-placeholder);
  }
}

// ------ 连接池 ------
.pool-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;

  .pool-stat-label {
    font-size: 13px;
    color: var(--td-text-color-secondary);
  }
}

.pool-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 24px;
  padding: 8px 0;

  .pool-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 0;
    border-bottom: 1px solid var(--td-component-stroke);

    .pool-label {
      font-size: 13px;
      color: var(--td-text-color-secondary);
    }

    .pool-value {
      font-size: 16px;
      font-weight: 600;

      &.accent  { color: var(--td-brand-color); }
      &.success { color: var(--td-success-color); }
      &.danger  { color: var(--td-error-color); }
    }
  }
}

// ------ 请求状态分布 ------
.status-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 24px;
  padding: 8px 0;

  .status-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 0;
    border-bottom: 1px solid var(--td-component-stroke);

    .status-label {
      font-size: 13px;
      color: var(--td-text-color-secondary);
    }

    .status-value {
      font-size: 16px;
      font-weight: 600;

      &.primary { color: var(--td-brand-color); }
      &.success { color: var(--td-success-color); }
      &.warning { color: var(--td-warning-color); }
      &.danger  { color: var(--td-error-color); }
    }
  }
}

// ------ 延迟 ------
.latency-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
  padding: 16px 0;

  .latency-item {
    text-align: center;
    padding: 12px 8px;
    border-radius: 6px;
    background: var(--td-bg-color-container);

    .latency-label {
      display: block;
      font-size: 12px;
      color: var(--td-text-color-placeholder);
      margin-bottom: 6px;
    }

    .latency-value {
      display: block;
      font-size: 18px;
      font-weight: 700;

      &.success { color: var(--td-success-color); }
      &.warning { color: var(--td-warning-color); }
      &.danger  { color: var(--td-error-color); }
      &.accent  { color: var(--td-brand-color); }
    }
  }
}

// ------ 运行时 ------
.runtime-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
  padding: 16px 0;

  .runtime-item {
    text-align: center;
    padding: 12px 8px;
    border-radius: 6px;
    background: var(--td-bg-color-container);

    .runtime-label {
      display: block;
      font-size: 12px;
      color: var(--td-text-color-placeholder);
      margin-bottom: 6px;
    }

    .runtime-value {
      display: block;
      font-size: 18px;
      font-weight: 700;

      &.warning { color: var(--td-warning-color); }
    }
  }
}
</style>
