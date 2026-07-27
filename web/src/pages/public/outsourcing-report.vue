<template>
  <div class="public-report-page">
    <!-- 加载中 -->
    <div v-if="loading" class="loading-container">
      <t-loading size="large" text="验证分享链接中..." />
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="error-container">
      <t-result theme="error" title="链接已失效" :description="error">
        <template #extra>
          <t-button theme="primary" @click="retryLoad">重新验证</t-button>
        </template>
      </t-result>
    </div>

    <!-- 成功状态 -->
    <div v-else class="report-content">
      <!-- 外发信息头部 -->
      <div class="report-header">
        <div class="header-brand">
          <img :src="logoFullUrl" alt="logo" class="logo-img" />
        </div>
        <p class="header-subtitle">
          来自 {{ record.company_name || '本公司' }} 的 {{ record.outsource_type === 'MOLD' ? '模具' : '生产' }} 订单
        </p>
      </div>

      <!-- 外发记录信息卡片 -->
      <t-card :bordered="true" class="info-card">
        <t-descriptions :column="2" bordered>
          <t-descriptions-item v-if="isMold && record.mold" label="模具编号" :span="2">
            <span class="mold-code-highlight">{{ record.mold?.mold_code || '-' }}</span>
          </t-descriptions-item>
          <t-descriptions-item label="外发单号">{{ record.record_code }}</t-descriptions-item>
          <t-descriptions-item label="外发类型">
            <t-tag :theme="record.outsource_type === 'MOLD' ? 'warning' : 'primary'" variant="light">
              {{ record.outsource_type === 'MOLD' ? '模具外发' : '生产外发' }}
            </t-tag>
          </t-descriptions-item>
          <t-descriptions-item label="加工商">{{ record.processor?.name || '-' }}</t-descriptions-item>
          <t-descriptions-item label="外发数量">{{ record.quantity }}</t-descriptions-item>
          <t-descriptions-item label="状态">
            <t-tag :theme="getStatusTheme(record.status)">{{ getStatusText(record.status) }}</t-tag>
          </t-descriptions-item>
          <t-descriptions-item label="外发日期">{{ formatDate(record.outsource_date) }}</t-descriptions-item>
          <t-descriptions-item v-if="record.expected_return_date" label="预计返回">{{ formatDate(record.expected_return_date) }}</t-descriptions-item>
          <t-descriptions-item v-if="record.work_order" label="工单编号">{{ record.work_order?.work_code || '-' }}</t-descriptions-item>
          <t-descriptions-item v-if="isMold && !record.mold" label="模具编号" :span="2">-</t-descriptions-item>
          <t-descriptions-item v-if="record.remark" label="备注" :span="2">{{ record.remark }}</t-descriptions-item>
        </t-descriptions>
      </t-card>

      <!-- 生产外发：数量进度提交 -->
      <t-card v-if="isProduction && record.status !== 'RETURNED' && record.status !== 'CANCELLED'" :bordered="true" class="form-card" title="提交生产进度">
        <div class="progress-overview">
          <div class="progress-row">
            <span class="progress-label">已完成</span>
            <span class="progress-value">{{ completedQty }} / {{ record.quantity || 0 }}</span>
            <span class="progress-percent">{{ progressPercent }}%</span>
          </div>
          <t-progress
            :percentage="progressPercent"
            :stroke-width="12"
            :color="{ from: '#0052d9', to: '#00a870' }"
            style="margin-bottom: 20px"
          />
        </div>
        <t-form ref="formRef" :data="formData" :rules="formRules" label-width="100px" @submit="submitReport">
          <t-form-item label="良品数量" name="ok_qty">
            <t-input-number v-model="formData.ok_qty" :min="0" placeholder="请输入良品数量" style="width: 100%" />
          </t-form-item>
          <t-form-item label="不良品数量" name="bad_qty">
            <t-input-number v-model="formData.bad_qty" :min="0" placeholder="请输入不良数量" style="width: 100%" />
          </t-form-item>
          <t-form-item label="不良原因" name="defect_reason">
            <t-textarea v-model="formData.defect_reason" placeholder="如有不良品，请说明原因（选填）" :autosize="{ minRows: 2, maxRows: 4 }" />
          </t-form-item>
          <t-form-item label="备注" name="remark">
            <t-textarea v-model="formData.remark" placeholder="其他补充说明（选填）" :autosize="{ minRows: 2, maxRows: 4 }" />
          </t-form-item>
          <t-form-item>
            <t-space>
              <t-button theme="primary" type="submit" :loading="submitting">提 交</t-button>
              <t-button theme="default" @click="resetForm">重置</t-button>
            </t-space>
          </t-form-item>
        </t-form>
      </t-card>

      <!-- 模具外发：工序进度 -->
      <t-card v-if="isMold && record.status !== 'RETURNED' && record.status !== 'CANCELLED'" :bordered="true" class="form-card" title="工序进度">
        <div v-if="groupedProcesses.length > 0" class="process-groups">
          <template v-for="(group, groupIndex) in groupedProcesses" :key="'g-' + group.groupId">
            <div class="process-group-card">
              <div class="group-card-header">
                <div class="group-card-title">
                  <span class="group-badge">{{ group.groupId }}</span>
                  <span>{{ group.groupName }}</span>
                  <span class="group-hint">（可并行）</span>
                </div>
                <t-tag v-if="groupIndex > 0" theme="default" size="small">
                  上一组完成后开始
                </t-tag>
              </div>

              <div class="group-card-processes">
                <div
                  v-for="tpl in group.templates"
                  :key="tpl.template_id"
                  class="process-item"
                  :class="tpl.is_completed === 1 ? 'process-completed' : (tpl.is_completed === 2 ? 'process-skipped' : 'process-pending')"
                >
                  <div class="process-item-header">
                    <span class="process-item-name">{{ tpl.process_name }}</span>
                    <t-tag
                      v-if="tpl.is_completed === 1"
                      theme="success"
                      size="small"
                    >
                      已完成
                    </t-tag>
                    <t-tag
                      v-else-if="tpl.is_completed === 2"
                      theme="default"
                      size="small"
                      variant="light-outline"
                    >
                      已跳过
                    </t-tag>
                    <t-tag
                      v-else-if="groupIndex === 0 || allPrevGroupsCompleted(groupIndex)"
                      theme="primary"
                      size="small"
                    >
                      进行中
                    </t-tag>
                    <t-tag v-else theme="default" size="small">待开工</t-tag>
                  </div>
                  <div class="process-item-action">
                    <t-button
                      v-if="tpl.is_completed === 1"
                      theme="success"
                      variant="outline"
                      size="small"
                      @click="toggleProcess(tpl)"
                    >
                      ✓ 已完成（点击撤销）
                    </t-button>
                    <template v-else-if="tpl.is_completed === 2">
                      <t-button theme="default" variant="outline" size="small" disabled>
                        ⊗ 已跳过
                      </t-button>
                    </template>
                    <t-button
                      v-else
                      theme="primary"
                      variant="outline"
                      size="small"
                      @click="toggleProcess(tpl)"
                    >
                      ✓ 完成
                    </t-button>
                    <t-button
                      v-if="tpl.is_completed === 0"
                      theme="default"
                      variant="outline"
                      size="small"
                      @click="skipProcess(tpl)"
                    >
                      ⊗ 跳过
                    </t-button>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
        <div v-else class="process-empty-hint">暂无工序数据</div>
      </t-card>

      <!-- 已结束提示 -->
      <t-card v-if="record.status === 'RETURNED' || record.status === 'CANCELLED'" :bordered="true" class="completed-card">
        <div class="completed-content">
          <t-icon name="check-circle" size="48px" style="color: var(--td-success-color)" />
          <div class="completed-title">当前工单已完结，无需操作</div>
          <div class="completed-desc">外发任务已结束，加工商无需再提交进度。如有疑问请联系发单方。</div>
        </div>
      </t-card>

      <!-- 历史记录（仅生产外发且未结束时显示） -->
      <t-card v-else-if="isProduction && reports.length > 0" :bordered="true" class="reports-card" title="历史记录">
        <t-table
          :data="reports"
          :columns="reportColumns"
          row-key="id"
          size="small"
          :stripe="false"
          :bordered="false"
        >
          <template #report_date="{ row }">
            {{ formatDate(row.report_date) }}
          </template>
        </t-table>
      </t-card>

      <!-- 提交成功提示 -->
      <div v-if="submitSuccess" class="success-banner">
        <t-alert theme="success" message="提交成功！" close @close="submitSuccess = false" />
        <div class="success-action">
          <t-button theme="primary" @click="handleContinue">继续提交</t-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { MessagePlugin } from 'tdesign-vue-next';
