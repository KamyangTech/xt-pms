<template>
  <div class="basic-data-container">
    <!-- 操作栏 -->
    <t-card class="toolbar-card">
      <div class="toolbar-content">
        <div class="toolbar-left">
          <t-tabs v-model="activeTab" @change="handleTabChange">
            <t-tab-panel value="customer" label="客户信息" />
            <t-tab-panel value="processor" label="加工商管理" />
            <t-tab-panel value="supplier" label="供应商管理" />
            <t-tab-panel value="warehouse" label="仓库管理" />
          </t-tabs>
        </div>
        <div class="toolbar-actions">
          <t-button theme="primary" @click="handleAdd">
            <template #icon><add-icon /></template>
            {{ addBtnLabel }}
          </t-button>
          <t-button variant="outline" @click="handleRefresh">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <!-- 客户表格 -->
      <t-table
        v-if="activeTab === 'customer'"
        :data="customerTableData"
        :columns="customerColumns"
        :loading="customerLoading"
        row-key="id"
        :stripe="true"
      >
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '正常' : '停用' }}</t-tag>
        </template>
        <template #create_time="{ row }">{{ formatDateTime(row.create_time) }}</template>
        <template #update_time="{ row }">{{ row.update_time ? formatDateTime(row.update_time) : '-' }}</template>
        <template #operation="{ row }">
          <t-link theme="primary" @click="handleCustomerEdit(row)" style="margin-right: 8px">编辑</t-link>
          <t-popconfirm content="确认删除该客户？" @confirm="handleCustomerDelete(row)">
            <t-link theme="danger">删除</t-link>
          </t-popconfirm>
        </template>
      </t-table>

      <!-- 加工商表格 -->
      <t-table
        v-if="activeTab === 'processor'"
        :data="processorTableData"
        :columns="processorColumns"
        :loading="processorLoading"
        row-key="id"
        :stripe="true"
      >
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '正常' : '停用' }}</t-tag>
        </template>
        <template #create_time="{ row }">{{ formatDateTime(row.create_time) }}</template>
        <template #operation="{ row }">
          <t-link theme="primary" @click="handleProcessorEdit(row)" style="margin-right: 8px">编辑</t-link>
          <t-popconfirm content="确认删除该加工商？" @confirm="handleProcessorDelete(row)">
            <t-link theme="danger">删除</t-link>
          </t-popconfirm>
        </template>
      </t-table>

      <!-- 供应商表格 -->
      <t-table
        v-if="activeTab === 'supplier'"
        :data="supplierTableData"
        :columns="supplierColumns"
        :loading="supplierLoading"
        row-key="id"
        :stripe="true"
      >
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '正常' : '停用' }}</t-tag>
        </template>
        <template #create_time="{ row }">{{ formatDateTime(row.create_time) }}</template>
        <template #operation="{ row }">
          <t-link theme="primary" @click="handleSupplierEdit(row)" style="margin-right: 8px">编辑</t-link>
          <t-popconfirm content="确认删除该供应商？" @confirm="handleSupplierDelete(row)">
            <t-link theme="danger">删除</t-link>
          </t-popconfirm>
        </template>
      </t-table>

      <!-- 仓库表格 -->
      <t-table
        v-if="activeTab === 'warehouse'"
        :data="warehouseTableData"
        :columns="warehouseColumns"
        :loading="warehouseLoading"
        row-key="id"
        :stripe="true"
      >
        <template #status="{ row }">
          <t-tag :theme="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '正常' : '停用' }}</t-tag>
        </template>
        <template #create_time="{ row }">{{ formatDateTime(row.create_time) }}</template>
        <template #update_time="{ row }">{{ row.update_time ? formatDateTime(row.update_time) : '-' }}</template>
        <template #operation="{ row }">
          <t-link theme="primary" @click="handleWarehouseEdit(row)" style="margin-right: 8px">编辑</t-link>
          <t-popconfirm content="确认删除该仓库？" @confirm="handleWarehouseDelete(row)">
            <t-link theme="danger">删除</t-link>
          </t-popconfirm>
        </template>
      </t-table>
    </t-card>

    <!-- ========== 客户编辑对话框 ========== -->
    <t-dialog v-model:visible="customerDialogVisible" :title="customerDialogTitle" width="500px" placement="center" :close-on-overlay-click="false">
      <t-form ref="customerFormRef" :data="customerForm" :rules="customerRules" label-width="100px">
        <t-form-item label="客户编号" name="cusCode"><t-input v-model="customerForm.cusCode" placeholder="请输入客户编号" /></t-form-item>
        <t-form-item label="客户名称" name="cusName"><t-input v-model="customerForm.cusName" placeholder="请输入客户名称" /></t-form-item>
        <t-form-item label="联系人" name="contact"><t-input v-model="customerForm.contact" placeholder="请输入联系人" /></t-form-item>
        <t-form-item label="联系电话" name="phone"><t-input v-model="customerForm.phone" placeholder="请输入联系电话" /></t-form-item>
        <t-form-item label="电子邮箱" name="email"><t-input v-model="customerForm.email" placeholder="请输入电子邮箱" /></t-form-item>
        <t-form-item label="客户地址" name="address"><t-input v-model="customerForm.address" placeholder="请输入客户地址" /></t-form-item>
        <t-form-item label="备注" name="remark"><t-textarea v-model="customerForm.remark" placeholder="请输入备注" :rows="3" /></t-form-item>
        <t-form-item label="状态" name="status">
          <t-radio-group v-model="customerForm.status"><t-radio :value="1">正常</t-radio><t-radio :value="0">停用</t-radio></t-radio-group>
        </t-form-item>
      </t-form>
      <template #footer>
        <t-button theme="default" @click="customerDialogVisible = false">取消</t-button>
        <t-button theme="primary" style="margin-left: 8px" @click="handleCustomerSubmit">确定</t-button>
      </template>
    </t-dialog>

    <!-- ========== 加工商编辑对话框 ========== -->
    <t-dialog v-model:visible="processorDialogVisible" :header="processorDialogTitle" width="500px" placement="center" :close-on-overlay-click="false">
      <t-form ref="processorFormRef" :data="processorForm" :rules="processorRules" label-width="100px">
        <t-form-item label="加工商编号" name="processorCode"><t-input v-model="processorForm.processorCode" placeholder="请输入加工商编号" /></t-form-item>
        <t-form-item label="加工商名称" name="processorName"><t-input v-model="processorForm.processorName" placeholder="请输入加工商名称" /></t-form-item>
        <t-form-item label="联系人" name="contact"><t-input v-model="processorForm.contact" placeholder="请输入联系人" /></t-form-item>
        <t-form-item label="联系电话" name="phone"><t-input v-model="processorForm.phone" placeholder="请输入联系电话" /></t-form-item>
        <t-form-item label="电子邮箱" name="email"><t-input v-model="processorForm.email" placeholder="请输入邮箱" /></t-form-item>
        <t-form-item label="地址" name="address"><t-input v-model="processorForm.address" placeholder="请输入地址" /></t-form-item>
        <t-form-item label="加工类型" name="processType"><t-input v-model="processorForm.processType" placeholder="如：CNC加工" /></t-form-item>
        <t-form-item label="备注" name="remark"><t-input v-model="processorForm.remark" placeholder="备注" /></t-form-item>
        <t-form-item label="状态" name="status">
          <t-radio-group v-model="processorForm.status"><t-radio :value="1">正常</t-radio><t-radio :value="0">停用</t-radio></t-radio-group>
        </t-form-item>
      </t-form>
      <template #footer>
        <t-button theme="default" @click="processorDialogVisible = false">取消</t-button>
        <t-button theme="primary" style="margin-left: 8px" @click="handleProcessorSubmit">确定</t-button>
      </template>
    </t-dialog>

    <!-- ========== 供应商编辑对话框 ========== -->
    <t-dialog v-model:visible="supplierDialogVisible" :header="supplierDialogTitle" width="500px" placement="center" :close-on-overlay-click="false">
      <t-form ref="supplierFormRef" :data="supplierForm" :rules="supplierRules" label-width="100px">
        <t-form-item label="供应商编号" name="supCode"><t-input v-model="supplierForm.supCode" placeholder="请输入供应商编号" /></t-form-item>
        <t-form-item label="供应商名称" name="supName"><t-input v-model="supplierForm.supName" placeholder="请输入供应商名称" /></t-form-item>
        <t-form-item label="联系人" name="contact"><t-input v-model="supplierForm.contact" placeholder="请输入联系人" /></t-form-item>
        <t-form-item label="联系电话" name="phone"><t-input v-model="supplierForm.phone" placeholder="请输入联系电话" /></t-form-item>
        <t-form-item label="电子邮箱" name="email"><t-input v-model="supplierForm.email" placeholder="请输入邮箱" /></t-form-item>
        <t-form-item label="地址" name="address"><t-input v-model="supplierForm.address" placeholder="请输入地址" /></t-form-item>
        <t-form-item label="备注" name="remark"><t-textarea v-model="supplierForm.remark" placeholder="请输入备注" :rows="3" /></t-form-item>
        <t-form-item label="状态" name="status">
          <t-radio-group v-model="supplierForm.status"><t-radio :value="1">正常</t-radio><t-radio :value="0">停用</t-radio></t-radio-group>
        </t-form-item>
      </t-form>
      <template #footer>
        <t-button theme="default" @click="supplierDialogVisible = false">取消</t-button>
        <t-button theme="primary" style="margin-left: 8px" @click="handleSupplierSubmit">确定</t-button>
      </template>
    </t-dialog>

    <!-- ========== 仓库编辑对话框 ========== -->
    <t-dialog v-model:visible="warehouseDialogVisible" :title="warehouseDialogTitle" width="500px" placement="center" :close-on-overlay-click="false">
      <t-form ref="warehouseFormRef" :data="warehouseForm" :rules="warehouseRules" label-width="100px">
        <t-form-item label="仓库编码" name="whCode"><t-input v-model="warehouseForm.whCode" placeholder="请输入仓库编码" :disabled="warehouseIsEdit" /></t-form-item>
        <t-form-item label="仓库名称" name="whName"><t-input v-model="warehouseForm.whName" placeholder="请输入仓库名称" /></t-form-item>
        <t-form-item label="仓库地址" name="address"><t-input v-model="warehouseForm.address" placeholder="请输入仓库地址" /></t-form-item>
        <t-form-item label="管理员" name="manager"><t-input v-model="warehouseForm.manager" placeholder="请输入管理员" /></t-form-item>
        <t-form-item label="联系电话" name="phone"><t-input v-model="warehouseForm.phone" placeholder="请输入联系电话" /></t-form-item>
        <t-form-item label="备注" name="remark"><t-textarea v-model="warehouseForm.remark" placeholder="请输入备注" :rows="3" /></t-form-item>
        <t-form-item label="状态" name="status">
          <t-radio-group v-model="warehouseForm.status"><t-radio :value="1">正常</t-radio><t-radio :value="0">停用</t-radio></t-radio-group>
        </t-form-item>
      </t-form>
      <template #footer>
        <t-button theme="default" @click="warehouseDialogVisible = false">取消</t-button>
        <t-button theme="primary" style="margin-left: 8px" @click="handleWarehouseSubmit">确定</t-button>
      </template>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { TabValue } from 'tdesign-vue-next';
