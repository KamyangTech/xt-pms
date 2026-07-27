<template>
  <div class="role-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="formData"
      @search="handleSearch"
      @reset="handleReset"
    />

    <t-card class="toolbar-card">
      <div class="toolbar-content">
        <span class="toolbar-title">角色管理</span>
        <div class="toolbar-actions">
          <t-button theme="default" variant="outline" @click="fetchRoleList">
            <template #icon><t-icon name="refresh" /></template>
            刷新
          </t-button>
          <t-button v-permission="'system:role:create'" theme="primary" @click="handleAdd">
            <template #icon><t-icon name="add" /></template>
            新增角色
          </t-button>
        </div>
      </div>
    </t-card>

    <t-card :bordered="false">
      <t-table
        :data="roleList"
        :columns="columns"
        :loading="loading"
        row-key="id"
        hover
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
          <t-link
            v-permission="'system:role:update'"
            theme="primary"
            :disabled="row.role_type === 1"
            @click="row.role_type === 1 ? undefined : handleEdit(row)"
            style="margin-right: 8px"
          >编辑</t-link>
          <t-link
            v-permission="'system:role:assign-perm'"
            theme="primary"
            :disabled="row.role_type === 1"
            @click="row.role_type === 1 ? undefined : handleAssignMenu(row)"
            style="margin-right: 8px"
          >分配菜单权限</t-link>
          <t-popconfirm
            :content="row.role_type >= 1 ? '内置角色不可删除' : '确认删除该角色？'"
            :disabled="row.role_type >= 1"
            @confirm="handleDelete(row)"
          >
            <t-link
              v-permission="'system:role:delete'"
              :theme="row.role_type >= 1 ? 'default' : 'danger'"
              :disabled="row.role_type >= 1"
            >删除</t-link>
          </t-popconfirm>
        </template>
      </t-table>
    </t-card>

    <!-- 新增/编辑角色对话框 -->
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

    <!-- 分配菜单权限对话框 -->
    <t-dialog
      v-model:visible="menuDialogVisible"
      :title="`分配菜单权限 - ${currentRole?.role_name || ''}`"
      width="500px"
      :confirm-btn="{ content: '保存', loading: menuSubmitting }"
      @confirm="handleMenuSubmit"
      @cancel="menuDialogVisible = false"
     :close-on-overlay-click="false">
      <div v-if="menuLoading" style="text-align:center;padding:40px">加载中...</div>
      <t-tree
        v-else
        ref="menuTreeRef"
        :data="allMenuTree"
        :keys="{ label: 'menu_name', value: 'id', children: 'children' }"
        :value="checkedMenuIds"
        :checkable="true"
        :check-strictly="false"
        :expand-all="false"
        :hover="true"
        :activatable="true"
        @change="handleMenuCheckChange"
      />
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref } from 'vue';
import type { TreeOptionData } from 'tdesign-vue-next';

import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';

import {
  getRoleList,
  createRole,
  updateRole,
  deleteRole,
  type RoleItem,
  type CreateRoleParams,
  type UpdateRoleParams,
} from '@/api/role-management';
import { getRoleMenuTree, assignRoleMenus, getRoleMenus, type MenuItem } from '@/api/menu';
import { formatDateTime } from '@/utils/date';

defineOptions({ name: 'SystemRole' });

const loading = ref(false);
const roleList = ref<RoleItem[]>([]);
const roleDialogVisible = ref(false);
const roleSubmitting = ref(false);
const isEdit = ref(false);
const roleFormRef = ref();

// 分配菜单对话框
const menuDialogVisible = ref(false);
const menuLoading = ref(false);
const menuSubmitting = ref(false);
const menuTreeRef = ref();
const allMenuTree = ref<MenuItem[]>([]);
const checkedMenuIds = ref<(string | number)[]>([]);
const currentRole = ref<RoleItem | null>(null);

// 搜索
const formData = reactive({ role_name: '' });

const columns = [
  { colKey: 'role_name', title: '角色名称', width: 160, fixed: 'left' as const },
  { colKey: 'remark', title: '备注', width: 200 },
  { colKey: 'role_type', title: '类型', width: 100 },
  { colKey: 'data_scope', title: '数据权限', width: 130 },
  { colKey: 'status', title: '状态', width: 80 },
  { colKey: 'create_time', title: '创建时间', width: 180 },
  { colKey: 'operation', title: '操作', width: 260, fixed: 'right' as const },
];

// 角色表单
const roleFormModel = reactive<CreateRoleParams & { id?: number; data_scope?: number; status?: number }>({
  role_name: '',
  remark: '',
  data_scope: 1,
  status: 1,
});

const roleFormRules = {
  role_name: [{ required: true, message: '请输入角色名称', trigger: 'blur' as const }],
};

const roleDialogTitle = computed(() => (isEdit.value ? '编辑角色' : '新增角色'));

// 筛选栏
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'role_name', placeholder: '请输入角色名称' },
];

