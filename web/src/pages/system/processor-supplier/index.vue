<template>
  <div class="processor-supplier-container">
    <!-- 搜索筛选 - 放在Tabs外面，根据当前Tab动态切换筛选字段 -->
    <SearchFilter
      :filters="currentFilterConfig"
      :model-value="currentFormData"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card">
      <div class="toolbar-content">
        <div class="toolbar-left">
          <t-tabs v-model="activeTab" @change="handleTabChange">
            <t-tab-panel value="processor" label="加工商管理" />
            <t-tab-panel value="supplier" label="供应商管理" />
          </t-tabs>
          <span v-if="currentSelectedKeys.length > 0" class="batch-count">
            已选择 {{ currentSelectedKeys.length }} 项
          </span>
        </div>
        <div class="toolbar-actions">
          <t-popconfirm v-if="currentSelectedKeys.length > 0" :content="`确认批量删除所选${ activeTab === 'processor' ? '加工商' : '供应商' }？`" @confirm="handleBatchDelete">
            <t-button theme="danger" variant="outline">批量删除</t-button>
          </t-popconfirm>
          <t-button theme="default" variant="outline" @click="handleRefresh">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
          <t-button theme="default" variant="outline" @click="handleColumnSetting">
            <template #icon><setting-icon /></template>
            列设置
          </t-button>
          <t-button theme="primary" @click="handleAdd">
            <template #icon><add-icon /></template>
            {{ activeTab === 'processor' ? '新增加工商' : '新增供应商' }}
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">

      <!-- 加工商表格 -->
      <t-table
        v-if="activeTab === 'processor'"
        :data="processorTableData"
        :columns="displayProcessorColumns"
        row-key="id"
        :loading="processorLoading"
        :selected-row-keys="processorSelectedKeys"
        @select-change="(keys: (string|number)[]) => processorSelectedKeys = keys"
        :pagination="processorPagination"
        :stripe="true"
        hover
        @page-change="rehandleProcessorPageChange"
      >
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '正常' : '停用' }}
          </t-tag>
        </template>
        <template #create_time="{ row }">
          {{ formatDateTime(row.create_time) }}
        </template>
        <template #update_time="{ row }">
          {{ row.update_time ? formatDateTime(row.update_time) : '-' }}
        </template>
        <template #operation="{ row }">
          <t-space>
            <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
            <t-popconfirm content="确认删除该加工商？" @confirm="handleDelete(row)">
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>

      <!-- 供应商表格 -->
      <t-table
        v-if="activeTab === 'supplier'"
        :data="supplierTableData"
        :columns="displaySupplierColumns"
        row-key="id"
        :loading="supplierLoading"
        :selected-row-keys="supplierSelectedKeys"
        @select-change="(keys: (string|number)[]) => supplierSelectedKeys = keys"
        :pagination="supplierPagination"
        :stripe="true"
        hover
        @page-change="rehandleSupplierPageChange"
      >
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '正常' : '停用' }}
          </t-tag>
        </template>
        <template #create_time="{ row }">
          {{ formatDateTime(row.create_time) }}
        </template>
        <template #update_time="{ row }">
          {{ row.update_time ? formatDateTime(row.update_time) : '-' }}
        </template>
        <template #operation="{ row }">
          <t-space>
            <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
            <t-popconfirm content="确认删除该供应商？" @confirm="handleDelete(row)">
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <!-- 加工商表单弹窗 -->
    <t-dialog
      v-model:visible="processorDialogVisible"
      :header="processorDialogTitle"
      width="600px"
      placement="center"
     :close-on-overlay-click="false">
      <t-form ref="processorFormRef" :data="processorForm" :rules="processorRules" label-width="100px">
        <t-row :gutter="[16, 16]" style="margin-left: 0; margin-right: 0;">
          <t-col :span="12">
            <t-form-item label="加工商编号" name="processorCode">
              <t-input v-model="processorForm.processorCode" placeholder="请输入加工商编号" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="加工商名称" name="processorName">
              <t-input v-model="processorForm.processorName" placeholder="请输入加工商名称" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="联系人" name="contact">
              <t-input v-model="processorForm.contact" placeholder="请输入联系人" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="联系电话" name="phone">
              <t-input v-model="processorForm.phone" placeholder="请输入联系电话" />
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="电子邮箱" name="email">
              <t-input v-model="processorForm.email" placeholder="请输入邮箱" />
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="地址" name="address">
              <t-input v-model="processorForm.address" placeholder="请输入地址" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="加工类型" name="processType">
              <t-input v-model="processorForm.processType" placeholder="如：CNC加工、线割、放电等" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="状态" name="status">
              <t-radio-group v-model="processorForm.status">
                <t-radio :value="1">正常</t-radio>
                <t-radio :value="0">停用</t-radio>
              </t-radio-group>
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="备注" name="remark">
              <t-textarea v-model="processorForm.remark" placeholder="请输入备注" :rows="3" />
            </t-form-item>
          </t-col>
        </t-row>
      </t-form>
      <template #footer>
        <t-button theme="default" @click="handleProcessorCancel">取消</t-button>
        <t-button theme="primary" style="margin-left: 8px" @click="handleProcessorConfirm">确定</t-button>
      </template>
    </t-dialog>

    <!-- 供应商表单弹窗 -->
    <t-dialog
      v-model:visible="supplierDialogVisible"
      :header="supplierDialogTitle"
      width="600px"
      placement="center"
     :close-on-overlay-click="false">
      <t-form ref="supplierFormRef" :data="supplierForm" :rules="supplierRules" label-width="100px">
        <t-row :gutter="[16, 16]" style="margin-left: 0; margin-right: 0;">
          <t-col :span="12">
            <t-form-item label="供应商编号" name="supCode">
              <t-input v-model="supplierForm.supCode" placeholder="请输入供应商编号" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="供应商名称" name="supName">
              <t-input v-model="supplierForm.supName" placeholder="请输入供应商名称" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="联系人" name="contact">
              <t-input v-model="supplierForm.contact" placeholder="请输入联系人" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="联系电话" name="phone">
              <t-input v-model="supplierForm.phone" placeholder="请输入联系电话" />
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="电子邮箱" name="email">
              <t-input v-model="supplierForm.email" placeholder="请输入邮箱" />
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="地址" name="address">
              <t-input v-model="supplierForm.address" placeholder="请输入地址" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="状态" name="status">
              <t-radio-group v-model="supplierForm.status">
                <t-radio :value="1">正常</t-radio>
                <t-radio :value="0">停用</t-radio>
              </t-radio-group>
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="备注" name="remark">
              <t-textarea v-model="supplierForm.remark" placeholder="请输入备注" :rows="3" />
            </t-form-item>
          </t-col>
        </t-row>
      </t-form>
      <template #footer>
        <t-button theme="default" @click="handleSupplierCancel">取消</t-button>
        <t-button theme="primary" style="margin-left: 8px" @click="handleSupplierConfirm">确定</t-button>
      </template>
    </t-dialog>

    <!-- 列设置对话框 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="currentAllColumns"
      :default-columns="currentDefaultColumns"
      :storage-key="currentStorageKey"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { TabValue } from 'tdesign-vue-next';
