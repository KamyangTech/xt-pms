<template>
  <div class="dept-container">
    <!-- 操作栏 -->
    <t-card class="toolbar-card">
      <div class="toolbar-content">
        <span class="toolbar-title">部门管理</span>
        <div class="toolbar-actions">
          <t-button variant="outline" @click="fetchDeptList">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
          <t-button variant="outline" @click="handleColumnSetting">
            <template #icon><setting-icon /></template>
            列设置
          </t-button>
          <t-button theme="primary" @click="handleAdd">
            <template #icon><add-icon /></template>
            新增部门
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <t-enhanced-table
        ref="tableRef"
        :data="tableData"
        :columns="displayColumns"
        :loading="loading"
        row-key="id"
        :tree="{ childrenKey: 'children', defaultExpandAll: true }"
        :stripe="true"
      >
        <template #create_time="{ row }">
          {{ formatDateTime(row.create_time) }}
        </template>
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '正常' : '停用' }}
          </t-tag>
        </template>
        <template #operation="{ row }">
          <t-link theme="primary" @click="handleEdit(row)" style="margin-right: 8px">编辑</t-link>
          <t-popconfirm :content="`确认删除「${row.dept_name}」？`" @confirm="handleDelete(row)">
            <t-link theme="danger">删除</t-link>
          </t-popconfirm>
        </template>
      </t-enhanced-table>
    </t-card>

    <!-- 新增/编辑对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :title="dialogTitle"
      width="600px"
      placement="center"
      :footer="null"
     :close-on-overlay-click="false">
      <t-form ref="formRef" :data="formModel" :rules="formRules" label-width="100px">
        <t-form-item label="父部门" name="parent_id">
          <t-select v-model="formModel.parent_id" placeholder="请选择父部门（不选则为顶级部门）" clearable>
            <t-option :value="0" label="无（顶级部门）" />
            <t-option 
              v-for="dept in flatDeptList" 
              :key="dept.id" 
              :value="dept.id" 
              :label="dept.label" 
            />
          </t-select>
        </t-form-item>
        <t-form-item label="部门名称" name="dept_name">
          <t-input v-model="formModel.dept_name" placeholder="请输入部门名称" />
        </t-form-item>
        <t-form-item label="排序" name="sort">
          <t-input-number v-model="formModel.sort" :min="0" placeholder="请输入排序" />
        </t-form-item>
        <t-form-item label="负责人" name="leader">
          <t-input v-model="formModel.leader" placeholder="请输入负责人" />
        </t-form-item>
        <t-form-item label="联系电话" name="phone">
          <t-input v-model="formModel.phone" placeholder="请输入联系电话" />
        </t-form-item>
        <t-form-item label="状态" name="status">
          <t-radio-group v-model="formModel.status">
            <t-radio :value="1">正常</t-radio>
            <t-radio :value="0">停用</t-radio>
          </t-radio-group>
        </t-form-item>
        <div style="text-align: right; margin-top: 16px;">
          <t-button theme="default" @click="dialogVisible = false" style="margin-right: 8px">取消</t-button>
          <t-button theme="primary" @click="handleSubmit">确定</t-button>
        </div>
      </t-form>
    </t-dialog>

    <!-- 列设置对话框 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="system_dept_table_columns"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { AddIcon, RefreshIcon, SettingIcon } from 'tdesign-icons-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref } from 'vue';

import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';

import { createDept, deleteDept, getDeptList, updateDept, type CreateDeptParams, type DeptItem, type UpdateDeptParams } from '@/api/dept-management';
import { formatDateTime } from '@/utils/date';

defineOptions({
  name: 'SystemDept',
});

const loading = ref(false);
const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref();

// 表格数据
const tableData = ref<DeptItem[]>([]);

// 表格列定义（基础列，不含操作列）
const baseColumns = [
  { colKey: 'dept_name', title: '部门名称', width: 200, fixed: 'left' as const },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'leader', title: '负责人', width: 120 },
  { colKey: 'phone', title: '联系电话', width: 150 },
  { colKey: 'sort', title: '排序', width: 100 },
  { colKey: 'create_time', title: '创建时间', width: 180 },
];

// 完整列定义（含操作列）
const columnsWithOperation = [
  ...baseColumns,
  { colKey: 'operation', title: '操作', width: 250, fixed: 'right' as const },
];

// 列设置相关
const allColumns: ColumnOption[] = [
  { colKey: 'dept_name', title: '部门名称', pinned: true },
  { colKey: 'status', title: '状态' },
  { colKey: 'leader', title: '负责人' },
  { colKey: 'phone', title: '联系电话' },
  { colKey: 'sort', title: '排序' },
  { colKey: 'create_time', title: '创建时间' },
];

const DEFAULT_COLUMNS = ['dept_name', 'status', 'leader', 'phone', 'sort', 'create_time'];

const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([]);

// 动态列配置
const displayColumns = computed(() => {
  // 如果还没加载列配置，先用默认全部列
  if (selectedColumns.value.length === 0) return columnsWithOperation;

  const orderedColumns = selectedColumns.value
    .map(colKey => columnsWithOperation.find(col => col.colKey === colKey))
    .filter(col => col !== undefined) as typeof columnsWithOperation;

  // 操作列始终在最后
  const operationCol = columnsWithOperation.find(col => col.colKey === 'operation');
  if (operationCol) {
    orderedColumns.push(operationCol);
  }
  return orderedColumns;
});

