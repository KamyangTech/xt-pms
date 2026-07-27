<template>
  <div class="business-config-container">
    <!-- 操作栏卡片 -->
    <t-card class="toolbar-card">
      <div class="page-toolbar">
        <span class="toolbar-title">业务配置</span>
        <div class="toolbar-actions">
          <span class="group-label">业务大类：</span>
          <t-select
            v-model="selectedGroup"
            style="width: 160px;"
            placeholder="请选择大类"
            @change="handleGroupChange"
          >
            <t-option
              v-for="group in categoryGroups"
              :key="group.key"
              :value="group.key"
              :label="group.label"
            />
          </t-select>
          <div v-if="selectedGroup && activeTab !== 'process-template'" class="add-btn-wrapper" :title="isProtectedType ? '当前字典类型由系统固件控制，不允许新增' : ''">
            <t-button theme="primary" :disabled="isProtectedType" @click="handleAdd">
              <template #icon><add-icon /></template>
              新增
            </t-button>
          </div>
          <div v-if="activeTab === 'process-template'">
            <t-button theme="primary" @click="handlePtAdd">
              <template #icon><add-icon /></template>
              新增
            </t-button>
          </div>
          <t-button theme="default" @click="handleRefresh">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 卡片：Tab切换 + 数据表格 -->
    <t-card :bordered="false">
      <!-- 子分类 Tabs（根据选择的大类动态渲染） -->
      <template v-if="selectedGroup && currentGroupTabs.length > 0">
        <t-tabs v-model="activeTab" @change="handleTabChange">
          <t-tab-panel
            v-for="tab in currentGroupTabs"
            :key="tab.value"
            :value="tab.value"
            :label="tab.label"
          >
            <div v-if="tab.value === 'process-template'" class="tab-content">
              <t-table
                :data="ptTableData"
                :columns="ptColumns"
                :loading="ptLoading"
                row-key="id"
                :stripe="true"
              >
                <template #status="{ row }">
                  <t-tag :theme="row.status === 1 ? 'success' : 'default'">
                    {{ row.status === 1 ? '启用' : '禁用' }}
                  </t-tag>
                </template>
                <template #created_at="{ row }">{{ formatDateTime(row.created_at) }}</template>
                <template #operation="{ row }">
                  <t-link theme="primary" @click="handlePtEdit(row)" style="margin-right: 8px">编辑</t-link>
                  <t-popconfirm content="确认删除该工序？" @confirm="handlePtDelete(row)">
                    <t-link theme="danger">删除</t-link>
                  </t-popconfirm>
                </template>
              </t-table>
            </div>
            <div v-else class="tab-content">
              <t-table
                :data="dictDataList"
                :columns="tableColumns"
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
                <template #needs_delivery_unlock="{ row }">
                  <t-tag :theme="row.needs_delivery_unlock === 1 ? 'warning' : 'default'" variant="light">
                    {{ row.needs_delivery_unlock === 1 ? '是' : '否' }}
                  </t-tag>
                </template>
                <template #status="{ row }">
                  <t-tag :theme="row.status === 1 ? 'success' : 'default'">
                    {{ row.status === 1 ? '启用' : '禁用' }}
                  </t-tag>
                </template>
                <template #operation="{ row }">
                  <t-space>
                    <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
                    <t-popconfirm v-if="!isRowValueProtected(row.dictValue)" content="确认删除？" @confirm="handleDelete(row)">
                      <t-link theme="danger">删除</t-link>
                    </t-popconfirm>
                    <t-link v-else theme="danger" disabled style="cursor: not-allowed; opacity: 0.4;" :title="`「${row.dictLabel}」由系统固件控制，不可删除`">删除</t-link>
                  </t-space>
                </template>
              </t-table>
            </div>
          </t-tab-panel>
        </t-tabs>
      </template>
      <div v-else class="empty-hint">请先选择一个业务大类</div>
    </t-card>

    <!-- 新增/编辑对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :title="dialogTitle"
      width="600px"
      placement="center"
     :close-on-overlay-click="false">
      <t-form ref="formRef" :data="formModel" :rules="formRules" label-width="100px">
        <t-row :gutter="[16, 0]" style="margin-left: 0; margin-right: 0;">
          <t-col :span="12">
            <t-form-item label="显示标签" name="dictLabel">
              <t-input v-model="formModel.dictLabel" placeholder="请输入显示标签" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="字典值" name="dictValue">
              <t-input v-model="formModel.dictValue" placeholder="请输入字典值" :disabled="(isEdit && isProtectedType) || isEditingProtectedValue" />
              <template v-if="isEdit && (isProtectedType || isEditingProtectedValue)" #help>
                <span style="color: var(--td-warning-color); font-size: 12px;">
                  {{ isProtectedType ? protectedTypeHint : '该字典值为系统默认值，不可修改' }}
                </span>
              </template>
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="颜色" name="color">
              <DictColorPicker v-model="formModel.color" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="排序" name="sort">
              <t-input-number v-model="formModel.sort" :min="0" placeholder="请输入排序" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="状态" name="status">
              <t-select v-model="formModel.status" placeholder="请选择状态" :disabled="isEdit && isProtectedType">
                <t-option :label="'启用'" :value="1" />
                <t-option :label="'禁用'" :value="0" />
              </t-select>
            </t-form-item>
          </t-col>
          <!-- 款项类型专属：交付后解锁开关 -->
          <t-col v-if="activeTab === 'payment-type'" :span="12">
            <t-form-item label="交付后解锁" name="needs_delivery_unlock">
              <t-switch
                v-model="formModel.needs_delivery_unlock"
                :custom-value="[1, 0]"
              >
                <template #label="slotProps">
                  {{ slotProps.value === 1 ? '是' : '否' }}
                </template>
              </t-switch>
              <div style="margin-left: 8px; font-size: 12px; color: var(--td-text-color-secondary);">
                开启后，该款项类型在交付前默认锁定，交付完成后自动解锁可收款
              </div>
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="备注" name="remark">
              <t-textarea v-model="formModel.remark" placeholder="请输入备注" :rows="3" />
            </t-form-item>
          </t-col>
        </t-row>
      </t-form>
      <template #footer>
        <t-button theme="default" @click="handleCancel">取消</t-button>
        <t-button theme="primary" style="margin-left: 8px" @click="handleSubmit">确定</t-button>
      </template>
    </t-dialog>

    <!-- ========== 工序模板编辑对话框 ========== -->
    <t-dialog
      v-model:visible="ptDialogVisible"
      :header="ptDialogTitle"
      width="600px"
      placement="center"
      @confirm="handlePtSubmit"
      @cancel="handlePtCancel"
     :close-on-overlay-click="false">
      <t-form ref="ptFormRef" :model="ptFormModel" :rules="ptFormRules" :label-width="100" style="margin-top: 16px">
        <t-form-item label="分组" name="groupName">
          <t-select v-model="ptFormModel.groupId" placeholder="请选择分组" @change="handlePtGroupChange">
            <t-option v-for="item in ptGroupOptions" :key="item.value" :label="item.label" :value="item.value" />
          </t-select>
        </t-form-item>
        <t-form-item label="工序名称" name="processName">
          <t-input v-model="ptFormModel.processName" placeholder="请输入工序名称" />
        </t-form-item>
        <t-form-item label="排序号" name="sort">
          <t-input-number v-model="ptFormModel.sort" :min="1" placeholder="请输入排序号" style="width: 100%" />
        </t-form-item>
        <t-form-item label="状态" name="status">
          <t-radio-group v-model="ptFormModel.status">
            <t-radio :value="1">启用</t-radio>
            <t-radio :value="2">禁用</t-radio>
          </t-radio-group>
        </t-form-item>
        <t-form-item label="备注" name="remark">
          <t-textarea v-model="ptFormModel.remark" placeholder="请输入备注信息" :maxlength="500" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { TableProps, TabValue } from 'tdesign-vue-next';
