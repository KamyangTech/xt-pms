<template>
  <div class="settings-container">
    <t-card class="header-card">
      <div class="header-content">
        <span class="page-title">系统设置</span>
        <div class="header-actions">
          <t-button variant="outline" @click="handleClearCache">
            <template #icon><delete-icon /></template>
            清除缓存
          </t-button>
          <t-button variant="outline" @click="handleRestart">
            <template #icon><poweroff-icon /></template>
            重启
          </t-button>
        </div>
      </div>
    </t-card>

    <t-card :bordered="false">
      <t-tabs v-model="activeTab" @change="handleTabChange">
        <!-- ========== 系统信息配置 Tab ========== -->
        <t-tab-panel value="system-info" label="系统信息配置">
          <div class="tab-content">
            <t-card title="基本设置" :bordered="false" class="setting-card">
              <t-form
                ref="systemInfoFormRef"
                :data="systemInfoForm"
                :rules="systemInfoFormRules"
                label-width="120px"
                style="max-width: 600px"
              >
                <t-form-item label="系统名称" name="systemName">
                  <t-input
                    v-model="systemInfoForm.systemName"
                    placeholder="请输入系统名称"
                  />
                </t-form-item>
                <t-form-item label="页面图标">
                  <t-upload
                    v-model="faviconFiles"
                    theme="image"
                    :max="1"
                    accept=".ico,.png,.svg"
                    tips="建议尺寸 64x64，支持 ico/png/svg 格式"
                    :request-method="handleFaviconUpload"
                    :before-upload="handleBeforeFaviconUpload"
                    @remove="handleFaviconRemove"
                  />
                </t-form-item>
                <t-form-item>
                  <t-button
                    theme="primary"
                    :loading="systemInfoSaving"
                    @click="handleSaveSystemInfo"
                  >
                    保存配置
                  </t-button>
                </t-form-item>
              </t-form>
            </t-card>
          </div>
        </t-tab-panel>

        <!-- ========== 文件清理 Tab ========== -->
        <t-tab-panel value="file-cleanup" label="文件清理">
          <div class="tab-content">
            <t-card
              title="文件清理"
              description="清理系统中不再引用的孤文件，节省存储空间。"
              :bordered="false"
              class="setting-card"
            >
              <template #actions>
                <t-button theme="danger" :loading="isCleaning" @click="handleCleanup">
                  立即清理
                </t-button>
              </template>
              <t-form label-width="120px" style="max-width: 600px; padding-top: 8px;">
                <t-form-item
                  label="保留天数"
                  help="系统会自动清理超过保留天数且未被引用的文件，可设置范围为 1-365 天。"
                >
                  <t-input-number
                    v-model="retentionDays"
                    :min="1"
                    :max="365"
                    suffix="天"
                    style="width: 200px"
                  />
                </t-form-item>
                <t-form-item style="margin-top: var(--td-comp-margin-xxl);">
                  <t-button theme="primary" :loading="saving" @click="handleSaveConfig">
                    保存清理配置
                  </t-button>
                </t-form-item>
              </t-form>
            </t-card>
          </div>
        </t-tab-panel>

        <!-- ========== 通知配置 Tab ========== -->
        <t-tab-panel value="notification" label="通知配置">
          <div class="tab-content">
            <t-card :bordered="false" class="setting-card">
              <t-form
                ref="notifyFormRef"
                :data="notifyForm"
                label-width="120px"
                style="max-width: 600px; margin-bottom: 24px"
              >
                <t-form-item label="通知开关">
                  <t-switch v-model="notifyForm.enabled" />
                </t-form-item>
                <t-form-item label="通知方式">
                  <t-checkbox-group v-model="notifyForm.channels">
                    <t-checkbox value="dingtalk">钉钉</t-checkbox>
                    <t-checkbox value="wechat">企业微信</t-checkbox>
                    <t-checkbox value="email">邮件</t-checkbox>
                    <t-checkbox value="sms">短信</t-checkbox>
                  </t-checkbox-group>
                </t-form-item>
                <t-form-item label="Webhook URL">
                  <t-input
                    v-model="notifyForm.webhookUrl"
                    placeholder="请输入 Webhook 地址"
                  />
                </t-form-item>
                <t-form-item label="通知频率限制">
                  <t-select v-model="notifyForm.rateLimit" style="width: 200px">
                    <t-option value="1" label="每分钟 1 条" />
                    <t-option value="5" label="每分钟 5 条" />
                    <t-option value="10" label="每分钟 10 条" />
                    <t-option value="0" label="不限制" />
                  </t-select>
                </t-form-item>
                <t-form-item label="通知路由规则">
                  <div class="rule-config-trigger" @click="configPanelVisible = true">
                    <div class="rule-config-summary">
                      <t-tag theme="primary" variant="light">{{ configTableData.length }} 条规则</t-tag>
                      <span class="rule-config-desc">点击配置通知接收规则</span>
                    </div>
                    <t-icon name="chevron-right" class="rule-config-arrow" />
                  </div>
                </t-form-item>
                <t-form-item>
                  <t-button
                    theme="primary"
                    :loading="notifySaving"
                    @click="handleSaveNotify"
                  >
                    保存配置
                  </t-button>
                </t-form-item>
              </t-form>
            </t-card>
          </div>
        </t-tab-panel>

        <!-- ========== 登录与安全 Tab ========== -->
        <t-tab-panel value="security" label="登录与安全">
          <div class="tab-content">
            <t-card :bordered="false" class="setting-card">
              <t-form label-width="150px" style="max-width: 620px; margin-bottom: 24px">
                <t-form-item label="Token 有效期">
                  <t-input v-model="securityForm.tokenExpire" placeholder="请输入" style="width: 200px" />
                  <span style="margin-left: 8px; color: var(--td-text-color-secondary)">秒</span>
                </t-form-item>
                <t-form-item label="最小密码长度">
                  <t-input-number
                    v-model="securityForm.passwordMinLen"
                    :min="4"
                    :max="32"
                    style="width: 200px"
                  />
                </t-form-item>
                <t-form-item label="密码复杂度">
                  <t-space size="16px">
                    <t-checkbox v-model="securityForm.passwordRequireNumber">必须包含数字</t-checkbox>
                    <t-checkbox v-model="securityForm.passwordRequireUpper">必须包含大写字母</t-checkbox>
                    <t-checkbox v-model="securityForm.passwordRequireSpecial">必须包含特殊字符</t-checkbox>
                  </t-space>
                </t-form-item>
                <t-form-item label="登录失败上限" help="连续失败超过该次数后账号将被锁定">
                  <t-input-number
                    v-model="securityForm.maxLoginAttempts"
                    :min="1"
                    :max="20"
                    suffix="次"
                    style="width: 200px"
                  />
                </t-form-item>
                <t-form-item label="账号锁定时间" help="0 表示需管理员手动解锁">
                  <t-input v-model="securityForm.lockDuration" placeholder="请输入" style="width: 200px" />
                  <span style="margin-left: 8px; color: var(--td-text-color-secondary)">秒</span>
                </t-form-item>
                <t-form-item label="Cookie Secure">
                  <t-switch v-model="securityForm.cookieSecure" />
                  <span style="margin-left: 8px; color: var(--td-text-color-secondary); white-space: nowrap;">公网 HTTPS 部署建议开启</span>
                </t-form-item>
                <t-form-item label="WebSocket Secure">
                  <t-switch v-model="securityForm.websocketSecure" />
                  <span style="margin-left: 8px; color: var(--td-text-color-secondary); white-space: nowrap;">公网 HTTPS 部署建议开启</span>
                </t-form-item>
                <t-form-item>
                  <t-button
                    theme="primary"
                    :loading="securitySaving"
                    @click="handleSaveSecurityConfig"
                  >
                    保存配置
                  </t-button>
                </t-form-item>
              </t-form>
            </t-card>
          </div>
        </t-tab-panel>

        <!-- ========== 更新升级 Tab ========== -->
        <t-tab-panel value="update" label="更新升级">
          <UpdateTab />
        </t-tab-panel>

        <!-- ========== 关于 Tab ========== -->
        <t-tab-panel value="about" label="关于">
          <div class="tab-content">
            <t-card :bordered="false" class="setting-card about-card">
              <div class="about-section">
                <div class="about-logo">
                  <t-icon name="logo-github-filled" size="48px" />
                </div>
                <div class="about-title">{{ systemTitle }}</div>
                <div class="about-version">v{{ appVersion }}</div>
                <div class="about-desc">
                  一套基于 Vue 3 + TDesign + Go 的生产管理解决方案，
                  涵盖订单管理、库存管理、生产报工、系统配置等功能模块。
                </div>

                <t-divider />

                <div class="about-info">
                  <div class="about-info-row">
                    <span class="about-label">前端框架</span>
                    <span class="about-value">Vue 3 + TypeScript</span>
                  </div>
                  <div class="about-info-row">
                    <span class="about-label">UI 组件库</span>
                    <span class="about-value">TDesign Vue Next</span>
                  </div>
                  <div class="about-info-row">
                    <span class="about-label">后端框架</span>
                    <span class="about-value">Go + Gin</span>
                  </div>
                  <div class="about-info-row">
                    <span class="about-label">数据库</span>
                    <span class="about-value">MySQL / PostgreSQL</span>
                  </div>
                </div>
              </div>
            </t-card>
          </div>
        </t-tab-panel>
      </t-tabs>
    </t-card>
    <!-- 通知路由规则配置抽屉 -->
    <t-drawer
      v-model:visible="configPanelVisible"
      header="通知路由规则"
      size="860px"
      :footer="false"
      @close="configPanelVisible = false"
    >
      <div class="rule-panel">
        <t-alert theme="info" message="配置哪些类别的通知发送给什么角色或用户。例如：将模具通知配置给模具主管角色，所有模具相关通知自动发送给该角色的所有用户。" style="margin-bottom: 16px" />

        <div class="rule-panel-header">
          <t-button @click="handleAddConfig">
            <template #icon><t-icon name="add" /></template>
            新增规则
          </t-button>
        </div>

        <t-table
          :data="configTableData"
          :columns="configColumns"
          row-key="id"
          :loading="configLoading"
          :pagination="configPagination"
          :hover="true"
          @page-change="onConfigPageChange"
        >
          <template #type_pattern="{ row }">
            <t-tag variant="light">{{ getTypePatternLabel(row.type_pattern) }}</t-tag>
          </template>
          <template #target_type="{ row }">
            <t-tag :theme="row.target_type === 'role' ? 'primary' : 'success'" variant="light">
              {{ row.target_type === 'role' ? '角色' : '用户' }}
            </t-tag>
          </template>
          <template #target_id="{ row }">
            {{ getTargetName(row) }}
          </template>
          <template #enabled="{ row }">
            <t-switch :value="row.enabled === 1" @change="(val: boolean) => handleToggleConfig(row, val)" />
          </template>
          <template #create_time="{ row }">
            {{ formatDateTime(row.create_time) }}
          </template>
          <template #op="{ row }">
            <t-space>
              <t-link theme="primary" @click="handleEditConfig(row)">编辑</t-link>
              <t-link theme="danger" @click="handleDeleteConfig(row)">删除</t-link>
            </t-space>
          </template>
        </t-table>
      </div>
    </t-drawer>

    <!-- 通知配置编辑对话框 -->
    <t-dialog
      v-model:visible="configDialogVisible"
      :header="configIsEdit ? '编辑通知规则' : '新增通知规则'"
      :on-confirm="handleConfigDialogConfirm"
     :close-on-overlay-click="false">
      <t-form v-if="configDialogVisible" ref="configFormRef" :data="configFormData" :rules="configFormRules" label-align="right" :label-width="100">
        <t-form-item label="通知类型" name="type_pattern">
          <t-select
            v-model="configFormData.type_pattern"
            filterable
            allow-create
            placeholder="选择通知类别"
            style="width: 100%"
          >
            <t-option
              v-for="opt in typePatternOptions"
              :key="opt.value"
              :value="opt.value"
              :label="opt.label"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="目标类型" name="target_type">
          <t-radio-group v-model="configFormData.target_type">
            <t-radio value="role">角色</t-radio>
            <t-radio value="user">用户</t-radio>
          </t-radio-group>
        </t-form-item>
        <t-form-item label="目标" name="target_id">
          <t-select
            v-model="configFormData.target_id"
            :options="configTargetOptions"
            filterable
            placeholder="选择目标"
            style="width: 100%"
          />
        </t-form-item>
        <t-form-item label="启用" name="enabled">
          <t-switch v-model="configFormData.enabled" :custom-value="[1, 0]" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue';
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next';