import { AddIcon, RefreshIcon } from 'tdesign-icons-vue-next';

import { getCustomerList, createCustomer, updateCustomer, deleteCustomer } from '@/api/customer';
import { getProcessorList, createProcessor, updateProcessor, deleteProcessor } from '@/api/processor';
import { getSupplierList, createSupplier, updateSupplier, deleteSupplier } from '@/api/supplier';
import { getWarehouseList, createWarehouse, updateWarehouse, deleteWarehouse } from '@/api/warehouse';
import { formatDateTime } from '@/utils/date';

defineOptions({ name: 'BasicData' });

const phoneReg = /^1[3-9]\d{9}$/;
const emailReg = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

const activeTab = ref('customer');

const addBtnLabel = computed(() => {
  const labels: Record<string, string> = {
    customer: '新增客户',
    processor: '新增加工商',
    supplier: '新增供应商',
    warehouse: '新增仓库',
  };
  return labels[activeTab.value] || '新增';
});

const handleTabChange = (value: TabValue) => {
  activeTab.value = value as string;
  loadTabData();
};

const handleAdd = () => {
  switch (activeTab.value) {
    case 'customer': handleCustomerAdd(); break;
    case 'processor': handleProcessorAdd(); break;
    case 'supplier': handleSupplierAdd(); break;
    case 'warehouse': handleWarehouseAdd(); break;
  }
};