import { AddIcon, RefreshIcon, SettingIcon } from 'tdesign-icons-vue-next';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import { formatDateTime } from '@/utils/date';
import {
  getProcessorList,
  createProcessor,
  updateProcessor,
  deleteProcessor,
  batchDeleteProcessors
} from '@/api/processor';
import {
  getSupplierList,
  createSupplier,
  updateSupplier,
  deleteSupplier,
  batchDeleteSuppliers
} from '@/api/supplier';

defineOptions({
  name: 'ProcessorSupplierIndex',
});

// 手机号正则
const phoneReg = /^1[3-9]\d{9}$/;
// 邮箱正则
const emailReg = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

// 当前激活的tab
const activeTab = ref('processor');

// ==================== 当前激活 Tab 动态代理 ====================

// 当前搜索表单
const currentFormData = computed(() =>
  activeTab.value === 'processor' ? processorFormData : supplierFormData
);

// 当前筛选配置
const currentFilterConfig = computed<FilterItem[]>(() =>
  activeTab.value === 'processor' ? processorFilterConfig : supplierFilterConfig
);

// 当前列设置
const currentAllColumns = computed<ColumnOption[]>(() =>
  activeTab.value === 'processor' ? processorAllColumns : supplierAllColumns
);
const currentDefaultColumns = computed(() =>
  activeTab.value === 'processor' ? PROCESSOR_DEFAULT_COLUMNS : SUPPLIER_DEFAULT_COLUMNS
);
const currentStorageKey = computed(() =>
  activeTab.value === 'processor' ? 'processor_table_columns' : 'supplier_table_columns'
);