const appVersion = __APP_VERSION__;import type { PageInfo, PrimaryTableCol, TabValue } from 'tdesign-vue-next';
import { DeleteIcon, PoweroffIcon } from 'tdesign-icons-vue-next';
import { request } from '@/utils/request';
import { triggerCleanupOrphanFiles, getOrphanFileCount, getCleanupConfig, updateCleanupConfig, getSystemInfo, updateSystemInfo, getSecurityConfig, updateSecurityConfig } from '@/api/system/settings';
import { restartService } from '@/api/system/update';
import { getWebSocketClient } from '@/utils/websocket';
import {
  getNotificationConfigs,
  createNotificationConfig,
  updateNotificationConfig,
  deleteNotificationConfig,
  getAllNotificationTypes,
  type NotificationConfigItem,
} from '@/api/notification-config';
import { getRoleList, type RoleItem } from '@/api/role-management';
import { getUserList, type UserItem } from '@/api/user-management';
import { NOTIFICATION_TYPE_LABELS } from '@/constants';
import { formatDateTime } from '@/utils/date';
import UpdateTab from './components/UpdateTab.vue';

defineOptions({
  name: 'SystemSettings',
});

// ========== 顶部操作按钮 ==========
const systemTitle = computed(() => {
  return systemInfoForm.systemName || '生产管理系统 (xt-pms)';
});