import logoFullUrl from '@/assets/assets-logo-full.svg?url';
import { request } from '@/utils/request';
import dayjs from 'dayjs';

defineOptions({
  name: 'OutsourcingReport',
});

const route = useRoute();

const loading = ref(true);
const error = ref('');
const submitting = ref(false);
const submitSuccess = ref(false);
const record = ref<any>({});
const reports = ref<any[]>([]);

// 历史记录表格列
const reportColumns = [
  { colKey: 'report_date', title: '报工日期', width: 160 },
  { colKey: 'ok_qty', title: '良品数量', width: 100 },
  { colKey: 'bad_qty', title: '不良数量', width: 100 },
  { colKey: 'defect_reason', title: '不良原因', ellipsis: true },
  { colKey: 'remark', title: '备注', ellipsis: true },
];

// 进度计算
const completedQty = computed(() => {
  return reports.value.reduce((sum, r) => sum + (r.ok_qty || 0) + (r.bad_qty || 0), 0);
});
const progressPercent = computed(() => {
  const total = record.value.quantity || 0;
  if (total === 0) return 0;
  return Math.min(100, Math.round((completedQty.value / total) * 100));
});

// 外发类型判断
const isProduction = computed(() => record.value.outsource_type === 'PRODUCTION');
const isMold = computed(() => record.value.outsource_type === 'MOLD');

