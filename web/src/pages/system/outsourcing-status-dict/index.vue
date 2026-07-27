<template>
  <div class="outsourcing-status-dict-container">
    <!-- 操作栏卡片 -->
    <t-card class="toolbar-card">
      <div class="page-toolbar">
        <span class="toolbar-title">外发状态字典管理</span>
        <div class="toolbar-actions">
          <t-space>
            <t-button theme="default" @click="handleRefresh">
              <template #icon><refresh-icon /></template>
              刷新
            </t-button>
          </t-space>
        </div>
      </div>
    </t-card>

    <!-- 状态列表 -->
    <t-card :bordered="false">
      <div class="tab-content">
        <t-alert theme="info" message="系统内置状态，仅支持修改颜色和备注" style="margin-bottom: 16px;" />
        <t-table
          :data="dictDataList"
          :columns="dataColumns"
          row-key="id"
          :loading="loading"
          :stripe="true"
        >
          <template #color="{ row }">
            <t-tag 
              :style="{ 
                background: getDictColorHex(row.color) + ' !important', 
                color: (isLightColor(row.color) ? 'rgba(0, 0, 0, 0.9)' : '#fff') + ' !important', 
                borderColor: getDictColorHex(row.color) + ' !important'
              }"
            >
              {{ getDictColorName(row.color) }}
            </t-tag>
          </template>
          <template #operation="{ row }">
            <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
          </template>
        </t-table>
      </div>
    </t-card>

    <!-- 编辑对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      header="编辑状态显示属性"
      width="520px"
      placement="center"
      :confirm-btn="{ content: '确定', theme: 'primary' }"
      :cancel-btn="{ content: '取消', theme: 'default' }"
      @confirm="handleSubmit"
     :close-on-overlay-click="false">
      <t-form ref="formRef" :data="formModel" :rules="formRules" label-width="80px">
        <!-- 只读信息卡片 -->
        <t-card :bordered="true" :shadow="false" style="margin-bottom: 24px; background-color: var(--td-bg-color-secondarycontainer);">
          <t-row :gutter="[16, 16]">
            <t-col :span="8">
              <div class="info-item">
                <div class="info-label">状态名称</div>
                <div class="info-value">{{ formModel.dictLabel }}</div>
              </div>
            </t-col>
            <t-col :span="16">
              <div class="info-item">
                <div class="info-label">字典值</div>
                <div class="info-value">
                  <t-tag variant="light" theme="primary" style="font-size: 16px; font-weight: 600;">{{ formModel.dictValue }}</t-tag>
                </div>
              </div>
            </t-col>
          </t-row>
        </t-card>

        <t-form-item label="颜色" name="color">
          <t-select
            v-model="formModel.color"
            placeholder="请选择状态颜色"
            clearable
          >
            <template #valueDisplay="{ value }">
              <div v-if="value" class="color-selected">
                <span class="color-swatch" :style="{ background: getDictColorHex(value) }" />
                <span class="color-name">{{ getDictColorName(value) }}</span>
                <t-tag size="small" variant="light" theme="default" class="color-hex">{{ getDictColorHex(value) }}</t-tag>
              </div>
              <span v-else class="placeholder-text">请选择颜色</span>
            </template>

            <t-option-group
              v-for="group in colorOptionGroups"
              :key="group.group"
              :label="group.group"
            >
              <t-option
                v-for="opt in group.options"
                :key="opt.value"
                :value="opt.value"
                class="color-option"
              >
                <div class="color-option-content">
                  <span class="color-swatch-option" :style="{ background: opt.hex }" />
                  <span class="color-option-label">{{ opt.label.split('（')[0] }}</span>
                  <t-tag size="small" variant="light" theme="default" class="color-hex-tag">{{ opt.hex }}</t-tag>
                </div>
              </t-option>
            </t-option-group>
          </t-select>
        </t-form-item>

        <t-form-item label="备注" name="remark">
          <t-textarea 
            v-model="formModel.remark" 
            placeholder="请输入备注信息（可选）" 
            :rows="3"
            :maxlength="200"
            :show-limit="true"
            clearable
          />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { TableProps } from 'tdesign-vue-next';