const handleClearCache = () => {
  // 清除 localStorage 中的字典版本等缓存
  const keysToRemove: string[] = [];
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (key && (key.startsWith('dict_') || key.startsWith('pms_'))) {
      keysToRemove.push(key);
    }
  }
  keysToRemove.forEach((k) => localStorage.removeItem(k));
  // 也清除 sessionStorage
  sessionStorage.clear();
  MessagePlugin.success(`已清除 ${keysToRemove.length + 1} 项缓存（含 SessionStorage）`);
};

const handleRestart = async () => {
  const confirmed = await new Promise<boolean>((resolve) => {
    DialogPlugin.confirm({
      header: '确认重启',
      body: '重启期间服务将暂时不可用，确定继续吗？',
      confirmBtn: { theme: 'warning', content: '确认重启' },
      cancelBtn: '取消',
      onConfirm: () => resolve(true),
      onCancel: () => resolve(false),
      onClose: () => resolve(false),
    });
  });
  if (!confirmed) return;

  try {
    const result = await restartService();
    MessagePlugin.success(result.message || '服务即将重启');
    setTimeout(() => {
      window.location.reload();
    }, 8000);
  } catch (e: any) {
    MessagePlugin.error(e?.message || '重启失败');
  }
};

// ========== Tab 切换 ==========
const activeTab = ref('system-info');

