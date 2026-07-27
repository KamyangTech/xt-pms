<template>
  <div class="stock-out-form-container">
    <!-- 返回按钮 -->
    <div class="back-button-container">
      <t-button theme="default" variant="text" @click="handleBack">
        <template #icon><t-icon name="chevron-left" /></template>
        返回
      </t-button>
    </div>

    <t-form
      ref="formRef"
      :data="formData"
      :rules="rules"
      label-align="top"
      :label-width="100"
      @reset="onReset"
      @submit="onSubmit"
    >
      <div class="form-basic-container">
        <div class="form-basic-item">
          <!-- 基本信息 -->
          <div class="form-basic-container-title">
            {{ isEditMode ? '基础信息（不可修改）' : '基本信息' }}
          </div>
          <t-row class="row-gap" :gutter="[32, 24]">
            <!-- 编辑模式：显示基础信息 -->
            <template v-if="isEditMode">
              <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
                <t-form-item label="出库单号">
                  <t-input :value="stockOutData?.stock_code" disabled />
                </t-form-item>
              </t-col>
              <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
                <t-form-item label="模具编号">
                  <t-input :value="getMoldCode()" disabled />
                </t-form-item>
              </t-col>
              <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
                <t-form-item label="产品名称">
                  <t-input :value="stockOutData?.product_name" disabled />
                </t-form-item>
              </t-col>
            </template>

            <!-- 新建模式：自动填充模式 -->
            <template v-else-if="isAutoFillMode">
              <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
                <t-form-item label="模具编号">
                  <t-input :value="autoFilledMoldInfo" disabled />
                </t-form-item>
              </t-col>
              <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
                <t-form-item label="产品名称">
                  <t-input :value="autoFilledProductInfo" disabled />
                </t-form-item>
              </t-col>
            </template>

            <!-- 新建模式：手动填写模式 -->
            <template v-else>
              <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
                <t-form-item label="模具类型" name="moldType">
                  <t-radio-group v-model="formData.moldType" @change="fetchCurrentStock">
                    <t-radio-button value="internal">自有模具</t-radio-button>
                    <t-radio-button value="external">外来模具</t-radio-button>
                  </t-radio-group>
                </t-form-item>
              </t-col>

              <t-col
                v-if="formData.moldType === 'internal'"
                :span="12"
                :xs="12"
                :sm="8"
                :md="8"
                :lg="6"
                :xl="4"
              >
                <t-form-item label="选择模具" name="moldId">
                  <t-select
                    v-model="formData.moldId"
                    placeholder="请选择模具"
                    filterable
                    @change="handleMoldChange"
                  >
                    <t-option
                      v-for="mold in moldList"
                      :key="mold.id"
                      :label="`${mold.mold_code} - ${mold.mold_name}`"
                      :value="mold.id"
                    />
                  </t-select>
                </t-form-item>
              </t-col>

              <t-col v-else :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
                <t-form-item label="外来模具编号" name="externalMoldCode">
                  <t-input
                    v-model="formData.externalMoldCode"
                    placeholder="请输入外来模具编号"
                    @blur="fetchCurrentStock"
                  />
                </t-form-item>
              </t-col>

              <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
                <t-form-item label="选择产品" name="productId">
                  <t-select
                    v-if="!isAutoFillMode"
                    v-model="formData.productId"
                    placeholder="请选择产品"
                    filterable
                    @change="handleProductChange"
                  >
                    <t-option
                      v-for="(product, idx) in filteredProductList"
                      :key="`stock-${idx}`"
                      :label="product.product_name || product.pro_name"
                      :value="product.product_name || product.pro_name"
                    />
                  </t-select>
                  <t-input
                    v-else
                    :value="autoFilledProductInfo"
                    disabled
                    placeholder="无产品信息"
                  />
                </t-form-item>
              </t-col>
            </template>
          </t-row>

          <!-- 出库信息 -->
          <div v-if="!isEditMode" class="form-basic-container-title form-title-gap">出库信息</div>
          <t-row v-if="!isEditMode" class="row-gap" :gutter="[32, 24]">
            <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
              <t-form-item label="出库仓库" name="warehouseId">
                <t-select
                  v-model="formData.warehouseId"
                  placeholder="请选择仓库"
                  @change="fetchCurrentStock"
                >
                  <t-option
                    v-for="warehouse in warehouseList"
                    :key="warehouse.id"
                    :value="warehouse.id"
                    :label="getWarehouseLabel(warehouse)"
                  >
                    <div class="warehouse-option">
                      <span class="warehouse-option-name">{{ warehouse.wh_name }}</span>
                      <span class="warehouse-option-stock">
                        库存数量：{{ warehouseStockMap[warehouse.id] ?? '—' }}
                      </span>
                    </div>
                  </t-option>
                </t-select>
              </t-form-item>
            </t-col>

            <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
              <t-form-item label="发货数量" name="quantity" :help="stockHelpText">
                <t-input-number
                  v-model="formData.quantity"
                  :min="1"
                  :max="currentStock > 0 ? currentStock : undefined"
                  placeholder="请输入出库数量"
                  style="width: 100%"
                />
              </t-form-item>
            </t-col>

            <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
              <t-form-item label="批次号">
                <t-input v-model="formData.batchNo" placeholder="系统自动生成" disabled />
              </t-form-item>
            </t-col>
          </t-row>

          <!-- 收货信息 -->
          <div class="form-basic-container-title form-title-gap">收货信息</div>
          <t-row class="row-gap" :gutter="[32, 24]">
            <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
              <t-form-item label="关联订单">
                <t-select
                  v-model="formData.orderId"
                  placeholder="请选择关联订单（可选，不选则为直接出库）"
                  filterable
                  clearable
                  @change="handleOrderChange"
                >
                  <t-option
                    v-for="order in orderList"
                    :key="order.id"
                    :label="order.order_code"
                    :value="order.id"
                  />
                </t-select>
              </t-form-item>
            </t-col>

            <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
              <t-form-item label="接收客户" name="customerName">
                <t-select
                  v-model="formData.customerName"
                  placeholder="请选择或输入客户名称"
                  filterable
                  clearable
                  allow-create
                >
                  <t-option
                    v-for="(cus, idx) in customerList"
                    :key="`cus-${idx}`"
                    :label="cus"
                    :value="cus"
                  />
                </t-select>
              </t-form-item>
            </t-col>

            <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
              <t-form-item label="发货日期" name="deliveryDate">
                <t-date-picker
                  v-model="formData.deliveryDate"
                  theme="primary"
                  mode="date"
                  placeholder="请选择发货日期"
                  format="YYYY-MM-DD"
                  style="width: 100%"
                />
              </t-form-item>
            </t-col>
          </t-row>

          <!-- 发货方式 -->
          <div class="form-basic-container-title form-title-gap">发货方式</div>
          <t-row class="row-gap" :gutter="[32, 24]">
            <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
              <t-form-item label="发货方式" name="deliveryMethod">
                <t-select v-model="formData.deliveryMethod" placeholder="请选择发货方式" clearable>
                  <t-option
                    v-for="item in deliveryMethodDict"
                    :key="item.dict_value"
                    :label="item.dict_label"
                    :value="item.dict_value"
                  />
                </t-select>
              </t-form-item>
            </t-col>

            <t-col
              v-if="formData.deliveryMethod === 'EXPRESS'"
              :span="12"
              :xs="12"
              :sm="8"
              :md="8"
              :lg="6"
              :xl="4"
            >
              <t-form-item label="快递公司">
                <t-select v-model="formData.expressCompany" placeholder="请选择快递公司" clearable>
                  <t-option
                    v-for="item in expressCompanyDict"
                    :key="item.dict_value"
                    :label="item.dict_label"
                    :value="item.dict_value"
                  />
                </t-select>
              </t-form-item>
            </t-col>

            <t-col
              v-if="formData.deliveryMethod === 'EXPRESS'"
              :span="12"
              :xs="12"
              :sm="8"
              :md="8"
              :lg="6"
              :xl="4"
            >
              <t-form-item label="快递单号" help="仅做保存不校验，推荐复制单号/扫描输入">
                <t-input
                  v-model="formData.expressNo"
                  placeholder="请输入或扫描快递单号"
                  clearable
                  @blur="handleRecognizeExpress"
                />
              </t-form-item>
            </t-col>

            <t-col
              v-if="formData.deliveryMethod === 'LALAMOVE'"
              :span="12"
              :xs="12"
              :sm="8"
              :md="8"
              :lg="6"
              :xl="4"
            >
              <t-form-item label="司机电话">
                <t-input
                  v-model="formData.driverPhone"
                  placeholder="请输入司机联系电话（非必填）"
                />
              </t-form-item>
            </t-col>

            <t-col
              v-if="formData.deliveryMethod === 'LALAMOVE'"
              :span="12"
              :xs="12"
              :sm="8"
              :md="8"
              :lg="6"
              :xl="4"
            >
              <t-form-item label="车牌号">
                <t-input v-model="formData.driverPlate" placeholder="请输入车牌号（非必填）" />
              </t-form-item>
            </t-col>
          </t-row>

          <!-- 照片上传 -->
          <div class="form-basic-container-title form-title-gap">发货照片</div>
          <div style="margin-bottom: var(--td-comp-margin-xxl)">
            <file-uploader
              ref="photoUploaderRef"
              v-model="formData.photos"
              :auto-save="false"
              :image-only="true"
              :max-count="5"
              :multiple="true"
              category="stock-out"
              upload-btn-text="上传发货照片"
              show-empty-hint
              empty-text="暂无发货照片"
            />
          </div>

          <!-- 备注 -->
          <div class="form-basic-container-title form-title-gap">备注</div>
          <t-row class="row-gap" :gutter="[32, 24]">
            <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
              <t-form-item label="备注">
                <t-textarea
                  v-model="formData.remark"
                  placeholder="请输入备注（可选）"
                  :rows="3"
                  maxlength="500"
                />
              </t-form-item>
            </t-col>
          </t-row>
        </div>
      </div>

      <!-- 底部提交区域 -->
      <div class="form-submit-container">
        <div class="form-submit-sub">
          <div class="form-submit-left">
            <t-button theme="primary" :loading="submitting" type="submit">
              {{ isEditMode ? '保存' : '提交送货单' }}
            </t-button>
            <t-button theme="default" type="reset">取消</t-button>
          </div>
        </div>
      </div>
    </t-form>
  </div>