// 当前选中行
const currentSelectedKeys = computed(() =>
  activeTab.value === 'processor' ? processorSelectedKeys.value : supplierSelectedKeys.value
);

// ==================== 加工商相关 ====================
const processorLoading = ref(false);
const processorTableData = ref<any[]>([]);
const processorSelectedKeys = ref<(string | number)[]>([]);

const processorFormData = reactive({
  processorCode: '',
  processorName: '',
  contact: '',
  phone: '',
  status: undefined as number | undefined,
});

const processorPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

// 加工商对话框
const processorDialogVisible = ref(false);
const processorDialogTitle = ref('新增加工商');
const processorFormRef = ref();
const processorForm = reactive({
  id: null as number | null,
  processorCode: '',
  processorName: '',
  contact: '',
  phone: '',
  email: '',
  address: '',
  processType: '',
  remark: '',
  status: 1,
});

const processorRules = {
  processorCode: [{ required: true, message: '请输入加工商编号', trigger: 'blur' as const }],
  processorName: [{ required: true, message: '请输入加工商名称', trigger: 'blur' as const }],
  phone: [{ pattern: phoneReg, message: '请输入正确的手机号', trigger: 'blur' as const }],
  email: [{ pattern: emailReg, message: '请输入正确的邮箱地址', trigger: 'blur' as const }],
};

// 加工商表格列配置
const processorColumns = [
  { colKey: 'row-select', type: 'multiple' as const, width: 50, fixed: 'left' as const },
  { colKey: 'processor_code', title: '加工商编号', width: 150, fixed: 'left' as const },
  { colKey: 'processor_name', title: '加工商名称', width: 150 },
  { colKey: 'process_type', title: '加工类型', width: 150 },
  { colKey: 'contact', title: '联系人', width: 120 },
  { colKey: 'phone', title: '联系电话', width: 150 },
  { colKey: 'email', title: '邮箱', width: 180 },
  { colKey: 'address', title: '地址', ellipsis: true as const },
  { colKey: 'remark', title: '备注', ellipsis: true as const },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'create_time', title: '创建时间', width: 180 },
  { colKey: 'update_time', title: '更新时间', width: 180 },
  { colKey: 'operation', title: '操作', width: 150, fixed: 'right' as const },
];

// 动态列
const displayProcessorColumns = computed(() => {
  return processorColumns.filter(col =>
    col.colKey === 'row-select' || col.colKey === 'operation' || processorSelectedColumns.value.includes(col.colKey)
  );
});

// 可选列
const processorAllColumns: ColumnOption[] = [
  { colKey: 'processor_code', title: '加工商编号', pinned: true },
  { colKey: 'processor_name', title: '加工商名称' },
  { colKey: 'process_type', title: '加工类型' },
  { colKey: 'contact', title: '联系人' },
  { colKey: 'phone', title: '联系电话' },
  { colKey: 'email', title: '邮箱' },
  { colKey: 'address', title: '地址' },
  { colKey: 'remark', title: '备注' },
  { colKey: 'status', title: '状态' },
  { colKey: 'create_time', title: '创建时间' },
  { colKey: 'update_time', title: '更新时间' },
];

const PROCESSOR_DEFAULT_COLUMNS = ['processor_code', 'processor_name', 'process_type', 'contact', 'phone', 'email', 'address', 'status', 'create_time'];
const PROCESSOR_PINNED_COLUMNS = ['processor_code'];