import { RefreshIcon } from 'tdesign-icons-vue-next';
import {
  getDictDataByType,
  updateDictData,
  clearDictCache,
} from '@/api/dict';
import {
  getDictColorHex,
  getDictColorName,
  isLightColor,
  getAllColorOptions,
} from '@/utils/dict-color';

defineOptions({
  name: 'OutsourcingStatusDict',
});

// 表格相关
const loading = ref(false);
const dictDataList = ref<any[]>([]);

// 编辑对话框相关
const dialogVisible = ref(false);
const formRef = ref();

// 颜色选项
const colorOptionGroups = getAllColorOptions();

const formModel = reactive({
  id: undefined as number | undefined,
  dictLabel: '',
  dictValue: '',
  color: 'default',
  remark: '',
});

const formRules = {};

// 表格列配置
const dataColumns: TableProps['columns'] = [
  { colKey: 'sort', title: '排序', width: 80 },
  { colKey: 'dictLabel', title: '状态名称', width: 100 },
  { colKey: 'dictValue', title: '字典值', width: 150 },
  { colKey: 'color', title: '颜色', width: 120, cell: 'color' as any },
  { colKey: 'remark', title: '备注', width: 200, ellipsis: true },
  { colKey: 'operation', title: '操作', width: 120, fixed: 'right' as const },
];

const loadDictData = async () => {
  loading.value = true;
  try {
    const res = await getDictDataByType({ dict_type: 'outsourcing-status' });
    dictDataList.value = (res.list || []).map((item: any) => ({
      id: item.id,
      dictLabel: item.dict_label,
      dictValue: item.dict_value,
      color: item.color,
      sort: item.sort,
      remark: item.remark,
    }));
  } catch (error) {
    console.error('加载字典数据失败:', error);
    MessagePlugin.error('加载字典数据失败');
  } finally {
    loading.value = false;
  }
};

// 编辑（只允许编辑颜色和备注）
const handleEdit = (row: any) => {
  Object.assign(formModel, {
    id: row.id,
    dictLabel: row.dictLabel,
    dictValue: row.dictValue,
    color: row.color || 'default',
    remark: row.remark || '',
  });
  setTimeout(() => {
    formRef.value?.clearValidate?.();
  }, 100);
  dialogVisible.value = true;
};

// 提交表单（只提交 color 和 remark）
const handleSubmit = async () => {
  if (!formModel.id) {
    MessagePlugin.error('未找到要编辑的状态');
    return;
  }

  try {
    const submitData: any = {
      id: Number(formModel.id),
      color: formModel.color,
      remark: formModel.remark,
    };

    await updateDictData(submitData);
    MessagePlugin.success('更新成功');
    dialogVisible.value = false;
    loadDictData();
  } catch (error) {
    MessagePlugin.error('更新失败');
  }
};

// 刷新
const handleRefresh = async () => {
  try {
    await clearDictCache();
  } catch {
    // 清除缓存失败不影响重新加载
  }
  loadDictData();
};

onMounted(async () => {
  loadDictData();
});
</script>

<style scoped>
.outsourcing-status-dict-container {
  padding: 16px;
}

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

/* 编辑弹窗信息卡片样式 */
.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-label {
  font-size: 12px;
  color: var(--td-text-color-placeholder);
  line-height: 1.5;
}

.info-value {
  font-size: 14px;
  color: var(--td-text-color-primary);
  line-height: 1.5;
  display: flex;
  align-items: center;
}

/* 颜色选择器选项样式 */
.color-option {
  padding: 6px 8px;
}

.color-option-content {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.color-option-label {
  flex: 1;
  font-size: 14px;
}

.color-hex-tag {
  opacity: 0.6;
  font-size: 11px !important;
}

.color-swatch {
  display: inline-block;
  width: 18px;
  height: 18px;
  border-radius: var(--td-radius-small);
  border: 1px solid rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
}

.color-swatch-option {
  display: inline-block;
  width: 20px;
  height: 20px;
  border-radius: var(--td-radius-small);
  border: 1px solid rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
}

.color-selected {
  display: flex;
  align-items: center;
  gap: 8px;
}

.color-name {
  font-size: 14px;
  font-weight: 500;
}

.color-hex {
  opacity: 0.5;
  font-size: 11px !important;
  margin-left: auto;
}

.placeholder-text {
  color: var(--td-text-color-placeholder);
}

:deep(.t-card__body) {
  padding: 16px;
}
</style>
