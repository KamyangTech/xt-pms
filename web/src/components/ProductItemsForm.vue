<template>
  <div class="product-items-form">
    <!-- 客户选择器（独立生产订单使用） -->
    <t-form-item v-if="showCustomerSelector" label="客户" :name="customerFieldName">
      <t-select
        v-model="selectedCustomerId"
        placeholder="请选择客户"
        clearable
        filterable
        @change="handleCustomerChange"
      >
        <t-option
          v-for="customer in customerList"
          :key="customer.id"
          :value="customer.id"
          :label="customer.name"
        />
      </t-select>
    </t-form-item>

    <!-- 模具来源选择 -->
    <t-form-item v-if="showMoldSourceSelector" label="模具来源" :name="moldSourceFieldName">
      <t-select 
        v-model="moldSourceType" 
        placeholder="请选择模具来源"
        clearable
        :options="moldSourceOptions"
        @change="handleMoldSourceChange"
      />
    </t-form-item>

    <!-- 产品配套明细表格 -->
    <div class="items-header">
      <span>产品配套明细</span>
      <t-button 
        v-if="showAddButton" 
        theme="primary" 
        variant="outline" 
        size="small" 
        @click="handleAddItem"
      >
        <template #icon><add-icon /></template>
        添加产品
      </t-button>
    </div>

    <t-table
      v-if="items.length > 0"
      :data="items"
      :columns="columns"
      row-key="id"
      bordered
    >
      <template #product_name="{ row, rowIndex }">
        <t-input
          v-model="row.product_name"
          :placeholder="getProductNamePlaceholder(row)"
          :disabled="isDisabled || isProductNameDisabled(row)"
          :class="{ 'readonly-input': isProductNameDisabled(row) }"
        />
      </template>

      <template #mold_selector="{ row, rowIndex }">
        <!-- ✨ 修改：单选模式，一个产品只绑定一套模具 -->
        <t-select
          v-if="moldSourceType === 'SELF_MADE'"
          v-model="row.mold_id"
          :placeholder="row.mold_code ? '可更换模具' : '请选择模具'"
          clearable
          filterable
          :disabled="isDisabled"
          @change="(value: any) => handleMoldChange(value, row, rowIndex)"
        >
          <t-option
            v-for="mold in availableMoldList"
            :key="mold.id"
            :value="mold.id"
            :label="mold.label || `${mold.mold_code} - ${mold.mold_name || ''}`"
          />
        </t-select>
        <t-input
          v-else
          v-model="row.external_mold_code"
          placeholder="请输入外来模具编号"
          clearable
          :disabled="isDisabled"
          @blur="() => handleExternalMoldCodeBlur(row, rowIndex)"
        />
      </template>

      <template #plan_qty="{ row, rowIndex }">
        <t-input 
          v-model="row.plan_qty" 
          type="number"
          :min="1"
          placeholder="计划数量" 
          :disabled="isDisabled"
        />
      </template>

      <template #product_image="{ row, rowIndex }" v-if="showImageUpload">
        <FileUploader
          :ref="(el: any) => setProductImageRef(el, rowIndex)"
          :model-value="row.product_image_url ? [{ id: '', key: `img-${rowIndex}`, url: row.product_image_url, fileName: '', fileSize: 0, fileType: '.jpg' }] : []"
          :auto-save="false"
          :max-count="1"
          :multiple="false"
          image-only
          :show-empty-hint="false"
          :show-actions="!isDisabled"
          :disabled="isDisabled"
          inline
          category="product-images"
          upload-btn-text=""
          @update:model-value="(files: any[]) => {
            row.product_image_url = files.length > 0 ? files[0].url : '';
          }"
        />
      </template>

      <template #plastic_material_id="{ row, rowIndex }" v-if="showPlasticMaterial">
        <t-select
          v-model="row.plastic_material_id"
          placeholder="请选择塑胶原料"
          clearable
          filterable
          :disabled="isDisabled"
          :options="plasticMaterialOptions"
        />
      </template>

      <template #process_params="{ row, rowIndex }">
        <t-input
          v-model="row.process_params"
          placeholder="请输入工艺参数"
          :disabled="isDisabled"
        />
      </template>

      <template #remark="{ row, rowIndex }">
        <t-textarea
          v-model="row.remark"
          placeholder="请输入备注"
          :autosize="{ minRows: 1, maxRows: 2 }"
          :disabled="isDisabled"
        />
      </template>

      <template #op="{ rowIndex }">
        <t-link 
          theme="danger" 
          @click="handleDeleteItem(rowIndex)" 
          :disabled="isDisabled || items.length <= 1"
        >
          删除
        </t-link>
      </template>
    </t-table>

    <div v-else class="empty-tip">
      <t-empty description="暂无产品配套明细，请点击“添加产品”按钮添加" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { AddIcon, MoveIcon } from 'tdesign-icons-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { getMoldList } from '@/api/reference';
