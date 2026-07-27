<template>
  <div class="receiving-account-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="formData"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 工具栏 -->
    <t-card class="toolbar-card">
      <div class="toolbar-content">
        <div class="toolbar-left">
          <span class="toolbar-title">收款账户管理</span>
          <span v-if="selectedRowKeys.length > 0" class="batch-count">
            已选择 {{ selectedRowKeys.length }} 项
          </span>
        </div>
        <div class="toolbar-actions">
          <t-popconfirm v-if="selectedRowKeys.length > 0" content="确认批量删除所选收款账户？" @confirm="handleBatchDelete">
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
            新增收款账户
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <t-table
        :data="tableData"
        :columns="displayColumns"
        :loading="loading"
        row-key="id"
        :selected-row-keys="selectedRowKeys"
        @select-change="handleSelectChange"
        :pagination="pagination"
        :stripe="true"
        hover
        @page-change="rehandlePageChange"
      >
        <template #account_type="{ row }">
          <t-tag :theme="getTypeTheme(row.account_type)" variant="light">
            {{ getTypeLabel(row.account_type) }}
          </t-tag>
        </template>
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'default'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </t-tag>
        </template>
        <template #create_time="{ row }">
          {{ formatDateTime(row.create_time) }}
        </template>
        <template #update_time="{ row }">
          {{ row.update_time ? formatDateTime(row.update_time) : '-' }}
        </template>
        <template #operation="{ row }">
          <t-link theme="primary" @click="handleEdit(row)" style="margin-right: 8px">编辑</t-link>
          <t-popconfirm content="确认删除该收款账户？" @confirm="handleDelete(row)">
            <t-link theme="danger">删除</t-link>
          </t-popconfirm>
        </template>
      </t-table>
    </t-card>

    <!-- 新增/编辑对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :header="isEdit ? '编辑收款账户' : '新增收款账户'"
      width="600px"
      placement="center"
     :close-on-overlay-click="false">
      <t-form ref="formRef" :data="formModel" :rules="rules" label-width="100px">
        <t-form-item label="账户名称" name="account_name">
          <t-input v-model="formModel.account_name" placeholder="如：公司公账、法人银行卡" />
        </t-form-item>
        <t-form-item label="账号" name="account_no">
          <t-input v-model="formModel.account_no" placeholder="请输入银行账号" />
        </t-form-item>
        <t-form-item label="开户行">
          <t-input v-model="formModel.bank_name" placeholder="请输入开户行名称" />
        </t-form-item>
        <t-form-item label="账户类型">
          <t-select v-model="formModel.account_type" placeholder="请选择账户类型" clearable>
            <t-option value="company" label="对公账户" />
            <t-option value="personal" label="个人银行卡" />
            <t-option value="wechat" label="微信" />
            <t-option value="alipay" label="支付宝" />
          </t-select>
        </t-form-item>
        <t-row :gutter="[16, 0]">
          <t-col :span="12">
            <t-form-item label="排序">
              <t-input-number v-model="formModel.sort" :min="0" placeholder="排序号" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="状态">
              <t-radio-group v-model="formModel.status">
                <t-radio :value="1">启用</t-radio>
                <t-radio :value="0">禁用</t-radio>
              </t-radio-group>
            </t-form-item>
          </t-col>
        </t-row>
        <t-form-item label="备注">
          <t-textarea v-model="formModel.remark" placeholder="可选备注信息" :maxlength="500" :rows="3" />
        </t-form-item>
      </t-form>
      <template #footer>
        <t-button theme="default" @click="handleCancel">取消</t-button>
        <t-button theme="primary" style="margin-left: 8px" @click="handleSubmit">确定</t-button>
      </template>
    </t-dialog>

    <!-- 列设置对话框 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="receiving_account_table_columns"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { AddIcon, RefreshIcon, SettingIcon } from 'tdesign-icons-vue-next';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';
import {
  getReceivingAccountList,
  createReceivingAccount,
  updateReceivingAccount,
  deleteReceivingAccount,
  batchDeleteReceivingAccounts,
} from '@/api/receiving-account';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import { formatDateTime } from '@/utils/date';

defineOptions({
  name: 'ReceivingAccount',
});

