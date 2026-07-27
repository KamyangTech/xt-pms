<template>
  <div class="warehouse-container">
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
          <span class="toolbar-title">仓库管理</span>
          <span v-if="selectedRowKeys.length > 0" class="batch-count">
            已选择 {{ selectedRowKeys.length }} 项
          </span>
        </div>
        <div class="toolbar-actions">
          <t-popconfirm v-if="selectedRowKeys.length > 0" content="确认批量删除所选仓库？" @confirm="handleBatchDelete">
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
            新增仓库
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
          <t-link theme="primary" @click="handleEdit(row)" style="margin-right: 8px">编辑</t-link>
          <t-popconfirm content="确认删除该仓库？" @confirm="handleDelete(row)">
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
            <t-form-item label="仓库编码" name="whCode">
              <t-input v-model="formModel.whCode" placeholder="请输入仓库编码" :disabled="isEdit" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="仓库名称" name="whName">
              <t-input v-model="formModel.whName" placeholder="请输入仓库名称" />
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="仓库地址" name="address">
              <t-input v-model="formModel.address" placeholder="请输入仓库地址" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="管理员" name="manager">
              <t-input v-model="formModel.manager" placeholder="请输入管理员" />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="联系电话" name="phone">
              <t-input v-model="formModel.phone" placeholder="请输入联系电话" />
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

    <!-- 列设置对话框 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="warehouse_table_columns"
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
import { getWarehouseList, createWarehouse, updateWarehouse, deleteWarehouse, batchDeleteWarehouses } from '@/api/warehouse';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import { formatDateTime } from '@/utils/date';

defineOptions({
  name: 'SystemWarehouse',
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
  whCode: '',
  whName: '',
  manager: '',
  status: undefined as number | undefined,
});

// 对话框相关
const dialogVisible = ref(false);
const isEdit = ref(false);
const dialogTitle = ref('新增仓库');
const formRef = ref();

const formModel = reactive({
  id: undefined,
  whCode: '',
  whName: '',
  address: '',
  manager: '',
  phone: '',
  remark: '',
  status: 1,
});

// 手机号正则
const phoneReg = /^1[3-9]\d{9}$/;

const formRules = {
  whCode: [{ required: true, message: '请输入仓库编码', trigger: 'blur' as const }],
  whName: [{ required: true, message: '请输入仓库名称', trigger: 'blur' as const }],
  phone: [
    {
      pattern: phoneReg,
      message: '请输入正确的手机号',
      trigger: 'blur' as const,
    },
  ],
};

// 表格列配置
const columns: PrimaryTableCol[] = [
  { colKey: 'row-select', type: 'multiple', width: 50, fixed: 'left' },
  { colKey: 'id', title: 'ID', width: 80 },
  { colKey: 'whCode', title: '仓库编码', width: 150, fixed: 'left' },
  { colKey: 'whName', title: '仓库名称', width: 200 },
  { colKey: 'address', title: '仓库地址', width: 200 },
  { colKey: 'manager', title: '管理员', width: 120 },
  { colKey: 'phone', title: '联系电话', width: 150 },
  { colKey: 'remark', title: '备注', ellipsis: true },
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
  { colKey: 'whCode', title: '仓库编码', pinned: true },
  { colKey: 'whName', title: '仓库名称' },
  { colKey: 'address', title: '仓库地址' },
  { colKey: 'manager', title: '管理员' },
  { colKey: 'phone', title: '联系电话' },
  { colKey: 'remark', title: '备注' },
  { colKey: 'status', title: '状态' },
  { colKey: 'create_time', title: '创建时间' },
  { colKey: 'update_time', title: '更新时间' },
];

// 默认显示的列
const DEFAULT_COLUMNS = ['whCode', 'whName', 'address', 'manager', 'phone', 'status', 'create_time'];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([]);

const fetchWarehouseList = async () => {
  loading.value = true;
  try {
    const res = await getWarehouseList({
      page: pagination.current,
      page_size: pagination.pageSize,
      ...formData,
    });
    
    const list = res.list || [];
    tableData.value = list.map((item: any) => ({
      id: item.id,
      whCode: item.wh_code,
      whName: item.wh_name,
      address: item.address,
      manager: item.manager,
      phone: item.phone,
      remark: item.remark,
      status: item.status,
      create_time: item.create_time,
      update_time: item.update_time,
    }));
    pagination.total = res.total || 0;

    // 清除已选行（数据刷新后选中状态失效）
    selectedRowKeys.value = [];
  } catch (error) {
    MessagePlugin.error('获取仓库列表失败');
  } finally {
    loading.value = false;
  }
};

// 搜索
const handleSearch = () => {
  pagination.current = 1;
  fetchWarehouseList();
};

const handleReset = () => {
  formData.whCode = '';
  formData.whName = '';
  formData.manager = '';
  formData.status = undefined;
  pagination.current = 1;
  fetchWarehouseList();
};

// 刷新
const handleRefresh = () => {
  fetchWarehouseList();
  MessagePlugin.success('刷新成功');
};

// 分页
const rehandlePageChange = (curr: any) => {
  pagination.current = curr.current;
  pagination.pageSize = curr.pageSize;
  fetchWarehouseList();
};

// 新增
const handleAdd = () => {
  isEdit.value = false;
  dialogTitle.value = '新增仓库';
  resetForm();
  dialogVisible.value = true;
};

// 编辑
const handleEdit = (row: any) => {
  isEdit.value = true;
  dialogTitle.value = '编辑仓库';
  Object.assign(formModel, {
    id: Number(row.id),
    whCode: row.whCode,
    whName: row.whName,
    address: row.address,
    manager: row.manager,
    phone: row.phone,
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
    await deleteWarehouse(row.id);
    MessagePlugin.success('删除成功');
    fetchWarehouseList();
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
    await batchDeleteWarehouses(ids);
    MessagePlugin.success(`已删除 ${ids.length} 个仓库`);
    fetchWarehouseList();
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
      wh_code: formModel.whCode,
      wh_name: formModel.whName,
      address: formModel.address,
      manager: formModel.manager,
      phone: formModel.phone,
      remark: formModel.remark,
      status: formModel.status,
    };
    
    // 编辑时添加 ID
    if (isEdit.value && formModel.id) {
      submitData.id = Number(formModel.id);
    }
    
    if (isEdit.value) {
      await updateWarehouse(submitData);
      MessagePlugin.success('更新成功');
    } else {
      await createWarehouse(submitData);
      MessagePlugin.success('创建成功');
    }
    
    dialogVisible.value = false;
    fetchWarehouseList();
  } catch (error: any) {
    MessagePlugin.error(error.message || (isEdit.value ? '更新失败' : '创建失败'));
  }
};

const resetForm = () => {
  Object.assign(formModel, {
    id: undefined,
    whCode: '',
    whName: '',
    address: '',
    manager: '',
    phone: '',
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
    const saved = localStorage.getItem('warehouse_table_columns');
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
  { type: 'input', field: 'whCode', placeholder: '请输入仓库编码' },
  { type: 'input', field: 'whName', placeholder: '请输入仓库名称' },
  { type: 'input', field: 'manager', placeholder: '请输入管理员' },
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
  fetchWarehouseList();
  loadColumnConfig();
});
</script>

<style scoped>
.warehouse-container {
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
