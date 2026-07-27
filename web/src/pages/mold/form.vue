<template>
  <div class="mold-form-page">
    <div class="back-button-container">
      <t-button theme="default" @click="goBack" variant="text">
        <template #icon><t-icon name="chevron-left" /></template>
        返回
      </t-button>
    </div>
    <t-loading :loading="loading" size="small" text="加载中...">
    <t-form
      ref="formRef"
      class="mold-form"
    :data="formModel"
    :rules="activeRules"
    :label-width="100"
    label-align="top"
    @reset="onReset"
    @submit="onSubmit"
  >
    <div class="form-basic-container">
      <div class="form-basic-item">
        <div class="form-basic-container-title">模具信息</div>
        
        <!-- 模具基础信息 -->
        <t-row class="row-gap" :gutter="[32, 24]">
          <t-col :span="6">
            <t-form-item label="模具编号" name="moldCode">
              <t-input 
                v-model="formModel.moldCode" 
                placeholder="请输入模具编号" 
                :readonly="isEdit && !moldCodeUnlocked"
                :class="{ 'locked-input': isEdit && !moldCodeUnlocked }"
              >
                <template #suffix v-if="isEdit">
                  <t-button 
                    v-if="!moldCodeUnlocked" 
                    theme="default" 
                    size="small" 
                    @click="unlockMoldCode"
                    style="margin-left: 8px"
                  >
                    <template #icon><lock-off-icon /></template>
                    解锁
                  </t-button>
                  <t-button 
                    v-else 
                    theme="default" 
                    size="small" 
                    @click="lockMoldCode"
                    style="margin-left: 8px"
                  >
                    <template #icon><lock-on-icon /></template>
                    锁定
                  </t-button>
                </template>
              </t-input>
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="产品名称" name="moldName">
              <t-input v-model="formModel.moldName" placeholder="请输入产品名称" />
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="客户" name="customerId">
              <t-select v-model="formModel.customerId" placeholder="请选择客户" :clearable="false">
                <t-option 
                  v-for="customer in customerList" 
                  :key="customer.id" 
                  :label="customer.cus_name" 
                  :value="Number(customer.id)" 
                />
              </t-select>
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="开始生产日期" name="productionDate">
              <t-date-picker 
                v-model="formModel.productionDate" 
                placeholder="请选择开始生产日期" 
                format="YYYY-MM-DD"
                style="width: 100%"
              />
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="生命周期状态" name="lifeCycleStatus">
              <t-select v-model="formModel.lifeCycleStatus" placeholder="请选择状态" :clearable="false">
                <t-option
                  v-for="item in processStatusList"
                  :key="item.dict_value"
                  :label="item.dict_label"
                  :value="String(item.dict_value)"
                />
              </t-select>
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="模胚情况" name="moldType">
              <t-select v-model="formModel.moldType" placeholder="请选择类型">
                <t-option
                  v-for="item in moldTypeList"
                  :key="item.dict_value"
                  :label="item.dict_label"
                  :value="item.dict_value"
                />
              </t-select>
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="模具工艺" name="moldProcess">
              <t-select v-model="formModel.moldProcess" placeholder="请选择工艺">
                <t-option
                  v-for="item in moldProcessList"
                  :key="item.dict_value"
                  :label="item.dict_label"
                  :value="item.dict_value"
                />
              </t-select>
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="材质" name="materialType">
              <t-select v-model="formModel.materialType" placeholder="请选择材质">
                <t-option
                  v-for="item in materialTypeList"
                  :key="item.dict_value"
                  :label="item.dict_label"
                  :value="item.dict_value"
                />
              </t-select>
            </t-form-item>
          </t-col>
          <t-col :span="6">
            <t-form-item label="模腔数" name="cavityCount">
              <t-select
                v-model="formModel.cavityCount"
                placeholder="请选择模腔数"
                :options="cavityCountOptions"
              />
            </t-form-item>
          </t-col>
        </t-row>

        <!-- 模具图片 -->
        <div class="form-basic-container-title form-title-gap">模具图片</div>
        
        <t-form-item>
          <div class="upload-section-wrapper">
            <FileUploader
              ref="imageUploaderRef"
              v-model="fileUploadList"
              :auto-save="false"
              :support-primary="true"
              :image-only="true"
              :max-count="9"
              :multiple="true"
              category="mold"
              show-empty-hint
              empty-text="暂无模具图片"
              upload-btn-text="上传模具图片"
            />
          </div>
        </t-form-item>

        <!-- 备注 -->
        <div class="form-basic-container-title form-title-gap">其他信息</div>
        
        <t-form-item label="备注">
          <t-textarea v-model="formModel.remark" :height="124" placeholder="请输入备注" />
        </t-form-item>
      </div>
    </div>

    <div class="form-submit-container">
      <div class="form-submit-sub">
        <div class="form-submit-left">
          <t-button theme="primary" class="form-submit-confirm" type="submit">
            {{ isEdit ? '更新' : '确认' }}
          </t-button>
          <t-button type="reset" class="form-submit-cancel" theme="default" variant="base">
            取消
          </t-button>
        </div>
      </div>
    </div>
  </t-form>
  </t-loading>
  </div>
