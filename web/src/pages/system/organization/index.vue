<template>
  <div class="organization-container">
    <!-- 操作栏卡片 -->
    <t-card class="toolbar-card">
      <div class="page-toolbar">
        <span class="toolbar-title">组织与权限</span>
        <div class="toolbar-actions">
          <t-space>
            <template v-if="activeTab === 'user'">
              <t-button v-permission="'system:user:create'" theme="primary" @click="handleUserAdd">
                <template #icon><add-icon /></template>
                新增用户
              </t-button>
            </template>

            <template v-if="activeTab === 'role'">
              <t-button v-permission="'system:role:create'" theme="primary" @click="handleRoleAdd">
                <template #icon><add-icon /></template>
                新增角色
              </t-button>
            </template>

            <template v-if="activeTab === 'dept'">
              <t-button theme="primary" @click="handleDeptAdd">
                <template #icon><add-icon /></template>
                新增部门
              </t-button>
            </template>

            <!-- 刷新按钮统一在最右侧 -->
            <t-button variant="outline" @click="handleRefresh">
              <template #icon><refresh-icon /></template>
              刷新
            </t-button>
          </t-space>
        </div>
      </div>
    </t-card>

    <!-- 主内容区 - Tabs -->
    <t-card :bordered="false">
      <t-tabs v-model="activeTab" @change="handleTabChange">
        <!-- 公司信息 Tab -->
        <t-tab-panel value="company" label="公司信息">
          <div class="tab-content">
            <t-form
              ref="companyFormRef"
              :data="companyFormData"
              :rules="companyFormRules"
              label-width="100px"
              style="max-width: 600px; margin-top: 16px;"
            >
              <t-form-item label="公司名称" name="company_name">
                <t-input
                  v-model="companyFormData.company_name"
                  placeholder="请输入公司全称"
                  :maxlength="100"
                />
              </t-form-item>
              <t-form-item label="公司简称" name="short_name">
                <t-input
                  v-model="companyFormData.short_name"
                  placeholder="请输入公司简称"
                  :maxlength="50"
                />
              </t-form-item>
              <t-form-item label="英文名称" name="english_name">
                <t-input
                  v-model="companyFormData.english_name"
                  placeholder="请输入英文全称"
                  :maxlength="200"
                />
              </t-form-item>
              <t-form-item label="英文简称" name="english_short_name">
                <t-input
                  v-model="companyFormData.english_short_name"
                  placeholder="请输入英文简称"
                  :maxlength="100"
                />
              </t-form-item>
              <t-form-item label="公司 Logo" name="logo">
                <t-upload
                  v-model="companyLogoFiles"
                  theme="image"
                  :max="1"
                  accept=".png,.jpg,.jpeg,.gif,.svg"
                  :request-method="handleCompanyLogoUpload"
                  :abridge-name="[8, 6]"
                  tips="建议尺寸 336*44px，支持 JPG/PNG/GIF/SVG 格式"
                  @remove="handleCompanyLogoRemove"
                />
              </t-form-item>
            </t-form>
            <div style="margin-top: 24px; padding-left: 100px;">
              <t-button theme="primary" @click="handleCompanySave">
                <template #icon><check-icon /></template>
                保存
              </t-button>
            </div>
          </div>
        </t-tab-panel>

        <!-- 用户管理 Tab -->
        <t-tab-panel value="user" label="用户管理">
          <div class="tab-content">
            <t-table
              :data="userTableData"
              :columns="userDisplayColumns"
              :loading="userLoading"
              row-key="id"
              :stripe="true"
            >
              <template #status="{ row }">
                <t-tag :theme="row.status === 1 ? 'success' : 'danger'">
                  {{ row.status === 1 ? '正常' : '禁用' }}
                </t-tag>
              </template>
              <template #roles="{ row }">
                <div v-if="row.roles && row.roles.length > 0" class="role-tags">
                  <t-tag v-for="role in row.roles" :key="role.id" theme="primary" variant="light">
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
                <t-link v-permission="'system:user:update'" theme="primary" @click="handleUserEdit(row)" style="margin-right: 8px">编辑</t-link>
                <t-link v-permission="'system:user:update'" theme="primary" @click="handleUserResetPassword(row)" style="margin-right: 8px">重置密码</t-link>
                <t-popconfirm content="确认删除该用户？" @confirm="handleUserDelete(row)">
                  <t-link v-permission="'system:user:delete'" theme="danger">删除</t-link>
                </t-popconfirm>
              </template>
            </t-table>
          </div>
        </t-tab-panel>

        <!-- 角色管理 Tab -->
        <t-tab-panel value="role" label="角色管理">
          <div class="tab-content">
            <t-table
              :data="roleList"
              :columns="roleColumns"
              :loading="roleLoading"
              row-key="id"
              :stripe="true"
            >
              <template #role_type="{ row }">
                <t-tag v-if="row.role_type === 1" theme="warning" variant="light">超级管理员</t-tag>
                <t-tag v-else-if="row.role_type === 2" theme="primary" variant="light">业务内置</t-tag>
                <t-tag v-else theme="default" variant="light">自定义</t-tag>
              </template>
              <template #data_scope="{ row }">
                <span v-if="row.data_scope === 1">全部数据</span>
                <span v-else-if="row.data_scope === 2">本部门及以下</span>
                <span v-else-if="row.data_scope === 3">本部门</span>
                <span v-else-if="row.data_scope === 4">仅本人</span>
              </template>
              <template #status="{ row }">
                <t-tag :theme="row.status === 1 ? 'success' : 'danger'" variant="light">
                  {{ row.status === 1 ? '启用' : '禁用' }}
                </t-tag>
              </template>
              <template #create_time="{ row }">
                <span>{{ formatDateTime(row.create_time) }}</span>
              </template>
              <template #operation="{ row }">
                <t-link v-permission="'system:role:update'" theme="primary" :disabled="row.role_type === 1" @click="row.role_type === 1 ? undefined : handleRoleEdit(row)" style="margin-right: 8px">编辑</t-link>
                <t-link v-permission="'system:role:assign-perm'" theme="primary" :disabled="row.role_type === 1" @click="row.role_type === 1 ? undefined : handleRoleAssignMenu(row)" style="margin-right: 8px">分配菜单权限</t-link>
                <t-popconfirm :content="row.role_type >= 1 ? '内置角色不可删除' : '确认删除该角色？'" :disabled="row.role_type >= 1" @confirm="handleRoleDelete(row)">
                  <t-link v-permission="'system:role:delete'" :theme="row.role_type >= 1 ? 'default' : 'danger'" :disabled="row.role_type >= 1">删除</t-link>
                </t-popconfirm>
              </template>
            </t-table>
          </div>
        </t-tab-panel>

        <!-- 部门管理 Tab -->
        <t-tab-panel value="dept" label="部门管理">
          <div class="tab-content">
            <t-enhanced-table
              ref="deptTableRef"
              :data="deptTableData"
              :columns="deptDisplayColumns"
              :loading="deptLoading"
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
                <t-link theme="primary" @click="handleDeptEdit(row)" style="margin-right: 8px">编辑</t-link>
                <t-popconfirm :content="`确认删除「${row.dept_name}」？`" @confirm="handleDeptDelete(row)">
                  <t-link theme="danger">删除</t-link>
                </t-popconfirm>
              </template>
            </t-enhanced-table>
          </div>
        </t-tab-panel>
      </t-tabs>
    </t-card>

    <!-- ========== 用户编辑对话框 ========== -->
    <t-dialog
      v-model:visible="userDialogVisible"
      :title="userDialogTitle"
      width="600px"
      placement="center"
      @confirm="handleUserSubmit"
      @cancel="userDialogVisible = false"
     :close-on-overlay-click="false">
      <t-form ref="userFormRef" :data="userFormModel" :rules="userFormRules" label-width="100px">
        <t-form-item label="用户名" name="username">
          <t-input v-model="userFormModel.username" placeholder="请输入用户名" />
        </t-form-item>
        <t-form-item v-if="!userIsEdit" label="密码" name="password">
          <t-input v-model="userFormModel.password" type="password" placeholder="请输入密码" />
        </t-form-item>
        <t-form-item label="姓名" name="real_name">
          <t-input v-model="userFormModel.real_name" placeholder="请输入姓名" />
        </t-form-item>
        <t-form-item label="手机号" name="phone">
          <t-input v-model="userFormModel.phone" placeholder="请输入手机号" />
        </t-form-item>
        <t-form-item label="邮箱" name="email">
          <t-input v-model="userFormModel.email" placeholder="请输入邮箱" />
        </t-form-item>
        <t-form-item label="职位" name="position">
          <t-input v-model="userFormModel.position" placeholder="请输入职位" />
        </t-form-item>
        <t-form-item label="入职日期" name="hire_date">
          <t-date-picker v-model="userFormModel.hire_date" placeholder="请选择入职日期" format="YYYY-MM-DD" clearable />
        </t-form-item>
        <t-form-item label="部门" name="dept_id">
          <t-select v-model="userFormModel.dept_id" placeholder="请选择部门" clearable>
            <t-option v-for="dept in deptList" :key="dept.id" :value="dept.id" :label="dept.dept_name" />
          </t-select>
        </t-form-item>
        <t-form-item label="角色" name="role_ids">
          <t-select v-model="userFormModel.role_ids" placeholder="请选择角色" multiple clearable>
            <t-option v-for="role in roleListOpt" :key="role.id" :value="role.id" :label="role.role_name" />
          </t-select>
        </t-form-item>
        <t-form-item label="状态" name="status">
          <t-select v-model="userFormModel.status" placeholder="请选择状态">
            <t-option :value="1" label="正常" />
            <t-option :value="0" label="禁用" />
          </t-select>
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- ========== 用户重置密码对话框 ========== -->
    <t-dialog
      v-model:visible="userResetPwdVisible"
      title="重置密码"
      width="400px"
      placement="center"
      :confirm-btn="{ content: '确定', loading: userResetPwdLoading }"
      @confirm="handleUserResetPwdSubmit"
      @cancel="handleUserResetPwdCancel"
     :close-on-overlay-click="false">
      <t-form ref="userResetPwdFormRef" :data="userResetPwdModel" :rules="userResetPwdRules" label-width="100px">
        <t-form-item label="用户名">
          <span>{{ userResetPwdModel.username }}</span>
        </t-form-item>
        <t-form-item label="新密码" name="new_password">
          <t-input v-model="userResetPwdModel.new_password" type="password" placeholder="请输入新密码" />
        </t-form-item>
        <t-form-item label="确认密码" name="confirm_password">
          <t-input v-model="userResetPwdModel.confirm_password" type="password" placeholder="请再次输入新密码" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- ========== 角色编辑对话框 ========== -->
    <t-dialog
      v-model:visible="roleDialogVisible"
      :title="roleDialogTitle"
      width="500px"
      :confirm-btn="{ content: '确定', loading: roleSubmitting }"
      @confirm="handleRoleSubmit"
      @cancel="roleDialogVisible = false"
     :close-on-overlay-click="false">
      <t-form ref="roleFormRef" :data="roleFormModel" :rules="roleFormRules" label-width="100px">
        <t-form-item label="角色名称" name="role_name">
          <t-input v-model="roleFormModel.role_name" placeholder="请输入角色名称" />
        </t-form-item>
        <t-form-item label="角色备注" name="remark">
          <t-input v-model="roleFormModel.remark" placeholder="请输入备注" />
        </t-form-item>
        <t-form-item label="数据权限" name="data_scope">
          <t-select v-model="roleFormModel.data_scope">
            <t-option :value="1" label="全部数据" />
            <t-option :value="2" label="本部门及以下" />
            <t-option :value="3" label="本部门" />
            <t-option :value="4" label="仅本人" />
          </t-select>
        </t-form-item>
        <t-form-item label="状态" name="status">
          <t-radio-group v-model="roleFormModel.status">
            <t-radio-button :value="1">启用</t-radio-button>
            <t-radio-button :value="0">禁用</t-radio-button>
          </t-radio-group>
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- ========== 角色分配菜单权限对话框 ========== -->
    <t-dialog
      v-model:visible="roleMenuDialogVisible"
      :title="`分配菜单权限 - ${roleCurrentRole?.role_name || ''}`"
      width="500px"
      :confirm-btn="{ content: '保存', loading: roleMenuSubmitting }"
      @confirm="handleRoleMenuSubmit"
      @cancel="roleMenuDialogVisible = false"
     :close-on-overlay-click="false">
      <div v-if="roleMenuLoading" style="text-align:center;padding:40px">加载中...</div>
      <t-tree
        v-else
        ref="roleMenuTreeRef"
        :data="roleAllMenuTree"
        :keys="{ label: 'menu_name', value: 'id', children: 'children' }"
        :value="roleCheckedMenuIds"
        :checkable="true"
        :check-strictly="false"
        :expand-all="false"
        :hover="true"
        :activatable="true"
        @change="handleRoleMenuCheckChange"
      />
    </t-dialog>

    <!-- ========== 部门编辑对话框 ========== -->
    <t-dialog
      v-model:visible="deptDialogVisible"
      :title="deptDialogTitle"
      width="600px"
      placement="center"
      :footer="null"
     :close-on-overlay-click="false">
      <t-form ref="deptFormRef" :data="deptFormModel" :rules="deptFormRules" label-width="100px">
        <t-form-item label="父部门" name="parent_id">
          <t-select v-model="deptFormModel.parent_id" placeholder="请选择父部门（不选则为顶级部门）" clearable>
            <t-option :value="0" label="无（顶级部门）" />
            <t-option v-for="dept in deptFlatList" :key="dept.id" :value="dept.id" :label="dept.label" />
          </t-select>
        </t-form-item>
        <t-form-item label="部门名称" name="dept_name">
          <t-input v-model="deptFormModel.dept_name" placeholder="请输入部门名称" />
        </t-form-item>
        <t-form-item label="排序" name="sort">
          <t-input-number v-model="deptFormModel.sort" :min="0" placeholder="请输入排序" />
        </t-form-item>
        <t-form-item label="负责人" name="leader">
          <t-input v-model="deptFormModel.leader" placeholder="请输入负责人" />
        </t-form-item>
        <t-form-item label="联系电话" name="phone">
          <t-input v-model="deptFormModel.phone" placeholder="请输入联系电话" />
        </t-form-item>
        <t-form-item label="状态" name="status">
          <t-radio-group v-model="deptFormModel.status">
            <t-radio :value="1">正常</t-radio>
            <t-radio :value="0">停用</t-radio>
          </t-radio-group>
        </t-form-item>
        <div style="text-align: right; margin-top: 16px;">
          <t-button theme="default" @click="deptDialogVisible = false" style="margin-right: 8px">取消</t-button>
          <t-button theme="primary" @click="handleDeptSubmit">确定</t-button>
        </div>
      </t-form>
    </t-dialog>

  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { PrimaryTableCol, TabValue } from 'tdesign-vue-next';