import { useDictRaw } from '@/hooks/useDict';
import FileUploader from '@/components/common/FileUploader.vue';

const props = defineProps({
  modelValue: {
    type: Array as () => Array<{
      id?: string; // 唯一标识
      product_name: string;
      mold_id?: number; // 单模具（兼容旧数据）
      mold_ids?: number[]; // 多模具列表
      external_mold_code?: string;
      mold_code?: string; // 模具编号（带入的模具信息）
      mold_name?: string; // 模具名称（带入的模具信息）
      plan_qty: number;
      device_id?: number | string;
      product_image_url?: string;
      plastic_material_id?: number | string;
      process_params: string;
      remark: string;
    }>,
    default: () => [] as Array<{
      id?: string;
      product_name: string;
      mold_id?: number;
      mold_ids?: number[];
      external_mold_code?: string;
      mold_code?: string;
      mold_name?: string;
      plan_qty: number;
      device_id?: number | string;
      product_image_url?: string;
      plastic_material_id?: number | string;
      process_params: string;
      remark: string;
    }>,
  },
  // 是否显示客户选择器
  showCustomerSelector: {
    type: Boolean,
    default: false,
  },
  // 客户字段名
  customerFieldName: {
    type: String,
    default: 'customer_id',
  },
  // 预选客户ID（综合订单带入）
  prefilledCustomerId: {
    type: Number,
    default: undefined,
  },
  // 是否显示模具来源选择
  showMoldSourceSelector: {
    type: Boolean,
    default: true,
  },
  // 预选模具数据（综合订单直接带入）
  prefilledMoldItems: {
    type: Array as () => Array<{
      mold_id?: number;
      mold_code: string;
      mold_name: string;
      amount?: number;
      moldProcess?: number | string;
      moldType?: number | string;
      materialType?: number | string;
      remark?: string;
    }>,
    default: () => [] as Array<{
      mold_id?: number;
      mold_code: string;
      mold_name: string;
    }>,
  },
  // 模具来源字段名
  moldSourceFieldName: {
    type: String,
    default: 'mold_source_type',
  },
  // 是否显示模具选择器
  showMoldSelector: {
    type: Boolean,
    default: true,
  },
  // 是否显示图片上传
  showImageUpload: {
    type: Boolean,
    default: true,
  },
  // 是否显示塑胶原料
  showPlasticMaterial: {
    type: Boolean,
    default: true,
  },
  // 是否可编辑
  editable: {
    type: Boolean,
    default: true,
  },
  // 是否允许拖拽排序
  allowDrag: {
    type: Boolean,
    default: true,
  },
  // 是否允许删除
  allowDelete: {
    type: Boolean,
    default: true,
  },
  // 是否禁用
  disabled: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits<{
  (e: 'update:modelValue', value: any[]): void;
  (e: 'update:selectedCustomerId', value: number | undefined): void;
  (e: 'customer-change', value: number | undefined): void;
  (e: 'mold-source-type-change', value: string | undefined): void; // ✨ 添加模具来源变更事件
}>();

// 双向绑定
const items = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
});

// 选中的客户ID
const selectedCustomerId = computed({
  get: () => props.prefilledCustomerId,
  set: (value) => emit('update:selectedCustomerId', value),
});

// 模具来源
const moldSourceType = ref<string | undefined>(undefined);

// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: moldSourceOptions } = useDictRaw('mold-source', (items) =>
  (items as any[]).map((item: any) => ({
    label: item.dict_label,
    value: item.dict_value,
  })),
);
const { items: plasticMaterialOptions } = useDictRaw('plastic-material', (items) =>
  (items as any[]).map((item: any) => ({
    label: item.dict_label,
    value: item.dict_value,
  })),
);
const customerList = ref<any[]>([]);
const moldList = ref<any[]>([]);