import { AddIcon, RefreshIcon } from 'tdesign-icons-vue-next';
import {
  getDictDataByType,
  createDictData,
  updateBusinessDictData,
  deleteDictData,
  getDictTypes,
} from '@/api/dict';
import {
  getProcessTemplates,
  createProcessTemplate,
  updateProcessTemplate,
  deleteProcessTemplate,
} from '@/api/process-template';
import {
  getColorHex as getDictColorHex,
  getColorName as getDictColorName,
  isLightColor,
} from '@/config/dict-color-config';
import DictColorPicker from '@/components/DictColorPicker/index.vue';
import { formatDateTime } from '@/utils/date';
import { dictUtils } from '@/utils/dict';

defineOptions({
  name: 'BusinessConfigIndex',
});

// ============ 业务分类配置 ============
interface TabItem {
  value: string;
  label: string;
}

interface GroupConfig {
  key: string;
  label: string;
  tabs: TabItem[];
}

const categoryGroups: GroupConfig[] = [
  {
    key: 'mold',
    label: '模具信息',
    tabs: [
      { value: 'mold-process', label: '模具工艺管理' },
      { value: 'material-type', label: '材质类型管理' },
      { value: 'mold-type', label: '模胚情况管理' },
      { value: 'test-mode', label: '试模方式管理' },
    ],
  },
  {
    key: 'order',
    label: '订单/工单',
    tabs: [
      { value: 'mold-order-type', label: '订单/工单类型' },
      { value: 'process-template', label: '工序模板管理' },
    ],
  },
  {
    key: 'production',
    label: '生产信息',
    tabs: [
      { value: 'order-priority', label: '订单优先级' },
      { value: 'production-type', label: '生产类型' },
      { value: 'mold-source', label: '模具来源' },
      { value: 'production-equipment', label: '生产设备' },
      { value: 'responsible-employee', label: '负责员工' },
      { value: 'plastic-material', label: '塑胶原料' },
    ],
  },
  {
    key: 'inventory',
    label: '库存管理',
    tabs: [
      { value: 'delivery_method', label: '发货方式' },
      { value: 'express_company', label: '快递公司' },
      { value: 'out_reason', label: '出库原因' },
    ],
  },
  {
    key: 'finance',
    label: '财务',
    tabs: [
      { value: 'payment-method', label: '收款付款方式' },
      { value: 'receivable-status', label: '应收状态' },
      { value: 'payable-status', label: '应付状态' },
      { value: 'aging-category', label: '账龄分类' },
      { value: 'payment-type', label: '款项类型' },
    ],
  },
];