const processorSelectedColumns = ref<string[]>([]);

// 筛选栏配置
const processorFilterConfig: FilterItem[] = [
  { type: 'input', field: 'processorCode', placeholder: '请输入加工商编号' },
  { type: 'input', field: 'processorName', placeholder: '请输入加工商名称' },
  { type: 'input', field: 'contact', placeholder: '请输入联系人' },
  { type: 'input', field: 'phone', placeholder: '请输入联系电话' },
  {
    type: 'select',
    field: 'status',
    placeholder: '请选择状态',
    options: [
      { label: '正常', value: 1 },
      { label: '停用', value: 0 },
    ],
    showMore: true,
  },
];

// ==================== 供应商相关 ====================
const supplierLoading = ref(false);
const supplierTableData = ref<any[]>([]);
const supplierSelectedKeys = ref<(string | number)[]>([]);

const supplierFormData = reactive({
  supCode: '',
  supName: '',
  contact: '',
  phone: '',
  status: undefined as number | undefined,
});

const supplierPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

// 供应商对话框
const supplierDialogVisible = ref(false);
const supplierDialogTitle = ref('新增供应商');
const supplierFormRef = ref();
const supplierForm = reactive({
  id: null as number | null,
  supCode: '',
  supName: '',
  contact: '',
  phone: '',
  email: '',
  address: '',
  remark: '',
  status: 1,
});

const supplierRules = {
  supCode: [{ required: true, message: '请输入供应商编号', trigger: 'blur' as const }],
  supName: [{ required: true, message: '请输入供应商名称', trigger: 'blur' as const }],
  phone: [{ pattern: phoneReg, message: '请输入正确的手机号', trigger: 'blur' as const }],
  email: [{ pattern: emailReg, message: '请输入正确的邮箱地址', trigger: 'blur' as const }],
};

// 供应商表格列配置
const supplierColumns = [
  { colKey: 'row-select', type: 'multiple' as const, width: 50, fixed: 'left' as const },
  { colKey: 'sup_code', title: '供应商编号', width: 150, fixed: 'left' as const },
  { colKey: 'sup_name', title: '供应商名称', width: 150 },
  { colKey: 'contact', title: '联系人', width: 120 },
  { colKey: 'phone', title: '联系电话', width: 150 },
  { colKey: 'email', title: '邮箱', width: 180 },
  { colKey: 'address', title: '地址', ellipsis: true as const },
  { colKey: 'remark', title: '备注', ellipsis: true as const },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'create_time', title: '创建时间', width: 180 },
  { colKey: 'update_time', title: '更新时间', width: 180 },
  { colKey: 'operation', title: '操作', width: 150, fixed: 'right' as const },
];

// 动态列
const displaySupplierColumns = computed(() => {
  return supplierColumns.filter(col =>
    col.colKey === 'row-select' || col.colKey === 'operation' || supplierSelectedColumns.value.includes(col.colKey)
  );
});

// 可选列
const supplierAllColumns: ColumnOption[] = [
  { colKey: 'sup_code', title: '供应商编号', pinned: true },
  { colKey: 'sup_name', title: '供应商名称' },
  { colKey: 'contact', title: '联系人' },
  { colKey: 'phone', title: '联系电话' },
  { colKey: 'email', title: '邮箱' },
  { colKey: 'address', title: '地址' },
  { colKey: 'remark', title: '备注' },
  { colKey: 'status', title: '状态' },
  { colKey: 'create_time', title: '创建时间' },
  { colKey: 'update_time', title: '更新时间' },
];

const SUPPLIER_DEFAULT_COLUMNS = ['sup_code', 'sup_name', 'contact', 'phone', 'email', 'address', 'status', 'create_time'];
const SUPPLIER_PINNED_COLUMNS = ['sup_code'];

const supplierSelectedColumns = ref<string[]>([]);