</template>

<script setup lang="ts">
import type { FormRule, SubmitContext } from 'tdesign-vue-next';
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { MessagePlugin } from 'tdesign-vue-next';
import { LockOffIcon, LockOnIcon } from 'tdesign-icons-vue-next';
import { createMold, updateMold, getMoldById } from '@/api/mold';
import { getCustomerList, getDictTypes } from '@/api/reference';
import { getMoldImages, attachMoldFiles, deleteMoldImage, setPrimaryImage } from '@/api/mold/image';
import { uploadFiles } from '@/api/common/upload';
import FileUploader from '@/components/common/FileUploader.vue';

defineOptions({
  name: 'MoldForm',
});

const route = useRoute();
const router = useRouter();
const formRef = ref();
const imageUploaderRef = ref(); // 图片上传组件引用
const isEdit = ref(false);
const isSubmitting = ref(false); // 防止重复提交
const loading = ref(true); // 初始为 true，避免表单首次渲染时触发验证

// 客户列表
const customerList = ref<any[]>([]);

// 字典数据列表
const processStatusList = ref<any[]>([]); // 工序状态
const moldTypeList = ref<any[]>([]); // 模胚情况
const moldProcessList = ref<any[]>([]); // 模具工艺
const materialTypeList = ref<any[]>([]); // 材质

// 模具编号解锁状态
const moldCodeUnlocked = ref(false);

// 模具图片列表（FileUploader 的 UploadedFileItem 格式）
const fileUploadList = ref<any[]>([]);

// 保存原始数据，用于对比是否有变化
const originalData = ref<any>(null);

// 直接使用语义化字符串，不需要转换函数

// 表单数据（使用 reactive + :data 标准模式，参考 MoldOrderForm）
const formModel = reactive({
  id: undefined as number | undefined,
  moldCode: '',
  moldName: '',
  customerId: undefined as number | undefined,
  productionDate: '',
  lifeCycleStatus: 'DESIGNING',
  moldType: undefined as string | undefined, // 必选，不设默认值
  moldProcess: undefined as string | undefined, // 必选，不设默认值
  materialType: undefined as string | undefined, // 必选，不设默认值
  cavityCount: undefined as number | undefined,
  remark: '',
});

// 模腔数下拉选项（1-30）
const cavityCountOptions = Array.from({ length: 30 }, (_, i) => ({ label: `${i + 1} 腔`, value: i + 1 }));

// 表单验证规则（参考综合订单 OrderInfoForm 标准模式）
const FORM_RULES: Record<string, FormRule[]> = {
  moldCode: [{ required: true, message: '请输入模具编号', type: 'error' }],
  moldName: [{ required: true, message: '请输入产品名称', type: 'error' }],
  customerId: [{ required: true, message: '请选择客户', type: 'error' }],
  lifeCycleStatus: [{ required: true, message: '请选择生命周期状态', type: 'error' }],
  productionDate: [{ required: true, message: '请选择开始生产日期', type: 'error' }],
  moldType: [{ required: true, message: '请选择模胚情况', type: 'error' }],
  moldProcess: [{ required: true, message: '请选择模具工艺', type: 'error' }],
  materialType: [{ required: true, message: '请选择材质', type: 'error' }],
  cavityCount: [{ required: true, message: '请选择模腔数', type: 'error' }],
};

// 加载期间禁用验证，避免空字段触发红色边框
const activeRules = computed(() => (loading.value ? {} : FORM_RULES));