// 当前选中大类的 tab 列表
const currentGroupTabs = computed<TabItem[]>(() => {
  const group = categoryGroups.find(g => g.key === selectedGroup.value);
  return group?.tabs || [];
});

const getDictCode = (value: string): string => {
  const map: Record<string, string> = {
    'mold-order-type': 'mold_order_type',
  };
  return map[value] || value;
};

// ============ 状态 ============
const selectedGroup = ref('');
const activeTab = ref('');
const loading = ref(false);
const dictDataList = ref<any[]>([]);
const currentDictTypeID = ref<number | undefined>(undefined);

// 对话框
const dialogVisible = ref(false);
const isEdit = ref(false);
const dialogTitle = ref('');
const formRef = ref();


const formModel = reactive({
  id: undefined as number | undefined,
  dictLabel: '',
  dictValue: '',
  color: 'default',
  sort: 0,
  status: 1,
  needs_delivery_unlock: 0,
  remark: '',
});

const formRules = {
  dictLabel: [{ required: true, message: '请输入显示标签', trigger: 'submit' as const }],
  dictValue: [{ required: true, message: '请输入实际值', trigger: 'submit' as const }],
};

// 受保护字典类型列表（整个类型下所有字典值不可修改/删除）
const PROTECTED_DICT_TYPES = ['mold-order-type', 'production-type', 'mold-source', 'receivable-status', 'payable-status', 'aging-category', 'payment-type'];