// 表格相关
const loading = ref(false);
const tableData = ref<any[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

// 选中行
const selectedRowKeys = ref<(string | number)[]>([]);

// 搜索表单
const formData = reactive({
  account_name: '',
  account_no: '',
  bank_name: '',
  account_type: '',
  status: undefined as number | undefined,
  sort: undefined as number | undefined,
  remark: '',
});

// 对话框相关
const dialogVisible = ref(false);
const isEdit = ref(false);
const confirmLoading = ref(false);
const formRef = ref();
const formModel = reactive({
  id: undefined as number | undefined,
  account_name: '',
  account_no: '',
  bank_name: '',
  account_type: '',
  sort: 0,
  status: 1,
  remark: '',
});

const rules = {
  account_name: [{ required: true, message: '请输入账户名称', type: 'error' as const }],
  account_no: [{ required: true, message: '请输入账号', type: 'error' as const }],
};

// 表格列配置
const columns: PrimaryTableCol[] = [
  { colKey: 'row-select', type: 'multiple', width: 50, fixed: 'left' },
  { colKey: 'id', title: 'ID', width: 80 },
  { colKey: 'sort', title: '排序', width: 60 },
  { colKey: 'account_name', title: '账户名称', width: 160, fixed: 'left' },
  { colKey: 'account_no', title: '账号', width: 220 },
  { colKey: 'bank_name', title: '开户行', width: 200 },
  { colKey: 'account_type', title: '账户类型', width: 110 },
  { colKey: 'status', title: '状态', width: 80 },
  { colKey: 'remark', title: '备注', ellipsis: true, minWidth: 120 },
  { colKey: 'create_time', title: '创建时间', width: 180 },
  { colKey: 'update_time', title: '更新时间', width: 180 },
  { colKey: 'operation', title: '操作', width: 150, fixed: 'right' },
];

// 动态列配置
const displayColumns = computed(() => {
  return columns.filter(col =>
    col.colKey === 'row-select' || col.colKey === 'operation' || selectedColumns.value.includes(col.colKey)
  );
});

// 所有可选列配置
const allColumns: ColumnOption[] = [
  { colKey: 'account_name', title: '账户名称', pinned: true },
  { colKey: 'account_no', title: '账号' },
  { colKey: 'bank_name', title: '开户行' },
  { colKey: 'account_type', title: '账户类型' },
  { colKey: 'status', title: '状态' },
  { colKey: 'sort', title: '排序' },
  { colKey: 'remark', title: '备注' },
  { colKey: 'create_time', title: '创建时间' },
  { colKey: 'update_time', title: '更新时间' },
];

// 默认显示的列
const DEFAULT_COLUMNS = ['account_name', 'account_no', 'bank_name', 'account_type', 'status', 'remark', 'create_time'];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([]);

// 账户类型映射
const typeMap: Record<string, { label: string; theme: string }> = {
  company: { label: '对公账户', theme: 'primary' },
  personal: { label: '个人银行卡', theme: 'warning' },
  wechat: { label: '微信', theme: 'success' },
  alipay: { label: '支付宝', theme: 'default' },
};

const getTypeLabel = (type: string) => typeMap[type]?.label || type || '-';
const getTypeTheme = (type: string) => (typeMap[type]?.theme as any) || 'default';

const loadData = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.current,
      page_size: pagination.pageSize,
      ...formData,
    };
    // 移除空值
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === undefined) {
        delete params[key];
      }
    });
    const res = await getReceivingAccountList(params);
    const list = res.list || [];
    tableData.value = list.map((item: any) => ({
      id: item.id,
      account_name: item.account_name,
      account_no: item.account_no,
      bank_name: item.bank_name,
      account_type: item.account_type,
      sort: item.sort,
      status: item.status,
      remark: item.remark,
      create_time: item.create_time,
      update_time: item.update_time,
    }));
    pagination.total = res.total || 0;
    selectedRowKeys.value = [];
  } catch (error) {
    console.error('加载收款账户列表失败:', error);
    MessagePlugin.error('加载数据失败');
  } finally {
    loading.value = false;
  }
};

// 搜索
const handleSearch = () => {
  pagination.current = 1;
  loadData();
};

const handleReset = () => {
  formData.account_name = '';
  formData.account_no = '';
  formData.bank_name = '';
  formData.account_type = '';
  formData.status = undefined;
  pagination.current = 1;
  loadData();
};

// 刷新
const handleRefresh = () => {
  loadData();
  MessagePlugin.success('刷新成功');
};

// 分页
const rehandlePageChange = (curr: any) => {
  pagination.current = curr.current;
  pagination.pageSize = curr.pageSize;
  loadData();
};