import { AddIcon, CheckIcon, RefreshIcon } from 'tdesign-icons-vue-next';
import { request } from '@/utils/request';

import { getCompanyInfo, updateCompanyInfo } from '@/api/company-management';
import {
  getUserList, createUser, updateUser, deleteUser,
  resetUserPassword,
  type CreateUserParams, type UpdateUserParams, type UserItem,
} from '@/api/user-management';
import {
  getRoleList, createRole, updateRole, deleteRole,
  type RoleItem, type CreateRoleParams, type UpdateRoleParams,
} from '@/api/role-management';
import { getRoleMenuTree, assignRoleMenus, getRoleMenus, type MenuItem } from '@/api/menu';
import { getDeptList, createDept, updateDept, deleteDept, type CreateDeptParams, type DeptItem, type UpdateDeptParams } from '@/api/dept-management';
import { formatDateTime } from '@/utils/date';

defineOptions({ name: 'SystemOrganization' });

const phoneReg = /^1[3-9]\d{9}$/;
const emailReg = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

// ==================== 通用 ====================
const activeTab = ref('company');

const handleTabChange = (value: TabValue) => {
  activeTab.value = value as string;
  if (value === 'user') fetchUserList();
  else if (value === 'role') fetchRoleList();
  else if (value === 'dept') fetchDeptList();
  else if (value === 'company') fetchCompanyInfo();
};

