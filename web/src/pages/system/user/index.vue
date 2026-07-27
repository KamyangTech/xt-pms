<template>
  <div class="user-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="formData"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 工具栏 -->
    <t-card class="toolbar-card">
      <div class="toolbar-content">
        <span class="toolbar-title">用户管理</span>
        <div class="toolbar-actions">
          <t-button theme="default" variant="outline" @click="handleRefresh">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
          <t-button theme="default" variant="outline" @click="handleColumnSetting">
            <template #icon><setting-icon /></template>
            列设置
          </t-button>
          <t-button v-permission="'system:user:create'" theme="primary" @click="handleAdd">
            <template #icon><add-icon /></template>
            新增用户
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <!-- 批量操作栏 -->
      <div v-if="selectedRowKeys.length > 0" class="batch-bar">
        <span class="batch-count">已选择 {{ selectedRowKeys.length }} 项</span>
        <t-space>
          <t-button v-permission="'system:user:update'" size="small" theme="primary" variant="outline" @click="handleBatchEnable">批量启用</t-button>
          <t-button v-permission="'system:user:update'" size="small" theme="default" variant="outline" @click="handleBatchDisable">批量禁用</t-button>
          <t-popconfirm content="确认批量删除所选用户？" @confirm="handleBatchDelete">
            <t-button v-permission="'system:user:delete'" size="small" theme="danger" variant="outline">批量删除</t-button>
          </t-popconfirm>
        </t-space>
      </div>

      <t-table
        :data="tableData"
        :columns="displayColumns"
        :loading="loading"
        row-key="id"
        hover
        :selected-row-keys="selectedRowKeys"
        @select-change="handleSelectChange"
        :pagination="pagination"
        @page-change="onPageChange"
      >
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '正常' : '禁用' }}
          </t-tag>
        </template>
        <template #roles="{ row }">
          <div v-if="row.roles && row.roles.length > 0" class="role-tags">
            <t-tag
              v-for="role in row.roles"
              :key="role.id"
              theme="primary"
              variant="light"
            >
              {{ role.role_name }}
            </t-tag>
          </div>
          <span v-else class="no-role">-</span>
        </template>
        <template #create_time="{ row }">
          {{ formatDateTime(row.create_time) }}
        </template>
        <template #update_time="{ row }">
          {{ row.update_time ? formatDateTime(row.update_time) : '-' }}
        </template>
        <template #operation="{ row }">
          <t-link v-permission="'system:user:update'" theme="primary" @click="handleEdit(row)" style="margin-right: 8px">编辑</t-link>
          <t-link v-permission="'system:user:update'" theme="primary" @click="handleResetPassword(row)" style="margin-right: 8px">重置密码</t-link>
          <t-popconfirm content="确定解锁该用户账号？" @confirm="handleUnlockUser(row)">
            <t-link v-permission="'system:user:update'" theme="primary" style="margin-right: 8px">解锁</t-link>
          </t-popconfirm>
          <t-popconfirm content="确认删除该用户？" @confirm="handleDelete(row)">
            <t-link v-permission="'system:user:delete'" theme="danger">删除</t-link>
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
      :confirm-btn="isEdit ? { content: '确定' } : undefined"
      @confirm="handleSubmit"
      @cancel="dialogVisible = false"
     :close-on-overlay-click="false">
      <t-form ref="formRef" :data="formModel" :rules="formRules" label-width="100px">
        <t-form-item label="用户名" name="username">
          <t-input v-model="formModel.username" placeholder="请输入用户名" />
        </t-form-item>
        <t-form-item v-if="!isEdit" label="密码" name="password">
          <t-input v-model="formModel.password" type="password" placeholder="请输入密码" />
        </t-form-item>
        <t-form-item label="姓名" name="real_name">
          <t-input v-model="formModel.real_name" placeholder="请输入姓名" />
        </t-form-item>
        <t-form-item label="手机号" name="phone">
          <t-input v-model="formModel.phone" placeholder="请输入手机号" />
        </t-form-item>
        <t-form-item label="邮箱" name="email">
          <t-input v-model="formModel.email" placeholder="请输入邮箱" />
        </t-form-item>
        <t-form-item label="职位" name="position">
          <t-input v-model="formModel.position" placeholder="请输入职位" />
        </t-form-item>
        <t-form-item label="入职日期" name="hire_date">
          <t-date-picker 
            v-model="formModel.hire_date" 
            placeholder="请选择入职日期" 
            format="YYYY-MM-DD" 
            clearable 
          />
        </t-form-item>
        <t-form-item label="部门" name="dept_id">
          <t-select v-model="formModel.dept_id" placeholder="请选择部门" clearable>
            <t-option v-for="dept in deptList" :key="dept.id" :value="dept.id" :label="dept.dept_name" />
          </t-select>
        </t-form-item>
        <t-form-item label="角色" name="role_ids">
          <t-select 
            v-model="formModel.role_ids" 
            placeholder="请选择角色" 
            multiple
            clearable
          >
            <t-option 
              v-for="role in roleList" 
              :key="role.id" 
              :value="role.id" 
              :label="role.role_name" 
            />
          </t-select>
        </t-form-item>
        <t-form-item label="状态" name="status">
          <t-select v-model="formModel.status" placeholder="请选择状态">
            <t-option :value="1" label="正常" />
            <t-option :value="0" label="禁用" />
          </t-select>
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 列设置对话框 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="system_user_table_columns"
      @change="handleColumnsChange"
    />

    <!-- 重置密码对话框 -->
    <t-dialog
      v-model:visible="resetPwdVisible"
      title="重置密码"
      width="400px"
      placement="center"
      :confirm-btn="{ content: '确定', loading: resetPwdLoading }"
      @confirm="handleResetPwdSubmit"
      @cancel="handleResetPwdCancel"
     :close-on-overlay-click="false">
      <t-form ref="resetPwdFormRef" :data="resetPwdModel" :rules="resetPwdRules" label-width="100px">
        <t-form-item label="用户名">
          <span>{{ resetPwdModel.username }}</span>
        </t-form-item>
        <t-form-item label="新密码" name="new_password">
          <t-input v-model="resetPwdModel.new_password" type="password" placeholder="请输入新密码" />
        </t-form-item>
        <t-form-item label="确认密码" name="confirm_password">
          <t-input v-model="resetPwdModel.confirm_password" type="password" placeholder="请再次输入新密码" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { AddIcon, RefreshIcon, SettingIcon } from 'tdesign-icons-vue-next';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref } from 'vue';