// 模具工序相关
const processList = ref<any[]>([]);

const groupedProcesses = computed(() => {
  const groups: { groupId: number; groupName: string; templates: any[] }[] = [];
  const map = new Map<number, { groupId: number; groupName: string; templates: any[] }>();
  processList.value.forEach((tpl) => {
    const gid = tpl.group_id || 0;
    if (!map.has(gid)) {
      const g = { groupId: gid, groupName: tpl.group_name || '工序', templates: [] as any[] };
      map.set(gid, g);
      groups.push(g);
    }
    map.get(gid)!.templates.push(tpl);
  });
  return groups;
});

const toggleProcess = async (tpl: any) => {
  const newStatus = tpl.is_completed === 1 ? 0 : 1;
  try {
    const token = route.params.token as string;
    await request.post({
      url: `/public/outsourcing/${token}/process/toggle?template_id=${tpl.template_id}&status=${newStatus}`,
    });
    tpl.is_completed = newStatus;
    loadRecord();
  } catch (e: any) {
    // ignore
  }
};

const skipProcess = async (tpl: any) => {
  try {
    const token = route.params.token as string;
    await request.post({
      url: `/public/outsourcing/${token}/process/toggle?template_id=${tpl.template_id}&status=2`,
    });
    tpl.is_completed = 2;
    loadRecord();
  } catch (e: any) {
    // ignore
  }
};

const allPrevGroupsCompleted = (currentIndex: number): boolean => {
  for (let i = 0; i < currentIndex; i++) {
    const g = groupedProcesses.value[i];
    if (!g) continue;
    if (g.templates.some((t: any) => t.is_completed !== 1)) {
      return false;
    }
  }
  return true;
};

// 表单
const formRef = ref();
const formData = reactive({
  ok_qty: null as number | null,
  bad_qty: null as number | null,
  defect_reason: '',
  remark: '',
});

const formRules = {
  ok_qty: [{ required: true, message: '请输入良品数量', type: 'error' }],
};

// 状态映射
const statusMap: Record<string, { text: string; theme: string }> = {
  PENDING: { text: '待发出', theme: 'default' },
  SENT: { text: '已发出', theme: 'primary' },
  PRODUCING: { text: '生产中', theme: 'warning' },
  RETURNED: { text: '已返回', theme: 'success' },
  CANCELLED: { text: '已撤回', theme: 'danger' },
};

const getStatusText = (status: string) => statusMap[status]?.text || status;
const getStatusTheme = (status: string) => statusMap[status]?.theme || 'default';

const formatDate = (date: string | null) => {
  if (!date) return '-';
  return dayjs(date).format('YYYY-MM-DD HH:mm');
};

const loadRecord = async () => {
  loading.value = true;
  error.value = '';
  try {
    const token = route.params.token as string;
    const res = await request.get({ url: `/public/outsourcing/${token}` });
    record.value = res;
    reports.value = res.reports || [];
    processList.value = res.processes || [];
  } catch (e: any) {
    error.value = e?.message || '分享链接无效或已过期，请联系管理员获取新的链接。';
  } finally {
    loading.value = false;
  }
};