// 新增
const handleAdd = () => {
  isEdit.value = false;
  Object.assign(formModel, {
    id: undefined,
    account_name: '',
    account_no: '',
    bank_name: '',
    account_type: '',
    sort: 0,
    status: 1,
    remark: '',
  });
  formRef.value?.clearValidate?.();
  dialogVisible.value = true;
};

// 编辑
const handleEdit = (row: any) => {
  isEdit.value = true;
  Object.assign(formModel, {
    id: row.id,
    account_name: row.account_name,
    account_no: row.account_no,
    bank_name: row.bank_name,
    account_type: row.account_type,
    sort: row.sort,
    status: row.status,
    remark: row.remark,
  });
  setTimeout(() => {
    formRef.value?.clearValidate?.();
  }, 100);
  dialogVisible.value = true;
};

const handleDelete = async (row: any) => {
  try {
    await deleteReceivingAccount(row.id);
    MessagePlugin.success('删除成功');
    loadData();
  } catch (error: any) {
    MessagePlugin.error(error.message || '删除失败');
  }
};

// ==================== 批量操作 ====================

// 选中行变化
const handleSelectChange = (keys: (string | number)[]) => {
  selectedRowKeys.value = keys;
};

// 批量删除
const handleBatchDelete = async () => {
  const ids = selectedRowKeys.value.map(Number);
  try {
    await batchDeleteReceivingAccounts(ids);
    MessagePlugin.success(`已删除 ${ids.length} 个收款账户`);
    loadData();
  } catch {
    MessagePlugin.error('批量删除失败');
  }
};

// 提交
const handleSubmit = async () => {
  const validateResult = await formRef.value?.validate?.();
  if (validateResult === false) return;

  confirmLoading.value = true;
  try {
    const submitData: any = {
      account_name: formModel.account_name,
      account_no: formModel.account_no,
      bank_name: formModel.bank_name,
      account_type: formModel.account_type,
      sort: formModel.sort,
      status: formModel.status,
      remark: formModel.remark,
    };
    if (isEdit.value && formModel.id) {
      submitData.id = formModel.id;
      await updateReceivingAccount(submitData);
      MessagePlugin.success('更新成功');
    } else {
      await createReceivingAccount(submitData);
      MessagePlugin.success('创建成功');
    }
    dialogVisible.value = false;
    // 创建成功后重置搜索条件，避免旧的搜索条件过滤掉新数据
    handleReset();
  } catch (error: any) {
    MessagePlugin.error(error.message || '操作失败');
  } finally {
    confirmLoading.value = false;
  }
};

// 取消
const handleCancel = () => {
  dialogVisible.value = false;
};

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('receiving_account_table_columns');
    if (saved) {
      const savedData = JSON.parse(saved);
      if (Array.isArray(savedData)) {
        selectedColumns.value = savedData;
      } else if (savedData.visible) {
        selectedColumns.value = savedData.visible;
      } else {
        selectedColumns.value = [...DEFAULT_COLUMNS];
      }
    } else {
      selectedColumns.value = [...DEFAULT_COLUMNS];
    }
  } catch (error) {
    console.error('加载列配置失败:', error);
    selectedColumns.value = [...DEFAULT_COLUMNS];
  }
};

// 打开列设置对话框
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// 列配置变化
const handleColumnsChange = (cols: string[]) => {
  selectedColumns.value = cols;
};

// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'account_name', placeholder: '请输入账户名称' },
  { type: 'input', field: 'account_no', placeholder: '请输入账号' },
  { type: 'input', field: 'bank_name', placeholder: '请输入开户行' },
  {
    type: 'select',
    field: 'account_type',
    placeholder: '请选择账户类型',
    options: [
      { label: '对公账户', value: 'company' },
      { label: '个人银行卡', value: 'personal' },
      { label: '微信', value: 'wechat' },
      { label: '支付宝', value: 'alipay' },
    ],
    showMore: true,
  },
  {
    type: 'select',
    field: 'status',
    placeholder: '请选择状态',
    options: [
      { label: '启用', value: 1 },
      { label: '禁用', value: 0 },
    ],
    showMore: true,
  },
];

onMounted(() => {
  loadData();
  loadColumnConfig();
});
</script>

<style scoped>
.receiving-account-container {
  padding: 16px;
}

.toolbar-card {
  margin-bottom: 16px;
}

.toolbar-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.toolbar-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.batch-count {
  font-size: 14px;
  color: var(--td-brand-color, #0052d9);
  font-weight: 500;
}
</style>