</template>
<script setup lang="ts">
import type { FormInstanceFunctions, FormRule } from 'tdesign-vue-next';
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next';
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { uploadFiles } from '@/api/common/upload';
import { getProductionOrders } from '@/api/production/order';
import {
  getCurrentStock,
  getProductStockList,
  getStockOutById,
  recognizeExpressCompany,
  stockOutOrder,
  updateStockOut,
} from '@/api/production/stock';
import { getMoldList, getWarehouseList } from '@/api/reference';
import FileUploader from '@/components/common/FileUploader.vue';
import { useDictRaw } from '@/hooks/useDict';
import { formatDate } from '@/utils/date';

defineOptions({
  name: 'StockOutForm',
});

const route = useRoute();
const router = useRouter();

const formRef = ref<FormInstanceFunctions>();
const photoUploaderRef = ref();
const submitting = ref(false);

// 表单模式：'create' 或 'edit'
const formMode = ref<'create' | 'edit'>('create');
const stockOutId = ref<number>();
const stockOutData = ref<any>(null);

// 是否为自动填充模式（从产品库存列表跳转过来）
const isAutoFillMode = ref(false);
const autoFilledMoldInfo = ref('');
const autoFilledProductInfo = ref('');
const isProductDisabled = ref(false); // 是否禁用产品选择器