// 提交表单（统一处理新增和编辑，参考 MoldOrderForm 标准模式）
const onSubmit = async (ctx: SubmitContext) => {
  // 防止重复提交
  if (isSubmitting.value) {
    return;
  }

  if (ctx.validateResult === true) {
    try {
      isSubmitting.value = true;

      if (isEdit.value) {
        // ✨ 编辑模式逻辑
        const currentData: any = {
          id: formModel.id,
          mold_code: formModel.moldCode,
          mold_name: formModel.moldName,
          customer_id: formModel.customerId,
          production_date: formModel.productionDate,
          life_cycle_status: formModel.lifeCycleStatus,
          mold_type: formModel.moldType,
          mold_process: formModel.moldProcess,
          material_type: formModel.materialType,
          cavity_count: formModel.cavityCount,
          remark: formModel.remark,
        };

        // 对比数据是否有变化
        const hasChanges = compareData(originalData.value, currentData);
        
        let hasImageChanges = false;
        if (imageUploaderRef.value) {
          const pendingChanges = imageUploaderRef.value.getPendingChanges();
          hasImageChanges = pendingChanges.hasChanges;
        }
        
        if (!hasChanges && !hasImageChanges) {
          MessagePlugin.info('数据未发生变化，无需更新');
          isSubmitting.value = false;
          return;
        }

        // 1. 先更新模具基本信息
        if (hasChanges) {
          await updateMold(currentData);
        }

        // 2. 更新图片（如果有变更）
        if (hasImageChanges && imageUploaderRef.value && formModel.id) {
          const pending = imageUploaderRef.value.getPendingChanges();

          try {
            // 2a. 上传新文件（统一上传接口）
            let uploadedFileIds: string[] = [];
            if (pending.uploads.length > 0) {
              const results = await uploadFiles({
                files: pending.uploads,
                category: 'mold',
              });
              uploadedFileIds = results.map(r => r.id);
              // 关联到模具
              await attachMoldFiles(formModel.id, uploadedFileIds);
            }

            // 2b. 删除文件
            for (const idx of pending.deleteIndices) {
              const fileId = fileUploadList.value[idx]?.id;
              if (fileId) {
                await deleteMoldImage(fileId);
              }
            }

            // 2c. 设置主图（仅针对已有的文件）
            if (pending.primaryKey) {
              const match = (pending.primaryKey as string).match(/^uploaded-(\d+)$/);
              if (match) {
                const fileId = fileUploadList.value[parseInt(match[1])]?.id;
                if (fileId) {
                  await setPrimaryImage(formModel.id, fileId);
                }
              }
            }

            // 重新加载图片列表
            await loadMoldImages();
            imageUploaderRef.value?.resetPendingChanges();
          } catch (error: any) {
            console.error('图片更新失败:', error);
            MessagePlugin.warning('模具信息已更新，但图片更新失败，请重试');
            isSubmitting.value = false;
            return;
          }
        }

        MessagePlugin.success('更新成功');
      } else {
        // ✨ 新增模式逻辑
        const submitData: any = {
          mold_code: formModel.moldCode,
          mold_name: formModel.moldName,
          customer_id: formModel.customerId,
          production_date: formModel.productionDate,
          life_cycle_status: formModel.lifeCycleStatus,
          mold_type: formModel.moldType,
          mold_process: formModel.moldProcess,
          material_type: formModel.materialType,
          cavity_count: formModel.cavityCount,
          remark: formModel.remark,
        };

        const res = await createMold(submitData);
        const newMoldId = Number(res?.id || res?.data?.id);
        if (!newMoldId) {
          MessagePlugin.error('创建模具后获取ID失败，图片将无法关联');
        } else {
          formModel.id = newMoldId;

          // 处理图片上传（如果有选中的图片）
          if (imageUploaderRef.value) {
            const pendingChanges = imageUploaderRef.value.getPendingChanges();
            if (pendingChanges.uploads.length > 0) {
              try {
                const results = await uploadFiles({
                  files: pendingChanges.uploads,
                  category: 'mold',
                });
                const uploadedFileIds = results.map(r => r.id);
                await attachMoldFiles(newMoldId, uploadedFileIds);
              } catch (uploadError: any) {
                console.error('新建模具图片上传失败:', uploadError);
                MessagePlugin.warning('模具已创建，但图片上传失败，可后续编辑时重新上传');
              }
            }
          }
        }

        MessagePlugin.success('创建成功');

        // 新增成功后立即重置表单，防止 KeepAlive 缓存导致下次进入时显示旧数据
        formModel.id = undefined;
        formModel.moldCode = '';
        formModel.moldName = '';
        formModel.customerId = undefined;
        formModel.productionDate = '';
        formModel.lifeCycleStatus = 'DESIGNING';
        formModel.moldType = undefined;
        formModel.moldProcess = undefined;
        formModel.materialType = undefined;
        formModel.cavityCount = undefined;
        formModel.remark = '';
        fileUploadList.value = [];
        moldCodeUnlocked.value = false;
        originalData.value = null;
      }

      router.push('/mold/index');
    } catch (error: any) {
      console.error('提交失败:', error);
      if (error?.message) {
        MessagePlugin.error(error.message);
      } else {
        MessagePlugin.error(isEdit.value ? '更新失败' : '创建失败');
      }
    } finally {
      isSubmitting.value = false;
    }
  }
};