// 受保护字典数据值列表（仅特定字典值不可修改）
const PROTECTED_DICT_VALUES: Record<string, string[]> = {
  'order-priority': ['NORMAL'],
};

// 判断当前 tab 是否为整类型受保护
const isProtectedType = computed(() => PROTECTED_DICT_TYPES.includes(activeTab.value));

// 判断当前编辑行的字典值是否受保护
const isEditingProtectedValue = computed(() => {
  const protectedValues = PROTECTED_DICT_VALUES[activeTab.value];
  return protectedValues?.includes(formModel.dictValue) ?? false;
});

// 当前受保护的提示信息
const protectedTypeHint = computed(() => {
  const hints: Record<string, string> = {
    'mold-order-type': '该字典值由系统代码固件控制，不可修改；仅可编辑显示标签、颜色和备注等',
    'production-type': '该字典值影响外发加工商、外发款项等业务流程，不可修改；仅可编辑显示标签、颜色和备注等',
    'mold-source': '该字典值影响新建订单时选择已有模具或手动输入编号的交互逻辑，不可修改；仅可编辑显示标签、颜色和备注等',
    'receivable-status': '该字典值在前后端代码中大量硬编码引用，不可修改；仅可编辑显示标签、颜色和备注等',
    'payable-status': '该字典值在前后端代码中大量硬编码引用，不可修改；仅可编辑显示标签、颜色和备注等',
    'aging-category': '该字典值在后端账龄计算逻辑中硬编码，不可修改；仅可编辑显示标签、颜色和备注等',
    'payment-type': '该字典值在前后端代码中大量硬编码引用，不可修改；仅可编辑显示标签、颜色和备注等',
  };
  return hints[activeTab.value] || '';
});

// 判断某行字典值是否受保护（用于操作列判断是否可删除）
const isRowValueProtected = (dictValue: string) => {
  const protectedValues = PROTECTED_DICT_VALUES[activeTab.value];
  return isProtectedType.value || (protectedValues?.includes(dictValue) ?? false);
};

// 通用表格列配置
const dataColumns: TableProps['columns'] = [
  { colKey: 'sort', title: '排序', width: 80 },
  { colKey: 'dictLabel', title: '显示标签', width: 150 },
  { colKey: 'dictValue', title: '字典值', width: 140 },
  { colKey: 'color', title: '颜色', width: 120 },
  { colKey: 'status', title: '启用状态', width: 100 },
  { colKey: 'remark', title: '备注', width: 160 },
  { colKey: 'operation', title: '操作', width: 150, fixed: 'right' as const },
];

// 款项类型专属列配置（多一个 交付后解锁 列）
const paymentTypeColumns: TableProps['columns'] = [
  { colKey: 'sort', title: '排序', width: 80 },
  { colKey: 'dictLabel', title: '显示标签', width: 150 },
  { colKey: 'dictValue', title: '字典值', width: 140 },
  { colKey: 'color', title: '颜色', width: 120 },
  { colKey: 'needs_delivery_unlock', title: '交付后解锁', width: 120 },
  { colKey: 'status', title: '启用状态', width: 100 },
  { colKey: 'remark', title: '备注', width: 160 },
  { colKey: 'operation', title: '操作', width: 150, fixed: 'right' as const },
];

// 根据当前 tab 动态切换列配置
const tableColumns = computed(() => {
  return activeTab.value === 'payment-type' ? paymentTypeColumns : dataColumns;
});

// ============ 方法 ============
const handleGroupChange = () => {
  const tabs = currentGroupTabs.value;
  activeTab.value = tabs.length > 0 ? tabs[0].value : '';
  if (activeTab.value === 'process-template') loadPtData();
  else loadDictData();
};

const handleTabChange = (value: TabValue) => {
  activeTab.value = value as string;
  if (value === 'process-template') loadPtData();
  else loadDictData();
};