const isEditMode = computed(() => formMode.value === 'edit');

// 计算属性：获取当前产品名称（productId 现在存的是产品名字符串）
const currentProductName = computed(() => {
  if (isAutoFillMode.value) {
    return autoFilledProductInfo.value;
  }
  if (typeof formData.value.productId === 'string') {
    return formData.value.productId;
  }
  return '';
});

const stockHelpText = computed(() => {
  const hasProduct = !!formData.value.productId;
  if (!hasProduct) return '';

  // 自动填充模式（从库存页跳转）：只要有产品就显示，不需要仓库
  const hasWarehouse = !!formData.value.warehouseId;
  if (!isAutoFillMode.value && !hasWarehouse) return '';

  if (stockLoading.value) return '正在查询库存...';
  const qty = formData.value.quantity || 0;
  const remain = currentStock.value - qty;
  if (qty > 0 && currentStock.value > 0) {
    return `当前库存: ${currentStock.value}｜预计发货后剩余: ${remain}`;
  }
  return `当前库存: ${currentStock.value}`;
});

// 表单数据（productId 现在存的是产品名字符串）
const formData = ref({
  moldType: 'internal',
  moldId: undefined as number | undefined,
  externalMoldCode: '',
  productId: undefined as string | undefined,
  orderId: undefined as number | undefined,
  warehouseId: undefined as number | undefined,
  quantity: undefined as number | undefined,
  batchNo: '',
  customerName: '',
  deliveryDate: '',
  deliveryMethod: '',
  expressCompany: '',
  expressNo: '',
  driverPhone: '',
  driverPlate: '',
  photos: [] as any[],
  remark: '',
});

