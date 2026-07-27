<template>
  <div class="customer-container">
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
          <span class="toolbar-title">客户管理</span>
          <span v-if="selectedRowKeys.length > 0" class="batch-count">
            已选择 {{ selectedRowKeys.length }} 项
          </span>
        </div>
        <div class="toolbar-actions">
          <t-popconfirm v-if="selectedRowKeys.length > 0" content="确认批量删除所选客户？" @confirm="handleBatchDelete">
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
            新增客户
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
        <template #create_time="{ row }">
          {{ formatDateTime(row.create_time) }}
        </template>
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '正常' : '停用' }}
          </t-tag>
        </template>
        <template #update_time="{ row }">
          {{ row.update_time ? formatDateTime(row.update_time) : '-' }}
        </template>
        <template #operation="{ row }">
          <t-link theme="primary" @click="handleEdit(row)" style="margin-right: 8px">编辑</t-link>
          <t-popconfirm content="确认删除该客户？" @confirm="handleDelete(row)">
            <t-link theme="danger">删除</t-link>
          </t-popconfirm>
        </template>
      </t-table>
    </t-card>

    <!-- 新增/编辑对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :title="dialogTitle"
      width="600px"
      placement="center"
     :close-on-overlay-click="false">
      <t-form ref="formRef" :data="formModel" :rules="formRules" label-width="100px">
        <t-row :gutter="[16, 16]" style="margin-left: 0; margin-right: 0;">
          <t-col :span="12">
            <t-form-item label="客户编号" name="cusCode">
              <t-input v-model="formModel.cusCode" placeholder="请输入客户编号" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="客户名称" name="cusName">
              <t-input v-model="formModel.cusName" placeholder="请输入客户名称" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="联系人" name="contact">
              <t-input v-model="formModel.contact" placeholder="请输入联系人" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="联系电话" name="phone">
              <t-input v-model="formModel.phone" placeholder="请输入联系电话" />
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="电子邮箱" name="email">
              <t-input v-model="formModel.email" placeholder="请输入电子邮箱" />
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="客户地址" name="address">
              <t-input v-model="formModel.address" placeholder="请输入客户地址" />
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="备注" name="remark">
              <t-textarea v-model="formModel.remark" placeholder="请输入备注" :rows="3" />
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="状态" name="status">
              <t-radio-group v-model="formModel.status">
                <t-radio :value="1">正常</t-radio>
                <t-radio :value="0">停用</t-radio>
              </t-radio-group>
            </t-form-item>
          </t-col>
        </t-row>
      </t-form>
      <template #footer>
        <t-button theme="default" @click="handleCancel">取消</t-button>
        <t-button theme="primary" style="margin-left: 8px" @click="handleSubmit">确定</t-button>
      </template>
    </t-dialog>

    <!-- 列设置对话框 - 使用通用组件 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="customer_table_columns"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { AddIcon, RefreshIcon, SettingIcon } from 'tdesign-icons-vue-next';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';
import { getCustomerList, createCustomer, updateCustomer, deleteCustomer, batchDeleteCustomers } from '@/api/customer';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import { formatDateTime } from '@/utils/date';

defineOptions({
  name: 'CustomerIndex',
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
  cusCode: '',
  cusName: '',
  contact: '',
  phone: '',
  status: undefined as number | undefined,
});

// 对话框相关
const dialogVisible = ref(false);
const isEdit = ref(false);
const dialogTitle = ref('新增客户');
const formRef = ref();

const formModel = reactive({
  id: undefined,
  cusCode: '',
  cusName: '',
  contact: '',
  phone: '',
  email: '',
  address: '',
  remark: '',
  status: 1,
});

// 手机号正则
const phoneReg = /^1[3-9]\d{9}$/;
// 邮箱正则
const emailReg = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

const formRules = {
  cusCode: [{ required: true, message: '请输入客户编号', trigger: 'blur' as const }],
  cusName: [{ required: true, message: '请输入客户名称', trigger: 'blur' as const }],
  phone: [
    {
      pattern: phoneReg,
      message: '请输入正确的手机号',
      trigger: 'blur' as const,
    },
  ],
  email: [
    {
      pattern: emailReg,
      message: '请输入正确的邮箱地址',
      trigger: 'blur' as const,
    },
  ],
};

// 表格列配置
const columns: PrimaryTableCol[] = [
  { colKey: 'row-select', type: 'multiple', width: 50, fixed: 'left' },
  { colKey: 'id', title: 'ID', width: 80 },
  { colKey: 'cusCode', title: '客户编号', width: 150, fixed: 'left' },
  { colKey: 'cusName', title: '客户名称', width: 200 },
  { colKey: 'contact', title: '联系人', width: 120 },
  { colKey: 'phone', title: '联系电话', width: 150 },
  { colKey: 'email', title: '电子邮箱', width: 200 },
  { colKey: 'address', title: '客户地址', width: 200 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'create_time', title: '创建时间', width: 180 },
  { colKey: 'update_time', title: '更新时间', width: 180 },
  { colKey: 'operation', title: '操作', width: 150, fixed: 'right' },
];

// 动态列配置（根据用户选择）
const displayColumns = computed(() => {
  return columns.filter(col => 
    col.colKey === 'row-select' || col.colKey === 'operation' || selectedColumns.value.includes(col.colKey)
  );
});