const loadDictData = async () => {
  if (!activeTab.value) return;

  loading.value = true;
  try {
    const dictCode = getDictCode(activeTab.value);
    const typeRes = await getDictTypes({});
    const dictType = typeRes.list?.find((item: any) => item.dict_code === dictCode);

    if (dictType) {
      currentDictTypeID.value = dictType.id;
      const res = await getDictDataByType({ dict_type: dictCode });
      dictDataList.value = (res.list || []).map((item: any) => ({
        id: item.id,
        dict_type_id: item.dict_type_id,
        dictLabel: item.dict_label,
        dictValue: item.dict_value,
        color: item.color,
        sort: item.sort,
        status: item.status,
        needs_delivery_unlock: item.needs_delivery_unlock || 0,
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

const handleAdd = () => {
  isEdit.value = false;
  const currentTab = currentGroupTabs.value.find(t => t.value === activeTab.value);
  dialogTitle.value = `新增 - ${currentTab?.label || ''}`;
  Object.assign(formModel, {
    id: undefined,
    dictLabel: '',
    dictValue: '',
    color: 'default',
    sort: 0,
    status: 1,
    needs_delivery_unlock: 0,
    remark: '',
  });
  setTimeout(() => {
    formRef.value?.clearValidate?.();
  }, 100);
  dialogVisible.value = true;
};

const handleEdit = (row: any) => {
  isEdit.value = true;
  const currentTab = currentGroupTabs.value.find(t => t.value === activeTab.value);
  dialogTitle.value = `编辑 - ${currentTab?.label || ''}`;
  Object.assign(formModel, {
    id: row.id,
    dictLabel: row.dictLabel,
    dictValue: row.dictValue,
    color: row.color || 'default',
    sort: row.sort || 0,
    status: row.status !== undefined ? row.status : 1,
    needs_delivery_unlock: row.needs_delivery_unlock !== undefined ? row.needs_delivery_unlock : 0,
    remark: row.remark || '',
  });
  setTimeout(() => {
    formRef.value?.clearValidate?.();
  }, 100);
  dialogVisible.value = true;
};

const handleSubmit = async () => {
  try {
    const submitDictCode = getDictCode(activeTab.value);
    const submitData: any = {
      dict_type: submitDictCode,
      dict_label: formModel.dictLabel,
      dict_value: formModel.dictValue,
      color: formModel.color,
      sort: formModel.sort,
      status: formModel.status,
      needs_delivery_unlock: formModel.needs_delivery_unlock,
      remark: formModel.remark,
    };

    if (isEdit.value && formModel.id) {
      submitData.id = Number(formModel.id);
    }

    if (isEdit.value) {
      await updateBusinessDictData(submitData);
      MessagePlugin.success('更新成功');
    } else {
      await createDictData(submitData);
      MessagePlugin.success('创建成功');
    }

    dialogVisible.value = false;
    dictUtils.clearCache(submitDictCode);
    loadDictData();
  } catch (error) {
    MessagePlugin.error(isEdit.value ? '更新失败' : '创建失败');
  }
};

const handleCancel = () => {
  dialogVisible.value = false;
};

const handleDelete = async (row: any) => {
  try {
    await deleteDictData(row.id);
    MessagePlugin.success('删除成功');
    dictUtils.clearCache(getDictCode(activeTab.value));
    loadDictData();
  } catch (error) {
    MessagePlugin.error('删除失败');
  }
};

const handleRefresh = async () => {
  if (activeTab.value === 'process-template') loadPtData();
  else loadDictData();
};

// ==================== 工序模板管理 ====================
const ptLoading = ref(false);
const ptTableData = ref<any[]>([]);
const ptDialogVisible = ref(false);
const ptIsEdit = ref(false);
const ptDialogTitle = ref('新增工序');
const ptFormRef = ref();

const ptColumns = [
  { colKey: 'sort', title: '排序号', width: 100 },
  { colKey: 'group_name', title: '分组', width: 120 },
  { colKey: 'process_name', title: '工序名称', width: 150 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'remark', title: '备注', ellipsis: true as const, width: 200 },
  { colKey: 'created_at', title: '创建时间', width: 160 },
  { colKey: 'operation', title: '操作', width: 150, fixed: 'right' as const },
];

const ptGroupOptions = [
  { label: '准备阶段', value: 1 },
  { label: '加工阶段', value: 2 },
  { label: '后处理阶段', value: 3 },
  { label: '装配阶段', value: 4 },
];

const ptFormModel = reactive({
  id: undefined as number | undefined,
  groupId: 1,
  groupName: '准备阶段',
  processName: '',
  sort: 1,
  status: 1,
  remark: '',
});

const ptFormRules = {
  processName: [{ required: true, message: '请输入工序名称', trigger: 'blur' as const }],
  sort: [{ required: true, message: '请输入排序号', trigger: 'blur' as const }],
};

const handlePtGroupChange = (value: any) => {
  const numValue = typeof value === 'string' ? parseInt(value, 10) : Number(value);
  const group = ptGroupOptions.find(g => g.value === numValue);
  if (group) ptFormModel.groupName = group.label;
};

const loadPtData = async () => {
  ptLoading.value = true;
  try {
    const res = await getProcessTemplates();
    ptTableData.value = res || [];
  } catch { MessagePlugin.error('获取工序模板列表失败'); ptTableData.value = []; }
  finally { ptLoading.value = false; }
};

const handlePtAdd = () => {
  ptIsEdit.value = false;
  ptDialogTitle.value = '新增工序';
  Object.assign(ptFormModel, { id: undefined, groupId: 1, groupName: '准备阶段', processName: '', sort: 1, status: 1, remark: '' });
  ptFormRef.value?.reset();
  ptDialogVisible.value = true;
};

const handlePtEdit = (row: any) => {
  ptIsEdit.value = true;
  ptDialogTitle.value = '编辑工序';
  Object.assign(ptFormModel, {
    id: row.id, groupId: row.group_id || 1, groupName: row.group_name || '准备阶段',
    processName: row.process_name, sort: row.sort, status: row.status, remark: row.remark || '',
  });
  ptDialogVisible.value = true;
};

const handlePtDelete = async (row: any) => {
  try { await deleteProcessTemplate(row.id); MessagePlugin.success('删除成功'); loadPtData(); } catch { MessagePlugin.error('删除失败'); }
};

const handlePtSubmit = async () => {
  const valid = await ptFormRef.value?.validate();
  if (valid !== true) return;
  try {
    const data: any = {
      id: ptFormModel.id, group_id: ptFormModel.groupId, group_name: ptFormModel.groupName,
      process_name: ptFormModel.processName, sort: ptFormModel.sort, status: ptFormModel.status, remark: ptFormModel.remark,
    };
    if (ptIsEdit.value) { await updateProcessTemplate(data); MessagePlugin.success('更新成功'); }
    else { await createProcessTemplate(data); MessagePlugin.success('创建成功'); }
    ptDialogVisible.value = false;
    loadPtData();
  } catch { MessagePlugin.error(ptIsEdit.value ? '更新失败' : '创建失败'); }
};

const handlePtCancel = () => { ptDialogVisible.value = false; };

onMounted(() => {
  selectedGroup.value = categoryGroups[0].key;
  const tabs = categoryGroups[0].tabs;
  activeTab.value = tabs.length > 0 ? tabs[0].value : '';
  loadDictData();
});
</script>

<style scoped>
.business-config-container {
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
  align-items: center;
  gap: 8px;
}

.group-label {
  font-size: 14px;
  color: var(--td-text-color-primary);
  white-space: nowrap;
}

.add-btn-wrapper {
  display: inline-flex;
}

.tab-content {
  padding-top: 16px;
}

.empty-hint {
  text-align: center;
  padding: 64px 0;
  color: var(--td-text-color-placeholder);
  font-size: 14px;
}
</style>