// 下拉列表数据
const moldList = ref<any[]>([]);
const productList = ref<any[]>([]); // 全量产品列表（来自库存汇总）
const filteredProductList = ref<any[]>([]); // 根据模具过滤后的产品列表
const warehouseList = ref<any[]>([]);
const orderList = ref<any[]>([]);
const customerList = ref<string[]>([]); // 从订单中提取的唯一客户列表
// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: deliveryMethodDict } = useDictRaw('delivery_method');
const { items: expressCompanyDict } = useDictRaw('express_company');

// 当前库存数量
const currentStock = ref<number>(0);
const stockLoading = ref<boolean>(false);

// 每个仓库的库存数量缓存
const warehouseStockMap = ref<Record<number, number>>({});

// 获取仓库选项的展示标签（用于搜索过滤时也能匹配库存信息）
const getWarehouseLabel = (warehouse: any) => {
  const stock = warehouseStockMap.value[warehouse.id];
  return stock !== undefined ? `${warehouse.wh_name} - 库存数量：${stock}` : warehouse.wh_name;
};

// 表单验证规则
const rules: Record<string, FormRule[]> = {
  moldType: [
    {
      validator: () => {
        if (isEditMode.value || isAutoFillMode.value) return true;
        return !!formData.value.moldType;
      },
      message: '请选择模具类型',
      type: 'error',
    },
  ],
  moldId: [
    {
      validator: () => {
        if (isEditMode.value || isAutoFillMode.value) return true;
        if (formData.value.moldType === 'internal' && !formData.value.moldId) {
          return false;
        }
        return true;
      },
      message: '请选择模具',
      type: 'error',
    },
  ],
  externalMoldCode: [
    {
      validator: () => {
        if (isEditMode.value || isAutoFillMode.value) return true;
        if (formData.value.moldType === 'external' && !formData.value.externalMoldCode) {
          return false;
        }
        return true;
      },
      message: '请输入外来模具编号',
      type: 'error',
    },
  ],
  productId: [
    {
      validator: () => {
        if (isEditMode.value || isAutoFillMode.value) return true;
        return !!formData.value.productId;
      },
      message: '请选择产品',
      type: 'error',
    },
  ],
  warehouseId: [
    {
      validator: () => {
        if (isEditMode.value) return true;
        return !!formData.value.warehouseId;
      },
      message: '请选择仓库',
      type: 'error',
    },
  ],
  quantity: [
    {
      validator: () => {
        if (isEditMode.value) return true;
        return !!formData.value.quantity;
      },
      message: '请输入出库数量',
      type: 'error',
    },
    {
      validator: () => {
        if (isEditMode.value) return true;
        if (currentStock.value > 0 && formData.value.quantity > currentStock.value) {
          return false;
        }
        return true;
      },
      message: '出库数量不能超过当前库存',
      type: 'error',
    },
  ],
  customerName: [{ required: true, message: '请输入客户名称', type: 'error' }],
  deliveryDate: [{ required: true, message: '请选择发货日期', type: 'error' }],
  deliveryMethod: [{ required: true, message: '请选择发货方式', type: 'error' }],
};

