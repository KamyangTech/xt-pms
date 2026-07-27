<template>
  <div class="menu-container">
    <!-- 工具栏 -->
    <t-card class="toolbar-card">
      <div class="toolbar-content">
        <span class="toolbar-title">菜单管理</span>
        <div class="toolbar-actions">
          <t-button theme="default" variant="outline" @click="handleToggleExpand">
            {{ isAllExpanded ? '全部折叠' : '全部展开' }}
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 菜单树表格 -->
    <t-card :bordered="false">
      <t-enhanced-table
        ref="menuTableRef"
        :data="tableData"
        :columns="columns"
        :loading="loading"
        row-key="id"
        :tree="{ childrenKey: 'children' }"
        v-model:expanded-tree-nodes="expandedTreeNodes"
        :stripe="true"
        hover
        size="medium"
      >
        <template #menu_type="{ row }">
          <t-tag v-if="row.menu_type === 1" theme="default">目录</t-tag>
          <t-tag v-else-if="row.menu_type === 2" theme="primary">菜单</t-tag>
          <t-tag v-else-if="row.menu_type === 3" theme="warning">按钮</t-tag>
        </template>
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'danger'" variant="light">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </t-tag>
        </template>
        <template #operation="{ row }">
          <t-link v-permission="'system:menu:create'" theme="primary" @click="handleAddChild(row)" style="margin-right: 8px">添加子菜单</t-link>
          <t-link v-permission="'system:menu:update'" theme="primary" @click="handleEdit(row)" style="margin-right: 8px">编辑</t-link>
          <t-popconfirm content="确认删除该菜单？子菜单将被保留但失去父级" @confirm="handleDelete(row)">
            <t-link v-permission="'system:menu:delete'" theme="danger">删除</t-link>
          </t-popconfirm>
        </template>
      </t-enhanced-table>
    </t-card>

    <!-- 新增/编辑对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :title="dialogTitle"
      width="550px"
      placement="center"
      :confirm-btn="{ content: '确定', loading: submitting }"
      @confirm="handleSubmit"
      @cancel="dialogVisible = false"
     :close-on-overlay-click="false">
      <t-form ref="formRef" :data="formModel" :rules="formRules" label-width="100px">
        <t-form-item label="上级菜单" name="parent_id">
          <t-tree-select
            v-model="formModel.parent_id"
            :data="menuTreeOptions"
            placeholder="不选则为根菜单"
            clearable
            filterable
            value-mode="id"
            :keys="{ label: 'menu_name', value: 'id', children: 'children' }"
          />
        </t-form-item>
        <t-form-item label="菜单名称" name="menu_name">
          <t-input v-model="formModel.menu_name" placeholder="请输入菜单名称" />
        </t-form-item>
        <t-form-item label="菜单类型" name="menu_type">
          <t-radio-group v-model="formModel.menu_type">
            <t-radio-button :value="1">目录</t-radio-button>
            <t-radio-button :value="2">菜单</t-radio-button>
            <t-radio-button :value="3">按钮</t-radio-button>
          </t-radio-group>
        </t-form-item>
        <t-form-item label="权限标识" name="permission">
          <t-input v-model="formModel.permission" placeholder="如 system:user:list" />
        </t-form-item>
        <t-form-item v-if="formModel.menu_type !== 3" label="路由路径" name="path">
          <t-input v-model="formModel.path" placeholder="如 /system/user" />
        </t-form-item>
        <t-form-item v-if="formModel.menu_type !== 3" label="组件路径" name="component">
          <t-input v-model="formModel.component" placeholder="如 system/user/index" />
        </t-form-item>
        <t-form-item v-if="formModel.menu_type !== 3" label="图标" name="icon">
          <t-input v-model="formModel.icon" placeholder="如 setting" />
        </t-form-item>
        <t-form-item label="排序" name="sort">
          <t-input-number v-model="formModel.sort" :min="0" :max="999" />
        </t-form-item>
        <t-form-item label="状态" name="status">
          <t-radio-group v-model="formModel.status">
            <t-radio-button :value="1">启用</t-radio-button>
            <t-radio-button :value="0">禁用</t-radio-button>
          </t-radio-group>
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref } from 'vue';

import {
  createMenu,
  deleteMenu,
  getMenuTree,
  updateMenu,
  type CreateMenuParams,
  type MenuItem,
  type UpdateMenuParams,
} from '@/api/menu';

defineOptions({ name: 'SystemMenu' });

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref();
const menuTableRef = ref();
const tableData = ref<MenuItem[]>([]);
const rawMenuList = ref<MenuItem[]>([]);
const expandedTreeNodes = ref<(string | number)[]>([]);