const onReset = () => {
  // 编辑模式下，清理图片上传组件的待处理变更
  if (isEdit.value && imageUploaderRef.value) {
    imageUploaderRef.value.resetPendingChanges();
  }
  MessagePlugin.warning('取消操作');
  router.push('/mold/index');
};

// 解锁模具编号
const unlockMoldCode = () => {
  moldCodeUnlocked.value = true;
  formRef.value?.clearValidate?.('moldCode');
  MessagePlugin.warning('模具编号已解锁，请谨慎修改');
};

// 锁定模具编号
const lockMoldCode = () => {
  moldCodeUnlocked.value = false;
  formRef.value?.clearValidate?.('moldCode');
  MessagePlugin.success('模具编号已锁定');
};

const goBack = () => {
  router.back();
};

// 对比数据是否有变化
const compareData = (original: any, current: any): boolean => {
  if (!original || !current) {
    return true;
  }

  // 需要对比的字段列表
  const fieldsToCompare = [
    'mold_code',
    'mold_name',
    'customer_id',
    'production_date',
    'life_cycle_status',
    'mold_type',
    'mold_process',
    'material_type',
    'cavity_count',
    'remark'
  ];

  for (const field of fieldsToCompare) {
    const originalValue = original[field];
    const currentValue = current[field];

    if (originalValue === null && currentValue === undefined) continue;
    if (originalValue === undefined && currentValue === null) continue;
    if (originalValue === null && currentValue === null) continue;
    if (originalValue === undefined && currentValue === undefined) continue;

    // 比较值是否相同
    if (originalValue !== currentValue) {
      return true;
    }
  }

  return false;
};

const loadMoldImages = async () => {
  if (!formModel.id) return;
  try {
    const images = await getMoldImages(formModel.id);
    fileUploadList.value = (images || []).map((img: any) => ({
      id: String(img.id),
      key: `uploaded-${img.id}`,
      url: img.url,
      fileName: img.file_name || '',
      fileSize: img.file_size || 0,
      fileType: img.file_type || '.jpg',
      isPrimary: img.is_primary === 1,
    }));
  } catch (error) {
    console.error('加载模具图片列表失败:', error);
  }
};

const loadMoldDetail = async (id: number) => {
  try {
    const res = await getMoldById(id);
    if (res) {
      // 填充数据
      formModel.id = res.id;
      formModel.moldCode = res.mold_code || '';
      formModel.moldName = res.mold_name || '';
      formModel.customerId = res.customer_id ? Number(res.customer_id) : undefined;
      formModel.productionDate = res.production_date || '';
      formModel.lifeCycleStatus = res.life_cycle_status !== undefined && res.life_cycle_status !== null 
        ? String(res.life_cycle_status) 
        : 'DESIGNING';
      formModel.moldType = res.mold_type || 'NEW';
      formModel.moldProcess = res.mold_process || 'NORMAL';
      formModel.materialType = res.material_type || 'STEEL';
      formModel.cavityCount = res.cavity_count || undefined;
      formModel.remark = res.remark || '';
      
      // 保存原始数据用于后续对比
      originalData.value = {
        id: res.id,
        mold_code: res.mold_code || '',
        mold_name: res.mold_name || '',
        customer_id: res.customer_id ? Number(res.customer_id) : undefined,
        production_date: res.production_date || '',
        life_cycle_status: res.life_cycle_status !== undefined && res.life_cycle_status !== null 
          ? String(res.life_cycle_status) 
          : 'DESIGNING',
        mold_type: res.mold_type || 'NEW',
        mold_process: res.mold_process || 'NORMAL',
        material_type: res.material_type || 'STEEL',
        cavity_count: res.cavity_count || undefined,
        remark: res.remark || '',
      };
      
      await loadMoldImages();
    }
  } catch (error) {
    console.error('加载模具详情失败:', error);
    MessagePlugin.error('加载模具详情失败');
  }
};