const loadData = async () => {
  try {
    const [moldRes, stockProductRes, warehouseRes, orderRes] = await Promise.all([
      getMoldList({ page: 1, page_size: 1000 }),
      getProductStockList({ page: 1, page_size: 1000 }),
      getWarehouseList({ page: 1, page_size: 100 }),
      getProductionOrders({ page: 1, page_size: 100 }),
    ]);

    moldList.value = moldRes.list || [];
    warehouseList.value = warehouseRes.list || [];
    orderList.value = orderRes.list || [];

    // 从订单中提取唯一客户名称（用于客户下拉选择）
    const seenCustomers = new Set<string>();
    (orderRes.list || []).forEach((order: any) => {
      const name = order.customer?.cus_name || order.customer_name;
      if (name) {
        const trimmed = name.trim();
        if (!seenCustomers.has(trimmed)) {
          seenCustomers.add(trimmed);
        }
      }
    });
    customerList.value = Array.from(seenCustomers).sort(); // 按字母排序

    // 从库存汇总列表中提取唯一产品（与库存管理保持一致，按产品名称去重）
    const allStockProducts = stockProductRes.list || [];
    const seenNames = new Set<string>();
    productList.value = allStockProducts.filter((p: any) => {
      const name = p.product_name || p.pro_name;
      if (!name) return false;
      if (seenNames.has(name)) return false;
      seenNames.add(name);
      return true;
    });
    filteredProductList.value = productList.value;
  } catch (error) {
    console.error('加载数据失败:', error);
    MessagePlugin.error('加载数据失败');
  }
};

// 注：后端 getStockOutById 返回嵌套的 mold 对象（mold: { mold_code, mold_name, ...}），
//     同时也直接返回 external_mold_code 字段（外来模具时使用）。
const getMoldCode = () => {
  if (!stockOutData.value) return '-';
  return (
    stockOutData.value.mold?.mold_code ||
    stockOutData.value.mold_code ||
    stockOutData.value.external_mold_code ||
    '-'
  );
};

const parsePhotos = (photosStr: string) => {
  try {
    return photosStr ? JSON.parse(photosStr) : [];
  } catch {
    return [];
  }
};

const loadEditData = async () => {
  const id = route.params.id as string;
  if (!id) return;

  try {
    const res = await getStockOutById(Number(id));
    if (res) {
      stockOutData.value = res;
      stockOutId.value = Number(id);

      let parsedDeliveryDate = '';
      if (res.delivery_date) {
        const dateStr = res.delivery_date;
        if (dateStr.includes('T')) {
          parsedDeliveryDate = dateStr.split('T')[0];
        } else if (dateStr.includes(' ')) {
          parsedDeliveryDate = dateStr.split(' ')[0];
        } else {
          parsedDeliveryDate = dateStr;
        }
      }

      let customerName = res.customer_name || '';
      if (!customerName && res.order?.customer?.cus_name) {
        customerName = res.order.customer.cus_name;
      }

      // 编辑模式：productId 存储产品名称字符串
      formData.value = {
        moldType: 'internal',
        moldId: undefined,
        externalMoldCode: '',
        productId: res.product_name || undefined,
        orderId: res.order_id,
        warehouseId: res.warehouse_id,
        quantity: res.quantity,
        batchNo: res.batch_no || '',
        customerName,
        deliveryDate: parsedDeliveryDate,
        deliveryMethod: res.delivery_method || '',
        expressCompany: res.express_company || '',
        expressNo: res.express_no || '',
        driverPhone: res.driver_phone || '',
        driverPlate: res.driver_plate || '',
        photos: parsePhotos(res.photos || ''),
        remark: res.remark || '',
      };
    } else {
      MessagePlugin.error('送货单数据不存在');
      router.push('/production/stock-out');
    }
  } catch (error) {
    console.error('加载送货单数据失败:', error);
    MessagePlugin.error('加载送货单数据失败');
    router.push('/production/stock-out');
  }
};