const handleRefresh = () => loadTabData();

// ==================== 客户 ====================
const customerLoading = ref(false);
const customerTableData = ref<any[]>([]);
const customerDialogVisible = ref(false);
const customerDialogTitle = ref('新增客户');
const customerIsEdit = ref(false);
const customerFormRef = ref();

const customerColumns = [
  { colKey: 'cusCode', title: '客户编号', width: 140, fixed: 'left' as const },
  { colKey: 'cusName', title: '客户名称', width: 150 },
  { colKey: 'contact', title: '联系人', width: 100 },
  { colKey: 'phone', title: '联系电话', width: 140 },
  { colKey: 'email', title: '电子邮箱', width: 180 },
  { colKey: 'address', title: '地址', width: 200, ellipsis: true as const },
  { colKey: 'status', title: '状态', width: 80 },
  { colKey: 'create_time', title: '创建时间', width: 160 },
  { colKey: 'operation', title: '操作', width: 120, fixed: 'right' as const },
];

const customerForm = reactive({
  id: undefined as number | undefined,
  cusCode: '', cusName: '', contact: '', phone: '', email: '', address: '', remark: '', status: 1,
});

const customerRules = {
  cusCode: [{ required: true, message: '请输入客户编号', trigger: 'blur' as const }],
  cusName: [{ required: true, message: '请输入客户名称', trigger: 'blur' as const }],
  phone: [{ pattern: phoneReg, message: '请输入正确的手机号', trigger: 'blur' as const }],
  email: [{ pattern: emailReg, message: '请输入正确的邮箱地址', trigger: 'blur' as const }],
};