const handleTabChange = (value: TabValue) => {
  activeTab.value = value as string;
  // 切换到通知配置 Tab 时加载数据
  if (value === 'notification') {
    fetchConfigList();
    fetchTypePatternOptions();
  }
};

// ========== 文件清理 Tab ==========
const retentionDays = ref(7);
const isCleaning = ref(false);
const saving = ref(false);

const loadConfig = async () => {
  try {
    const res = await getCleanupConfig();
    retentionDays.value = res.retention_days || 7;
  } catch {
    // 默认7天
  }
};

const handleSaveConfig = async () => {
  saving.value = true;
  try {
    await updateCleanupConfig({ retention_days: retentionDays.value });
    MessagePlugin.success('清理配置已保存');
  } catch {
    MessagePlugin.error('保存失败');
  } finally {
    saving.value = false;
  }
};

const handleCleanup = async () => {
  // 先静默查询
  let count: number;
  try {
    count = (await getOrphanFileCount()).count;
  } catch {
    MessagePlugin.error('查询孤文件数量失败');
    return;
  }

  if (count === 0) {
    MessagePlugin.info('未发现需清理的孤文件');
    return;
  }

  // 有文件才弹确认框（使用 Promise 包装，确保等待用户操作）
  const confirmed = await new Promise<boolean>((resolve) => {
    DialogPlugin.confirm({
      header: '确认清理',
      body: `存在 ${count} 个孤文件，确定要清理吗？此操作不可撤销。`,
      confirmBtn: { theme: 'danger', content: '确认清理' },
      cancelBtn: '取消',
      onConfirm: () => resolve(true),
      onCancel: () => resolve(false),
      onClose: () => resolve(false),
    });
  });
  if (!confirmed) return;

  isCleaning.value = true;
  try {
    const res = await triggerCleanupOrphanFiles();
    MessagePlugin.success(res?.message || '清理任务已启动');
  } catch (error: any) {
    MessagePlugin.error(error?.message || '清理任务启动失败');
  } finally {
    isCleaning.value = false;
  }
};