const initPage = () => {
  if (route.path.includes('/edit/')) {
    formMode.value = 'edit';
    loadEditData();
  } else {
    formMode.value = 'create';
    const {
      mold_id,
      external_mold_code,
      product_id,
      mold_code,
      mold_name,
      product_name,
      customer_name,
      current_stock: queryCurrentStock,
    } = route.query;

    // 如果传递了 mold_id 或 external_mold_code，并且有 product_name，说明是从库存列表跳转的
    if ((mold_id || external_mold_code) && (product_id || product_name)) {
      isAutoFillMode.value = true;
      isProductDisabled.value = true;

      // 从 URL 参数预置当前库存，用于自动填充模式下立即显示库存提示
      if (queryCurrentStock) {
        currentStock.value = Number(queryCurrentStock);
      }

      if (mold_id) {
        formData.value.moldType = 'internal';
        formData.value.moldId = Number(mold_id);
      } else if (external_mold_code) {
        formData.value.moldType = 'external';
        formData.value.externalMoldCode = String(external_mold_code);
      }

      // 自动填充模式：使用 URL 参数中的产品名称
      if (product_name) {
        autoFilledProductInfo.value = String(product_name);
        // 同步设置 formData.productId，使 stockHelpText 能正确判断有产品
        formData.value.productId = String(product_name);
      }

      if (mold_code && mold_name) {
        autoFilledMoldInfo.value = `${mold_code} - ${mold_name}`;
      } else if (mold_code) {
        autoFilledMoldInfo.value = String(mold_code);
      } else if (external_mold_code) {
        autoFilledMoldInfo.value = String(external_mold_code);
      }

      if (customer_name) {
        formData.value.customerName = String(customer_name);
      } else if (mold_id) {
        // 从模具列表查询客户名称
        const selectedMold = moldList.value.find((m: any) => m.id === Number(mold_id));
        if (selectedMold) {
          const cusName = selectedMold.customer?.cus_name || selectedMold.customer?.name;
          if (cusName) {
            formData.value.customerName = cusName;
          }
        }
      }

      // 自动填充模式下，预加载所有仓库的产品库存
      nextTick(() => {
        fetchWarehouseStocks();
      });
    } else {
      // 手动填写模式
      isAutoFillMode.value = false;

      if (mold_id) {
        formData.value.moldType = 'internal';
        formData.value.moldId = Number(mold_id);
      } else if (external_mold_code) {
        formData.value.moldType = 'external';
        formData.value.externalMoldCode = String(external_mold_code);
      }

      // 手动模式如果有产品名参数，预填到下拉框
      if (product_name) {
        formData.value.productId = String(product_name);
      }
    }

    formData.value.deliveryDate = formatDate(new Date().toISOString());
  }
};

// 模具变更 → 过滤关联产品 + 自动带出客户
const handleMoldChange = async (skipClearProduct = false) => {
  if (!skipClearProduct) {
    formData.value.productId = undefined;
    currentStock.value = 0;
  }

  // 自动带出客户名称
  if (formData.value.moldId) {
    const selectedMold = moldList.value.find((m: any) => m.id === formData.value.moldId);
    if (selectedMold) {
      const customerName = selectedMold.customer?.cus_name || selectedMold.customer?.name;
      if (customerName) {
        formData.value.customerName = customerName;
      }
    }
  }

  if (!formData.value.moldId) {
    filteredProductList.value = productList.value;
  } else {
    // 按模具的 mold_id 过滤库存汇总中的产品
    filteredProductList.value = productList.value.filter(
      (p: any) => p.mold_id === formData.value.moldId,
    );
    if (filteredProductList.value.length === 0) {
      // 无匹配时显示全部（兼容库存数据没有 mold_id 的场景）
      filteredProductList.value = productList.value;
    }
  }

  fetchCurrentStock();
  fetchWarehouseStocks();
};

// 产品变更 → 自动回填模具 + 刷新库存查询
const handleProductChange = (productName: string) => {
  if (!productName) return;

  // 从产品列表中找到对应记录（来自库存汇总，包含 mold 信息）
  const product = productList.value.find(
    (p: any) => (p.product_name || p.pro_name) === productName,
  );

  if (product && !formData.value.moldId) {
    let matchedMold: any = null;

    // 策略1: 按 mold_id 匹配
    if (product.mold_id) {
      matchedMold = moldList.value.find((m: any) => m.id === product.mold_id);
    }

    // 策略2: 按 mold_code 匹配
    if (!matchedMold && product.mold_code) {
      matchedMold = moldList.value.find((m: any) => m.mold_code === product.mold_code);
    }

    // 策略3: 按产品名称匹配模具名称（很多场景下产品名即模具名）
    if (!matchedMold) {
      matchedMold = moldList.value.find(
        (m: any) => m.mold_name && m.mold_name.trim() === productName.trim(),
      );
    }

    // 策略4: 按产品名称匹配模具编号
    if (!matchedMold) {
      matchedMold = moldList.value.find(
        (m: any) => m.mold_code && m.mold_code.trim() === productName.trim(),
      );
    }

    if (matchedMold) {
      formData.value.moldId = matchedMold.id;
    }
  }

  // 自动带出客户名称：优先使用产品关联的订单客户（来自库存汇总），仅在客户未填写时填充
  if (product && !formData.value.customerName) {
    const customerName = product.customer_name || product.customer?.cus_name;
    if (customerName) {
      formData.value.customerName = customerName;
    }
  }

  // 程序化设置 v-model 不会触发 @change，手动刷新库存
  fetchCurrentStock();
  fetchWarehouseStocks();
};