// 可用模具列表（根据客户过滤）
const availableMoldList = computed(() => {
  // ✨ 综合订单模式：如果有带入的模具信息，优先使用
  if (props.prefilledMoldItems && props.prefilledMoldItems.length > 0) {
    return props.prefilledMoldItems.map((item, index) => ({
      id: item.mold_id || `temp_${index}`,
      mold_code: item.mold_code || '',
      mold_name: item.mold_name || '',
      customer_id: selectedCustomerId.value,
      label: `${item.mold_code || ''} - ${item.mold_name || ''}`.trim() || `模具 ${index + 1}`,
    }));
  }
  
  // ✨ 如果没有带入的模具信息（或独立生产订单模式）：从模具档案加载
  if (!selectedCustomerId.value) return [];
  return moldList.value.filter((mold: any) => mold.customer_id === selectedCustomerId.value);
});

// 是否禁用
const isDisabled = computed(() => props.disabled || !props.editable);

// 是否显示添加按钮
const showAddButton = computed(() => props.editable);

// 产品图片组件引用
const productImageRefs = ref<Map<number, any>>(new Map());

// 表格列配置
const columns = computed(() => {
  const baseColumns = [
    { colKey: 'product_name', title: '产品名称', width: 200 },
  ];

  if (props.showMoldSelector) {
    baseColumns.push({ colKey: 'mold_selector', title: '模具', width: 280 });
  }

  baseColumns.push(
    { colKey: 'plan_qty', title: '计划数量', width: 150 }
  );

  if (props.showImageUpload) {
    baseColumns.push({ colKey: 'product_image', title: '产品图片', width: 160 });
  }

  if (props.showPlasticMaterial) {
    baseColumns.push({ colKey: 'plastic_material_id', title: '塑胶原料', width: 150 });
  }

  baseColumns.push(
    { colKey: 'process_params', title: '工艺参数', width: 200 },
    { colKey: 'remark', title: '备注', width: 200 },
    { colKey: 'op', title: '操作', width: 80 }
  );

  return baseColumns;
});

const getProductNamePlaceholder = (row: any) => {
  // ✨ 综合订单模式（自制模具）：用户手动输入产品名称
  if (moldSourceType.value === 'SELF_MADE') {
    // ✨ 修改：根据 mold_id 判断是否已选择模具（单选模式）
    if (row.mold_id) {
      return '请输入产品名称';
    }
    // 如果没有模具信息，提示先选择模具
    return '请先选择模具';
  }
  
  // 客供模具模式：用户手动输入产品名称
  if (moldSourceType.value === 'CUSTOMER_PROVIDED') {
    return '请输入产品名称';
  }
  
  // 默认情况
  return '请先选择模具来源';
};

const isProductNameDisabled = (row: any) => {
  // ✨ 综合订单模式：产品名称始终可以编辑（由用户手动填写）
  // 不再根据 mold_id 禁用产品名称输入
  return false;
};

const setProductImageRef = (el: any, index: number) => {
  if (el) {
    productImageRefs.value.set(index, el);
  }
};

// 加载字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）

const loadCustomerList = async () => {
  try {
    // TODO: 调用实际 API
    // const res = await getCustomerList({ page: 1, page_size: 1000 });
    // customerList.value = res.list || [];
  } catch (error) {
    console.error('获取客户列表失败:', error);
  }
};

const loadMoldList = async () => {
  try {
    const res = await getMoldList({ page: 1, page_size: 1000 });
    moldList.value = res.list || [];
  } catch (error) {
    console.error('获取模具列表失败:', error);
  }
};

// 客户变更
const handleCustomerChange = (value: any) => {
  emit('customer-change', value as number | undefined);
};

// 模具来源变更
const handleMoldSourceChange = (value: any) => {
  // 综合订单模式下，不允许更改模具来源
  if (!props.showMoldSourceSelector) {
    return;
  }
  
  // ✨ 通知父组件模具来源变更
  emit('mold-source-type-change', value as string | undefined);
  
  items.value.forEach((item) => {
    // 只有当没有带入的模具信息时，才清空模具选择
    if (!item.mold_code) {
      item.mold_id = undefined;
      item.external_mold_code = undefined;
    }
    if (value === 'SELF_MADE' && !item.mold_code) {
      item.product_name = '';
    }
  });
};