import type { PrimaryTableCol } from 'tdesign-vue-next';

import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';

import { batchUpdateUserStatus, createUser, deleteUser, getUserList, resetUserPassword, unlockUser, updateUser, type CreateUserParams, type UpdateUserParams, type UserItem, type UserListParams } from '@/api/user-management';
import { getDeptList, type DeptItem } from '@/api/dept-management';
import { getRoleList, type RoleItem } from '@/api/role-management';
import { formatDateTime } from '@/utils/date';

defineOptions({
  name: 'SystemUser',
});

const loading = ref(false);
const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref();

// 分页
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
});

// 部门列表
const deptList = ref<DeptItem[]>([]);

// 角色列表
const roleList = ref<RoleItem[]>([]);

// 选中行
const selectedRowKeys = ref<(string | number)[]>([]);

// 搜索表单
const formData = reactive({
  username: '',
  realName: '',
  status: undefined as number | undefined,
  deptId: undefined as number | undefined,
});

// 表格数据
const tableData = ref<UserItem[]>([]);

// 基础列定义（不含操作列）
const baseColumns: PrimaryTableCol[] = [
  { colKey: 'id', title: 'ID', width: 80, fixed: 'left' as const },
  { colKey: 'username', title: '用户名', width: 120 },
  { colKey: 'real_name', title: '姓名', width: 100 },
  { colKey: 'phone', title: '手机号', width: 140 },
  { colKey: 'email', title: '邮箱', width: 220 },
  { colKey: 'position', title: '职位', width: 120 },
  { colKey: 'dept_name', title: '部门', width: 120 },
  { colKey: 'roles', title: '角色', width: 180 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'create_time', title: '创建时间', width: 180 },
  { colKey: 'update_time', title: '更新时间', width: 180 },
];