// 统一刷新按钮
const handleRefresh = () => {
  switch (activeTab.value) {
    case 'company': fetchCompanyInfo(); break;
    case 'user': fetchUserList(); break;
    case 'role': fetchRoleList(); break;
    case 'dept': fetchDeptList(); break;
  }
};

// ==================== 公司信息 ====================
const companyFormRef = ref();
const companyFormData = reactive({ company_name: '', short_name: '', english_name: '', english_short_name: '' });
const companyFormRules = { company_name: [{ required: true, message: '请输入公司名称', trigger: 'blur' as const }] };
// 公司 logo
const companyLogoFiles = ref<any[]>([]);
let companyLogoUrl = '';

const fetchCompanyInfo = async () => {
  try {
    const res = await getCompanyInfo();
    if (res) {
      companyFormData.company_name = res.company_name || '';
      companyFormData.short_name = res.short_name || '';
      companyFormData.english_name = res.english_name || '';
      companyFormData.english_short_name = res.english_short_name || '';
      companyLogoUrl = res.logo || '';
      // 回显已有 logo
      if (res.logo) {
        companyLogoFiles.value = [{
          name: 'company-logo',
          url: res.logo.startsWith('http') ? res.logo : `${import.meta.env.VITE_API_URL || ''}${res.logo}`,
          status: 'success',
        }];
      } else {
        companyLogoFiles.value = [];
      }
    }
  } catch {
    MessagePlugin.error('获取公司信息失败');
  }
};

