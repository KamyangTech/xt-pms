<template>
  <t-drawer
    v-model:visible="drawerVisible"
    :header="workOrder ? `工序详情 - ${workOrder.mold?.mold_code || '-'}` : '工序详情'"
    size="600px"
    :footer="false"
    @close="handleClose"
  >
    <div v-if="workOrder" class="process-drawer-content">
      <!-- 按分组展示工序 -->
      <div class="process-groups-drawer">
        <template v-for="(group, groupIndex) in groupedProcesses" :key="group.groupId">
          <div class="process-group-card">
            <div class="group-card-header">
              <div class="group-card-title">
                <span class="group-badge">{{ group.groupId }}</span>
                <span>{{ group.groupName }}</span>
                <span class="group-hint">（可并行）</span>
              </div>
              <t-tag v-if="groupIndex > 0" theme="default" size="small">
                <t-icon name="arrow-down" /> 上一组完成后开始
              </t-tag>
            </div>
            
            <div class="group-card-processes">
              <div
                v-for="(template, templateIndex) in group.templates"
                :key="template.id"
                class="process-item"
                :class="getProcessItemClass(workOrder, Number(template.id), templateIndex, group.templates, groupIndex)"
              >
                <div class="process-item-header">
                  <span class="process-item-name">{{ template.process_name }}</span>
                  <t-tag
                    v-if="getProcessStatusValue(workOrder, template.id) === PROCESS_STATUS.COMPLETED"
                    theme="success"
                    size="small"
                  >
                    已完成
                  </t-tag>
                  <t-tag
                    v-else-if="getProcessStatusValue(workOrder, template.id) === PROCESS_STATUS.SKIPPED"
                    theme="default"
                    size="small"
                    variant="light-outline"
                  >
                    已跳过
                  </t-tag>
                  <t-tag
                    v-else-if="isGroupCanBeStarted(workOrder, groupIndex)"
                    theme="primary"
                    size="small"
                  >
                    进行中
                  </t-tag>
                  <t-tag
                    v-else
                    theme="default"
                    size="small"
                  >
                    待开工
                  </t-tag>
                </div>
                <div class="process-item-action">
                  <!-- 未完成状态：显示操作按钮 -->
                  <div v-if="getProcessStatusValue(workOrder, template.id) === PROCESS_STATUS.PENDING" class="action-buttons">
                    <t-tooltip :content="getMutexTip(workOrder) || (!isCurrentProcess(workOrder, template.id) ? '需等待上一组完成后开始' : '')" placement="top">
                      <t-button
                        theme="success"
                        variant="outline"
                        size="small"
                        shape="round"
                        :disabled="isWorkOrderLocked(workOrder) || workOrder?.is_cancelled === 1 || !isCurrentProcess(workOrder, template.id)"
                        @click="toggleProcessStatus(workOrder, template.id, PROCESS_STATUS.COMPLETED)"
                        class="action-btn-complete"
                      >
                        <template #icon><check-icon /></template>
                        完成
                      </t-button>
                    </t-tooltip>
                    <t-tooltip :content="getMutexTip(workOrder) || (!isCurrentProcess(workOrder, template.id) ? '需等待上一组完成后开始' : '')" placement="top">
                      <t-button
                        theme="default"
                        variant="outline"
                        size="small"
                        shape="round"
                        :disabled="isWorkOrderLocked(workOrder) || workOrder?.is_cancelled === 1 || !isCurrentProcess(workOrder, template.id)"
                        @click="toggleProcessStatus(workOrder, template.id, PROCESS_STATUS.SKIPPED)"
                        class="action-btn-skip"
                      >
                        <template #icon><close-circle-icon /></template>
                        跳过
                      </t-button>
                    </t-tooltip>
                  </div>
                  
                  <!-- 已完成状态：显示取消按钮 -->
                  <div v-else-if="getProcessStatusValue(workOrder, template.id) === PROCESS_STATUS.COMPLETED" class="action-buttons">
                    <t-tooltip :content="getMutexTip(workOrder) || (getStatusLevel(workOrder?.status) >= WORK_ORDER_STATUS.COMPLETED ? '工单已进入后期阶段，禁止回退工序操作' : '')" placement="top">
                      <t-button
                        theme="default"
                        variant="text"
                        size="small"
                        :disabled="isWorkOrderLocked(workOrder) || workOrder?.is_cancelled === 1 || getStatusLevel(workOrder?.status) >= WORK_ORDER_STATUS.COMPLETED"
                        @click="toggleProcessStatus(workOrder, template.id, PROCESS_STATUS.PENDING)"
                        class="action-btn-cancel"
                      >
                        撤销
                      </t-button>
                    </t-tooltip>
                  </div>
                  
                  <!-- 已跳过状态：显示取消按钮 -->
                  <div v-else-if="getProcessStatusValue(workOrder, template.id) === PROCESS_STATUS.SKIPPED" class="action-buttons">
                    <t-tooltip :content="getMutexTip(workOrder) || (getStatusLevel(workOrder?.status) >= WORK_ORDER_STATUS.COMPLETED ? '工单已进入后期阶段，禁止回退工序操作' : '')" placement="top">
                      <t-button
                        theme="default"
                        variant="text"
                        size="small"
                        :disabled="isWorkOrderLocked(workOrder) || workOrder?.is_cancelled === 1 || getStatusLevel(workOrder?.status) >= WORK_ORDER_STATUS.COMPLETED"
                        @click="toggleProcessStatus(workOrder, template.id, PROCESS_STATUS.PENDING)"
                        class="action-btn-cancel"
                      >
                        撤销
                      </t-button>
                    </t-tooltip>
                  </div>
                  
                  <span v-if="getProcessCompletedAt(workOrder, template.id)" class="process-item-time">
                    {{ formatCompletedTime(getProcessCompletedAt(workOrder, template.id)) }}
                  </span>
                </div>
              </div>
            </div>
          </div>
          
          <!-- 组间分隔箭头 -->
          <div v-if="groupIndex < groupedProcesses.length - 1" class="group-arrow-divider">
            <t-icon name="arrow-down" size="24px" />
          </div>
        </template>
      </div>

      <!-- 进度统计 -->
      <t-card class="process-stats" :bordered="false">
        <div class="stats-container">
          <div class="stat-item">
            <div class="stat-value">{{ getCompletedCount(workOrder) }}</div>
            <div class="stat-label">已完成</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ getInProgressCount(workOrder) }}</div>
            <div class="stat-label">进行中</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ getPendingCount(workOrder) }}</div>
            <div class="stat-label">待开工</div>
          </div>
        </div>
      </t-card>
    </div>
  </t-drawer>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next';