// 表单模型
const formModel = reactive<CreateDeptParams & { id?: number }>({
  dept_name: '',
  parent_id: 0,
  sort: 0,
  leader: '',
  phone: '',
  status: 1,
});

// 表单验证规则
const formRules = {
  dept_name: [{ required: true, message: '请输入部门名称', trigger: 'blur' as const }],
};

const dialogTitle = ref('新增部门');

// 最大部门层级
const MAX_DEPT_LEVEL = 3;

// 将树形数据扁平化为带前缀标签的列表（用于父部门下拉选择，已达最大层级的部门不可再选）
const flatDeptList = computed(() => {
  const result: { id: number; label: string }[] = [];
  const flatten = (nodes: DeptItem[], depth: number) => {
    for (const node of nodes) {
      if (isEdit.value && node.id === formModel.id) continue; // 排除自身
      const prefix = depth > 0 ? '  '.repeat(depth) + '└ ' : '';
      result.push({ id: node.id, label: prefix + node.dept_name });
      // 只有未达最大层级的部门才递归展开其子部门（已达3级则不可再添加子级）
      if (node.children && node.children.length > 0 && depth < MAX_DEPT_LEVEL - 1) {
        flatten(node.children, depth + 1);
      }
    }
  };
  flatten(tableData.value, 0);
  return result;
});

const fetchDeptList = async () => {
  loading.value = true;
  try {
    const res = await getDeptList();
    tableData.value = res.list || [];
  } catch (error) {
    console.error('获取部门列表失败:', error);
    MessagePlugin.error('获取部门列表失败');
  } finally {
    loading.value = false;
  }
};

// 新增
const handleAdd = () => {
  isEdit.value = false;
  dialogTitle.value = '新增部门';
  resetForm();
  dialogVisible.value = true;
};

// 编辑
const handleEdit = (row: DeptItem) => {
  isEdit.value = true;
  dialogTitle.value = '编辑部门';
  Object.assign(formModel, {
    id: row.id,
    dept_name: row.dept_name,
    parent_id: row.parent_id,
    sort: row.sort,
    leader: row.leader,
    phone: row.phone,
    status: row.status,
  });
  dialogVisible.value = true;
};

const handleDelete = async (row: DeptItem) => {
  try {
    await deleteDept(row.id);
    MessagePlugin.success('删除成功');
    fetchDeptList();
  } catch (error: any) {
    // 显示后端返回的具体错误信息（如：存在子部门或关联用户）
    const message = error?.response?.data?.message || error?.message || '删除失败';
    MessagePlugin.error(message);
  }
};

// 提交表单
const handleSubmit = async () => {
  const valid = await formRef.value?.validate();
  if (valid !== true) return;

  try {
    if (isEdit.value && formModel.id) {
      // 编辑
      const updateData: UpdateDeptParams = {
        id: formModel.id,
        dept_name: formModel.dept_name,
        parent_id: formModel.parent_id,
        sort: formModel.sort,
        leader: formModel.leader,
        phone: formModel.phone,
        status: formModel.status,
      };
      await updateDept(updateData);
      MessagePlugin.success('更新成功');
    } else {
      // 新增
      const createData: CreateDeptParams = {
        dept_name: formModel.dept_name,
        parent_id: formModel.parent_id,
        sort: formModel.sort,
        leader: formModel.leader,
        phone: formModel.phone,
        status: formModel.status,
      };
      await createDept(createData);
      MessagePlugin.success('创建成功');
    }
    dialogVisible.value = false;
    fetchDeptList();
  } catch (error: any) {
    console.error('操作失败:', error);
    const message = error?.response?.data?.message || error?.message || (isEdit.value ? '更新失败' : '创建失败');
    MessagePlugin.error(message);
  }
};

const resetForm = () => {
  Object.assign(formModel, {
    id: undefined,
    dept_name: '',
    parent_id: 0,
    sort: 0,
    leader: '',
    phone: '',
    status: 1,
  });
  formRef.value?.reset();
};

onMounted(() => {
  // 从 localStorage 加载列配置
  try {
    const saved = localStorage.getItem('system_dept_table_columns');
    if (saved) {
      const parsed = JSON.parse(saved);
      const cols = parsed.columns || parsed.visible || parsed || [];
      selectedColumns.value = Array.isArray(cols) ? cols : [];
    }
  } catch {
    // 无保存配置，使用默认
  }
  fetchDeptList();
});

// 列设置
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

const handleColumnsChange = (newColumns: string[]) => {
  selectedColumns.value = newColumns;
};
</script>

<style lang="less" scoped>
// 树形表格：叶节点也保留图标占位宽度，确保与有子节点的行文字对齐
:deep(.t-table__tree-leaf-node .t-table__tree-op-icon) {
  min-width: 16px;
  margin-right: var(--td-comp-margin-s);
}

.dept-container {
  padding: 16px;

  .toolbar-card {
    margin-bottom: 16px;

    .toolbar-content {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .toolbar-title {
        font-size: 16px;
        font-weight: 500;
      }

      .toolbar-actions {
        display: flex;
        gap: 8px;
      }
    }
  }
}
</style>