// 上传 logo
const handleCompanyLogoUpload = async (file: any) => {
  const formData = new FormData();
  formData.append('file', file.raw);
  try {
    const res = await request.post<{ url: string }>({ url: '/system/company/logo', data: formData, headers: { 'Content-Type': 'multipart/form-data' } });
    companyLogoUrl = res.url;
    return { status: 'success', url: res.url };
  } catch {
    MessagePlugin.error('Logo 上传失败');
    return { status: 'fail' as const };
  }
};

// 移除 logo
const handleCompanyLogoRemove = async () => {
  try {
    await request.delete({ url: '/system/company/logo' });
    companyLogoUrl = '';
    companyLogoFiles.value = [];
    MessagePlugin.success('已恢复默认 Logo');
  } catch {
    MessagePlugin.error('操作失败');
  }
};

const handleCompanySave = async () => {
  const valid = await companyFormRef.value?.validate();
  if (valid !== true) return;
  try {
    await updateCompanyInfo({
      company_name: companyFormData.company_name.trim(),
      short_name: companyFormData.short_name.trim(),
      english_name: companyFormData.english_name.trim(),
      english_short_name: companyFormData.english_short_name.trim(),
      logo: companyLogoUrl,
    });
    MessagePlugin.success('保存成功');
    fetchCompanyInfo();
  } catch {
    MessagePlugin.error('保存失败');
  }
};