const loadCustomerList = async () => {
  try {
    const res = await getCustomerList();
    customerList.value = res.list || [];
  } catch (error) {
    console.error('获取客户列表失败:', error);
  }
};

const loadDictData = async () => {
  try {
    const res = await getDictTypes();
    const dictList = res.list || [];
    
    // 根据字典编码分类存储
    dictList.forEach((dict: any) => {
      switch (dict.dict_code) {
        case 'mold-status':  // 模具状态（工序状态）
          processStatusList.value = dict.data_list || [];
          break;
        case 'mold-type':  // 模胚情况
          moldTypeList.value = dict.data_list || [];
          break;
        case 'mold-process':  // 模具工艺
          moldProcessList.value = dict.data_list || [];
          break;
        case 'material-type':  // 材质
          materialTypeList.value = dict.data_list || [];
          break;
      }
    });
  } catch (error) {
    console.error('加载字典数据失败:', error);
  }
};

const initEditMode = async (idParam: string | undefined) => {
  // 先开启 loading 禁用验证，避免切换模式时重置字段触发 validateOnChange
  loading.value = true;
  if (idParam) {
    isEdit.value = true;
    formModel.id = undefined;
    formModel.moldCode = '';
    formModel.moldName = '';
    formModel.customerId = undefined;
    formModel.productionDate = '';
    formModel.lifeCycleStatus = 'DESIGNING';
    formModel.moldType = undefined;
    formModel.moldProcess = undefined;
    formModel.materialType = undefined;
    formModel.cavityCount = undefined;
    formModel.remark = '';
    fileUploadList.value = [];
    moldCodeUnlocked.value = false;
    originalData.value = null;
    // 编辑模式：加载数据（保证加载动画至少显示 300ms，避免一闪而过）
    await Promise.all([
      loadMoldDetail(Number(idParam)),
      new Promise((r) => setTimeout(r, 300)),
    ]);
    // 编辑模式：动态更新页面标题
    route.meta.title = { zh_CN: '编辑模具档案', en_US: 'Edit Mold Archive' };
    loading.value = false;
    await nextTick();
    formRef.value?.clearValidate?.();
  } else {
    isEdit.value = false;
    // 新增模式：重置表单状态
    formModel.id = undefined;
    formModel.moldCode = '';
    formModel.moldName = '';
    formModel.customerId = undefined;
    formModel.productionDate = '';
    formModel.lifeCycleStatus = 'DESIGNING';
    formModel.moldType = undefined;
    formModel.moldProcess = undefined;
    formModel.materialType = undefined;
    formModel.cavityCount = undefined;
    formModel.remark = '';
    fileUploadList.value = [];
    moldCodeUnlocked.value = false;
    originalData.value = null;
    route.meta.title = { zh_CN: '新增模具档案', en_US: 'New Mold Archive' };
    loading.value = false;
    await nextTick();
    formRef.value?.clearValidate?.();
  }
};

onMounted(async () => {
  await loadCustomerList();
  await loadDictData();

  const idParam = route.query.id;
  await initEditMode(typeof idParam === 'string' ? idParam : undefined);
});

// 监听路由 query.id 变化，当从 A 模具编辑切换到 B 模具编辑时重新加载数据
watch(
  () => route.query.id,
  async (newId, oldId) => {
    // 仅在已挂载后、且 id 参数确实变化时重新加载
    if (newId !== oldId) {
      await initEditMode(typeof newId === 'string' ? newId : undefined);
    }
  },
);

</script>

<style lang="less" scoped>
@import '@/pages/form/base/index.less';

/* 上传占位提示样式 */
.upload-placeholder-hint {
  width: 100%;
}

/* 锁定状态下的输入框样式 */
.locked-input :deep(.t-input__inner) {
  background-color: var(--td-bg-color-container, #f5f5f5);
  cursor: not-allowed;
  color: var(--td-text-color-disabled, rgba(0, 0, 0, 0.36));
}

/* 上传区域包装器 */
.upload-section-wrapper {
  width: 100%;
}

.back-button-container {
  margin-bottom: 16px;
}

</style>