// 完整列定义（含操作列）
const columnsWithOperation: PrimaryTableCol[] = [
  ...baseColumns,
  { colKey: 'operation', title: '操作', width: 180, fixed: 'right' as const },
];

// 列设置相关
const allColumns: ColumnOption[] = [
  { colKey: 'id', title: 'ID', pinned: true },
  { colKey: 'username', title: '用户名' },
  { colKey: 'real_name', title: '姓名' },
  { colKey: 'phone', title: '手机号' },
  { colKey: 'email', title: '邮箱' },
  { colKey: 'position', title: '职位' },
  { colKey: 'dept_name', title: '部门' },
  { colKey: 'roles', title: '角色' },
  { colKey: 'status', title: '状态' },
  { colKey: 'create_time', title: '创建时间' },
  { colKey: 'update_time', title: '更新时间' },
];

const DEFAULT_COLUMNS = ['id', 'username', 'real_name', 'phone', 'email', 'position', 'dept_name', 'roles', 'status', 'create_time'];

const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([]);

// 动态列配置
const displayColumns = computed(() => {
  const orderedColumns = selectedColumns.value
    .map(colKey => columnsWithOperation.find(col => col.colKey === colKey))
    .filter(col => col !== undefined);

  // 操作列始终在最后
  const operationCol = columnsWithOperation.find(col => col.colKey === 'operation');
  if (operationCol) {
    orderedColumns.push(operationCol);
  }

  return orderedColumns;
});

// 表单模型
const formModel = reactive<CreateUserParams & { id?: number }>({
  username: '',
  password: '',
  real_name: '',
  phone: '',
  email: '',
  position: '',
  hire_date: undefined,
  dept_id: undefined,
  status: 1,
  role_ids: [], // 角色ID数组
});

// 手机号正则
const phoneReg = /^1[3-9]\d{9}$/;

// 邮箱正则
const emailReg = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