// ==================== 用户管理 ====================
const userLoading = ref(false);
const userTableData = ref<UserItem[]>([]);
const userDialogVisible = ref(false);
const userIsEdit = ref(false);
const userDialogTitle = ref('新增用户');
const userFormRef = ref();
const userResetPwdVisible = ref(false);
const userResetPwdLoading = ref(false);
const userResetPwdFormRef = ref();

const userFormModel = reactive<CreateUserParams & { id?: number }>({
  username: '', password: '', real_name: '', phone: '', email: '', position: '',
  hire_date: undefined, dept_id: undefined, status: 1, role_ids: [],
});

const userResetPwdModel = reactive({
  user_id: 0, username: '', new_password: '', confirm_password: '',
});

const userResetPwdRules = {
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' as const },
    { min: 6, message: '密码长度不少于6位', trigger: 'blur' as const },
  ],
  confirm_password: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' as const },
    { validator: (val: string) => val === userResetPwdModel.new_password, message: '两次输入的密码不一致', trigger: 'blur' as const },
  ],
};

const userFormRules = computed(() => {
  const rules: Record<string, any[]> = {
    username: [{ required: true, message: '请输入用户名', trigger: 'blur' as const }],
    phone: [{ pattern: phoneReg, message: '请输入正确的手机号', trigger: 'blur' as const }],
    email: [{ pattern: emailReg, message: '请输入正确的邮箱地址', trigger: 'blur' as const }],
  };
  if (!userIsEdit.value) {
    rules.password = [
      { required: true, message: '请输入密码', trigger: 'blur' as const },
      { min: 6, message: '密码长度不少于6位', trigger: 'blur' as const },
    ];
  }
  return rules;
});