// 筛选栏配置
const supplierFilterConfig: FilterItem[] = [
  { type: 'input', field: 'supCode', placeholder: '请输入供应商编号' },
  { type: 'input', field: 'supName', placeholder: '请输入供应商名称' },
  { type: 'input', field: 'contact', placeholder: '请输入联系人' },
  { type: 'input', field: 'phone', placeholder: '请输入联系电话' },
  {
    type: 'select',
    field: 'status',
    placeholder: '请选择状态',
    options: [
      { label: '正常', value: 1 },
      { label: '停用', value: 0 },
    ],
    showMore: true,
  },
];

// ==================== 列设置相关 ====================
const columnSettingRef = ref<any>(null);

const handleColumnsChange = (cols: string[]) => {
  if (activeTab.value === 'processor') {
    processorSelectedColumns.value = cols;
  } else {
    supplierSelectedColumns.value = cols;
  }
};

// ==================== 通用方法 ====================

const handleTabChange = (value: TabValue) => {
  if (value === 'processor') {
    loadProcessorData();
  } else if (value === 'supplier') {
    loadSupplierData();
  }
};

// 搜索（根据当前 Tab 调用对应方法）
const handleSearch = () => {
  if (activeTab.value === 'processor') {
    handleProcessorSearch();
  } else {
    handleSupplierSearch();
  }
};

const handleReset = () => {
  if (activeTab.value === 'processor') {
    handleProcessorReset();
  } else {
    handleSupplierReset();
  }
};

// 刷新
const handleRefresh = () => {
  if (activeTab.value === 'processor') {
    handleProcessorRefresh();
  } else {
    handleSupplierRefresh();
  }
};

// 新增
const handleAdd = () => {
  if (activeTab.value === 'processor') {
    handleAddProcessor();
  } else {
    handleAddSupplier();
  }
};

// 编辑
const handleEdit = (row: any) => {
  if (activeTab.value === 'processor') {
    handleEditProcessor(row);
  } else {
    handleEditSupplier(row);
  }
};

const handleDelete = async (row: any) => {
  if (activeTab.value === 'processor') {
    await handleDeleteProcessor(row);
  } else {
    await handleDeleteSupplier(row);
  }
};

// 批量删除
const handleBatchDelete = async () => {
  if (activeTab.value === 'processor') {
    await handleProcessorBatchDelete();
  } else {
    await handleSupplierBatchDelete();
  }
};

// 列设置
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// ==================== 加工商数据方法 ====================
const loadProcessorData = async () => {
  processorLoading.value = true;
  try {
    const res = await getProcessorList({
      page: processorPagination.current,
      page_size: processorPagination.pageSize,
      ...processorFormData,
    });
    processorTableData.value = res.list || [];
    processorPagination.total = res.total || 0;
    processorSelectedKeys.value = [];
  } catch (error) {
    MessagePlugin.error('加载加工商数据失败');
  } finally {
    processorLoading.value = false;
  }
};

const handleProcessorSearch = () => {
  processorPagination.current = 1;
  loadProcessorData();
};

const handleProcessorReset = () => {
  Object.assign(processorFormData, {
    processorCode: '',
    processorName: '',
    contact: '',
    phone: '',
    status: undefined,
  });
  processorPagination.current = 1;
  loadProcessorData();
};

const handleProcessorRefresh = () => {
  localStorage.removeItem('processor_table_columns');
  processorSelectedColumns.value = [...PROCESSOR_DEFAULT_COLUMNS];
  loadProcessorData();
  MessagePlugin.success('刷新成功');
};

const rehandleProcessorPageChange = (pageInfo: any) => {
  processorPagination.current = pageInfo.current;
  processorPagination.pageSize = pageInfo.pageSize;
  loadProcessorData();
};

const handleAddProcessor = () => {
  processorDialogTitle.value = '新增加工商';
  Object.assign(processorForm, {
    id: null,
    processorCode: '',
    processorName: '',
    contact: '',
    phone: '',
    email: '',
    address: '',
    processType: '',
    remark: '',
    status: 1,
  });
  processorFormRef.value?.reset();
  processorDialogVisible.value = true;
};

