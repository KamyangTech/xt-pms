<template>
  <div class="status-dict-container">
    <!-- 操作栏卡片 -->
    <t-card class="toolbar-card">
      <div class="page-toolbar">
        <span class="toolbar-title">模具状态字典管理</span>
        <div class="toolbar-actions">
          <t-space>
            <t-button theme="default" @click="handleRefresh">
              <template #icon><refresh-icon /></template>
              刷新
            </t-button>
            <t-popconfirm
              content="确认恢复默认配置？此操作将重置状态字典和流转配置为初始值！"
              theme="danger"
              @confirm="handleRestoreDefault"
            >
              <t-button theme="danger" variant="outline">
                恢复默认配置
              </t-button>
            </t-popconfirm>
          </t-space>
        </div>
      </div>
    </t-card>

    <!-- Tabs选项卡 -->
    <t-card :bordered="false">
      <t-tabs v-model="activeTab" @change="handleTabChange">
        <!-- 模具档案状态 Tab -->
        <t-tab-panel value="mold-status" label="模具档案状态">
          <div class="tab-content">
            <t-alert theme="info" message="系统内置状态，仅支持修改颜色和备注" style="margin-bottom: 16px;" />
            <t-table
              :data="dictDataList"
              :columns="dataColumns"
              row-key="id"
              :loading="loading"
              :stripe="true"
            >
              <template #color="{ row }">
                <t-tag
                  :style="{
                    background: getDictColorHex(row.color) + ' !important',
                    color: (isLightColor(row.color) ? 'rgba(0, 0, 0, 0.9)' : '#fff') + ' !important',
                    borderColor: getDictColorHex(row.color) + ' !important'
                  }"
                >
                  {{ getDictColorName(row.color) }}
                </t-tag>
              </template>
              <template #allow_production="{ row }">
                <t-tag :theme="row.allow_production === 1 ? 'success' : 'default'">
                  {{ row.allow_production === 1 ? '是' : '否' }}
                </t-tag>
              </template>
              <template #operation="{ row }">
                <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
              </template>
            </t-table>
          </div>
        </t-tab-panel>

        <!-- 模具工单状态 Tab -->
        <t-tab-panel value="work-order-status" label="模具工单状态">
          <div class="tab-content">
            <t-alert theme="info" message="系统内置状态，仅支持修改颜色和备注" style="margin-bottom: 16px;" />
            <t-table
              :data="dictDataList"
              :columns="simpleDataColumns"
              row-key="id"
              :loading="loading"
              :stripe="true"
            >
              <template #color="{ row }">
                <t-tag
                  :style="{
                    background: getDictColorHex(row.color) + ' !important',
                    color: (isLightColor(row.color) ? 'rgba(0, 0, 0, 0.9)' : '#fff') + ' !important',
                    borderColor: getDictColorHex(row.color) + ' !important'
                  }"
                >
                  {{ getDictColorName(row.color) }}
                </t-tag>
              </template>
              <template #operation="{ row }">
                <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
              </template>
            </t-table>
          </div>
        </t-tab-panel>

        <!-- 模具订单状态 Tab -->
        <t-tab-panel value="mold-order-status" label="模具订单状态">
          <div class="tab-content">
            <t-alert theme="info" message="系统内置状态，仅支持修改颜色和备注" style="margin-bottom: 16px;" />
            <t-table
              :data="dictDataList"
              :columns="simpleDataColumns"
              row-key="id"
              :loading="loading"
              :stripe="true"
            >
              <template #color="{ row }">
                <t-tag
                  :style="{
                    background: getDictColorHex(row.color) + ' !important',
                    color: (isLightColor(row.color) ? 'rgba(0, 0, 0, 0.9)' : '#fff') + ' !important',
                    borderColor: getDictColorHex(row.color) + ' !important'
                  }"
                >
                  {{ getDictColorName(row.color) }}
                </t-tag>
              </template>
              <template #operation="{ row }">
                <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
              </template>
            </t-table>
          </div>
        </t-tab-panel>

        <!-- 状态流转配置 Tab -->
        <t-tab-panel value="status-flow-config" label="状态流转配置">
          <div class="tab-content">
            <t-alert theme="info" message="系统内置的基础流转配置，维持系统正常运转，不可修改" style="margin-bottom: 16px;" />
            <t-table
              :data="configList"
              :columns="flowColumns"
              row-key="id"
              :loading="flowLoading"
              :pagination="flowPagination"
              :stripe="true"
              @page-change="handleFlowPageChange"
            >
              <template #work_order_status="{ row }">
                <t-tag
                  v-if="row.work_order_status"
                  :theme="getStatusTheme('work_order', row.work_order_status)"
                  :style="getStatusTagStyle('work_order', row.work_order_status)"
                >
                  {{ getStatusLabel('work_order', row.work_order_status) }}
                </t-tag>
                <span v-else style="color: var(--td-text-color-placeholder);">-</span>
              </template>

              <template #mold_order_status="{ row }">
                <t-tag
                  v-if="row.mold_order_status"
                  :theme="getStatusTheme('mold_order', row.mold_order_status)"
                  :style="getStatusTagStyle('mold_order', row.mold_order_status)"
                >
                  {{ getStatusLabel('mold_order', row.mold_order_status) }}
                </t-tag>
                <span v-else style="color: var(--td-text-color-placeholder);">-</span>
              </template>

              <template #mold_status="{ row }">
                <t-tag
                  v-if="row.mold_status"
                  :theme="getStatusTheme('mold', row.mold_status)"
                  :style="getStatusTagStyle('mold', row.mold_status)"
                >
                  {{ getStatusLabel('mold', row.mold_status) }}
                </t-tag>
                <span v-else style="color: var(--td-text-color-placeholder);">-</span>
              </template>
            </t-table>
          </div>
        </t-tab-panel>
      </t-tabs>
    </t-card>

    <!-- 编辑对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      header="编辑状态显示属性"
      width="520px"
      placement="center"
      :confirm-btn="{ content: '确定', theme: 'primary' }"
      :cancel-btn="{ content: '取消', theme: 'default' }"
      @confirm="handleSubmit"
     :close-on-overlay-click="false">
      <t-form ref="formRef" :data="formModel" :rules="formRules" label-width="80px">
        <!-- 只读信息卡片 -->
        <t-card :bordered="true" :shadow="false" style="margin-bottom: 24px; background-color: var(--td-bg-color-secondarycontainer);">
          <t-row :gutter="[16, 16]">
            <t-col :span="3">
              <div class="info-item">
                <div class="info-label">状态名称</div>
                <div class="info-value">{{ formModel.dictLabel }}</div>
              </div>
            </t-col>
            <t-col :span="6">
              <div class="info-item">
                <div class="info-label">字典值</div>
                <div class="info-value">
                  <t-tag variant="light" theme="primary" style="font-size: 16px; font-weight: 600;">{{ formModel.dictValue }}</t-tag>
                </div>
              </div>
            </t-col>
          </t-row>
        </t-card>

        <!-- 颜色 -->
        <t-form-item label="颜色" name="color">
          <t-select
            v-model="formModel.color"
            placeholder="请选择状态颜色"
            clearable
          >
            <template #valueDisplay="{ value }">
              <div v-if="value" class="color-selected">
                <span class="color-swatch" :style="{ background: getDictColorHex(value) }" />
                <span class="color-name">{{ getDictColorName(value) }}</span>
                <t-tag size="small" variant="light" theme="default" class="color-hex">{{ getDictColorHex(value) }}</t-tag>
              </div>
              <span v-else class="placeholder-text">请选择颜色</span>
            </template>

            <t-option-group
              v-for="group in colorOptionGroups"
              :key="group.group"
              :label="group.group"
            >
              <t-option
                v-for="opt in group.options"
                :key="opt.value"
                :value="opt.value"
                class="color-option"
              >
                <div class="color-option-content">
                  <span class="color-swatch-option" :style="{ background: opt.hex }" />
                  <span class="color-option-label">{{ opt.label.split('（')[0] }}</span>
                  <t-tag size="small" variant="light" theme="default" class="color-hex-tag">{{ opt.hex }}</t-tag>
                </div>
              </t-option>
            </t-option-group>
          </t-select>
        </t-form-item>

        <t-form-item label="备注" name="remark">
          <t-textarea 
            v-model="formModel.remark" 
            placeholder="请输入备注信息（可选）" 
            :rows="3"
            :maxlength="200"
            :show-limit="true"
            clearable
          />
        </t-form-item>
      </t-form>
    </t-dialog>


  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { TableProps, TabValue } from 'tdesign-vue-next';