const userBaseColumns: PrimaryTableCol[] = [
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
const userColumnsWithOperation: PrimaryTableCol[] = [
  ...userBaseColumns,
  { colKey: 'operation', title: '操作', width: 180, fixed: 'right' as const },
];
const userDisplayColumns = userColumnsWithOperation;

// 部门/角色列表供用户表单下拉使用
const deptList = ref<DeptItem[]>([]);
const roleListOpt = ref<RoleItem[]>([]);

const fetchRoleListOpt = async () => {
  try { const res = await getRoleList(); roleListOpt.value = res.list || []; } catch {}
};

const fetchUserList = async () => {
  userLoading.value = true;
  try {
    const res = await getUserList({ page: 1, page_size: 9999 });
    userTableData.value = (res.list || []).map((user: UserItem) => {
      const dept = deptList.value.find(d => d.id === user.dept_id);
      return { ...user, dept_name: dept ? dept.dept_name : '-' };
    });
  } catch {
    MessagePlugin.error('获取用户列表失败');
  } finally {
    userLoading.value = false;
  }
};

const handleUserAdd = () => {
  userIsEdit.value = false;
  userDialogTitle.value = '新增用户';
  resetUserForm();
  userDialogVisible.value = true;
};

const handleUserEdit = (row: UserItem) => {
  userIsEdit.value = true;
  userDialogTitle.value = '编辑用户';
  const roleIds = row.roles?.map(r => r.id) || [];
  const deptId = row.dept_id && row.dept_id !== 0 ? row.dept_id : undefined;
  Object.assign(userFormModel, {
    id: row.id, username: row.username, real_name: row.real_name, phone: row.phone,
    email: row.email, position: row.position, hire_date: row.hire_date,
    dept_id: deptId, status: row.status, role_ids: roleIds,
  });
  userDialogVisible.value = true;
};

const handleUserDelete = async (row: UserItem) => {
  try { await deleteUser(row.id); MessagePlugin.success('删除成功'); fetchUserList(); } catch { MessagePlugin.error('删除失败'); }
};

const handleUserSubmit = async () => {
  const valid = await userFormRef.value?.validate();
  if (valid !== true) return;
  try {
    if (userIsEdit.value && userFormModel.id) {
      await updateUser({ id: userFormModel.id, username: userFormModel.username, real_name: userFormModel.real_name, phone: userFormModel.phone, email: userFormModel.email, position: userFormModel.position, hire_date: userFormModel.hire_date, dept_id: userFormModel.dept_id, status: userFormModel.status, role_ids: userFormModel.role_ids });
      MessagePlugin.success('更新成功');
    } else {
      await createUser({ username: userFormModel.username, password: userFormModel.password, real_name: userFormModel.real_name, phone: userFormModel.phone, email: userFormModel.email, position: userFormModel.position, hire_date: userFormModel.hire_date, dept_id: userFormModel.dept_id, status: userFormModel.status, role_ids: userFormModel.role_ids });
      MessagePlugin.success('创建成功');
    }
    userDialogVisible.value = false;
    fetchUserList();
  } catch { MessagePlugin.error(userIsEdit.value ? '更新失败' : '创建失败'); }
};

const resetUserForm = () => {
  Object.assign(userFormModel, { id: undefined, username: '', password: '', real_name: '', phone: '', email: '', position: '', hire_date: undefined, dept_id: undefined, status: 1, role_ids: [] });
  userFormRef.value?.reset();
};

const handleUserResetPassword = (row: UserItem) => {
  userResetPwdModel.user_id = row.id;
  userResetPwdModel.username = row.username;
  userResetPwdModel.new_password = '';
  userResetPwdModel.confirm_password = '';
  userResetPwdVisible.value = true;
};

const handleUserResetPwdCancel = () => { userResetPwdVisible.value = false; };

const handleUserResetPwdSubmit = async () => {
  const valid = await userResetPwdFormRef.value?.validate();
  if (valid !== true) return;
  userResetPwdLoading.value = true;
  try {
    await resetUserPassword(userResetPwdModel.user_id, userResetPwdModel.new_password);
    MessagePlugin.success('密码重置成功');
    userResetPwdVisible.value = false;
  } catch { MessagePlugin.error('密码重置失败'); }
  finally { userResetPwdLoading.value = false; }
};

// ==================== 角色管理 ====================
const roleLoading = ref(false);
const roleList = ref<RoleItem[]>([]);
const roleDialogVisible = ref(false);
const roleSubmitting = ref(false);
const roleIsEdit = ref(false);
const roleFormRef = ref();
const roleMenuDialogVisible = ref(false);
const roleMenuLoading = ref(false);
const roleMenuSubmitting = ref(false);
const roleMenuTreeRef = ref();
const roleAllMenuTree = ref<MenuItem[]>([]);
const roleCheckedMenuIds = ref<(string | number)[]>([]);
const roleCurrentRole = ref<RoleItem | null>(null);

const roleFormModel = reactive<CreateRoleParams & { id?: number; data_scope?: number; status?: number }>({
  role_name: '', remark: '', data_scope: 1, status: 1,
});

const roleFormRules = { role_name: [{ required: true, message: '请输入角色名称', trigger: 'blur' as const }] };

const roleDialogTitle = computed(() => (roleIsEdit.value ? '编辑角色' : '新增角色'));

const roleColumns = [
  { colKey: 'role_name', title: '角色名称', width: 160, fixed: 'left' as const },
  { colKey: 'remark', title: '备注', width: 200 },
  { colKey: 'role_type', title: '类型', width: 100 },
  { colKey: 'data_scope', title: '数据权限', width: 130 },
  { colKey: 'status', title: '状态', width: 80 },
  { colKey: 'create_time', title: '创建时间', width: 180 },
  { colKey: 'operation', title: '操作', width: 260, fixed: 'right' as const },
];

const fetchRoleList = async () => {
  roleLoading.value = true;
  try {
    const res = await getRoleList();
    roleList.value = res.list || [];
  } catch { MessagePlugin.error('获取角色列表失败'); }
  finally { roleLoading.value = false; }
};

const handleRoleAdd = () => {
  roleIsEdit.value = false;
  Object.assign(roleFormModel, { id: undefined, role_name: '', remark: '', data_scope: 1, status: 1 });
  roleFormRef.value?.reset();
  roleDialogVisible.value = true;
};

const handleRoleEdit = (row: RoleItem) => {
  roleIsEdit.value = true;
  Object.assign(roleFormModel, { id: row.id, role_name: row.role_name, remark: row.remark, data_scope: row.data_scope || 1, status: row.status });
  roleDialogVisible.value = true;
};

const handleRoleDelete = async (row: RoleItem) => {
  try { await deleteRole(row.id); MessagePlugin.success('删除成功'); fetchRoleList(); } catch { MessagePlugin.error('删除失败'); }
};

const handleRoleSubmit = async () => {
  const valid = await roleFormRef.value?.validate();
  if (valid !== true) return;
  roleSubmitting.value = true;
  try {
    if (roleIsEdit.value && roleFormModel.id) {
      await updateRole({ id: roleFormModel.id, role_name: roleFormModel.role_name, remark: roleFormModel.remark, data_scope: roleFormModel.data_scope, status: roleFormModel.status });
      MessagePlugin.success('更新成功');
    } else {
      await createRole({ role_name: roleFormModel.role_name, remark: roleFormModel.remark });
      MessagePlugin.success('创建成功');
    }
    roleDialogVisible.value = false;
    fetchRoleList();
  } catch { MessagePlugin.error(roleIsEdit.value ? '更新失败' : '创建失败'); }
  finally { roleSubmitting.value = false; }
};

const handleRoleAssignMenu = async (row: RoleItem) => {
  roleCurrentRole.value = row;
  roleMenuDialogVisible.value = true;
  roleMenuLoading.value = true;
  try {
    const [menuRes, roleMenusRes] = await Promise.all([getRoleMenuTree(), getRoleMenus(row.id)]);
    roleAllMenuTree.value = menuRes.list || [];
    roleCheckedMenuIds.value = roleMenusRes.menu_ids || [];
  } catch { MessagePlugin.error('获取菜单数据失败'); }
  finally { roleMenuLoading.value = false; }
};

const handleRoleMenuCheckChange = (value: (string | number)[]) => { roleCheckedMenuIds.value = value; };

function ensureAncestorIds(checked: number[], treeData: MenuItem[]): number[] {
  const idSet = new Set<number>(checked);
  const parentMap = new Map<number, number>();
  function buildMap(nodes: MenuItem[]) {
    for (const node of nodes) {
      if (node.children) {
        for (const child of node.children) parentMap.set(child.id, node.id);
        buildMap(node.children);
      }
    }
  }
  buildMap(treeData);
  for (const id of checked) {
    let pid = parentMap.get(id);
    while (pid && !idSet.has(pid)) { idSet.add(pid); pid = parentMap.get(pid); }
  }
  return Array.from(idSet);
}

const handleRoleMenuSubmit = async () => {
  if (!roleCurrentRole.value) return;
  roleMenuSubmitting.value = true;
  try {
    let menuIds = roleCheckedMenuIds.value.map(id => Number(id));
    menuIds = ensureAncestorIds(menuIds, roleAllMenuTree.value);
    await assignRoleMenus({ role_id: roleCurrentRole.value.id, menu_ids: menuIds });
    MessagePlugin.success('菜单权限分配成功');
    roleMenuDialogVisible.value = false;
  } catch { MessagePlugin.error('分配失败'); }
  finally { roleMenuSubmitting.value = false; }
};

// ==================== 部门管理 ====================
const deptLoading = ref(false);
const deptTableData = ref<DeptItem[]>([]);
const deptDialogVisible = ref(false);
const deptIsEdit = ref(false);
const deptDialogTitle = ref('新增部门');
const deptFormRef = ref();
const deptTableRef = ref();

const deptBaseColumns = [
  { colKey: 'dept_name', title: '部门名称', width: 200, fixed: 'left' as const },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'leader', title: '负责人', width: 120 },
  { colKey: 'phone', title: '联系电话', width: 150 },
  { colKey: 'sort', title: '排序', width: 100 },
  { colKey: 'create_time', title: '创建时间', width: 180 },
];
const deptColumnsWithOperation = [
  ...deptBaseColumns,
  { colKey: 'operation', title: '操作', width: 250, fixed: 'right' as const },
];
const deptDisplayColumns = deptColumnsWithOperation;