const loadCustomerData = async () => {
  customerLoading.value = true;
  try {
    const res = await getCustomerList({ page: 1, page_size: 9999 });
    customerTableData.value = (res.list || []).map((item: any) => ({
      ...item, cusCode: item.cus_code, cusName: item.cus_name,
    }));
  } catch { MessagePlugin.error('获取客户列表失败'); }
  finally { customerLoading.value = false; }
};

const handleCustomerAdd = () => {
  customerIsEdit.value = false;
  customerDialogTitle.value = '新增客户';
  Object.assign(customerForm, { id: undefined, cusCode: '', cusName: '', contact: '', phone: '', email: '', address: '', remark: '', status: 1 });
  customerFormRef.value?.reset();
  customerDialogVisible.value = true;
};

const handleCustomerEdit = (row: any) => {
  customerIsEdit.value = true;
  customerDialogTitle.value = '编辑客户';
  Object.assign(customerForm, {
    id: row.id, cusCode: row.cusCode, cusName: row.cusName, contact: row.contact,
    phone: row.phone, email: row.email, address: row.address, remark: row.remark, status: row.status,
  });
  setTimeout(() => customerFormRef.value?.clearValidate?.(), 100);
  customerDialogVisible.value = true;
};

const handleCustomerDelete = async (row: any) => {
  try { await deleteCustomer(row.id); MessagePlugin.success('删除成功'); loadCustomerData(); } catch { MessagePlugin.error('删除失败'); }
};

const handleCustomerSubmit = async () => {
  const valid = await customerFormRef.value?.validate();
  if (valid !== true) return;
  try {
    const data: any = { cus_code: customerForm.cusCode, cus_name: customerForm.cusName, contact: customerForm.contact, phone: customerForm.phone, email: customerForm.email, address: customerForm.address, remark: customerForm.remark, status: customerForm.status };
    if (customerIsEdit.value && customerForm.id) { data.id = customerForm.id; await updateCustomer(data); MessagePlugin.success('更新成功'); }
    else { await createCustomer(data); MessagePlugin.success('创建成功'); }
    customerDialogVisible.value = false;
    loadCustomerData();
  } catch { MessagePlugin.error(customerIsEdit.value ? '更新失败' : '创建失败'); }
};

