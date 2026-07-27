<template>
  <div class="product-management-page">
    <SearchFilter
      :filters="filterConfig"
      :model-value="filterForm"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">产品列表</span>
        <div class="toolbar-actions">
          <t-button theme="success" @click="handleSync" :loading="syncing">
            <template #icon><t-icon name="refresh" /></template>
            同步生产记录
          </t-button>
          <t-button theme="primary" @click="handleCreate">
            <template #icon><t-icon name="add" /></template>
            新增产品
          </t-button>
          <t-button theme="default" @click="handleColumnSetting">
            <template #icon><t-icon name="setting" /></template>
            列设置
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <t-table
        :data="tableData"
        :columns="displayColumns"
        row-key="id"
        :loading="loading"
        :pagination="pagination"
        @page-change="handlePageChange"
      >
        <template #product_image="{ row }">
          <img
            v-if="row.product_image_url"
            :src="getThumbnailUrl(row.product_image_url)"
            class="product-image-thumb"
            @click="handlePreviewImage(row)"
          />
          <div v-else class="product-image-placeholder">无图片</div>
        </template>
        <template #spec="{ row }">
          <t-button 
            v-if="row.spec_detail" 
            theme="primary" 
            variant="text" 
            size="small"
            @click="handleViewSpecDetail(row)"
          >
            <template #icon><t-icon name="browse" /></template>
            查看规格
          </t-button>
          <span v-else style="color: var(--td-text-color-placeholder)">-</span>
        </template>
        <template #mold_code="{ row }">
          <t-link 
            v-if="row.mold_code" 
            theme="primary" 
            @click="handleViewMoldDetail(row.mold_id)"
          >
            {{ row.mold_code }}
          </t-link>
          <span v-else style="color: var(--td-text-color-placeholder)">-</span>
        </template>
        <template #material_type="{ row }">
          <span v-if="row.spec_detail">
            {{ getMaterialType(row.spec_detail) }}
          </span>
          <span v-else style="color: var(--td-text-color-placeholder)">-</span>
        </template>
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'default'">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </t-tag>
        </template>
        <template #price="{ row }">
          <span>¥{{ row.price?.toFixed(2) || '0.00' }}</span>
        </template>
        <template #stock_info="{ row }">
          <div>
            <div>{{ row.current_stock || 0 }}</div>
            <div style="font-size: 12px; color: var(--td-text-color-secondary)">
              入: {{ row.total_in || 0 }} / 出: {{ row.total_out || 0 }}
            </div>
          </div>
        </template>
        <template #operate="{ row }">
          <t-space :size="8">
            <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
            <t-link theme="danger" @click="handleDelete(row)">删除</t-link>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <!-- 列设置对话框 - 使用通用组件 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="product_table_columns"
      @change="handleColumnsChange"
    />

    <!-- 新增/编辑产品对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :header="isEdit ? '编辑产品' : '新增产品'"
      width="820px"
      placement="center"
      class="product-edit-dialog"
      @close="handleClose"
     :close-on-overlay-click="false">
      <div class="product-edit-content">
        <!-- 左侧：表单 -->
        <div class="product-edit-left">
          <t-form :data="formData" :rules="rules" ref="formRef" label-width="80px">
            <t-form-item label="产品编码" name="pro_code">
              <t-input v-model="formData.pro_code" placeholder="留空自动生成">
                <template #suffix>
                  <t-button variant="text" size="small" @click="handleGenerateCode" :loading="generating">
                    自动生成
                  </t-button>
                </template>
              </t-input>
            </t-form-item>

            <t-form-item label="产品名称" name="pro_name">
              <t-input v-model="formData.pro_name" placeholder="请输入产品名称" />
            </t-form-item>

            <t-form-item label="规格" name="spec">
              <t-input v-model="formData.spec" placeholder="请输入产品规格" />
            </t-form-item>

            <t-form-item label="单位" name="unit">
              <t-input v-model="formData.unit" placeholder="如：个、件、箱" />
            </t-form-item>

            <t-row :gutter="[16, 0]">
              <t-col :span="12">
                <t-form-item label="销售单价" name="price" label-width="70px">
                  <t-input-number v-model="formData.price" :min="0" :precision="2" placeholder="0.00" style="width: 100%" />
                </t-form-item>
              </t-col>
              <t-col :span="12">
                <t-form-item label="产品重量" name="weight" label-width="70px">
                  <t-input-number v-model="formData.weight" :min="0" :precision="2" placeholder="0.00" style="width: 100%" />
                </t-form-item>
              </t-col>
            </t-row>

            <t-form-item label="详细规格" name="spec_detail">
              <div class="spec-action-row">
                <t-button theme="default" size="small" @click="handleOpenSpecDetail" :disabled="!formData.pro_name">
                  <template #icon><t-icon name="edit" /></template>
                  查看/编辑详细规格
                </t-button>
                <span v-if="hasSpecDetail" class="spec-status-text">
                  <t-icon name="check-circle" /> 已配置
                </span>
              </div>
            </t-form-item>

            <t-form-item label="状态" name="status">
              <t-radio-group v-model="formData.status">
                <t-radio :value="1">启用</t-radio>
                <t-radio :value="0">禁用</t-radio>
              </t-radio-group>
            </t-form-item>

            <t-form-item label="备注" name="remark">
              <t-textarea v-model="formData.remark" placeholder="请输入备注" :autosize="{ minRows: 2, maxRows: 4 }" />
            </t-form-item>
          </t-form>
        </div>

        <!-- 右侧：图片上传 -->
        <div class="product-edit-right">
          <div class="image-upload-section">
            <div class="image-upload-header">
              <span class="image-upload-title">产品图片</span>
              <span class="image-upload-subtitle">支持 JPG/PNG/GIF/WebP，不超过 5MB</span>
            </div>
            <div class="image-upload-body">
              <FileUploader
                ref="fileUploaderRef"
                :key="fileUploaderKey"
                :model-value="productImageModel"
                :auto-save="false"
                :image-only="true"
                :max-count="1"
                :multiple="false"
                :show-hint="false"
                :show-actions="true"
                category="product-images"
                upload-btn-text="选择图片"
                @change="handleFileUploadChange"
              />
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <t-button theme="default" @click="handleCancel">取消</t-button>
        <t-button theme="primary" @click="handleSubmit">确认</t-button>
      </template>
    </t-dialog>

    <!-- 查看规格详情弹窗（只读模式） -->
    <t-dialog
      v-model:visible="viewSpecDetailVisible"
      header="产品详细规格"
      width="1000px"
      placement="center"
      :footer="false"
     :close-on-overlay-click="false">
      <div class="spec-detail-container">
        <!-- 顶部操作栏 -->
        <div class="spec-view-actions">
          <t-button theme="primary" @click="handleEditSpecDetail">
            <template #icon><t-icon name="edit" /></template>
            编辑规格
          </t-button>
        </div>

        <!-- 产品基本信息 -->
        <div class="product-info-card">
          <div class="product-info-row">
            <div class="product-info-item">
              <span class="info-label">产品编号：</span>
              <span class="info-value">{{ currentProductInfo.pro_code || '-' }}</span>
            </div>
            <div class="product-info-item">
              <span class="info-label">产品名称：</span>
              <span class="info-value">{{ currentProductInfo.pro_name || '-' }}</span>
            </div>
            <div class="product-info-item">
              <span class="info-label">所属模具：</span>
              <span class="info-value mold-code" v-if="currentProductInfo.moldCode">
                <t-link theme="primary" @click="handleViewMoldDetail(currentProductInfo.moldId)">
                  {{ currentProductInfo.moldCode }}
                </t-link>
              </span>
              <span class="info-value" v-else>-</span>
            </div>
          </div>
        </div>

        <!-- 尺寸参数 -->
        <div class="spec-card">
          <div class="spec-card-header">
            <span class="spec-card-title">尺寸参数</span>
          </div>
          <div class="spec-card-body">
            <div class="spec-description-list">
              <div class="spec-description-item">
                <span class="spec-description-label">长(mm)</span>
                <span class="spec-description-value">{{ specDetailForm.size.length || '-' }}</span>
              </div>
              <div class="spec-description-item">
                <span class="spec-description-label">宽(mm)</span>
                <span class="spec-description-value">{{ specDetailForm.size.width || '-' }}</span>
              </div>
              <div class="spec-description-item">
                <span class="spec-description-label">高(mm)</span>
                <span class="spec-description-value">{{ specDetailForm.size.height || '-' }}</span>
              </div>
              <div class="spec-description-item">
                <span class="spec-description-label">壁厚(mm)</span>
                <span class="spec-description-value">{{ specDetailForm.size.wallThickness || '-' }}</span>
              </div>
              <div class="spec-description-item">
                <span class="spec-description-label">公差等级</span>
                <span class="spec-description-value">{{ specDetailForm.size.tolerance || '-' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 材质参数 -->
        <div class="spec-card">
          <div class="spec-card-header">
            <span class="spec-card-title">材质参数</span>
          </div>
          <div class="spec-card-body">
            <div class="spec-description-list">
              <div class="spec-description-item">
                <span class="spec-description-label">材质类型</span>
                <span class="spec-description-value">{{ specDetailForm.material.type || '-' }}</span>
              </div>
              <div class="spec-description-item">
                <span class="spec-description-label">材料牌号</span>
                <span class="spec-description-value">{{ specDetailForm.material.grade || '-' }}</span>
              </div>
              <div class="spec-description-item">
                <span class="spec-description-label">颜色</span>
                <span class="spec-description-value">{{ specDetailForm.material.color || '-' }}</span>
              </div>
              <div class="spec-description-item">
                <span class="spec-description-label">密度(g/cm³)</span>
                <span class="spec-description-value">{{ specDetailForm.material.density || '-' }}</span>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 工艺参数 -->
        <div class="spec-card">
          <div class="spec-card-header">
            <span class="spec-card-title">工艺参数</span>
          </div>
          <div class="spec-card-body">
            <div class="spec-description-list">
              <div class="spec-description-item">
                <span class="spec-description-label">注塑温度(℃)</span>
                <span class="spec-description-value">{{ specDetailForm.process.injectionTemp || '-' }}</span>
              </div>
              <div class="spec-description-item">
                <span class="spec-description-label">模具温度(℃)</span>
                <span class="spec-description-value">{{ specDetailForm.process.moldTemp || '-' }}</span>
              </div>
              <div class="spec-description-item">
                <span class="spec-description-label">成型周期(s)</span>
                <span class="spec-description-value">{{ specDetailForm.process.cycleTime || '-' }}</span>
              </div>
              <div class="spec-description-item">
                <span class="spec-description-label">缩水率(%)</span>
                <span class="spec-description-value">{{ specDetailForm.process.shrinkage || '-' }}</span>
              </div>
              <div class="spec-description-item">
                <span class="spec-description-label">表面处理</span>
                <span class="spec-description-value">{{ specDetailForm.process.surfaceTreatment || '-' }}</span>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 包装参数 -->
        <div class="spec-card">
          <div class="spec-card-header">
            <span class="spec-card-title">包装参数</span>
          </div>
          <div class="spec-card-body">
            <div class="spec-description-list">
              <div class="spec-description-item">
                <span class="spec-description-label">包装规格</span>
                <span class="spec-description-value">{{ specDetailForm.packaging.spec || '-' }}</span>
              </div>
              <div class="spec-description-item">
                <span class="spec-description-label">装箱数量(个)</span>
                <span class="spec-description-value">{{ specDetailForm.packaging.qtyPerBox || '-' }}</span>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 其他参数 -->
        <div class="spec-card">
          <div class="spec-card-header">
            <span class="spec-card-title">其他参数</span>
          </div>
          <div class="spec-card-body">
            <div class="spec-description-list">
              <div class="spec-description-item spec-description-item-full">
                <span class="spec-description-label">应用领域</span>
                <span class="spec-description-value">{{ specDetailForm.other.application || '-' }}</span>
              </div>
              <div class="spec-description-item spec-description-item-full">
                <span class="spec-description-label">备注</span>
                <span class="spec-description-value">{{ specDetailForm.other.remark || '-' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 自定义规格参数 -->
        <div class="spec-card" v-if="specDetailForm.customSpecs && specDetailForm.customSpecs.length > 0">
          <div class="spec-card-header">
            <span class="spec-card-title">自定义规格参数</span>
          </div>
          <div class="spec-card-body">
            <div class="spec-description-list">
              <div class="spec-description-item" v-for="(spec, index) in specDetailForm.customSpecs" :key="index">
                <span class="spec-description-label">{{ spec.name || '-' }}</span>
                <span class="spec-description-value">{{ spec.value || '-' }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </t-dialog>

    <!-- 图片预览 -->
  <t-image-viewer
    v-model:visible="imagePreviewVisible"
    :images="previewImages"
    :default-index="previewIndex"
    :z-index="10000"
  />

  <!-- 编辑规格详情弹窗（可编辑模式） -->
    <t-dialog
      v-model:visible="editSpecDetailVisible"
      header="编辑产品详细规格"
      width="1000px"
      placement="center"
      @confirm="handleSaveSpecDetail"
      @close="handleCloseSpecDetail"
     :close-on-overlay-click="false">
      <div class="spec-detail-container">
        <!-- 尺寸参数 -->
        <div class="spec-card">
          <div class="spec-card-header">
            <span class="spec-card-title">尺寸参数</span>
          </div>
          <div class="spec-card-body">
            <div class="spec-grid spec-grid-3col">
              <div class="spec-field">
                <label class="spec-label">长(mm)</label>
                <t-input-number v-model="specDetailForm.size.length" :min="0" placeholder="0" />
              </div>
              <div class="spec-field">
                <label class="spec-label">宽(mm)</label>
                <t-input-number v-model="specDetailForm.size.width" :min="0" placeholder="0" />
              </div>
              <div class="spec-field">
                <label class="spec-label">高(mm)</label>
                <t-input-number v-model="specDetailForm.size.height" :min="0" placeholder="0" />
              </div>
              <div class="spec-field">
                <label class="spec-label">壁厚(mm)</label>
                <t-input-number v-model="specDetailForm.size.wallThickness" :min="0" :precision="2" placeholder="0" />
              </div>
              <div class="spec-field">
                <label class="spec-label">公差等级</label>
                <t-input v-model="specDetailForm.size.tolerance" placeholder="如:±0.1" />
              </div>
            </div>
          </div>
        </div>
    
        <!-- 材质参数 -->
        <div class="spec-card">
          <div class="spec-card-header">
            <span class="spec-card-title">材质参数</span>
          </div>
          <div class="spec-card-body">
            <div class="spec-grid spec-grid-3col">
              <div class="spec-field">
                <label class="spec-label">材质类型</label>
                <t-select v-model="specDetailForm.material.type" placeholder="请选择" clearable filterable>
                  <t-option 
                    v-for="item in plasticMaterialDictList" 
                    :key="item.dict_value" 
                    :value="item.dict_value" 
                    :label="item.dict_label" 
                  />
                </t-select>
              </div>
              <div class="spec-field">
                <label class="spec-label">材料牌号</label>
                <t-input v-model="specDetailForm.material.grade" placeholder="如:T30S" />
              </div>
              <div class="spec-field">
                <label class="spec-label">颜色</label>
                <t-input v-model="specDetailForm.material.color" placeholder="如:黑色" />
              </div>
              <div class="spec-field">
                <label class="spec-label">密度(g/cm³)</label>
                <t-input-number v-model="specDetailForm.material.density" :min="0" :precision="3" placeholder="0" />
              </div>
            </div>
          </div>
        </div>
            
        <!-- 工艺参数 -->
        <div class="spec-card">
          <div class="spec-card-header">
            <span class="spec-card-title">工艺参数</span>
          </div>
          <div class="spec-card-body">
            <div class="spec-grid spec-grid-3col">
              <div class="spec-field">
                <label class="spec-label">注塑温度(℃)</label>
                <t-input-number v-model="specDetailForm.process.injectionTemp" :min="0" placeholder="0" />
              </div>
              <div class="spec-field">
                <label class="spec-label">模具温度(℃)</label>
                <t-input-number v-model="specDetailForm.process.moldTemp" :min="0" placeholder="0" />
              </div>
              <div class="spec-field">
                <label class="spec-label">成型周期(s)</label>
                <t-input-number v-model="specDetailForm.process.cycleTime" :min="0" placeholder="0" />
              </div>
              <div class="spec-field">
                <label class="spec-label">缩水率(%)</label>
                <t-input-number v-model="specDetailForm.process.shrinkage" :min="0" :precision="2" placeholder="0" />
              </div>
              <div class="spec-field">
                <label class="spec-label">表面处理</label>
                <t-select v-model="specDetailForm.process.surfaceTreatment" placeholder="请选择">
                  <t-option value="抛光" label="抛光" />
                  <t-option value="蚀纹" label="蚀纹" />
                  <t-option value="电镀" label="电镀" />
                  <t-option value="喷涂" label="喷涂" />
                </t-select>
              </div>
            </div>
          </div>
        </div>
            
        <!-- 包装参数 -->
        <div class="spec-card">
          <div class="spec-card-header">
            <span class="spec-card-title">包装参数</span>
          </div>
          <div class="spec-card-body">
            <div class="spec-grid spec-grid-3col">
              <div class="spec-field">
                <label class="spec-label">包装规格</label>
                <t-input v-model="specDetailForm.packaging.spec" placeholder="如:100个/箱" />
              </div>
              <div class="spec-field">
                <label class="spec-label">装箱数量(个)</label>
                <t-input-number v-model="specDetailForm.packaging.qtyPerBox" :min="0" placeholder="0" />
              </div>
            </div>
          </div>
        </div>
            
        <!-- 其他参数 -->
        <div class="spec-card">
          <div class="spec-card-header">
            <span class="spec-card-title">其他参数</span>
          </div>
          <div class="spec-card-body">
            <div class="spec-grid spec-grid-2col">
              <div class="spec-field spec-field-full">
                <label class="spec-label">应用领域</label>
                <t-input v-model="specDetailForm.other.application" placeholder="如:电子、家电、汽车等" />
              </div>
              <div class="spec-field spec-field-full">
                <label class="spec-label">备注</label>
                <t-textarea v-model="specDetailForm.other.remark" placeholder="其他特殊要求或说明" :autosize="{ minRows: 2, maxRows: 3 }" />
              </div>
            </div>
          </div>
        </div>
    
        <!-- 自定义规格参数 -->
        <div class="spec-card">
          <div class="spec-card-header">
            <span class="spec-card-title">自定义规格参数</span>
            <t-button theme="primary" size="small" @click="addCustomSpec">
              <template #icon><t-icon name="add" /></template>
              新增参数
            </t-button>
          </div>
          <div class="spec-card-body">
            <t-table
              v-if="specDetailForm.customSpecs && specDetailForm.customSpecs.length > 0"
              :data="specDetailForm.customSpecs"
              :columns="customSpecColumns"
              row-key="index"
              bordered
            >
              <template #name="{ row, rowIndex }">
                <t-input v-model="row.name" placeholder="请输入规格名称" />
              </template>
              <template #value="{ row, rowIndex }">
                <t-input v-model="row.value" placeholder="请输入规格参数" />
              </template>
              <template #operate="{ rowIndex }">
                <t-button theme="danger" variant="text" @click="removeCustomSpec(rowIndex)">
                  <template #icon><t-icon name="delete" /></template>
                  删除
                </t-button>
              </template>
            </t-table>
            <div v-else class="empty-tip">
              <t-icon name="info-circle" style="font-size: 32px; margin-bottom: 8px;" />
              <div>暂无自定义参数，点击“新增参数”添加</div>
            </div>
          </div>
        </div>
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next';
import { SettingIcon } from 'tdesign-icons-vue-next';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';
import ImagePlaceholder from '@/components/ImagePlaceholder/index.vue';
import { getProductList, createProduct, updateProduct, deleteProduct, generateProductCode, syncProducts } from '@/api/product';
import { uploadFiles, getThumbnailUrl, type UploadFileResult } from '@/api/common/upload';
import { useDictRaw } from '@/hooks/useDict';
import FileUploader from '@/components/common/FileUploader.vue';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';

defineOptions({
  name: 'ProductManagement',
});

const router = useRouter();

const loading = ref(false);
const dialogVisible = ref(false);
const isEdit = ref(false);
const generating = ref(false);
const syncing = ref(false);
const formRef = ref();
const fileUploaderRef = ref();
const fileUploaderKey = ref(0);

// FileUploader 的 model-value：已有图片链接转为组件需要的格式
// 注意：fileName 必须包含扩展名，组件通过 getFileExt(fileName) 判断是否为图片
// 关键：FileUploader 内部会对 url 再调用一次 getFileFullUrl()，如果传入相对路径会被拼接上 VITE_API_URL（API 端口）
//       导致 t-image 跨域请求被 CORP 拦截。所以这里必须用 window.location.origin 拼接完整同源 URL，
//       getFileFullUrl 看到 http:// 开头会原样返回。
// 注意：传原始 URL 而非 _thumb 缩略图，否则 t-image-viewer 预览大图也会是缩略图大小。
// 重要：当 FileUploader 有 pending 变更时返回空数组，避免旧图+新图同时显示两个。
const hasImagePending = ref(false);

const handleFileUploadChange = (changeInfo: any) => {
  hasImagePending.value = changeInfo.hasChanges;
};

const productImageModel = computed(() => {
  // 有 pending 变更时不传旧图，由 FileUploader 自行显示 pending 新文件
  if (hasImagePending.value) return [];

  const rawUrl = formData.product_image_url || formData.image_url || '';
  if (rawUrl) {
    // 用当前页面 origin 构建完整同源 URL
    const fullOriginUrl = rawUrl.startsWith('http') ? rawUrl : `${window.location.origin}${rawUrl}`;
    const ext = rawUrl.match(/\.(jpe?g|png|gif|webp)/i)?.[1] || 'jpg';
    const name = formData.pro_name || '产品';
    return [{
      key: 'existing-image',
      url: fullOriginUrl,
      fileName: `${name}.${ext}`,
      fileSize: 0,
      fileType: `.${ext}`,
    }];
  }
  return [];
});
const imagePreviewVisible = ref(false);
const previewImages = ref<string[]>([]);
const previewIndex = ref(0);

// 塑胶原料字典列表（useDictRaw 自动加载 + 全局缓存，只取启用项）
const { items: plasticMaterialDictList } = useDictRaw('plastic-material', (items) =>
  items.filter((item: any) => item.status === 1),
);

// 规格详情相关
const viewSpecDetailVisible = ref(false);  // 查看模式
const editSpecDetailVisible = ref(false);  // 编辑模式
const specDetailVisible = ref(false);  // 兼容旧代码
const hasSpecDetail = ref(false);
const currentProductInfo = reactive({
  id: undefined as number | undefined,
  pro_code: '',
  pro_name: '',
  moldCode: '',
  moldId: undefined as number | undefined,
});
const specDetailForm = reactive({
  size: {
    length: undefined,
    width: undefined,
    height: undefined,
    wallThickness: undefined,
    tolerance: '',
  },
  material: {
    type: '',
    grade: '',
    color: '',
    density: undefined,
  },
  process: {
    injectionTemp: undefined,
    moldTemp: undefined,
    cycleTime: undefined,
    shrinkage: undefined,
    surfaceTreatment: '',
  },
  packaging: {
    spec: '',
    qtyPerBox: undefined,
  },
  other: {
    application: '',
    remark: '',
  },
  customSpecs: [] as Array<{ name: string; value: string }>,
});

// 自定义规格参数表格列
const customSpecColumns = [
  {
    colKey: 'name',
    title: '规格名称',
    width: 70,
  },
  {
    colKey: 'value',
    title: '规格参数',
    ellipsis: true,
  },
  {
    colKey: 'operate',
    title: '操作',
    width: 30,
  },
];

const tableData = ref([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const filterForm = reactive({
  pro_code: '',
  pro_name: '',
});

const formData = reactive({
  id: undefined,
  pro_code: '',
  pro_name: '',
  spec: '',
  spec_detail: '',
  unit: '',
  price: 0,
  weight: 0,
  status: 1,
  remark: '',
  image_url: '',
  product_image_url: '',
});

const rules = {
  pro_name: [{ required: true, message: '请输入产品名称', type: 'error' as const }],
};

const columns = [
  { colKey: 'pro_code', title: '产品编码', width: 150, fixed: 'left' as const },
  { colKey: 'product_image', title: '图片', width: 120 },  // 图片列：可配置，不固定
  { colKey: 'pro_name', title: '产品名称', width: 150 },
  { colKey: 'mold_code', title: '模具编号', width: 150 },
  { colKey: 'spec', title: '规格', width: 180 },
  { colKey: 'material_type', title: '材质', width: 100 },
  { colKey: 'unit', title: '单位', width: 80 },
  { colKey: 'price', title: '单价', width: 120 },
  { colKey: 'weight', title: '重量', width: 100 },
  { colKey: 'stock_info', title: '当前库存', width: 200 },
  { colKey: 'status', title: '状态', width: 80 },
  { colKey: 'operate', title: '操作', width: 180, fixed: 'right' as const },
];

// 所有可选列配置（不包括操作列）
const allColumns: ColumnOption[] = [
  { colKey: 'pro_code', title: '产品编码', pinned: true },  // ✨ 固定列：不可隐藏/移动
  { colKey: 'product_image', title: '图片' },  // 可配置列
  { colKey: 'pro_name', title: '产品名称' },
  { colKey: 'mold_code', title: '模具编号' },
  { colKey: 'spec', title: '规格' },
  { colKey: 'material_type', title: '材质' },
  { colKey: 'price', title: '单价' },
  { colKey: 'weight', title: '重量' },
  { colKey: 'stock_info', title: '当前库存' },
  { colKey: 'status', title: '状态' },
];

// 默认显示的列（图片列默认第二）
const DEFAULT_COLUMNS = ['pro_code', 'product_image', 'pro_name', 'mold_code', 'spec', 'material_type', 'unit', 'price', 'weight', 'stock_info', 'status'];

const getMaterialType = (specDetail: string | undefined): string => {
  if (!specDetail) return '-';
  try {
    const detail = JSON.parse(specDetail);
    return detail?.material?.type || '-';
  } catch {
    return '-';
  }
};

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([]);

// 动态列配置（根据用户选择）
const displayColumns = computed(() => {
  // 根据 selectedColumns 的顺序来返回列配置
  const columnsWithOperation = [...columns];
  const orderedColumns = selectedColumns.value
    .map(colKey => columnsWithOperation.find(col => col.colKey === colKey))
    .filter(col => col !== undefined);
  
  // 将操作列添加到最后
  const operationCol = columnsWithOperation.find(col => col.colKey === 'operate');
  if (operationCol) {
    orderedColumns.push(operationCol);
  }
  
  return orderedColumns;
});

// 打开列设置对话框
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// 列配置变化时的回调
const handleColumnsChange = (newColumns: string[]) => {
  selectedColumns.value = newColumns;
};

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('product_table_columns');
    if (saved) {
      const savedData = JSON.parse(saved);
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

const loadProductList = async () => {
  loading.value = true;
  try {
    const res = await getProductList({
      page: pagination.current,
      page_size: pagination.pageSize,
      ...filterForm,
    });
    tableData.value = res.list || [];
    pagination.total = res.total || 0;
  } catch (error) {
    console.error('加载产品列表失败:', error);
    MessagePlugin.error('加载产品列表失败');
  } finally {
    loading.value = false;
  }
};

// 搜索
const handleSearch = () => {
  pagination.current = 1;
  loadProductList();
};

const handleReset = () => {
  filterForm.pro_code = '';
  filterForm.pro_name = '';
  pagination.current = 1;
  loadProductList();
};

// 分页变化
const handlePageChange = (pageInfo: any) => {
  pagination.current = pageInfo.current;
  pagination.pageSize = pageInfo.pageSize;
  loadProductList();
};

// 新增
const handleCreate = () => {
  isEdit.value = false;
  resetForm();
  fileUploaderKey.value++;
  dialogVisible.value = true;
};

// 编辑
const handleEdit = (row: any) => {
  isEdit.value = true;
  Object.assign(formData, row);
  
  if (row.spec_detail) {
    try {
      const parsed = JSON.parse(row.spec_detail);
      Object.assign(specDetailForm, parsed);
      hasSpecDetail.value = true;
    } catch (e) {
      console.error('解析规格详情失败:', e);
      hasSpecDetail.value = false;
    }
  } else {
    hasSpecDetail.value = false;
  }
  
  fileUploaderRef.value?.resetPendingChanges();
  fileUploaderKey.value++;
  dialogVisible.value = true;
};

const handleDelete = (row: any) => {
  DialogPlugin.confirm({
    header: '确认删除',
    body: `确定要删除产品"${row.pro_name}"吗？`,
    onConfirm: async () => {
      try {
        await deleteProduct(row.id);
        MessagePlugin.success('删除成功');
        loadProductList();
      } catch (error) {
        MessagePlugin.error('删除失败');
      }
    },
  });
};

const handleGenerateCode = async () => {
  generating.value = true;
  try {
    const res = await generateProductCode();
    formData.pro_code = res.product_code;
    MessagePlugin.success('编码已生成');
  } catch (error) {
    MessagePlugin.error('生成编码失败');
  } finally {
    generating.value = false;
  }
};

// 提交
const handleSubmit = async () => {
  const valid = await formRef.value?.validate();
  if (valid !== true) return;

  try {
    // 通过 FileUploader 统一组件处理图片上传
    const pending = fileUploaderRef.value?.getPendingChanges() || {};
    const pendingFiles: File[] = pending.uploads || [];
    const deleteIndices: number[] = pending.deleteIndices || [];

    if (pendingFiles.length > 0) {
      const results = await uploadFiles({ files: pendingFiles, category: 'product-images' });
      if (results?.length > 0) {
        formData.image_url = results[0].url;
        formData.product_image_url = results[0].url;
      }
    } else if (deleteIndices.includes(0)) {
      // 用户删除了已有的图片
      formData.image_url = '';
      formData.product_image_url = '';
    }
    // else: 既没有新文件也没有删除，保持原有 image_url
    
    const submitData = { ...formData };
    if (isEdit.value) {
      await updateProduct(submitData);
      MessagePlugin.success('更新成功');
    } else {
      await createProduct(submitData);
      MessagePlugin.success('创建成功');
    }
    dialogVisible.value = false;
    loadProductList();
  } catch (error: any) {
    MessagePlugin.error(error.message || (isEdit.value ? '更新失败' : '创建失败'));
  }
};

// 关闭对话框
const handleClose = () => {
  resetForm();
};

// 取消
const handleCancel = () => {
  dialogVisible.value = false;
  resetForm();
};

const resetForm = () => {
  formData.id = undefined;
  formData.pro_code = '';
  formData.pro_name = '';
  formData.spec = '';
  formData.spec_detail = '';
  formData.unit = '';
  formData.price = 0;
  formData.weight = 0;
  formData.status = 1;
  formData.remark = '';
  formData.image_url = '';
  formData.product_image_url = '';
  formRef.value?.reset();
  hasSpecDetail.value = false;
  
  fileUploaderRef.value?.resetPendingChanges();
  hasImagePending.value = false;
  
  specDetailForm.size = {
    length: undefined,
    width: undefined,
    height: undefined,
    wallThickness: undefined,
    tolerance: '',
  };
  specDetailForm.material = {
    type: '',
    grade: '',
    color: '',
    density: undefined,
  };
  specDetailForm.process = {
    injectionTemp: undefined,
    moldTemp: undefined,
    cycleTime: undefined,
    shrinkage: undefined,
    surfaceTreatment: '',
  };
  specDetailForm.packaging = {
    spec: '',
    qtyPerBox: undefined,
  };
  specDetailForm.other = {
    application: '',
    remark: '',
  };
  specDetailForm.customSpecs = [];
};

// 同步生产记录
const handleSync = async () => {
  syncing.value = true;
  try {
    await syncProducts();
    MessagePlugin.success('同步成功');
    loadProductList();
  } catch (error) {
    console.error('同步失败:', error);
    MessagePlugin.error('同步失败');
  } finally {
    syncing.value = false;
  }
};

// 图片预览
const handlePreviewImage = (row: any) => {
  if (row.product_image_url) {
    previewImages.value = [row.product_image_url];
    previewIndex.value = 0;
    imagePreviewVisible.value = true;
  }
};

// 打开规格详情弹窗（编辑模式）
const handleOpenSpecDetail = () => {
  // 如果有已保存的规格详情,解析JSON
  if (formData.spec_detail) {
    try {
      const parsed = JSON.parse(formData.spec_detail);
      Object.assign(specDetailForm, parsed);
      hasSpecDetail.value = true;
    } catch (e) {
      console.error('解析规格详情失败:', e);
    }
  } else {
    hasSpecDetail.value = false;
  }
  editSpecDetailVisible.value = true;  // 打开编辑模式
};

// 查看产品规格详情（从表格中点击查看）
const handleViewSpecDetail = async (row: any) => {
  if (!row.spec_detail) {
    MessagePlugin.warning('该产品暂无详细规格');
    return;
  }
  
  try {
    const parsed = JSON.parse(row.spec_detail);
    Object.assign(specDetailForm, parsed);
    hasSpecDetail.value = true;
    
    currentProductInfo.id = row.id;
    currentProductInfo.pro_code = row.pro_code || '';
    currentProductInfo.pro_name = row.pro_name || '';
    currentProductInfo.moldCode = row.mold_code || '';
    currentProductInfo.moldId = row.mold_id;
    
    viewSpecDetailVisible.value = true;  // 打开查看模式
  } catch (e) {
    console.error('解析规格详情失败:', e);
    MessagePlugin.error('规格数据解析失败');
  }
};

// 查看模具详情
const handleViewMoldDetail = (moldId?: number) => {
  if (!moldId) return;
  router.push(`/mold/detail/${moldId}`);
};

// 从查看模式切换到编辑模式
const handleEditSpecDetail = () => {
  viewSpecDetailVisible.value = false;
  editSpecDetailVisible.value = true;
};

const handleSaveSpecDetail = () => {
  // 将表单数据转为JSON
  const currentProduct = tableData.value.find(p => p.id === formData.id);
  if (currentProduct) {
    currentProduct.spec_detail = JSON.stringify(specDetailForm);
  }
  formData.spec_detail = JSON.stringify(specDetailForm);
  hasSpecDetail.value = true;
  editSpecDetailVisible.value = false;
  MessagePlugin.success('规格详情已保存');
  loadProductList();  // 刷新列表
};

// 关闭规格详情弹窗
const handleCloseSpecDetail = () => {
  viewSpecDetailVisible.value = false;
  editSpecDetailVisible.value = false;
  specDetailVisible.value = false;
};

// 新增自定义规格参数
const addCustomSpec = () => {
  specDetailForm.customSpecs.push({
    name: '',
    value: '',
  });
};

const removeCustomSpec = (index: number) => {
  specDetailForm.customSpecs.splice(index, 1);
};

// 加载塑胶原料字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'pro_code', placeholder: '请输入产品编码' },
  { type: 'input', field: 'pro_name', placeholder: '请输入产品名称' },
];

onMounted(() => {
  loadProductList();
  loadColumnConfig();  // 加载列配置
});
</script>

<style lang="less" scoped>
.product-management-page {
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

.toolbar-title {
  font-size: 16px;
  font-weight: 500;
}

.toolbar-actions {
  display: flex;
  gap: 8px;
}

.form-tip {
  font-size: 12px;
  color: var(--td-text-color-secondary);
  margin-top: 4px;
}

.spec-detail-container {
  max-height: 600px;
  overflow-y: auto;
  padding: 8px 0;
}

/* 产品基本信息卡片 */
.product-info-card {
  margin-bottom: 16px;
  padding: 16px;
  background-color: var(--td-bg-color-container);
  border: 1px solid var(--td-component-border);
  border-radius: 6px;
}

.product-info-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  align-items: center;
}

.product-info-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-label {
  font-size: 13px;
  color: var(--td-text-color-secondary);
  white-space: nowrap;
}

.info-value {
  font-size: 14px;
  color: var(--td-text-color-primary);
  font-weight: 500;
}

.mold-code {
  color: var(--td-brand-color);
  cursor: pointer;
}

.mold-code:hover {
  text-decoration: underline;
}

/* 查看模式操作栏 */
.spec-view-actions {
  margin-bottom: 12px;
  display: flex;
  justify-content: flex-end;
}

/* 描述列表样式 */
.spec-description-list {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.spec-description-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.spec-description-item-full {
  grid-column: 1 / -1;
}

.spec-description-label {
  font-size: 13px;
  color: var(--td-text-color-secondary);
  font-weight: 400;
}

.spec-description-value {
  font-size: 14px;
  color: var(--td-text-color-primary);
  font-weight: 500;
  padding: 8px 12px;
  background-color: var(--td-bg-color-container);
  border-radius: 4px;
  min-height: 20px;
  word-break: break-all;
}

.spec-card {
  margin-bottom: 16px;
  background-color: var(--td-bg-color-container);
  border: 1px solid var(--td-component-border);
  border-radius: 6px;
  overflow: hidden;
  transition: all 0.3s ease;
}

.spec-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.spec-card:last-child {
  margin-bottom: 0;
}

.spec-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: linear-gradient(135deg, var(--td-bg-color-secondarycontainer) 0%, var(--td-bg-color-container) 100%);
  border-bottom: 1px solid var(--td-component-border);
}

.spec-card-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.spec-card-body {
  padding: 16px;
}

.spec-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.spec-grid-3col {
  grid-template-columns: repeat(3, 1fr);
}

.spec-grid-2col {
  grid-template-columns: repeat(2, 1fr);
}

.spec-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.spec-field-full {
  grid-column: 1 / -1;
}

.spec-label {
  font-size: 13px;
  color: var(--td-text-color-secondary);
  font-weight: 400;
}

.empty-tip {
  text-align: center;
  padding: 40px 20px;
  color: var(--td-text-color-placeholder);
  font-size: 13px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

/* 优化输入框样式 */
.spec-field .t-input,
.spec-field .t-input-number,
.spec-field .t-select,
.spec-field .t-textarea {
  width: 100%;
}

/* 优化表格样式 */
.spec-card-body .t-table {
  margin-top: 0;
}

/* 自定义参数表格样式 */
.spec-card-body .t-table .t-table__row {
  height: auto;
}

.spec-card-body .t-table td {
  padding: 12px 16px !important;
  vertical-align: middle;
}

/* 表格内输入框样式 - 使用默认尺寸 */
.spec-card-body .t-table .t-input {
  width: 100%;
}

/* 表格内图片样式 - 参考生产订单 */
.product-image-thumb {
  width: 70px;
  height: 70px;
  object-fit: cover;
  border-radius: 4px;
  cursor: pointer;
  transition: opacity 0.2s;
  vertical-align: middle;
}

.product-image-thumb:hover {
  opacity: 0.8;
}

.product-image-placeholder {
  width: 70px;
  height: 70px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--td-text-color-placeholder);
  background-color: var(--td-bg-color-component);
  border-radius: 4px;
}

/* 表格单元垂直居中 */
::deep(.t-table td[data-colkey="product_image"]) {
  vertical-align: middle !important;
}

/* ==================== 产品编辑弹窗 - 左右分栏布局（参考收款弹窗） ==================== */
.product-edit-content {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 20px;
}

.product-edit-left {
  min-width: 0;
  padding-right: 20px;
  border-right: 1px solid var(--td-border-level-1-color);
}

.product-edit-right {
  min-width: 280px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 右侧图片上传区域 - 完全复刻收款弹窗 .file-upload-section */
.image-upload-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
}

.image-upload-header {
  padding-bottom: 8px;
  margin-bottom: 12px;
  border-bottom: 1px solid var(--td-border-level-1-color);
}

.image-upload-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.image-upload-subtitle {
  font-size: 11px;
  color: var(--td-text-color-placeholder);
  margin-top: 4px;
}

.image-upload-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.image-upload-body :deep(.upload-tip) {
  font-size: 11px;
}

/* 响应式：窄屏时单栏布局 */
@media (width <= 768px) {
  .product-edit-content {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  .product-edit-left {
    padding-right: 0;
    border-right: none;
    border-bottom: 1px solid var(--td-border-level-1-color);
    padding-bottom: 16px;
  }
  .product-edit-right {
    padding-left: 0;
  }
}

/* 规格操作行 */
.spec-action-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.spec-status-text {
  font-size: 12px;
  color: var(--td-success-color);
  white-space: nowrap;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
/* ==================== 产品编辑弹窗结束 ==================== */

.form-tip {
  font-size: 12px;
  color: var(--td-text-color-secondary);
  margin-top: 4px;
}
</style>