// 表单验证规则（根据新增/编辑动态调整）
const formRules = computed(() => {
  const rules: Record<string, any[]> = {
    username: [{ required: true, message: '请输入用户名', trigger: 'blur' as const }],
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

  // 新增时才校验密码
  if (!isEdit.value) {
    rules.password = [
      { required: true, message: '请输入密码', trigger: 'blur' as const },
      { min: 6, message: '密码长度不少于6位', trigger: 'blur' as const },
    ];
  }

  return rules;
});

const dialogTitle = ref('新增用户');

const fetchDeptList = async () => {
  try {
    const res = await getDeptList();
    deptList.value = res.list || [];
  } catch (error) {
    console.error('获取部门列表失败:', error);
  }
};

const fetchRoleList = async () => {
  try {
    const res = await getRoleList();
    roleList.value = res.list || [];
  } catch (error) {
    console.error('获取角色列表失败:', error);
  }
};

const fetchUserList = async () => {
  loading.value = true;
  try {
    const params: UserListParams = {
      page: pagination.current,
      page_size: pagination.pageSize,
    };
    if (formData.username) {
      params.username = formData.username;
    }
    if (formData.realName) {
      params.real_name = formData.realName;
    }
    if (formData.status !== undefined) {
      params.status = formData.status;
    }
    if (formData.deptId !== undefined) {
      params.dept_id = formData.deptId;
    }

    const res = await getUserList(params);
    tableData.value = res.list || [];
    pagination.total = res.total || 0;

    // 为每个用户添加部门名称
    tableData.value.forEach(user => {
      const dept = deptList.value.find(d => d.id === user.dept_id);
      user.dept_name = dept ? dept.dept_name : '-';
    });

    // 清除已选行（数据刷新后选中状态失效）
    selectedRowKeys.value = [];
  } catch (error) {
    console.error('获取用户列表失败:', error);
    MessagePlugin.error('获取用户列表失败');
  } finally {
    loading.value = false;
  }
};

// 搜索
const handleSearch = () => {
  pagination.current = 1;
  fetchUserList();
};

const handleReset = () => {
  formData.username = '';
  formData.realName = '';
  formData.status = undefined;
  formData.deptId = undefined;
  pagination.current = 1;
  fetchUserList();
};

// 分页变化
const onPageChange = (curr: any) => {
  pagination.current = curr.current;
  pagination.pageSize = curr.pageSize;
  fetchUserList();
};

// 刷新
const handleRefresh = () => {
  fetchUserList();
  MessagePlugin.success('刷新成功');
};

// 列设置
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

const handleColumnsChange = (cols: string[]) => {
  selectedColumns.value = cols;
};

// 新增
const handleAdd = () => {
  isEdit.value = false;
  dialogTitle.value = '新增用户';
  resetForm();
  dialogVisible.value = true;
};

// 编辑
const handleEdit = (row: UserItem) => {
  isEdit.value = true;
  dialogTitle.value = '编辑用户';
  
  const roleIds = row.roles?.map(role => role.id) || [];
  
  // 如果 dept_id 为 0 或 null，设置为 undefined
  const deptId = row.dept_id && row.dept_id !== 0 ? row.dept_id : undefined;
  
  Object.assign(formModel, {
    id: row.id,
    username: row.username,
    real_name: row.real_name,
    phone: row.phone,
    email: row.email,
    position: row.position,
    hire_date: row.hire_date,
    dept_id: deptId,
    status: row.status,
    role_ids: roleIds,
  });
  dialogVisible.value = true;
};

const handleDelete = async (row: UserItem) => {
  try {
    await deleteUser(row.id);
    MessagePlugin.success('删除成功');
    fetchUserList();
  } catch (error) {
    console.error('删除失败:', error);
    MessagePlugin.error('删除失败');
  }
};

// 解锁账号
const handleUnlockUser = async (row: UserItem) => {
  try {
    await unlockUser(row.id);
    MessagePlugin.success('账号已解锁');
    fetchUserList();
  } catch {
    MessagePlugin.error('解锁失败');
  }
};

// 提交表单
const handleSubmit = async () => {
  const valid = await formRef.value?.validate();
  if (valid !== true) return;

  try {
    if (isEdit.value && formModel.id) {
      // 编辑
      const updateData: UpdateUserParams = {
        id: formModel.id,
        username: formModel.username,
        real_name: formModel.real_name,
        phone: formModel.phone,
        email: formModel.email,
        position: formModel.position,
        hire_date: formModel.hire_date,
        dept_id: formModel.dept_id,
        status: formModel.status,
        role_ids: formModel.role_ids,
      };
      await updateUser(updateData);
      MessagePlugin.success('更新成功');
    } else {
      // 新增
      const createData: CreateUserParams = {
        username: formModel.username,
        password: formModel.password,
        real_name: formModel.real_name,
        phone: formModel.phone,
        email: formModel.email,
        position: formModel.position,
        hire_date: formModel.hire_date,
        dept_id: formModel.dept_id,
        status: formModel.status,
        role_ids: formModel.role_ids,
      };
      await createUser(createData);
      MessagePlugin.success('创建成功');
    }
    dialogVisible.value = false;
    fetchUserList();
  } catch (error) {
    console.error('操作失败:', error);
    MessagePlugin.error(isEdit.value ? '更新失败' : '创建失败');
  }
};

const resetForm = () => {
  Object.assign(formModel, {
    id: undefined,
    username: '',
    password: '',
    real_name: '',
    phone: '',
    email: '',
    position: '',
    hire_date: undefined,
    dept_id: undefined,
    status: 1,
    role_ids: [],
  });
  formRef.value?.reset();
};

// ==================== 批量操作 ====================

// 选中行变化
const handleSelectChange = (keys: (string | number)[]) => {
  selectedRowKeys.value = keys;
};

// 批量启用
const handleBatchEnable = async () => {
  const ids = selectedRowKeys.value.map(Number);
  try {
    await batchUpdateUserStatus({ ids, status: 1 });
    MessagePlugin.success(`已启用 ${ids.length} 个用户`);
    fetchUserList();
  } catch {
    MessagePlugin.error('批量启用失败');
  }
};

// 批量禁用
const handleBatchDisable = async () => {
  const ids = selectedRowKeys.value.map(Number);
  try {
    await batchUpdateUserStatus({ ids, status: 0 });
    MessagePlugin.success(`已禁用 ${ids.length} 个用户`);
    fetchUserList();
  } catch {
    MessagePlugin.error('批量禁用失败');
  }
};

// 批量删除
const handleBatchDelete = async () => {
  const ids = selectedRowKeys.value.map(Number);
  try {
    await Promise.all(ids.map(id => deleteUser(id)));
    MessagePlugin.success(`已删除 ${ids.length} 个用户`);
    fetchUserList();
  } catch {
    MessagePlugin.error('批量删除失败');
  }
};

// ==================== 重置密码 ====================

const resetPwdVisible = ref(false);
const resetPwdLoading = ref(false);
const resetPwdFormRef = ref();
const resetPwdModel = reactive({
  user_id: 0,
  username: '',
  new_password: '',
  confirm_password: '',
});
const resetPwdRules = {
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' as const },
    { min: 6, message: '密码长度不少于6位', trigger: 'blur' as const },
  ],
  confirm_password: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' as const },
    {
      validator: (val: string) => val === resetPwdModel.new_password,
      message: '两次输入的密码不一致',
      trigger: 'blur' as const,
    },
  ],
};