// ==================== 加工商 ====================
const processorLoading = ref(false);
const processorTableData = ref<any[]>([]);
const processorDialogVisible = ref(false);
const processorDialogTitle = ref('新增加工商');
const processorIsEdit = ref(false);
const processorFormRef = ref();

const processorColumns = [
  { colKey: 'processor_code', title: '加工商编号', width: 140, fixed: 'left' as const },
  { colKey: 'processor_name', title: '加工商名称', width: 150 },
  { colKey: 'process_type', title: '加工类型', width: 120 },
  { colKey: 'contact', title: '联系人', width: 100 },
  { colKey: 'phone', title: '联系电话', width: 140 },
  { colKey: 'address', title: '地址', width: 180, ellipsis: true as const },
  { colKey: 'status', title: '状态', width: 80 },
  { colKey: 'create_time', title: '创建时间', width: 160 },
  { colKey: 'operation', title: '操作', width: 120, fixed: 'right' as const },
];

const processorForm = reactive({
  id: null as number | null,
  processorCode: '', processorName: '', contact: '', phone: '', email: '',
  address: '', processType: '', remark: '', status: 1,
});

const processorRules = {
  processorCode: [{ required: true, message: '请输入加工商编号', trigger: 'blur' as const }],
  processorName: [{ required: true, message: '请输入加工商名称', trigger: 'blur' as const }],
  phone: [{ pattern: phoneReg, message: '请输入正确的手机号', trigger: 'blur' as const }],
  email: [{ pattern: emailReg, message: '请输入正确的邮箱地址', trigger: 'blur' as const }],
};

const loadProcessorData = async () => {
  processorLoading.value = true;
  try {
    const res = await getProcessorList({ page: 1, page_size: 9999 });
    processorTableData.value = res.list || [];
  } catch { MessagePlugin.error('加载加工商数据失败'); }
  finally { processorLoading.value = false; }
};

const handleProcessorAdd = () => {
  processorIsEdit.value = false;
  processorDialogTitle.value = '新增加工商';
  Object.assign(processorForm, { id: null, processorCode: '', processorName: '', contact: '', phone: '', email: '', address: '', processType: '', remark: '', status: 1 });
  processorFormRef.value?.reset();
  processorDialogVisible.value = true;
};

const handleProcessorEdit = (row: any) => {
  processorIsEdit.value = true;
  processorDialogTitle.value = '编辑加工商';
  Object.assign(processorForm, {
    id: row.id, processorCode: row.processor_code, processorName: row.processor_name,
    contact: row.contact, phone: row.phone, email: row.email, address: row.address,
    processType: row.process_type, remark: row.remark, status: row.status,
  });
  setTimeout(() => processorFormRef.value?.clearValidate?.(), 100);
  processorDialogVisible.value = true;
};

const handleProcessorDelete = async (row: any) => {
  try { await deleteProcessor(row.id); MessagePlugin.success('删除成功'); loadProcessorData(); } catch { MessagePlugin.error('删除失败'); }
};

const handleProcessorSubmit = async () => {
  const valid = await processorFormRef.value?.validate();
  if (valid !== true) return;
  try {
    const data: any = { processor_code: processorForm.processorCode, processor_name: processorForm.processorName, contact: processorForm.contact, phone: processorForm.phone, email: processorForm.email, address: processorForm.address, process_type: processorForm.processType, remark: processorForm.remark, status: processorForm.status };
    if (processorIsEdit.value && processorForm.id) { data.id = processorForm.id; await updateProcessor(data); MessagePlugin.success('更新成功'); }
    else { await createProcessor(data); MessagePlugin.success('创建成功'); }
    processorDialogVisible.value = false;
    loadProcessorData();
  } catch { MessagePlugin.error(processorIsEdit.value ? '更新失败' : '创建失败'); }
};

