<template>
  <div class="process-template-container">
    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">工序模板管理</span>
        <div class="toolbar-actions">
          <t-button theme="primary" @click="handleAdd">
            <template #icon><add-icon /></template>
            新增工序
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <t-table
        :data="tableData"
        :columns="columns"
        :loading="loading"
        row-key="id"
        :pagination="pagination"
        @page-change="rehandlePageChange"
      >
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'default'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </t-tag>
        </template>
        <template #created_at="{ row }">
          {{ formatDateTime(row.created_at) }}
        </template>
        <template #operation="{ row }">
          <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
          <t-popconfirm content="确认删除该工序？" @confirm="handleDelete(row)">
            <t-link theme="danger" style="margin-left: 8px">删除</t-link>
          </t-popconfirm>
        </template>
      </t-table>
    </t-card>

    <!-- 新增/编辑对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :header="dialogTitle"
      width="600px"
      placement="center"
      @confirm="handleSubmit"
      @cancel="handleCancel"
     :close-on-overlay-click="false">
      <t-form
        ref="formRef"
        :model="formModel"
        :rules="formRules"
        :label-width="100"
        style="margin-top: 16px"
      >
        <t-form-item label="分组" name="groupName">
          <t-select v-model="formModel.groupId" placeholder="请选择分组" @change="handleGroupChange">
            <t-option
              v-for="item in groupOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="工序名称" name="processName">
          <t-input v-model="formModel.processName" placeholder="请输入工序名称" />
        </t-form-item>
        <t-form-item label="排序号" name="sort">
          <t-input-number v-model="formModel.sort" :min="1" placeholder="请输入排序号" style="width: 100%" />
        </t-form-item>
        <t-form-item label="状态" name="status">
          <t-radio-group v-model="formModel.status">
            <t-radio :value="1">启用</t-radio>
            <t-radio :value="2">禁用</t-radio>
          </t-radio-group>
        </t-form-item>
        <t-form-item label="备注" name="remark">
          <t-textarea v-model="formModel.remark" placeholder="请输入备注信息" :maxlength="500" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { TableProps } from 'tdesign-vue-next';
import { AddIcon } from 'tdesign-icons-vue-next';
import { 
  getProcessTemplates, 
  createProcessTemplate, 
  updateProcessTemplate, 
  deleteProcessTemplate 
} from '@/api/process-template';
import { formatDateTime } from '@/utils/date';

defineOptions({
  name: 'ProcessTemplate',
});

// 表格相关
const loading = ref(false);
const tableData = ref<any[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

// 表格列配置
const columns: TableProps['columns'] = [
  { colKey: 'sort', title: '排序号', width: 100 },
  { colKey: 'group_name', title: '分组', width: 120 },
  { colKey: 'process_name', title: '工序名称', width: 150 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'remark', title: '备注', ellipsis: true, width: 200 },
  { colKey: 'created_at', title: '创建时间', width: 160 },
  { colKey: 'operation', title: '操作', width: 150, fixed: 'right' },
];

// 对话框相关
const dialogVisible = ref(false);
const isEdit = ref(false);
const dialogTitle = ref('新增工序');
const formRef = ref();

const formModel = reactive({
  id: undefined,
  groupId: 1,
  groupName: '准备阶段',
  processName: '',
  sort: 1,
  status: 1,
  remark: '',
});

const formRules = {
  groupName: [{ required: true, message: '请选择分组', trigger: 'change' as const }],
  processName: [{ required: true, message: '请输入工序名称', trigger: 'blur' as const }],
  sort: [{ required: true, message: '请输入排序号', trigger: 'blur' as const }],
};

// 分组选项
const groupOptions = [
  { label: '准备阶段', value: 1 },
  { label: '加工阶段', value: 2 },
  { label: '后处理阶段', value: 3 },
  { label: '装配阶段', value: 4 },
];

// 监听分组变化，自动设置分组名称
const handleGroupChange = (value: any) => {
  const numValue = typeof value === 'string' ? parseInt(value, 10) : Number(value);
  const group = groupOptions.find(g => g.value === numValue);
  if (group) {
    formModel.groupName = group.label;
  }
};

const getList = async () => {
  loading.value = true;
  try {
    const res = await getProcessTemplates();
    const allData = res || [];
    pagination.total = allData.length;
    
    // 前端分页
    const start = (pagination.current - 1) * pagination.pageSize;
    const end = start + pagination.pageSize;
    tableData.value = allData.slice(start, end);
  } catch (error) {
    console.error('获取工序模板列表失败:', error);
    MessagePlugin.error('获取工序模板列表失败');
    tableData.value = [];
    pagination.total = 0;
  } finally {
    loading.value = false;
  }
};

// 新增
const handleAdd = () => {
  isEdit.value = false;
  dialogTitle.value = '新增工序';
  Object.assign(formModel, {
    id: undefined,
    groupId: 1,
    groupName: '准备阶段',
    processName: '',
    sort: 1,
    status: 1,
    remark: '',
  });
  dialogVisible.value = true;
};

// 编辑
const handleEdit = (row: any) => {
  isEdit.value = true;
  dialogTitle.value = '编辑工序';
  Object.assign(formModel, {
    id: row.id,
    groupId: row.group_id || 1,
    groupName: row.group_name || '准备阶段',
    processName: row.process_name,
    sort: row.sort,
    status: row.status,
    remark: row.remark || '',
  });
  dialogVisible.value = true;
};

// 提交表单
const handleSubmit = async () => {
  const validateResult = await formRef.value?.validate();
  if (!validateResult) {
    return;
  }

  try {
    const data = {
      id: formModel.id,
      group_id: formModel.groupId,
      group_name: formModel.groupName,
      process_name: formModel.processName,
      sort: formModel.sort,
      status: formModel.status,
      remark: formModel.remark,
    };

    if (isEdit.value) {
      await updateProcessTemplate(data);
      MessagePlugin.success('更新成功');
    } else {
      await createProcessTemplate(data);
      MessagePlugin.success('创建成功');
    }

    dialogVisible.value = false;
    getList();
  } catch (error) {
    MessagePlugin.error(isEdit.value ? '更新失败' : '创建失败');
  }
};

// 取消
const handleCancel = () => {
  dialogVisible.value = false;
};

const handleDelete = async (row: any) => {
  try {
    await deleteProcessTemplate(row.id);
    MessagePlugin.success('删除成功');
    getList();
  } catch (error) {
    MessagePlugin.error('删除失败');
  }
};

// 分页变化
const rehandlePageChange = (curr: any) => {
  pagination.current = curr.current;
  pagination.pageSize = curr.pageSize;
  getList();
};

onMounted(() => {
  getList();
});
</script>

<style scoped>
.process-template-container {
  padding: 16px;
}

.search-card,
.toolbar-card {
  margin-bottom: 16px;
}

.toolbar-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
</style>