// 关联订单变更 - 自动带出客户名称
const handleOrderChange = (orderId: number | undefined) => {
  if (!orderId) return;

  const selectedOrder = orderList.value.find((order: any) => order.id === orderId);
  if (selectedOrder) {
    if (selectedOrder.customer?.cus_name) {
      formData.value.customerName = selectedOrder.customer.cus_name;
    } else if (selectedOrder.customer_name) {
      formData.value.customerName = selectedOrder.customer_name;
    }
  }
};

// 获取当前库存（以产品名称 + 仓库为主键，模具为可选过滤）
const fetchCurrentStock = async () => {
  if (isEditMode.value) return;

  if (!formData.value.warehouseId) return;
  if (!formData.value.productId) return;

  stockLoading.value = true;
  try {
    const params: any = {
      warehouse_id: formData.value.warehouseId,
      product_name: formData.value.productId, // 传产品名称而非 product_id
    };

    if (formData.value.moldType === 'internal' && formData.value.moldId) {
      params.mold_id = formData.value.moldId;
    } else if (formData.value.moldType === 'external' && formData.value.externalMoldCode) {
      params.external_mold_code = formData.value.externalMoldCode;
    }

    const res = await getCurrentStock(params);
    currentStock.value = res.current_stock || 0;
    // 同步更新仓库库存映射
    if (formData.value.warehouseId) {
      warehouseStockMap.value[formData.value.warehouseId] = currentStock.value;
    }
  } catch (error) {
    console.error('获取库存失败:', error);
    currentStock.value = 0;
  } finally {
    stockLoading.value = false;
  }
};

// 获取所有仓库的产品库存（用于仓库下拉选项展示）
const fetchWarehouseStocks = async () => {
  if (!formData.value.productId || warehouseList.value.length === 0) return;

  const paramsBase: any = {
    product_name: formData.value.productId,
  };
  if (formData.value.moldType === 'internal' && formData.value.moldId) {
    paramsBase.mold_id = formData.value.moldId;
  } else if (formData.value.moldType === 'external' && formData.value.externalMoldCode) {
    paramsBase.external_mold_code = formData.value.externalMoldCode;
  }

  try {
    const stocks = await Promise.all(
      warehouseList.value.map((wh) =>
        getCurrentStock({ ...paramsBase, warehouse_id: wh.id }),
      ),
    );
    warehouseList.value.forEach((wh, idx) => {
      warehouseStockMap.value[wh.id] = stocks[idx]?.current_stock || 0;
    });
  } catch (error) {
    console.error('获取各仓库库存失败:', error);
  }
};

// 监听产品/模具变化，自动加载各仓库库存（兼容手动选择路径）
watch(
  () => [formData.value.productId, formData.value.moldId, formData.value.externalMoldCode],
  () => {
    if (formData.value.productId) {
      fetchWarehouseStocks();
    }
  },
);
const handleRecognizeExpress = async () => {
  if (!formData.value.expressNo) return;

  try {
    const res = await recognizeExpressCompany(formData.value.expressNo);
    if (res.express_company) {
      formData.value.expressCompany = res.express_company;
      MessagePlugin.success(`已自动识别快递公司: ${res.express_company_name}`);
    }
  } catch (error) {
    console.error('识别快递公司失败:', error);
  }
};