// ==================== 供应商 ====================
const supplierLoading = ref(false);
const supplierTableData = ref<any[]>([]);
const supplierDialogVisible = ref(false);
const supplierDialogTitle = ref('新增供应商');
const supplierIsEdit = ref(false);
const supplierFormRef = ref();

const supplierColumns = [
  { colKey: 'sup_code', title: '供应商编号', width: 140, fixed: 'left' as const },
  { colKey: 'sup_name', title: '供应商名称', width: 150 },
  { colKey: 'contact', title: '联系人', width: 100 },
  { colKey: 'phone', title: '联系电话', width: 140 },
  { colKey: 'address', title: '地址', width: 180, ellipsis: true as const },
  { colKey: 'status', title: '状态', width: 80 },
  { colKey: 'create_time', title: '创建时间', width: 160 },
  { colKey: 'operation', title: '操作', width: 120, fixed: 'right' as const },
];

const supplierForm = reactive({
  id: null as number | null,
  supCode: '', supName: '', contact: '', phone: '', email: '', address: '', remark: '', status: 1,
});

const supplierRules = {
  supCode: [{ required: true, message: '请输入供应商编号', trigger: 'blur' as const }],
  supName: [{ required: true, message: '请输入供应商名称', trigger: 'blur' as const }],
  phone: [{ pattern: phoneReg, message: '请输入正确的手机号', trigger: 'blur' as const }],
  email: [{ pattern: emailReg, message: '请输入正确的邮箱地址', trigger: 'blur' as const }],
};

const loadSupplierData = async () => {
  supplierLoading.value = true;
  try {
    const res = await getSupplierList({ page: 1, page_size: 9999 });
    supplierTableData.value = res.list || [];
  } catch { MessagePlugin.error('加载供应商数据失败'); }
  finally { supplierLoading.value = false; }
};

const handleSupplierAdd = () => {
  supplierIsEdit.value = false;
  supplierDialogTitle.value = '新增供应商';
  Object.assign(supplierForm, { id: null, supCode: '', supName: '', contact: '', phone: '', email: '', address: '', remark: '', status: 1 });
  supplierFormRef.value?.reset();
  supplierDialogVisible.value = true;
};

const handleSupplierEdit = (row: any) => {
  supplierIsEdit.value = true;
  supplierDialogTitle.value = '编辑供应商';
  Object.assign(supplierForm, {
    id: row.id, supCode: row.sup_code, supName: row.sup_name, contact: row.contact,
    phone: row.phone, email: row.email, address: row.address, remark: row.remark, status: row.status,
  });
  setTimeout(() => supplierFormRef.value?.clearValidate?.(), 100);
  supplierDialogVisible.value = true;
};

const handleSupplierDelete = async (row: any) => {
  try { await deleteSupplier(row.id); MessagePlugin.success('删除成功'); loadSupplierData(); } catch { MessagePlugin.error('删除失败'); }
};

const handleSupplierSubmit = async () => {
  const valid = await supplierFormRef.value?.validate();
  if (valid !== true) return;
  try {
    const data: any = { sup_code: supplierForm.supCode, sup_name: supplierForm.supName, contact: supplierForm.contact, phone: supplierForm.phone, email: supplierForm.email, address: supplierForm.address, remark: supplierForm.remark, status: supplierForm.status };
    if (supplierIsEdit.value && supplierForm.id) { data.id = supplierForm.id; await updateSupplier(data); MessagePlugin.success('更新成功'); }
    else { await createSupplier(data); MessagePlugin.success('创建成功'); }
    supplierDialogVisible.value = false;
    loadSupplierData();
  } catch { MessagePlugin.error(supplierIsEdit.value ? '更新失败' : '创建失败'); }
};

// ==================== 仓库 ====================
const warehouseLoading = ref(false);
const warehouseTableData = ref<any[]>([]);
const warehouseDialogVisible = ref(false);
const warehouseDialogTitle = ref('新增仓库');
const warehouseIsEdit = ref(false);
const warehouseFormRef = ref();