import {
  RefreshIcon,
} from 'tdesign-icons-vue-next';
import {
  getDictDataByType,
  updateDictData,
  getDictTypes,
  clearDictCache,
  restoreDictSeed,
} from '@/api/dict';
import {
  getDictColorHex,
  getDictColorName,
  getDictColorTheme,
  getDictColorRenderConfig,
  isLightColor,
  getAllColorOptions,
  type ColorOption,
} from '@/utils/dict-color';
import {
  getStatusFlowConfigs,
} from '@/api/system/statusFlow';

defineOptions({
  name: 'StatusDictIndex',
});

// 当前激活的Tab
const activeTab = ref('mold-status');

// 表格相关
const loading = ref(false);
const dictDataList = ref<any[]>([]);

// 颜色选项（从统一配置动态生成，按功能色/中性色/扩展色分组）
const colorOptionGroups = getAllColorOptions();

// 编辑对话框相关
const dialogVisible = ref(false);
const formRef = ref();

// 表单模型（只用于编辑，只提交 color 和 remark）
const formModel = reactive({
  id: undefined as number | undefined,
  dictLabel: '',
  dictValue: '',
  color: 'default',
  allow_production: 0,
  remark: '',
});

const formRules = {};

// 表格列配置
const dataColumns: TableProps['columns'] = [
  { colKey: 'sort', title: '排序', width: 80 },
  { colKey: 'dictLabel', title: '状态名称', width: 150 },
  { colKey: 'dictValue', title: '字典值', width: 140 },
  { colKey: 'color', title: '颜色', width: 120 },
  { colKey: 'allow_production', title: '可用于生产', width: 120 },
  { colKey: 'remark', title: '备注', width: 200 },
  { colKey: 'operation', title: '操作', width: 80, fixed: 'right' as const },
];