// ========== 系统信息配置 Tab ==========
const systemInfoFormRef = ref();
const systemInfoSaving = ref(false);
const systemInfoLoading = ref(false);

const systemInfoForm = reactive({
  systemName: '',
  favicon_url: '',
});

const systemInfoFormRules = {
  systemName: [{ required: true, message: '请输入系统名称', trigger: 'blur' }],
};

// ========== Favicon ==========
const faviconURL = ref('');
const faviconVersion = ref(Date.now());
const faviconFiles = ref<any[]>([]);

const updateFaviconLink = (url: string) => {
  const link = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
  if (link) link.href = `${url}?v=${faviconVersion.value}`;
};

const handleBeforeFaviconUpload = (file: File) => {
  const name = file.name.toLowerCase();
  if (!name.endsWith('.ico') && !name.endsWith('.png') && !name.endsWith('.svg')) {
    MessagePlugin.warning('仅支持 .ico、.png 和 .svg 格式');
    return false;
  }
  return true;
};

const handleFaviconUpload = async (file: any) => {
  const formData = new FormData();
  formData.append('file', file.raw);
  try {
    const res = await request.post<{ url: string }>({ url: '/system/favicon', data: formData, headers: { 'Content-Type': 'multipart/form-data' } });
    faviconURL.value = res.url;
    faviconVersion.value = Date.now();
    updateFaviconLink(res.url);
    return { status: 'success', url: res.url };
  } catch {
    MessagePlugin.error('上传失败');
    return { status: 'fail' };
  }
};

const handleFaviconRemove = async () => {
  try {
    await request.delete({ url: '/system/favicon' });
    faviconURL.value = '';
    faviconVersion.value = Date.now();
    updateFaviconLink('/system-assets/logo.svg');
    MessagePlugin.success('已恢复默认 Favicon');
  } catch {
    MessagePlugin.error('操作失败');
  }
};

const loadSystemInfo = async () => {
  systemInfoLoading.value = true;
  try {
    const res = await getSystemInfo();
    systemInfoForm.systemName = res.system_name || '';
    systemInfoForm.favicon_url = res.favicon_url || '';
    faviconURL.value = res.favicon_url || '';
    if (res.favicon_url) {
      faviconFiles.value = [{ name: 'favicon', url: res.favicon_url, status: 'success' }];
    }
  } catch {
    systemInfoForm.systemName = '生产管理系统';
  } finally {
    systemInfoLoading.value = false;
  }
};

const handleSaveSystemInfo = async () => {
  const valid = await systemInfoFormRef.value?.validate?.();
  if (valid !== true) return;

  systemInfoSaving.value = true;
  try {
    await updateSystemInfo({
      system_name: systemInfoForm.systemName,
    });
    document.title = systemInfoForm.systemName;
    localStorage.setItem('pms_system_name', systemInfoForm.systemName);
    MessagePlugin.success('系统信息配置已保存');
  } catch {
    MessagePlugin.error('保存失败');
  } finally {
    systemInfoSaving.value = false;
  }
};