import { CheckIcon, CloseCircleIcon } from 'tdesign-icons-vue-next';
import dayjs from 'dayjs';
import { getActiveProcessTemplates } from '@/api/reference';
import { toggleProcess } from '@/api/process-template';
import { WORK_ORDER_STATUS, PROCESS_STATUS } from '@/constants';

interface Props {
  visible: boolean;
  workOrder: any;
}

interface Emits {
  (e: 'update:visible', value: boolean): void;
  (e: 'close'): void;
  (e: 'refresh', workOrderId: number): void;
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  workOrder: null,
});

const emit = defineEmits<Emits>();

const drawerVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});

const processTemplates = ref<any[]>([]);

// 关闭抽屉
const handleClose = () => {
  emit('close');
};

const loadProcessTemplates = async () => {
  try {
    const res = await getActiveProcessTemplates();
    processTemplates.value = res || [];
  } catch (error) {
    console.error('加载工序模板失败:', error);
    MessagePlugin.error('加载工序模板失败');
  }
};

// 按分组组织工序
const groupedProcesses = computed(() => {
  if (!processTemplates.value || processTemplates.value.length === 0) {
    return [];
  }
  
  const groups: Map<number, any> = new Map();
  
  processTemplates.value.forEach(template => {
    const groupId = template.group_id || 1;
    if (!groups.has(groupId)) {
      groups.set(groupId, {
        groupId,
        groupName: template.group_name || `第${groupId}组`,
        templates: [],
      });
    }
    groups.get(groupId).templates.push(template);
  });
  
  // 按 group_id 排序
  return Array.from(groups.values()).sort((a, b) => a.groupId - b.groupId) as Array<{
    groupId: number;
    groupName: string;
    templates: any[];
  }>;
});