const MAX_DEPT_LEVEL = 3;

const deptFlatList = computed(() => {
  const result: { id: number; label: string }[] = [];
  const flatten = (nodes: DeptItem[], depth: number) => {
    for (const node of nodes) {
      if (deptIsEdit.value && node.id === deptFormModel.id) continue;
      const prefix = depth > 0 ? '  '.repeat(depth) + '└ ' : '';
      result.push({ id: node.id, label: prefix + node.dept_name });
      if (node.children && node.children.length > 0 && depth < MAX_DEPT_LEVEL - 1) flatten(node.children, depth + 1);
    }
  };
  flatten(deptTableData.value, 0);
  return result;
});

const deptFormModel = reactive<CreateDeptParams & { id?: number }>({
  dept_name: '', parent_id: 0, sort: 0, leader: '', phone: '', status: 1,
});

const deptFormRules = { dept_name: [{ required: true, message: '请输入部门名称', trigger: 'blur' as const }] };

const fetchDeptList = async () => {
  deptLoading.value = true;
  try {
    const res = await getDeptList();
    deptTableData.value = res.list || [];
    deptList.value = res.list || []; // 同步更新用户表单的部门下拉
  } catch { MessagePlugin.error('获取部门列表失败'); }
  finally { deptLoading.value = false; }
};