// ========== 通知配置 Tab ==========
// 通知渠道配置（通用参数，后续对接第三方推送使用）
const notifyFormRef = ref();
const notifySaving = ref(false);
const notifyForm = reactive({
  enabled: true,
  channels: ['dingtalk'],
  webhookUrl: '',
  rateLimit: '5',
});
const handleSaveNotify = async () => {
  notifySaving.value = true;
  try {
    // TODO: 对接第三方推送 API
    MessagePlugin.success('通知配置已保存');
  } catch {
    MessagePlugin.error('保存失败');
  } finally {
    notifySaving.value = false;
  }
};

const configPanelVisible = ref(false);
const configLoading = ref(false);
const configTableData = ref<NotificationConfigItem[]>([]);
const configPagination = ref({ defaultPageSize: 20, total: 0, current: 1 });

const configDialogVisible = ref(false);
const configIsEdit = ref(false);
const configFormRef = ref();
const configFormData = ref<NotificationConfigItem>({
  id: 0, type_pattern: '', target_type: 'role', target_id: 0, enabled: 1,
  create_time: '', update_time: '',
});
const configFormRules = {
  type_pattern: [{ required: true, message: '请选择或输入通知类型', trigger: 'change' }],
  target_type: [{ required: true, message: '请选择目标类型', trigger: 'change' }],
  target_id: [{ required: true, message: '请选择目标', trigger: 'change' }],
};

// 角色和用户列表缓存
const configRoleList = ref<RoleItem[]>([]);
const configUserList = ref<UserItem[]>([]);
const configTargetNameMap = ref<Record<string, string>>({});

// 类型模式选项（从后端动态拉取）
const typePatternOptions = ref<{ value: string; label: string }[]>([]);
const typePatternLabelMap = ref<Record<string, string>>({});

function getTypePatternLabel(pattern: string): string {
  return typePatternLabelMap.value[pattern] || pattern;
}

async function fetchTypePatternOptions() {
  try {
    const types = await getAllNotificationTypes();
    const seen = new Set<string>();
    const options: { value: string; label: string }[] = [];
    const labelMap: Record<string, string> = {};
    for (const t of types) {
      const prefix = t.split(':')[0] + ':%';
      if (!seen.has(prefix)) {
        seen.add(prefix);
        const label = NOTIFICATION_TYPE_LABELS[t];
        const displayName = label || prefix;
        options.push({ value: prefix, label: displayName });
        labelMap[prefix] = displayName;
      }
    }
    // 先排序其他选项，最后再把"全部通知"插入到第一位
    options.sort((a, b) => a.label.localeCompare(b.label, 'zh'));
    options.unshift({ value: '%', label: '全部通知' });
    labelMap['%'] = '全部通知';
    typePatternOptions.value = options;
    typePatternLabelMap.value = labelMap;
  } catch (error) {
    console.error('获取通知类型列表失败', error);
    // 失败时保留"全部通知"选项兜底
    typePatternOptions.value = [{ value: '%', label: '全部通知' }];
    typePatternLabelMap.value = { '%': '全部通知' };
  }
}

// 目标选项
const configTargetOptions = computed(() => {
  if (configFormData.value.target_type === 'role') {
    return configRoleList.value
      .filter((r) => r.status === 1)
      .map((r) => ({ value: r.id, label: r.role_name }));
  }
  return configUserList.value
    .filter((u) => u.status === 1)
    .map((u) => ({ value: u.id, label: `${u.real_name || u.username}（${u.username}）` }));
});

const getTargetName = (row: NotificationConfigItem) => {
  const key = `${row.target_type}:${row.target_id}`;
  return configTargetNameMap.value[key] || `${row.target_type === 'role' ? '角色' : '用户'}#${row.target_id}`;
};

const configColumns: PrimaryTableCol[] = [
  { colKey: 'type_pattern', title: '通知类型', width: 130, cell: 'type_pattern' },
  { colKey: 'target_type', title: '目标类型', width: 80, cell: 'target_type' },
  { colKey: 'target_id', title: '目标', width: 130, cell: 'target_id' },
  { colKey: 'enabled', title: '启用', width: 100, cell: 'enabled' },
  { colKey: 'create_time', title: '创建时间', width: 200, cell: 'create_time' },
  { colKey: 'op', title: '操作', width: 120, cell: 'op' },
];