const getProcessStatusValue = (workOrder: any, templateId: number | string): number => {
  if (!workOrder?.process_records) return PROCESS_STATUS.PENDING;
  const record = workOrder.process_records.find((r: any) => r.template_id === Number(templateId));
  return record?.status || PROCESS_STATUS.PENDING;
};

const getProcessCompletedAt = (workOrder: any, templateId: number | string): string | null => {
  if (!workOrder?.process_records) return null;
  const record = workOrder.process_records.find((r: any) => r.template_id === Number(templateId));
  return record?.completed_at || null;
};

// 格式化完成时间
const formatCompletedTime = (time: string): string => {
  if (!time) return '';
  return dayjs(time).format('MM-DD HH:mm');
};

// 状态优先级映射（后端返回语义化字符串如 "COMPLETED"，前端常量是数字）
const STATUS_LEVEL_MAP: Record<string, number> = {
  NOT_STARTED: 1,
  PROCESSING: 2,
  COMPLETED: 3,
  TESTED: 4,
  STOCKED_IN: 5,
  DELIVERED: 6,
  CANCELLED: 7,
};

// 获取状态级别（兼容后端字符串和前端数字）
const getStatusLevel = (status: any): number => {
  if (typeof status === 'number') return status;
  return STATUS_LEVEL_MAP[String(status)] ?? -1;
};

const isWorkOrderLocked = (workOrder: any): boolean => {
  return workOrder?.is_cancelled === 1 || getStatusLevel(workOrder?.status) >= WORK_ORDER_STATUS.COMPLETED;
};

const getMutexTip = (workOrder: any): string => {
  if (workOrder?.is_cancelled === 1) {
    return '工单已取消，无法操作工序';
  }
  if (getStatusLevel(workOrder?.status) >= WORK_ORDER_STATUS.COMPLETED) {
    return '工单已进入后期阶段，禁止回退工序操作';
  }
  return '';
};

// 判断是否为当前可操作的工序
const isCurrentProcess = (workOrder: any, templateId: number | string): boolean => {
  if (!workOrder || !processTemplates.value) return false;
  
  const template = processTemplates.value.find((t: any) => t.id === Number(templateId));
  if (!template) return false;
  
  const groupId = template.group_id || 1;
  
  if (groupId === 1) return true;
  
  const prevGroup = groupedProcesses.value.find((g: any) => g.groupId === groupId - 1);
  if (!prevGroup) return true;
  
  return prevGroup.templates.every((t: any) => {
    const status = getProcessStatusValue(workOrder, t.id);
    return status === PROCESS_STATUS.COMPLETED || status === PROCESS_STATUS.SKIPPED;
  });
};

const isGroupCanBeStarted = (workOrder: any, groupIndex: number): boolean => {
  if (groupIndex === 0) return true;
  
  const prevGroup = groupedProcesses.value[groupIndex - 1];
  if (!prevGroup) return true;
  
  return prevGroup.templates.every((t: any) => {
    const status = getProcessStatusValue(workOrder, t.id);
    return status === PROCESS_STATUS.COMPLETED || status === PROCESS_STATUS.SKIPPED;
  });
};

const getProcessItemClass = (
  workOrder: any,
  templateId: number | string,
  templateIndex: number,
  templates: any[],
  groupIndex: number
): string => {
  const status = getProcessStatusValue(workOrder, templateId);
  const classes: string[] = [];
  
  if (status === PROCESS_STATUS.COMPLETED) {
    classes.push('process-item-completed');
  } else if (status === PROCESS_STATUS.SKIPPED) {
    classes.push('process-item-skipped');
  } else if (isCurrentProcess(workOrder, templateId)) {
    classes.push('process-item-active');
  } else {
    classes.push('process-item-pending');
  }
  
  return classes.join(' ');
};