// 工单和订单状态的列配置（无"可用于生产"列）
const simpleDataColumns: TableProps['columns'] = [
  { colKey: 'sort', title: '排序', width: 80 },
  { colKey: 'dictLabel', title: '状态名称', width: 150 },
  { colKey: 'dictValue', title: '字典值', width: 140 },
  { colKey: 'color', title: '颜色', width: 120 },
  { colKey: 'remark', title: '备注', width: 200 },
  { colKey: 'operation', title: '操作', width: 80, fixed: 'right' as const },
];

// Tab切换
const handleTabChange = (value: TabValue) => {
  activeTab.value = value as string;
  if (value === 'status-flow-config') {
    getFlowList();
  } else {
    loadDictData();
  }
};

const loadDictData = async () => {
  if (!activeTab.value) return;

  loading.value = true;
  try {
    const typeRes = await getDictTypes({});
    const dictType = typeRes.list?.find((item: any) => item.dict_code === activeTab.value);

    if (dictType) {
      const res = await getDictDataByType({ dict_type: activeTab.value });
      dictDataList.value = (res.list || []).map((item: any) => ({
        id: item.id,
        dict_type_id: item.dict_type_id,
        dictLabel: item.dict_label,
        dictValue: item.dict_value,
        color: item.color,
        sort: item.sort,
        status: item.status,
        allow_production: item.allow_production || 0,
        is_locked: item.is_locked || 0,
        remark: item.remark,
      }));
    } else {
      MessagePlugin.warning('未找到对应的字典类型，请先在字典类型管理中创建');
      dictDataList.value = [];
    }
  } catch (error) {
    console.error('加载字典数据失败:', error);
    MessagePlugin.error('加载字典数据失败');
  } finally {
    loading.value = false;
  }
};

// 编辑（只允许编辑颜色和备注）
const handleEdit = (row: any) => {
  Object.assign(formModel, {
    id: row.id,
    dictLabel: row.dictLabel,
    dictValue: row.dictValue,
    color: row.color || 'default',
    allow_production: row.allow_production !== undefined ? row.allow_production : 0,
    remark: row.remark || '',
  });
  setTimeout(() => {
    formRef.value?.clearValidate?.();
  }, 100);
  dialogVisible.value = true;
};

// 刷新
const handleRefresh = async () => {
  try {
    await clearDictCache();
  } catch {
    // 清除缓存失败不影响重新加载
  }
  loadDictData();
  getFlowList();
};

// 恢复默认配置
const handleRestoreDefault = async () => {
  try {
    await restoreDictSeed();
    MessagePlugin.success('已恢复默认配置');
    // 刷新当前 tab 数据
    if (activeTab.value === 'status-flow-config') {
      getFlowList();
    } else {
      loadDictData();
    }
  } catch (error) {
    MessagePlugin.error('恢复默认配置失败');
  }
};

// 提交表单（只提交 color 和 remark）
const handleSubmit = async () => {
  if (!formModel.id) {
    MessagePlugin.error('未找到要编辑的状态');
    return;
  }

  try {
    const submitData: any = {
      id: Number(formModel.id),
      color: formModel.color,
      remark: formModel.remark,
    };

    await updateDictData(submitData);
    MessagePlugin.success('更新成功');
    dialogVisible.value = false;
    loadDictData();
  } catch (error) {
    MessagePlugin.error('更新失败');
  }
};

// ============ 状态流转配置相关 ============

const flowLoading = ref(false);
const configList = ref<any[]>([]);

const flowDictData = ref<any>({
  work_order: [],
  mold_order: [],
  mold: [],
});

const flowPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const flowColumns: TableProps['columns'] = [
  { colKey: 'priority', title: '优先级', width: 100 },
  { colKey: 'work_order_status', title: '工单状态', width: 150 },
  { colKey: 'mold_order_status', title: '订单状态', width: 150 },
  { colKey: 'mold_status', title: '模具状态', width: 150 },
  { colKey: 'description', title: '说明', ellipsis: true },
];

const getStatusLabel = (type: string, statusValue: string) => {
  if (!type || !statusValue) return '';
  const statusList = flowDictData.value[type] || [];
  const status = statusList.find((s: any) => s.value === statusValue);
  return status ? status.label : statusValue;
};

const getStatusTheme = (type: string, statusValue: string): 'default' | 'primary' | 'success' | 'warning' | 'danger' => {
  if (!type || !statusValue) return 'default';

  const statusList = flowDictData.value[type] || [];
  if (statusList.length === 0) return 'default';

  const status = statusList.find((s: any) => s.value === statusValue);
  if (!status) return 'default';

  const color = status.color || 'default';
  return getDictColorTheme(color);
};

const getStatusTagStyle = (type: string, statusValue: string) => {
  if (!type || !statusValue) return {};

  const statusList = flowDictData.value[type] || [];
  if (statusList.length === 0) return {};

  const status = statusList.find((s: any) => s.value === statusValue);
  if (!status || !status.color) return {};

  return (getDictColorRenderConfig(status.color).style || {});
};

const loadFlowDictData = async () => {
  try {
    const res = await getDictTypes();
    const dictList = res.list || [];

    dictList.forEach((dict: any) => {
      switch (dict.dict_code) {
        case 'work-order-status':
          flowDictData.value.work_order = (dict.data_list || []).map((item: any) => ({
            value: item.dict_value,
            label: item.dict_label,
            color: item.color || 'default',
          }));
          break;
        case 'mold-order-status':
          flowDictData.value.mold_order = (dict.data_list || []).map((item: any) => ({
            value: item.dict_value,
            label: item.dict_label,
            color: item.color || 'default',
          }));
          break;
        case 'mold-status':
          flowDictData.value.mold = (dict.data_list || []).map((item: any) => ({
            value: item.dict_value,
            label: item.dict_label,
            color: item.color || 'default',
          }));
          break;
      }
    });
  } catch (error) {
    console.error('加载字典数据失败:', error);
  }
};

const getFlowList = async () => {
  flowLoading.value = true;
  try {
    const res = await getStatusFlowConfigs();
    const list = res.list || res || [];
    configList.value = list;
    flowPagination.total = list.length;
  } catch (error) {
    MessagePlugin.error('获取列表失败');
  } finally {
    flowLoading.value = false;
  }
};

const handleFlowPageChange = (pageInfo: any) => {
  flowPagination.current = pageInfo.current;
  flowPagination.pageSize = pageInfo.pageSize;
  getFlowList();
};

onMounted(async () => {
  await loadFlowDictData();
  loadDictData();
  if (activeTab.value === 'status-flow-config') {
    getFlowList();
  }
});
</script>

<style scoped>
.status-dict-container {
  padding: 16px;
}

.page-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.toolbar-card {
  margin-bottom: 16px;
}

.toolbar-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.toolbar-actions {
  display: flex;
  gap: 8px;
}

.tab-content {
  padding-top: 16px;
}

.toolbar {
  margin-bottom: 16px;
}

/* 编辑弹窗信息卡片样式 */
.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-label {
  font-size: 12px;
  color: var(--td-text-color-placeholder);
  line-height: 1.5;
}

.info-value {
  font-size: 14px;
  color: var(--td-text-color-primary);
  line-height: 1.5;
  display: flex;
  align-items: center;
}

.text-placeholder {
  color: var(--td-text-color-placeholder);
}

/* 颜色选择器选项样式 */
.color-option {
  padding: 6px 8px;
}

.color-option-content {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.color-option-label {
  flex: 1;
  font-size: 14px;
}

.color-hex-tag {
  opacity: 0.6;
  font-size: 11px !important;
}

/* 颜色色块 */
.color-swatch {
  display: inline-block;
  width: 18px;
  height: 18px;
  border-radius: var(--td-radius-small);
  border: 1px solid rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
}

.color-swatch-option {
  display: inline-block;
  width: 20px;
  height: 20px;
  border-radius: var(--td-radius-small);
  border: 1px solid rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
}

/* 选中值展示 */
.color-selected {
  display: flex;
  align-items: center;
  gap: 8px;
}

.color-name {
  font-size: 14px;
  font-weight: 500;
}

.color-hex {
  opacity: 0.5;
  font-size: 11px !important;
  margin-left: auto;
}

.placeholder-text {
  color: var(--td-text-color-placeholder);
}

/* Dialog 内卡片自适应间距 */
:deep(.t-card__body) {
  padding: 16px;
}
</style>