// 模具选择变更（支持多选）
const handleMoldChange = (value: any, item: any, index: number) => {
  const moldIds = Array.isArray(value) ? value : (value ? [value] : []);

  if (moldIds.length > 0) {
    // 查找选中的模具信息并保存到 item 上（用 mold_code_display 避免被 watch 清除）
    const selectedMoldId = moldIds[0];
    const selectedMold = availableMoldList.value.find(m => m.id === selectedMoldId);
    if (selectedMold) {
      item.mold_code_display = selectedMold.mold_code;
      item.mold_name_display = selectedMold.mold_name;
    }
  } else {
    item.mold_code_display = '';
    item.mold_name_display = '';
  }
};

// 外来模具编号失焦
const handleExternalMoldCodeBlur = async (item: any, index: number) => {
  if (!item.external_mold_code) return;

  try {
    // TODO: 调用 API 查询外来模具档案
    // const res = await getExternalMoldArchive(item.external_mold_code);
    // if (res) {
    //   item.product_name = res.product_name || '';
    // }
  } catch (error) {
    console.error('查询外来模具档案失败:', error);
  }
};

const handleAddItem = () => {
  let newItem: any = {
    id: `item_${Date.now()}`, // 添加唯一标识
    product_name: '',
    mold_id: undefined,
    mold_ids: [], // 多模具列表
    external_mold_code: undefined,
    plan_qty: undefined,
    device_id: undefined,
    product_image_url: '',
    plastic_material_id: undefined,
    process_params: '',
    remark: '',
  };
  
  // ✨ 综合订单模式下，自动绑定模具信息
  if (props.prefilledMoldItems && props.prefilledMoldItems.length > 0) {    
    // 如果只有 1 套模具，自动绑定
    if (props.prefilledMoldItems.length === 1) {
      const moldItem = props.prefilledMoldItems[0];
      
      // ✨ 只使用有效的数字模具ID
      const validMoldId = moldItem.mold_id && !isNaN(Number(moldItem.mold_id)) 
        ? Number(moldItem.mold_id) 
        : undefined;
      
      newItem.mold_code = moldItem.mold_code || '';
      newItem.mold_name = moldItem.mold_name || '';
      newItem.mold_id = validMoldId;
      newItem.mold_ids = validMoldId ? [validMoldId] : [];
    } else {
      // 如果有多套模具，初始化为空，让用户选择
      newItem.mold_code = '';
      newItem.mold_name = '';
      newItem.mold_id = undefined;
      newItem.mold_ids = [];
    }
  }
  
  items.value.push(newItem);
};

const handleDeleteItem = (index: number) => {
  if (items.value.length <= 1) {
    MessagePlugin.warning('至少保留一项产品明细');
    return;
  }
  items.value.splice(index, 1);
};

// 监听模具选择变化，清除带入的模具信息
watch(
  () => items.value,
  (newItems) => {
    newItems.forEach((item) => {
      // 如果用户选择了系统模具（mold_id），则清除带入的模具信息
      if (item.mold_id && item.mold_code) {
        item.mold_code = undefined;
        item.mold_name = undefined;
      }
    });
  },
  { deep: true }
);

onMounted(() => {
  // 字典已通过 useDictRaw 自动加载
  if (props.showCustomerSelector) {
    loadCustomerList();
  }
  loadMoldList();
});

// 监听模具信息变化，自动设置模具来源
watch(
  () => [items.value, props.prefilledMoldItems, props.showMoldSourceSelector],
  () => {
    // 综合订单模式：如果有带入的模具信息，自动设置模具来源为自制
    if (!props.showMoldSourceSelector && props.prefilledMoldItems && props.prefilledMoldItems.length > 0) {
      moldSourceType.value = 'SELF_MADE';
      // ✨ 通知父组件模具来源变更
      emit('mold-source-type-change', 'SELF_MADE');
    } else if (items.value.length > 0 && items.value[0].mold_code) {
      // 其他模式：如果有带入的模具信息，设置模具来源为自带模具
      moldSourceType.value = 'SELF_MADE';
      // ✨ 通知父组件模具来源变更
      emit('mold-source-type-change', 'SELF_MADE');
    }
  },
  { immediate: true, deep: true }
);

// 暴露方法
defineExpose({
  items,
  productImageRefs,
});
</script>

<style scoped lang="less">
.product-items-form {
  width: 100%;
}

.items-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.empty-tip {
  padding: 40px 0;
  text-align: center;
}

.readonly-input {
  :deep(.t-input) {
    background-color: var(--td-bg-color-secondarycontainer);
    cursor: not-allowed;
  }
}

.amount-text {
  font-weight: 600;
  color: var(--td-text-color-primary);
}

</style>