async function fetchConfigTargets() {
  try {
    const [roleRes, userRes] = await Promise.all([
      getRoleList(),
      getUserList({ page_size: 200 }),
    ]);
    configRoleList.value = roleRes.list || [];
    configUserList.value = (userRes.list || []) as UserItem[];
    const map: Record<string, string> = {};
    for (const r of configRoleList.value) map[`role:${r.id}`] = r.role_name;
    for (const u of configUserList.value) map[`user:${u.id}`] = `${u.real_name || u.username}（${u.username}）`;
    configTargetNameMap.value = map;
  } catch (error) {
    console.error('获取角色/用户列表失败', error);
  }
}

async function fetchConfigList() {
  configLoading.value = true;
  try {
    const res = await getNotificationConfigs({
      page: configPagination.value.current,
      page_size: configPagination.value.defaultPageSize,
    });
    configTableData.value = res.list || [];
    configPagination.value.total = res.total || 0;
  } catch (error) {
    console.error('获取通知配置列表失败', error);
  } finally {
    configLoading.value = false;
  }
}

function onConfigPageChange(pageInfo: PageInfo) {
  configPagination.value.current = pageInfo.current;
  fetchConfigList();
}

function handleAddConfig() {
  configIsEdit.value = false;
  // target_id 不设置默认值，让用户自己选择
  configFormData.value = { id: 0, type_pattern: '', target_type: 'role', target_id: null as any, enabled: 1, create_time: '', update_time: '' };
  configDialogVisible.value = true;
}

function handleEditConfig(row: NotificationConfigItem) {
  configIsEdit.value = true;
  configFormData.value = { ...row };
  configDialogVisible.value = true;
}

function handleDeleteConfig(row: NotificationConfigItem) {
  const confirmDia = DialogPlugin.confirm({
    header: '确认删除',
    body: '确定要删除该通知规则吗？',
    onConfirm: async () => {
      try {
        await deleteNotificationConfig(row.id);
        MessagePlugin.success('删除成功');
        fetchConfigList();
      } catch {
        MessagePlugin.error('删除失败');
      } finally {
        confirmDia.hide();
      }
    },
    onClose: () => confirmDia.hide(),
  });
}

async function handleToggleConfig(row: NotificationConfigItem, val: boolean) {
  try {
    await updateNotificationConfig({ id: row.id, enabled: val ? 1 : 0 });
    row.enabled = val ? 1 : 0;
  } catch {
    MessagePlugin.error('更新失败');
  }
}

async function handleConfigDialogConfirm() {
  const valid = await configFormRef.value?.validate();
  if (valid !== true) return;
  try {
    if (configIsEdit.value) {
      await updateNotificationConfig(configFormData.value);
      MessagePlugin.success('更新成功');
    } else {
      await createNotificationConfig(configFormData.value);
      MessagePlugin.success('创建成功');
    }
    configDialogVisible.value = false;
    fetchConfigList();
  } catch {
    MessagePlugin.error(configIsEdit.value ? '更新失败' : '创建失败');
  }
}

// ========== 登录与安全 Tab ==========
const securitySaving = ref(false);
const securityForm = reactive({
  tokenExpire: 7200,
  maxLoginAttempts: 5,
  lockDuration: 900,
  passwordMinLen: 6,
  passwordRequireNumber: true,
  passwordRequireUpper: false,
  passwordRequireSpecial: false,
  cookieSecure: false,
  websocketSecure: false,
});

const loadSecurityConfig = async () => {
  try {
    const res = await getSecurityConfig();
    securityForm.tokenExpire = res.token_expire || 7200;
    securityForm.maxLoginAttempts = res.max_login_attempts || 5;
    securityForm.lockDuration = res.lock_duration || 900;
    securityForm.passwordMinLen = res.password_min_len || 6;
    securityForm.passwordRequireNumber = !!res.password_require_number;
    securityForm.passwordRequireUpper = !!res.password_require_upper;
    securityForm.passwordRequireSpecial = !!res.password_require_special;
    securityForm.cookieSecure = !!res.cookie_secure;
    securityForm.websocketSecure = !!res.websocket_secure;
    // 同步到已初始化的 WebSocket 客户端
    const wsClient = getWebSocketClient();
    if (wsClient) {
      wsClient.setSecure(!!res.websocket_secure);
    }
  } catch {
    // 使用默认值
  }
};