// 所有可选列配置（不包括操作列）
const allColumns: ColumnOption[] = [
  { colKey: 'cusCode', title: '客户编号', pinned: true },  // ✨ 固定列：不可隐藏/移动
  { colKey: 'cusName', title: '客户名称' },
  { colKey: 'contact', title: '联系人' },
  { colKey: 'phone', title: '联系电话' },
  { colKey: 'email', title: '电子邮箱' },
  { colKey: 'address', title: '客户地址' },
  { colKey: 'status', title: '状态' },
  { colKey: 'create_time', title: '创建时间' },
  { colKey: 'update_time', title: '更新时间' },
];

// 默认显示的列
const DEFAULT_COLUMNS = ['cusCode', 'cusName', 'contact', 'phone', 'email', 'address', 'status', 'create_time'];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([]);

const getCustomerListFunc = async () => {
  loading.value = true;
  try {
    const res = await getCustomerList({
      page: pagination.current,
      page_size: pagination.pageSize,
      ...formData,
    });
    
    const list = res.list || [];
    tableData.value = list.map((item: any) => ({
      id: item.id,
      cusCode: item.cus_code,
      cusName: item.cus_name,
      contact: item.contact,
      phone: item.phone,
      email: item.email,
      address: item.address,
      remark: item.remark,
      status: item.status,
      create_time: item.create_time,
      update_time: item.update_time,
    }));
    pagination.total = res.total || 0;

    // 清除已选行（数据刷新后选中状态失效）
    selectedRowKeys.value = [];
  } catch (error) {
    MessagePlugin.error('获取客户列表失败');
  } finally {
    loading.value = false;
  }
};

// 搜索
const handleSearch = () => {
  pagination.current = 1;
  getCustomerListFunc();
};

const handleReset = () => {
  formData.cusCode = '';
  formData.cusName = '';
  formData.contact = '';
  formData.phone = '';
  formData.status = undefined;
  pagination.current = 1;
  getCustomerListFunc();
};

// 刷新
const handleRefresh = () => {
  getCustomerListFunc();
  MessagePlugin.success('刷新成功');
};

// 分页
const rehandlePageChange = (curr: any) => {
  pagination.current = curr.current;
  pagination.pageSize = curr.pageSize;
  getCustomerListFunc();
};

// 新增
const handleAdd = () => {
  isEdit.value = false;
  dialogTitle.value = '新增客户';
  resetForm();
  dialogVisible.value = true;
};

// 编辑
const handleEdit = (row: any) => {
  isEdit.value = true;
  dialogTitle.value = '编辑客户';
  Object.assign(formModel, {
    id: Number(row.id),  // 确保 ID 是数字类型
    cusCode: row.cusCode,
    cusName: row.cusName,
    contact: row.contact,
    phone: row.phone,
    email: row.email,
    address: row.address,
    remark: row.remark,
    status: row.status,
  });
  setTimeout(() => {
    formRef.value?.clearValidate?.();
  }, 100);
  dialogVisible.value = true;
};

const handleDelete = async (row: any) => {
  try {
    await deleteCustomer(row.id);
    MessagePlugin.success('删除成功');
    getCustomerListFunc();
  } catch (error) {
    MessagePlugin.error('删除失败');
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
    await batchDeleteCustomers(ids);
    MessagePlugin.success(`已删除 ${ids.length} 个客户`);
    getCustomerListFunc();
  } catch {
    MessagePlugin.error('批量删除失败');
  }
};

// 提交表单
const handleSubmit = async () => {
  const valid = await formRef.value?.validate();
  if (valid !== true) return;

  try {
    const submitData: any = {
      cus_code: formModel.cusCode,
      cus_name: formModel.cusName,
      contact: formModel.contact,
      phone: formModel.phone,
      email: formModel.email,
      address: formModel.address,
      remark: formModel.remark,
      status: formModel.status,
    };
    
    // 编辑时添加 ID
    if (isEdit.value && formModel.id) {
      submitData.id = Number(formModel.id);
    }
    
    if (isEdit.value) {
      await updateCustomer(submitData);
      MessagePlugin.success('更新成功');
    } else {
      await createCustomer(submitData);
      MessagePlugin.success('创建成功');
    }
    
    dialogVisible.value = false;
    getCustomerListFunc();
  } catch (error) {
    MessagePlugin.error(isEdit.value ? '更新失败' : '创建失败');
  }
};

const resetForm = () => {
  Object.assign(formModel, {
    id: undefined,
    cusCode: '',
    cusName: '',
    contact: '',
    phone: '',
    email: '',
    address: '',
    remark: '',
    status: 1,
  });
  formRef.value?.reset();
};

// 取消
const handleCancel = () => {
  dialogVisible.value = false;
  formRef.value?.reset();
};

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('customer_table_columns');
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

// 列配置变化时的回调
const handleColumnsChange = (columns: string[]) => {
  selectedColumns.value = columns;
};

// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'cusCode', placeholder: '请输入客户编号' },
  { type: 'input', field: 'cusName', placeholder: '请输入客户名称' },
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

onMounted(() => {
  getCustomerListFunc();
  loadColumnConfig();
});
</script>

<style scoped>
.customer-container {
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

/* 列设置样式 */
.column-setting-content {
  padding: 16px 0;
}

.column-item {
  padding: 8px 0;
  border-bottom: 1px solid var(--td-border-level-1-color, #e7e7e7);
  
  &:last-child {
    border-bottom: none;
  }
}

.column-setting-footer {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