const handleResetPassword = (row: UserItem) => {
  resetPwdModel.user_id = row.id;
  resetPwdModel.username = row.username;
  resetPwdModel.new_password = '';
  resetPwdModel.confirm_password = '';
  resetPwdVisible.value = true;
};

const handleResetPwdCancel = () => {
  resetPwdVisible.value = false;
};

const handleResetPwdSubmit = async () => {
  const valid = await resetPwdFormRef.value?.validate();
  if (valid !== true) return;

  resetPwdLoading.value = true;
  try {
    await resetUserPassword(resetPwdModel.user_id, resetPwdModel.new_password);
    MessagePlugin.success('密码重置成功');
    resetPwdVisible.value = false;
  } catch {
    MessagePlugin.error('密码重置失败');
  } finally {
    resetPwdLoading.value = false;
  }
};

// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'username', placeholder: '请输入用户名' },
  { type: 'input', field: 'realName', placeholder: '请输入姓名' },
  {
    type: 'select',
    field: 'status',
    placeholder: '请选择状态',
    options: [
      { label: '正常', value: 1 },
      { label: '禁用', value: 0 },
    ],
    showMore: true,
  },
  {
    type: 'select',
    field: 'deptId',
    placeholder: '请选择部门',
    options: deptList,
    labelKey: 'dept_name',
    valueKey: 'id',
    showMore: true,
  },
];

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('system_user_table_columns');
    if (saved) {
      const savedData = JSON.parse(saved);
      // ColumnSettingDialog 保存的是 { order, visible } 对象
      if (savedData.visible && Array.isArray(savedData.visible)) {
        selectedColumns.value = savedData.visible;
      } else if (Array.isArray(savedData)) {
        selectedColumns.value = savedData;
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

onMounted(() => {
  loadColumnConfig();
  fetchUserList();
  fetchDeptList();
  fetchRoleList();
});
</script>

<style lang="less" scoped>
.user-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;

  :deep(.search-card) {
    margin-bottom: 0;
  }

  .toolbar-card {
    margin-bottom: 0;

    .toolbar-content {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .toolbar-title {
        font-size: 16px;
        font-weight: 500;
        color: var(--td-text-color-primary);
      }

      .toolbar-actions {
        display: flex;
        gap: 8px;
      }
    }
  }
}

// 批量操作栏
.batch-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  margin-bottom: 8px;
  background-color: var(--td-bg-color-container-hover, #f5f5f5);
  border-radius: var(--td-radius-default, 6px);

  .batch-count {
    font-size: 14px;
    color: var(--td-brand-color, #0052d9);
    font-weight: 500;
  }
}

.no-role {
  color: var(--td-text-color-disabled, rgba(0, 0, 0, 0.36));
}

.role-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
</style>