// 切换工序状态
const toggleProcessStatus = async (workOrder: any, templateId: number | string, targetStatus: number) => {
  if (!workOrder?.id) {
    MessagePlugin.error('工单信息不存在');
    return;
  }
  
  const statusText = targetStatus === PROCESS_STATUS.COMPLETED ? '完成' : targetStatus === PROCESS_STATUS.SKIPPED ? '跳过' : '撤销';
  
  const dialog = DialogPlugin.confirm({
    header: '确认操作',
    body: `确定要${statusText}该工序吗？`,
    onConfirm: async () => {
      try {
        await toggleProcess({
          work_order_id: workOrder.id,
          process_template_id: Number(templateId),
          is_completed: targetStatus,
        });
        MessagePlugin.success(`${statusText}成功`);
        
        // 触发刷新事件
        emit('refresh', workOrder.id);
      } catch (error: any) {
        MessagePlugin.error(error.message || `${statusText}失败`);
      }
      dialog.destroy();
    },
  });
};

const getCompletedCount = (workOrder: any): number => {
  if (!workOrder?.process_records) return 0;
  return workOrder.process_records.filter((r: any) => r.status === PROCESS_STATUS.COMPLETED).length;
};

const getInProgressCount = (workOrder: any): number => {
  if (!workOrder?.process_records) return 0;
  return workOrder.process_records.filter((r: any) => r.status === PROCESS_STATUS.PENDING).length;
};

const getPendingCount = (workOrder: any): number => {
  if (!workOrder?.process_records) return 0;
  return workOrder.process_records.filter((r: any) => r.status === PROCESS_STATUS.PENDING).length;
};

onMounted(() => {
  loadProcessTemplates();
});

defineExpose({
  loadProcessTemplates,
});
</script>

<style scoped>
/* 工序详情抽屉样式 */
.process-drawer-content {
  padding: 16px;
}

.process-groups-drawer {
  margin-bottom: 16px;
}

.process-group-card {
  border: 1px solid var(--td-border-level-1-color);
  border-radius: 8px;
  margin-bottom: 12px;
  overflow: hidden;
  background: var(--td-bg-color-container);
}

.group-card-header {
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--td-border-level-1-color);
  margin-bottom: 0;
}

.group-card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.group-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: var(--td-brand-color);
  color: #fff;
  border-radius: 50%;
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
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
  padding: 0 16px 16px 16px;
}

.process-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background: var(--td-bg-color-container-hover);
  border-radius: 6px;
  transition: all 0.3s ease;
  border: 1px solid var(--td-border-level-1-color);
}

/* 已完成工序 - 绿色边框 */
.process-item-completed {
  border-color: var(--td-success-color) !important;
}

/* 跳过工序 - 灰色虚线边框 */
.process-item-skipped {
  border-color: var(--td-gray-color-5) !important;
  border-style: dashed !important;
  opacity: 0.7;
}

/* 进行中工序 - 蓝色边框 */
.process-item-active {
  border-color: var(--td-brand-color) !important;
}

/* 待开工工序 - 灰色边框 */
.process-item-pending {
  border-color: var(--td-gray-color-3) !important;
}

.process-item:hover {
  background: var(--td-bg-color-container-active);
}

.process-item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.process-item-name {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
  flex: 1;
}

.process-item-action {
  margin-top: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.action-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.action-btn-complete,
.action-btn-skip {
  min-width: 80px;
  padding: 4px 16px !important;
  font-size: 13px !important;
  border-radius: 20px !important;
  transition: all 0.2s ease;
}

.action-btn-complete:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(27, 200, 138, 0.3);
}

.action-btn-skip:hover {
  background-color: var(--td-bg-color-container-active, #ebebeb) !important;
  transform: translateY(-1px);
}

.action-btn-cancel {
  font-size: 12px !important;
  color: var(--td-text-color-secondary) !important;
  padding: 4px 8px !important;
  transition: all 0.2s ease;
}

.action-btn-cancel:hover {
  color: var(--td-brand-color) !important;
}

.process-item-time {
  font-size: 14px;
  font-weight: 500;
  color: var(--td-text-color-primary);
  white-space: nowrap;
}

.group-arrow-divider {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 8px 0;
  color: var(--td-text-color-placeholder);
}

.process-stats {
  margin-top: 16px;
}

.stats-container {
  display: flex;
  justify-content: space-around;
  text-align: center;
}

.stat-item {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--td-brand-color);
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: var(--td-text-color-secondary);
}
</style>