// 提交报工
const submitReport = async ({ validateResult }: any) => {
  if (validateResult !== true) return;

  if (!formData.ok_qty && !formData.bad_qty) {
    MessagePlugin.warning('请至少填写良品或不良品数量');
    return;
  }

  submitting.value = true;
  try {
    const token = route.params.token as string;
    await request.post({
      url: `/public/outsourcing/${token}/report`,
      data: {
        ok_qty: formData.ok_qty,
        bad_qty: formData.bad_qty,
        defect_reason: formData.defect_reason,
        remark: formData.remark,
      },
    });
    submitSuccess.value = true;
    formData.ok_qty = null;
    formData.bad_qty = null;
    formData.defect_reason = '';
    formData.remark = '';
    // 刷新报工记录
    loadRecord();
  } catch (e: any) {
    MessagePlugin.error(e?.message || '提交失败，请稍后重试');
  } finally {
    submitting.value = false;
  }
};

const resetForm = () => {
  formData.ok_qty = null;
  formData.bad_qty = null;
  formData.defect_reason = '';
  formData.remark = '';
};

const retryLoad = () => {
  loadRecord();
};

const handleContinue = () => {
  submitSuccess.value = false;
};

onMounted(() => {
  loadRecord();
});

watch(
  () => route.params.token,
  (newToken) => {
    if (newToken && route.name === 'outsourcingReport') {
      loadRecord();
    }
  },
);
</script>

<style scoped>
.public-report-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #e8ecf1 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.loading-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 300px;
}

.error-container {
  max-width: 500px;
  width: 100%;
}

.report-content {
  max-width: 800px;
  width: 100%;
}

.report-header {
  text-align: center;
  margin-bottom: 28px;
  padding-top: 8px;
}

.header-brand {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 16px;
}

.logo-img {
  height: 32px;
  width: auto;
  max-width: 100%;
}

.header-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0 0 12px 0;
}

.header-subtitle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--td-text-color-secondary);
  margin: 0;
  padding: 4px 12px;
  background: var(--td-bg-color-container-hover, #f5f5f5);
  border-radius: 12px;
}

.info-card {
  margin-bottom: 20px;
  border-radius: 12px;
  overflow: hidden;
}

.info-card :deep(.t-card__body) {
  padding: 0;
}

.form-card {
  margin-bottom: 20px;
  border-radius: 12px;
}

.form-card :deep(.t-form__item) {
  margin-bottom: 20px;
}

.form-card :deep(.t-form__item:last-child) {
  margin-bottom: 0;
}

.form-card :deep(.t-input-number) {
  width: 100%;
}

.completed-card {
  border-radius: 12px;
}

.success-banner {
  margin-top: 16px;
}

.progress-overview {
  margin-bottom: 8px;
}

.progress-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 12px;
}

.progress-label {
  font-size: 14px;
  color: var(--td-text-color-secondary);
}

.progress-value {
  font-size: 20px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.progress-percent {
  font-size: 14px;
  color: var(--td-brand-color);
  font-weight: 500;
}

/* 模具工序 */
.process-groups {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.process-group-card {
  background: var(--td-bg-color-container, #fff);
  border: 1px solid var(--td-border-level-1-color, #e7e7e7);
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 12px;
}

.group-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--td-bg-color-container-hover, #f5f5f5);
  border-bottom: 1px solid var(--td-border-level-1-color, #e7e7e7);
}

.group-card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
}

.group-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--td-brand-color, #0052d9);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
}

.group-hint {
  font-size: 12px;
  color: var(--td-text-color-placeholder);
  font-weight: normal;
}

.group-card-processes {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 12px;
  padding: 12px 16px;
}

.process-item {
  border: 1px solid var(--td-border-level-2-color, #e7e7e7);
  border-radius: 8px;
  padding: 12px;
  background: #fff;
  transition: all 0.2s;
}

.process-item.process-completed {
  background: #f0faf5;
  border-color: #b3ebd2;
}

.process-item.process-skipped {
  background: #fafafa;
  border-color: #e0e0e0;
  opacity: 0.7;
}

.process-item.process-pending {
  background: #fff;
  border-color: #e0e0e0;
}

.process-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.process-item-name {
  font-size: 14px;
  font-weight: 500;
}

.process-item-action {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.process-empty-hint {
  text-align: center;
  color: var(--td-text-color-placeholder);
  padding: 24px;
}

.completed-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 24px 16px;
  text-align: center;
}

.completed-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.mold-code-highlight {
  font-size: 18px;
  font-weight: 700;
  color: var(--td-brand-color);
  letter-spacing: 0.5px;
}

.completed-desc {
  font-size: 14px;
  color: var(--td-text-color-secondary);
  line-height: 1.6;
}

.success-action {
  margin-top: 12px;
  display: flex;
  justify-content: flex-start;
}
</style>