const fetchRoleList = async () => {
  loading.value = true;
  try {
    const res = await getRoleList();
    roleList.value = res.list || [];
  } catch {
    MessagePlugin.error('获取角色列表失败');
  } finally {
    loading.value = false;
  }
};

const fetchMenuTree = async () => {
  try {
    const res = await getRoleMenuTree();
    allMenuTree.value = res.list || [];
  } catch {
    MessagePlugin.error('获取菜单树失败');
  }
};

const handleSearch = () => { fetchRoleList(); };
const handleReset = () => { fetchRoleList(); };

// 新增角色
const handleAdd = () => {
  isEdit.value = false;
  Object.assign(roleFormModel, { id: undefined, role_name: '', remark: '', data_scope: 1, status: 1 });
  roleFormRef.value?.reset();
  roleDialogVisible.value = true;
};

// 编辑角色
const handleEdit = (row: RoleItem) => {
  isEdit.value = true;
  Object.assign(roleFormModel, {
    id: row.id,
    role_name: row.role_name,
    remark: row.remark,
    data_scope: row.data_scope || 1,
    status: row.status,
  });
  roleDialogVisible.value = true;
};

const handleDelete = async (row: RoleItem) => {
  try {
    await deleteRole(row.id);
    MessagePlugin.success('删除成功');
    fetchRoleList();
  } catch {
    MessagePlugin.error('删除失败');
  }
};

// 提交角色
const handleRoleSubmit = async () => {
  const valid = await roleFormRef.value?.validate();
  if (valid !== true) return;
  roleSubmitting.value = true;
  try {
    if (isEdit.value && roleFormModel.id) {
      await updateRole({
        id: roleFormModel.id,
        role_name: roleFormModel.role_name,
        remark: roleFormModel.remark,
        data_scope: roleFormModel.data_scope,
        status: roleFormModel.status,
      });
      MessagePlugin.success('更新成功');
    } else {
      await createRole({
        role_name: roleFormModel.role_name,
        remark: roleFormModel.remark,
      });
      MessagePlugin.success('创建成功');
    }
    roleDialogVisible.value = false;
    fetchRoleList();
  } catch {
    MessagePlugin.error(isEdit.value ? '更新失败' : '创建失败');
  } finally {
    roleSubmitting.value = false;
  }
};

// 分配菜单权限 - 打开对话框
const handleAssignMenu = async (row: RoleItem) => {
  currentRole.value = row;
  menuDialogVisible.value = true;
  menuLoading.value = true;

  try {
    // 并行获取菜单树和已选菜单
    const [menuRes, roleMenusRes] = await Promise.all([
      getRoleMenuTree(),
      getRoleMenus(row.id),
    ]);
    allMenuTree.value = menuRes.list || [];
    checkedMenuIds.value = roleMenusRes.menu_ids || [];
  } catch {
    MessagePlugin.error('获取菜单数据失败');
  } finally {
    menuLoading.value = false;
  }
};

// 菜单勾选变化
const handleMenuCheckChange = (value: (string | number)[]) => {
  checkedMenuIds.value = value;
};

/**
 * 补全选中节点的所有祖先ID
 * TDesign树 check-strictly=false 时，勾选父节点可能只返回叶子节点，
 * 需要手动补全中间的菜单级(type=2)节点，确保 :list 读权限不丢失
 */
function ensureAncestorIds(checked: number[], treeData: MenuItem[]): number[] {
  const idSet = new Set<number>(checked);
  const parentMap = new Map<number, number>();
  function buildMap(nodes: MenuItem[]) {
    for (const node of nodes) {
      if (node.children) {
        for (const child of node.children) {
          parentMap.set(child.id, node.id);
        }
        buildMap(node.children);
      }
    }
  }
  buildMap(treeData);
  // 对每个已选中的节点，向上追溯所有祖先
  for (const id of checked) {
    let pid = parentMap.get(id);
    while (pid && !idSet.has(pid)) {
      idSet.add(pid);
      pid = parentMap.get(pid);
    }
  }
  return Array.from(idSet);
}

// 提交菜单分配
const handleMenuSubmit = async () => {
  if (!currentRole.value) return;
  menuSubmitting.value = true;
  try {
    let menuIds = checkedMenuIds.value.map(id => Number(id));
    // 自动补全祖先菜单ID，确保 :list 权限不被遗漏
    menuIds = ensureAncestorIds(menuIds, allMenuTree.value);
    await assignRoleMenus({ role_id: currentRole.value.id, menu_ids: menuIds });
    MessagePlugin.success('菜单权限分配成功');
    menuDialogVisible.value = false;
  } catch {
    MessagePlugin.error('分配失败');
  } finally {
    menuSubmitting.value = false;
  }
};

onMounted(() => {
  fetchRoleList();
  fetchMenuTree();
});
</script>

<style lang="less" scoped>
.role-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;

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
</style>