const handleDeptAdd = () => {
  deptIsEdit.value = false;
  deptDialogTitle.value = '新增部门';
  resetDeptForm();
  deptDialogVisible.value = true;
};

const handleDeptEdit = (row: DeptItem) => {
  deptIsEdit.value = true;
  deptDialogTitle.value = '编辑部门';
  Object.assign(deptFormModel, { id: row.id, dept_name: row.dept_name, parent_id: row.parent_id, sort: row.sort, leader: row.leader, phone: row.phone, status: row.status });
  deptDialogVisible.value = true;
};

const handleDeptDelete = async (row: DeptItem) => {
  try { await deleteDept(row.id); MessagePlugin.success('删除成功'); fetchDeptList(); }
  catch (error: any) { MessagePlugin.error(error?.response?.data?.message || error?.message || '删除失败'); }
};

const handleDeptSubmit = async () => {
  const valid = await deptFormRef.value?.validate();
  if (valid !== true) return;
  try {
    if (deptIsEdit.value && deptFormModel.id) {
      await updateDept({
        id: deptFormModel.id, dept_name: deptFormModel.dept_name, parent_id: deptFormModel.parent_id,
        sort: deptFormModel.sort, leader: deptFormModel.leader, phone: deptFormModel.phone, status: deptFormModel.status,
      });
      MessagePlugin.success('更新成功');
    } else {
      await createDept({
        dept_name: deptFormModel.dept_name, parent_id: deptFormModel.parent_id, sort: deptFormModel.sort,
        leader: deptFormModel.leader, phone: deptFormModel.phone, status: deptFormModel.status,
      });
      MessagePlugin.success('创建成功');
    }
    deptDialogVisible.value = false;
    fetchDeptList();
  } catch (error: any) { MessagePlugin.error(error?.response?.data?.message || error?.message || (deptIsEdit.value ? '更新失败' : '创建失败')); }
};

const resetDeptForm = () => {
  Object.assign(deptFormModel, { id: undefined, dept_name: '', parent_id: 0, sort: 0, leader: '', phone: '', status: 1 });
  deptFormRef.value?.reset();
};

// ==================== 初始化 ====================
onMounted(async () => {
  await Promise.all([fetchCompanyInfo(), fetchDeptList(), fetchRoleListOpt()]);
  // 默认Tab加载对应数据
  if (activeTab.value === 'user') fetchUserList();
  else if (activeTab.value === 'role') fetchRoleList();
  else if (activeTab.value === 'dept') fetchDeptList();
});
</script>

<style scoped lang="less">
.organization-container {
  padding: 16px;

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
}
</style>