const columns = [
  { colKey: 'menu_name', title: '菜单名称', width: 250, fixed: 'left' as const },
  { colKey: 'permission', title: '权限标识', width: 200 },
  { colKey: 'menu_type', title: '类型', width: 100 },
  { colKey: 'icon', title: '图标', width: 100 },
  { colKey: 'sort', title: '排序', width: 80 },
  { colKey: 'status', title: '状态', width: 80 },
  { colKey: 'operation', title: '操作', width: 280, fixed: 'right' as const },
];

// 表单数据
const formModel = reactive<CreateMenuParams & { id?: number }>({
  parent_id: undefined,
  menu_name: '',
  menu_type: 2,
  permission: '',
  path: '',
  component: '',
  icon: '',
  sort: 0,
  status: 1,
});

const formRules = {
  menu_name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' as const }],
  status: [{ required: true, message: '请选择状态', trigger: 'blur' as const }],
};

const dialogTitle = computed(() => (isEdit.value ? '编辑菜单' : '新增子菜单'));

// 扁平化菜单列表转为树选择器选项
const menuTreeOptions = computed(() => rawMenuList.value);

const fetchMenuTree = async () => {
  loading.value = true;
  try {
    const res = await getMenuTree();
    rawMenuList.value = res.list || [];
    tableData.value = res.list || [];
  } catch {
    MessagePlugin.error('获取菜单列表失败');
  } finally {
    loading.value = false;
  }
};

// 获取所有有子节点的菜单ID（用于判断展开状态）
const getParentIds = (menus: MenuItem[]): number[] => {
  const ids: number[] = [];
  const collect = (items: MenuItem[]) => {
    for (const item of items) {
      if (item.children && item.children.length > 0) {
        ids.push(item.id);
        collect(item.children);
      }
    }
  };
  collect(menus);
  return ids;
};

const isAllExpanded = computed(() => {
  const allParentIds = getParentIds(tableData.value);
  return allParentIds.length > 0 && allParentIds.every(id => expandedTreeNodes.value.includes(id));
});

// 展开/折叠切换
const handleToggleExpand = () => {
  if (isAllExpanded.value) {
    // 当前是全部展开，执行全部折叠
    menuTableRef.value?.foldAll();
  } else {
    // 当前不是全部展开，执行全部展开
    menuTableRef.value?.expandAll();
  }
};

// 新增子菜单
const handleAddChild = (row: MenuItem) => {
  isEdit.value = false;
  resetForm();
  formModel.parent_id = row.id;
  dialogVisible.value = true;
};

// 编辑
const handleEdit = (row: MenuItem) => {
  isEdit.value = true;
  Object.assign(formModel, {
    id: row.id,
    parent_id: row.parent_id || undefined,
    menu_name: row.menu_name,
    menu_type: row.menu_type,
    permission: row.permission,
    path: row.path,
    component: row.component,
    icon: row.icon,
    sort: row.sort,
    status: row.status,
  });
  dialogVisible.value = true;
};

const handleDelete = async (row: MenuItem) => {
  try {
    await deleteMenu(row.id);
    MessagePlugin.success('删除成功');
    fetchMenuTree();
  } catch {
    MessagePlugin.error('删除失败');
  }
};

// 提交
const handleSubmit = async () => {
  const valid = await formRef.value?.validate();
  if (valid !== true) return;

  submitting.value = true;
  try {
    if (isEdit.value && formModel.id) {
      const data: UpdateMenuParams = {
        id: formModel.id,
        menu_name: formModel.menu_name,
        menu_type: formModel.menu_type,
        permission: formModel.permission,
        sort: formModel.sort,
        status: formModel.status,
      };
      if (formModel.parent_id) data.parent_id = formModel.parent_id;
      if (formModel.path) data.path = formModel.path;
      if (formModel.component) data.component = formModel.component;
      if (formModel.icon) data.icon = formModel.icon;
      await updateMenu(data);
      MessagePlugin.success('更新成功');
    } else {
      const data: CreateMenuParams = {
        parent_id: formModel.parent_id,
        menu_name: formModel.menu_name,
        menu_type: formModel.menu_type,
        permission: formModel.permission,
        path: formModel.path,
        component: formModel.component,
        icon: formModel.icon,
        sort: formModel.sort,
        status: formModel.status,
      };
      await createMenu(data);
      MessagePlugin.success('创建成功');
    }
    dialogVisible.value = false;
    fetchMenuTree();
  } catch {
    MessagePlugin.error(isEdit.value ? '更新失败' : '创建失败');
  } finally {
    submitting.value = false;
  }
};

const resetForm = () => {
  Object.assign(formModel, {
    id: undefined,
    parent_id: undefined,
    menu_name: '',
    menu_type: 2,
    permission: '',
    path: '',
    component: '',
    icon: '',
    sort: 0,
    status: 1,
  });
  formRef.value?.reset();
};

onMounted(() => {
  fetchMenuTree();
});
</script>

<style lang="less" scoped>
.menu-container {
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