const warehouseColumns = [
  { colKey: 'wh_code', title: '仓库编码', width: 120, fixed: 'left' as const },
  { colKey: 'wh_name', title: '仓库名称', width: 150 },
  { colKey: 'address', title: '地址', width: 200, ellipsis: true as const },
  { colKey: 'manager', title: '管理员', width: 100 },
  { colKey: 'phone', title: '联系电话', width: 140 },
  { colKey: 'status', title: '状态', width: 80 },
  { colKey: 'create_time', title: '创建时间', width: 160 },
  { colKey: 'operation', title: '操作', width: 120, fixed: 'right' as const },
];

const warehouseForm = reactive({
  id: undefined as number | undefined,
  whCode: '', whName: '', address: '', manager: '', phone: '', remark: '', status: 1,
});

const warehouseRules = {
  whCode: [{ required: true, message: '请输入仓库编码', trigger: 'blur' as const }],
  whName: [{ required: true, message: '请输入仓库名称', trigger: 'blur' as const }],
  phone: [{ pattern: phoneReg, message: '请输入正确的手机号', trigger: 'blur' as const }],
};

const loadWarehouseData = async () => {
  warehouseLoading.value = true;
  try {
    const res = await getWarehouseList({ page: 1, page_size: 9999 });
    warehouseTableData.value = (res.list || []).map((item: any) => ({
      ...item, wh_code: item.wh_code, wh_name: item.wh_name,
    }));
  } catch { MessagePlugin.error('获取仓库列表失败'); }
  finally { warehouseLoading.value = false; }
};

const handleWarehouseAdd = () => {
  warehouseIsEdit.value = false;
  warehouseDialogTitle.value = '新增仓库';
  Object.assign(warehouseForm, { id: undefined, whCode: '', whName: '', address: '', manager: '', phone: '', remark: '', status: 1 });
  warehouseFormRef.value?.reset();
  warehouseDialogVisible.value = true;
};

const handleWarehouseEdit = (row: any) => {
  warehouseIsEdit.value = true;
  warehouseDialogTitle.value = '编辑仓库';
  Object.assign(warehouseForm, {
    id: row.id, whCode: row.wh_code, whName: row.wh_name, address: row.address,
    manager: row.manager, phone: row.phone, remark: row.remark, status: row.status,
  });
  setTimeout(() => warehouseFormRef.value?.clearValidate?.(), 100);
  warehouseDialogVisible.value = true;
};

const handleWarehouseDelete = async (row: any) => {
  try { await deleteWarehouse(row.id); MessagePlugin.success('删除成功'); loadWarehouseData(); } catch { MessagePlugin.error('删除失败'); }
};

const handleWarehouseSubmit = async () => {
  const valid = await warehouseFormRef.value?.validate();
  if (valid !== true) return;
  try {
    const data: any = { wh_code: warehouseForm.whCode, wh_name: warehouseForm.whName, address: warehouseForm.address, manager: warehouseForm.manager, phone: warehouseForm.phone, remark: warehouseForm.remark, status: warehouseForm.status };
    if (warehouseIsEdit.value && warehouseForm.id) { data.id = warehouseForm.id; await updateWarehouse(data); MessagePlugin.success('更新成功'); }
    else { await createWarehouse(data); MessagePlugin.success('创建成功'); }
    warehouseDialogVisible.value = false;
    loadWarehouseData();
  } catch { MessagePlugin.error(warehouseIsEdit.value ? '更新失败' : '创建失败'); }
};

// ==================== 通用 ====================
const loadTabData = () => {
  switch (activeTab.value) {
    case 'customer': loadCustomerData(); break;
    case 'processor': loadProcessorData(); break;
    case 'supplier': loadSupplierData(); break;
    case 'warehouse': loadWarehouseData(); break;
  }
};

onMounted(() => { loadCustomerData(); });
</script>

<style scoped lang="less">
.basic-data-container {
  padding: 16px;

  .toolbar-card {
    margin-bottom: 16px;

    .toolbar-content {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .toolbar-left {
        display: flex;
        align-items: center;
        gap: 12px;
      }

      .toolbar-actions {
        display: flex;
        gap: 8px;
      }
    }
  }
}
</style>