// 提交表单
const onSubmit = async ({ validateResult }: any) => {
  if (validateResult !== true) return;

  submitting.value = true;
  try {
    if (isEditMode.value) {
      const existingPhotoUrls = (formData.value.photos || []).map((p: any) => p.url || p);
      const pending = photoUploaderRef.value?.getPendingChanges() || {};
      const pendingFiles = pending.uploads || [];

      let allUrls = [...existingPhotoUrls];
      if (pendingFiles.length > 0 && stockOutId.value) {
        const results = await uploadFiles({ files: pendingFiles, category: 'stock-out' });
        if (results && results.length > 0) {
          allUrls = [...allUrls, ...results.map((r) => r.url)];
        }
      }

      const params: any = {
        customer_name: formData.value.customerName || undefined,
        delivery_date: formData.value.deliveryDate || undefined,
        delivery_method: formData.value.deliveryMethod || undefined,
        express_company: formData.value.expressCompany || undefined,
        express_no: formData.value.expressNo || undefined,
        photos: JSON.stringify(allUrls),
        driver_phone: formData.value.driverPhone || undefined,
        driver_plate: formData.value.driverPlate || undefined,
        remark: formData.value.remark || undefined,
      };

      await updateStockOut(stockOutId.value!, params);
      MessagePlugin.success('送货单更新成功');
    } else {
      const pending = photoUploaderRef.value?.getPendingChanges() || {};
      const pendingFiles = pending.uploads || [];

      let allUrls: string[] = [];
      if (pendingFiles.length > 0) {
        const results = await uploadFiles({ files: pendingFiles, category: 'stock-out' });
        if (results && results.length > 0) {
          allUrls = results.map((r) => r.url);
        }
      }

      const params: any = {
        order_id: formData.value.orderId || Number(route.query.order_id || 0),
        warehouse_id: formData.value.warehouseId,
        product_name: currentProductName.value,
        quantity: formData.value.quantity,
        batch_no: formData.value.batchNo,
        customer_name: formData.value.customerName,
        delivery_date: formData.value.deliveryDate,
        delivery_method: formData.value.deliveryMethod,
        express_company: formData.value.expressCompany || undefined,
        express_no: formData.value.expressNo || undefined,
        photos: allUrls.length > 0 ? JSON.stringify(allUrls) : undefined,
        driver_phone: formData.value.driverPhone || undefined,
        driver_plate: formData.value.driverPlate || undefined,
        remark: formData.value.remark || undefined,
      };

      if (formData.value.moldType === 'internal') {
        params.mold_id = formData.value.moldId;
      } else {
        params.external_mold_code = formData.value.externalMoldCode;
      }

      await stockOutOrder(params);
      MessagePlugin.success('送货单创建成功');

      formData.value = {
        moldType: 'internal',
        moldId: undefined,
        externalMoldCode: '',
        productId: undefined,
        orderId: undefined,
        warehouseId: undefined,
        quantity: undefined as number | undefined,
        batchNo: '',
        customerName: '',
        deliveryDate: '',
        deliveryMethod: '',
        expressCompany: '',
        expressNo: '',
        driverPhone: '',
        driverPlate: '',
        photos: [] as any[],
        remark: '',
      };

      const dialog = DialogPlugin.confirm({
        header: '提示',
        body: '是否立即打印送货单？',
        confirmBtn: '立即打印',
        cancelBtn: '稍后打印',
        onConfirm: () => {
          dialog.destroy();
          router.push('/production/stock-out');
        },
        onCancel: () => {
          dialog.destroy();
          router.push('/production/stock-out');
        },
      });
      return;
    }

    router.push('/production/stock-out');
  } catch (error: any) {
    console.error(isEditMode.value ? '更新送货单失败:' : '创建送货单失败:', error);
    const errorMsg =
      error.response?.data?.message ||
      error.message ||
      (isEditMode.value ? '更新送货单失败' : '创建送货单失败');
    MessagePlugin.error(errorMsg);
  } finally {
    submitting.value = false;
  }
};

const onReset = () => {
  MessagePlugin.warning('取消操作');
  router.push('/production/stock-out');
};

const handleBack = () => {
  const fromPage = route.query.from as string;
  if (fromPage === 'stock' || fromPage === 'stock-detail') {
    router.go(-1);
  } else {
    router.push('/production/stock-out');
  }
};

onMounted(async () => {
  await loadData(); // 等待模具列表加载完成
  initPage(); // 此时 moldList 已有数据
  // 如果页面初始化后已有产品信息，预加载各仓库库存
  if (formData.value.productId) {
    fetchWarehouseStocks();
  }
});
</script>
<style lang="less" scoped>
@import '@/pages/form/base/index.less';

.stock-out-create-container {
  padding: 0;
}

.back-button-container {
  margin-bottom: 16px;
  padding: 0 16px;
}
</style>

<style lang="less">
/* 仓库下拉选项样式（下拉菜单渲染在 body，需要全局样式） */
.warehouse-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.warehouse-option-name {
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.warehouse-option-stock {
  font-size: 12px;
  color: var(--td-brand-color);
  margin-left: 16px;
  flex-shrink: 0;
}
</style>