const handleEditProcessor = (row: any) => {
  processorDialogTitle.value = '编辑加工商';
  Object.assign(processorForm, {
    id: row.id,
    processorCode: row.processor_code,
    processorName: row.processor_name,
    contact: row.contact,
    phone: row.phone,
    email: row.email,
    address: row.address,
    processType: row.process_type,
    remark: row.remark,
    status: row.status,
  });
  setTimeout(() => {
    processorFormRef.value?.clearValidate?.();
  }, 100);
  processorDialogVisible.value = true;
};

const handleDeleteProcessor = async (row: any) => {
  try {
    await deleteProcessor(row.id);
    MessagePlugin.success('删除成功');
    loadProcessorData();
  } catch (error) {
    MessagePlugin.error('删除失败');
  }
};

const handleProcessorBatchDelete = async () => {
  const ids = processorSelectedKeys.value.map(Number);
  try {
    await batchDeleteProcessors(ids);
    MessagePlugin.success(`已删除 ${ids.length} 个加工商`);
    loadProcessorData();
  } catch {
    MessagePlugin.error('批量删除失败');
  }
};

const handleProcessorConfirm = async () => {
  const valid = await processorFormRef.value?.validate();
  if (valid !== true) return;

  try {
    const submitData: any = {
      processor_code: processorForm.processorCode,
      processor_name: processorForm.processorName,
      contact: processorForm.contact,
      phone: processorForm.phone,
      email: processorForm.email,
      address: processorForm.address,
      process_type: processorForm.processType,
      remark: processorForm.remark,
      status: processorForm.status,
    };

    if (processorForm.id) {
      submitData.id = processorForm.id;
      await updateProcessor(submitData);
      MessagePlugin.success('更新成功');
    } else {
      await createProcessor(submitData);
      MessagePlugin.success('创建成功');
    }
    processorDialogVisible.value = false;
    loadProcessorData();
  } catch (error) {
    MessagePlugin.error(processorForm.id ? '更新失败' : '创建失败');
  }
};

const handleProcessorCancel = () => {
  processorDialogVisible.value = false;
  processorFormRef.value?.reset();
};

// ==================== 供应商数据方法 ====================
const loadSupplierData = async () => {
  supplierLoading.value = true;
  try {
    const res = await getSupplierList({
      page: supplierPagination.current,
      page_size: supplierPagination.pageSize,
      ...supplierFormData,
    });
    supplierTableData.value = res.list || [];
    supplierPagination.total = res.total || 0;
    supplierSelectedKeys.value = [];
  } catch (error) {
    MessagePlugin.error('加载供应商数据失败');
  } finally {
    supplierLoading.value = false;
  }
};

const handleSupplierSearch = () => {
  supplierPagination.current = 1;
  loadSupplierData();
};

const handleSupplierReset = () => {
  Object.assign(supplierFormData, {
    supCode: '',
    supName: '',
    contact: '',
    phone: '',
    status: undefined,
  });
  supplierPagination.current = 1;
  loadSupplierData();
};

const handleSupplierRefresh = () => {
  localStorage.removeItem('supplier_table_columns');
  supplierSelectedColumns.value = [...SUPPLIER_DEFAULT_COLUMNS];
  loadSupplierData();
  MessagePlugin.success('刷新成功');
};

const rehandleSupplierPageChange = (pageInfo: any) => {
  supplierPagination.current = pageInfo.current;
  supplierPagination.pageSize = pageInfo.pageSize;
  loadSupplierData();
};

const handleAddSupplier = () => {
  supplierDialogTitle.value = '新增供应商';
  Object.assign(supplierForm, {
    id: null,
    supCode: '',
    supName: '',
    contact: '',
    phone: '',
    email: '',
    address: '',
    remark: '',
    status: 1,
  });
  supplierFormRef.value?.reset();
  supplierDialogVisible.value = true;
};

const handleEditSupplier = (row: any) => {
  supplierDialogTitle.value = '编辑供应商';
  Object.assign(supplierForm, {
    id: row.id,
    supCode: row.sup_code,
    supName: row.sup_name,
    contact: row.contact,
    phone: row.phone,
    email: row.email,
    address: row.address,
    remark: row.remark,
    status: row.status,
  });
  setTimeout(() => {
    supplierFormRef.value?.clearValidate?.();
  }, 100);
  supplierDialogVisible.value = true;
};