const handleSaveSecurityConfig = async () => {
  if (securityForm.tokenExpire < 300) {
    MessagePlugin.warning('Token 有效期不能少于 5 分钟');
    return;
  }
  securitySaving.value = true;
  try {
    await updateSecurityConfig({
      token_expire: securityForm.tokenExpire,
      max_login_attempts: securityForm.maxLoginAttempts,
      lock_duration: securityForm.lockDuration,
      password_min_len: securityForm.passwordMinLen,
      password_require_number: securityForm.passwordRequireNumber,
      password_require_upper: securityForm.passwordRequireUpper,
      password_require_special: securityForm.passwordRequireSpecial,
      cookie_secure: securityForm.cookieSecure,
      websocket_secure: securityForm.websocketSecure,
    });
    // 同步 WebSocket Secure 标志到客户端
    const wsClient = getWebSocketClient();
    if (wsClient) {
      wsClient.setSecure(securityForm.websocketSecure);
    }
    MessagePlugin.success('安全配置已保存，新登录的用户将使用新的 Token 有效期');
  } catch {
    MessagePlugin.error('保存失败');
  } finally {
    securitySaving.value = false;
  }
};


onMounted(() => {
  loadConfig();
  loadSystemInfo();
  loadSecurityConfig();
  fetchConfigTargets();
  fetchConfigList();
  fetchTypePatternOptions();
});
</script>

<style scoped lang="less">
.settings-container {
  margin: var(--td-comp-margin-l);
}

.header-card {
  margin-bottom: var(--td-comp-margin-l);
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.header-actions {
  display: flex;
  gap: 8px;
}

.tab-content {
  padding-top: 16px;
}

.setting-card {
  margin-bottom: var(--td-comp-margin-l);
}

.setting-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.setting-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.setting-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.setting-title {
  font-size: var(--td-font-size-title-medium);
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.setting-desc {
  font-size: var(--td-font-size-body-medium);
  color: var(--td-text-color-secondary);
  line-height: 1.6;
  max-width: 640px;
}

.setting-actions {
  display: flex;
  gap: 12px;
}

.form-wrapper {
  padding-top: 8px;
}

/* 关于页面样式 */
.about-card {
  display: flex;
  justify-content: center;
}

.about-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px 0;
  text-align: center;
}

.about-logo {
  margin-bottom: 16px;
  color: var(--td-brand-color);
}

.about-title {
  font-size: var(--td-font-size-headline-medium);
  font-weight: 600;
  color: var(--td-text-color-primary);
  margin-bottom: 8px;
}

.about-version {
  font-size: var(--td-font-size-body-large);
  color: var(--td-text-color-secondary);
  margin-bottom: 16px;
}

.about-desc {
  font-size: var(--td-font-size-body-medium);
  color: var(--td-text-color-secondary);
  line-height: 1.6;
  max-width: 500px;
  margin-bottom: 16px;
}

.about-info {
  width: 100%;
  max-width: 400px;
}

/* 通知路由规则触发区域 - 与 t-input/t-select 等表单控件保持统一 */
.rule-config-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0 8px;
  height: var(--td-comp-size-m);
  border: 1px solid var(--td-component-border);
  border-radius: var(--td-radius-default);
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
  background: var(--td-bg-color-container);
  box-sizing: border-box;
}
.rule-config-trigger:hover {
  border-color: var(--td-brand-color);
  box-shadow: 0 0 0 2px var(--td-brand-color-focus);
}
.rule-config-summary {
  display: flex;
  align-items: center;
  gap: 8px;
}
.rule-config-desc {
  font-size: var(--td-font-size-body-small);
  color: var(--td-text-color-placeholder);
}
.rule-config-arrow {
  color: var(--td-text-color-placeholder);
  font-size: 16px;
}

/* 规则配置抽屉 */
.rule-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.rule-panel-header {
  display: flex;
  justify-content: flex-end;
}

.about-info-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: var(--td-font-size-body-medium);
}

.about-label {
  color: var(--td-text-color-secondary);
}

.about-value {
  color: var(--td-text-color-primary);
  font-weight: 500;
}

</style>