const handleDeleteSupplier = async (row: any) => {
  try {
    await deleteSupplier(row.id);
    MessagePlugin.success('删除成功');
    loadSupplierData();
  } catch (error) {
    MessagePlugin.error('删除失败');
  }
};

const handleSupplierBatchDelete = async () => {
  const ids = supplierSelectedKeys.value.map(Number);
  try {
    await batchDeleteSuppliers(ids);
    MessagePlugin.success(`已删除 ${ids.length} 个供应商`);
    loadSupplierData();
  } catch {
    MessagePlugin.error('批量删除失败');
  }
};

const handleSupplierConfirm = async () => {
  const valid = await supplierFormRef.value?.validate();
  if (valid !== true) return;

  try {
    const submitData: any = {
      sup_code: supplierForm.supCode,
      sup_name: supplierForm.supName,
      contact: supplierForm.contact,
      phone: supplierForm.phone,
      email: supplierForm.email,
      address: supplierForm.address,
      remark: supplierForm.remark,
      status: supplierForm.status,
    };

    if (supplierForm.id) {
      submitData.id = supplierForm.id;
      await updateSupplier(submitData);
      MessagePlugin.success('更新成功');
    } else {
      await createSupplier(submitData);
      MessagePlugin.success('创建成功');
    }
    supplierDialogVisible.value = false;
    loadSupplierData();
  } catch (error) {
    MessagePlugin.error(supplierForm.id ? '更新失败' : '创建失败');
  }
};

const handleSupplierCancel = () => {
  supplierDialogVisible.value = false;
  supplierFormRef.value?.reset();
};

const loadColumnConfig = () => {
  // 加工商
  try {
    const saved = localStorage.getItem('processor_table_columns');
    if (saved) {
      const data = JSON.parse(saved);
      processorSelectedColumns.value = Array.isArray(data) ? data : (data.visible || [...PROCESSOR_DEFAULT_COLUMNS]);
    } else {
      processorSelectedColumns.value = [...PROCESSOR_DEFAULT_COLUMNS];
    }
  } catch {
    processorSelectedColumns.value = [...PROCESSOR_DEFAULT_COLUMNS];
  }
  // 确保固定列始终可见
  PROCESSOR_PINNED_COLUMNS.forEach(key => {
    if (!processorSelectedColumns.value.includes(key)) {
      processorSelectedColumns.value.unshift(key);
    }
  });

  // 供应商
  try {
    const saved = localStorage.getItem('supplier_table_columns');
    if (saved) {
      const data = JSON.parse(saved);
      supplierSelectedColumns.value = Array.isArray(data) ? data : (data.visible || [...SUPPLIER_DEFAULT_COLUMNS]);
    } else {
      supplierSelectedColumns.value = [...SUPPLIER_DEFAULT_COLUMNS];
    }
  } catch {
    supplierSelectedColumns.value = [...SUPPLIER_DEFAULT_COLUMNS];
  }
  // 确保固定列始终可见
  SUPPLIER_PINNED_COLUMNS.forEach(key => {
    if (!supplierSelectedColumns.value.includes(key)) {
      supplierSelectedColumns.value.unshift(key);
    }
  });
};

onMounted(() => {
  loadProcessorData();
  loadColumnConfig();
});
</script>

<style scoped lang="less">
.processor-supplier-container {
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingLR-xl);

  .toolbar-card {
    margin-bottom: var(--td-comp-margin-xl);

    .toolbar-content {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .toolbar-left {
        display: flex;
        align-items: center;
        gap: 12px;
      }

      .toolbar-actions {
        display: flex;
        align-items: center;
        gap: var(--td-comp-margin-s);
        flex-shrink: 0;
      }
    }
  }
}

.batch-count {
  font-size: 14px;
  color: var(--td-brand-color, #0052d9);
  font-weight: 500;
  white-space: nowrap;
}
</style>
