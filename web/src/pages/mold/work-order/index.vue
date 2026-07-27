<template>
  <div class="mold-work-order-container">
    <search-filter
      :filters="filterConfig"
      :model-value="formData"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">模具工单管理</span>
        <div class="toolbar-actions">
          <t-button theme="default" @click="handleColumnSetting">
            <template #icon><setting-icon /></template>
            列设置
          </t-button>
          <t-button theme="primary" @click="handleAdd">
            <template #icon><add-icon /></template>
            新增工单
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <t-table
        :data="tableData"
        :columns="displayColumns"
        :loading="loading"
        row-key="id"
        :pagination="pagination"
        :stripe="false"
        @page-change="rehandlePageChange"
      >
        <template #order_type="{ row }">
          <order-tag kind="type" v-bind="getOrderTypeTagProps(row.order_type)">
            {{ getOrderTypeText(row.order_type) }}
          </order-tag>
        </template>
        <template #[`mold_order.order_code`]="{ row }">
          <div
            v-if="row.mold_order?.order_code"
            style="display: flex; align-items: center; gap: 8px"
          >
            <t-link @click="handleViewMoldOrder(row.mold_order.id)">
              {{ row.mold_order.order_code }}
            </t-link>
            <t-tag
              v-if="row.mold_order.priority"
              shape="round"
              size="small"
              :theme="getPriorityTheme(row.mold_order.priority)"
              :style="getPriorityStyle(row.mold_order.priority)"
            >
              {{ getPriorityText(row.mold_order.priority) }}
            </t-tag>
          </div>
          <span v-else>-</span>
        </template>
        <template #[`mold.mold_code`]="{ row }">
          <t-link
            v-if="row.mold?.mold_code"
            class="mold-code-link"
            @click="handleViewMold(row.mold.mold_code)"
          >
            {{ row.mold?.mold_code }}
          </t-link>
          <span v-else>-</span>
        </template>
        <template #[`mold.customer.cus_name`]="{ row }">
          <span v-if="row.mold?.customer?.cus_name">{{ row.mold.customer.cus_name }}</span>
          <span v-else>-</span>
        </template>
        <!-- 工序进度展示（分组并行） -->
        <template #processes="{ row }">
          <div class="process-summary" @click="showProcessDetail(row)">
            <!-- 第一行：工序胶囊（居中对齐的基准） -->
            <div class="process-row-capsules">
              <template
                v-for="(group, groupIndex) in groupedProcesses"
                :key="`group-${group.groupId}`"
              >
                <!-- 工序标签组 -->
                <div class="group-processes-list">
                  <template v-for="(template, templateIndex) in group.templates" :key="template.id">
                    <span
                      class="process-name-item"
                      :class="getProcessNameClass(row, template.id, templateIndex)"
                    >
                      {{ getProcessShortName(template.process_name) }}
                    </span>
                  </template>
                </div>

                <!-- 分组间的箭头 -->
                <span v-if="groupIndex < groupedProcesses.length - 1" class="group-arrow">→</span>
              </template>
            </div>
          </div>
        </template>
        <template #progress="{ row }">
          <t-progress
            :percentage="calculateProgress(row)"
            :status="calculateProgress(row) === 100 ? 'success' : 'active'"
            style="width: 100px"
          />
        </template>
        <template #status="{ row }">
          <t-tag
            shape="round"
            :theme="getStatusTheme(row.status)"
            :style="getStatusStyle(row.status)"
          >
            {{ getStatusText(row.status) }}
          </t-tag>
        </template>
        <template #plan_start="{ row }">
          {{ formatDate(row.plan_start) }}
        </template>
        <template #plan_end="{ row }">
          {{ formatDate(row.plan_end) }}
        </template>
        <template #operation="{ row }">
          <div style="width: 180px">
            <t-space :size="8" break-line>
              <t-link
                theme="primary"
                :disabled="isWorkOrderLocked(row) || row.is_cancelled === 1"
                @click="handleEdit(row)"
              >
                编辑
              </t-link>

              <t-tooltip :content="getTryTestDisabledTip(row)">
                <t-link
                  theme="primary"
                  :disabled="!canTryTest(row) || isWorkOrderLocked(row) || row.is_cancelled === 1"
                  @click="handleTryTest(row)"
                >
                  试模
                </t-link>
              </t-tooltip>

              <t-link
                v-if="canStockIn(row)"
                theme="success"
                :disabled="isWorkOrderLocked(row) || row.is_cancelled === 1"
                @click="handleStockIn(row)"
              >
                入库
              </t-link>

              <!-- 未开始状态显示删除，加工中状态显示取消 -->
              <t-popconfirm
                v-if="getStatusLevel(row.status) < WORK_ORDER_STATUS.PROCESSING"
                content="确认删除该工单吗？"
                @confirm="handleDelete(row)"
              >
                <t-link theme="danger">删除</t-link>
              </t-popconfirm>
              <t-link
                v-else-if="getStatusLevel(row.status) === WORK_ORDER_STATUS.PROCESSING"
                theme="warning"
                :disabled="isWorkOrderLocked(row) || row.is_cancelled === 1"
                @click="handleCancelWorkOrder(row)"
              >
                取消
              </t-link>
            </t-space>
          </div>
        </template>
      </t-table>
    </t-card>

    <!-- 新增/编辑对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :header="dialogTitle"
      width="800px"
      @confirm="handleSubmit"
      @cancel="handleCancel"
      @opened="handleDialogOpened"
     :close-on-overlay-click="false">
      <t-form
        ref="formRef"
        :model="formModel"
        :rules="formRules"
        :label-width="100"
        style="margin-top: 16px"
      >
        <t-row :gutter="24" style="margin-left: 0; margin-right: 0">
          <t-col :span="12">
            <t-form-item label="工单号" name="workCode" style="margin-bottom: 16px">
              <t-input v-model="formModel.workCode" placeholder="系统自动生成" disabled />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="订单类型" name="orderType" style="margin-bottom: 16px">
              <t-select
                v-model="formModel.orderType"
                placeholder="请选择订单类型"
                :disabled="isEdit"
                @change="handleOrderTypeChange"
              >
                <t-option
                  v-for="type in orderTypeList"
                  :key="type.value"
                  :label="type.label"
                  :value="Number(type.value)"
                />
              </t-select>
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="模具订单" name="moldOrderId" style="margin-bottom: 16px">
              <div style="display: flex; gap: 8px; align-items: center; width: 100%">
                <t-select
                  v-model="formModel.moldOrderId"
                  placeholder="请选择订单（可选）"
                  clearable
                  :disabled="isEdit && moldOrderLocked"
                  :class="{ 'locked-input': isEdit && moldOrderLocked }"
                  style="flex: 1"
                  @change="handleMoldOrderChange"
                >
                  <t-option
                    v-for="order in filteredOrderList"
                    :key="order.id"
                    :label="order.order_code"
                    :value="order.id"
                  />
                </t-select>
                <t-button
                  v-if="isEdit"
                  :theme="moldOrderLocked ? 'default' : 'primary'"
                  variant="outline"
                  size="small"
                  @click="toggleMoldOrderLock"
                >
                  {{ moldOrderLocked ? '解锁' : '锁定' }}
                </t-button>
              </div>
            </t-form-item>
          </t-col>
          <t-col
            v-if="
              formModel.orderType === ORDER_TYPE.REPAIR || formModel.orderType === ORDER_TYPE.MODIFY
            "
            :span="12"
          >
            <t-form-item label="模具" name="moldId" style="margin-bottom: 16px">
              <t-select v-model="formModel.moldId" placeholder="请选择模具" :disabled="isEdit">
                <t-option
                  v-for="mold in moldList"
                  :key="mold.id"
                  :label="mold.mold_code"
                  :value="mold.id"
                />
              </t-select>
            </t-form-item>
          </t-col>
        </t-row>

        <!-- 新模工单 - 新模具信息 -->
        <t-divider v-if="formModel.orderType === ORDER_TYPE.NEW_MOLD" align="left"
          >模具信息</t-divider
        >
        <t-row
          v-if="formModel.orderType === ORDER_TYPE.NEW_MOLD"
          :gutter="24"
          style="margin-left: 0; margin-right: 0"
        >
          <t-col :span="12">
            <t-form-item label="模具编号" style="margin-bottom: 16px">
              <t-input
                v-if="!isEdit"
                v-model="formModel.newMoldCode"
                placeholder="请输入模具编号"
              />
              <t-input
                v-else
                v-model="existingMoldInfo.moldCode"
                disabled
                :class="{ 'locked-input': isEdit }"
              />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="产品名称" style="margin-bottom: 16px">
              <t-input
                v-if="!isEdit"
                v-model="formModel.newMoldName"
                placeholder="请输入产品名称"
              />
              <t-input
                v-else
                v-model="existingMoldInfo.moldName"
                disabled
                :class="{ 'locked-input': isEdit }"
              />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="客户" style="margin-bottom: 16px">
              <t-select
                v-if="!isEdit"
                v-model="formModel.customerId"
                placeholder="请选择客户"
                clearable
              >
                <t-option
                  v-for="customer in customerList"
                  :key="customer.id"
                  :label="customer.name"
                  :value="customer.id"
                />
              </t-select>
              <t-input
                v-else
                v-model="existingMoldInfo.customerName"
                disabled
                :class="{ 'locked-input': isEdit }"
              />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="模具类型" style="margin-bottom: 16px">
              <t-select
                v-if="!isEdit"
                v-model="formModel.newMoldType"
                placeholder="请选择模具类型"
                clearable
              >
                <t-option label="注塑模" :value="1" />
                <t-option label="冲压模" :value="2" />
                <t-option label="压铸模" :value="3" />
              </t-select>
              <t-input
                v-else
                v-model="existingMoldInfo.moldTypeText"
                disabled
                :class="{ 'locked-input': isEdit }"
              />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="材质类型" style="margin-bottom: 16px">
              <t-select
                v-if="!isEdit"
                v-model="formModel.newMaterialType"
                placeholder="请选择材质类型"
                clearable
              >
                <t-option label="钢材" :value="1" />
                <t-option label="铝材" :value="2" />
                <t-option label="铜材" :value="3" />
              </t-select>
              <t-input
                v-else
                v-model="existingMoldInfo.materialTypeText"
                disabled
                :class="{ 'locked-input': isEdit }"
              />
            </t-form-item>
          </t-col>
        </t-row>

        <t-row :gutter="24" style="margin-left: 0; margin-right: 0">
          <t-col :span="12">
            <t-form-item label="计划开始时间" name="planStart" style="margin-bottom: 16px">
              <t-date-picker
                v-model="formModel.planStart"
                placeholder="请选择日期"
                format="YYYY-MM-DD HH:mm:ss"
                show-time
              />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="计划结束时间" name="planEnd" style="margin-bottom: 16px">
              <t-date-picker
                v-model="formModel.planEnd"
                placeholder="请选择日期"
                format="YYYY-MM-DD HH:mm:ss"
                show-time
              />
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="备注" name="remark">
              <t-textarea v-model="formModel.remark" placeholder="请输入备注" :rows="3" />
            </t-form-item>
          </t-col>
        </t-row>
      </t-form>
    </t-dialog>

    <!-- 试模管理对话框 -->
    <t-dialog
      v-model:visible="tryTestDialogVisible"
      header="试模记录"
      width="900px"
      :footer="false"
      @close="tryTestDialogVisible = false"
     :close-on-overlay-click="false">
      <div class="process-content">
        <t-tooltip
          :content="getTryTestDisabledTip(currentWorkOrder) || getMutexTip(currentWorkOrder)"
          placement="top"
        >
          <t-button
            theme="primary"
            :disabled="
              !canTryTest(currentWorkOrder) ||
              isWorkOrderLocked(currentWorkOrder) ||
              currentWorkOrder?.is_cancelled === 1
            "
            style="margin-bottom: 16px"
            @click="handleAddTryTest"
          >
            <template #icon><add-icon /></template>
            新增试模
          </t-button>
        </t-tooltip>
        <t-table
          :data="tryTestList"
          :columns="tryTestColumns"
          row-key="id"
          :pagination="{ pageSize: 0, total: 0 }"
        >
          <template #operation="{ row }">
            <t-tooltip :content="getMutexTip(currentWorkOrder)" placement="top">
              <t-link
                theme="primary"
                :disabled="
                  isWorkOrderLocked(currentWorkOrder) || currentWorkOrder?.is_cancelled === 1
                "
                @click="handleEditTryTest(row)"
              >
                编辑
              </t-link>
            </t-tooltip>
            <t-tooltip :content="getMutexTip(currentWorkOrder)" placement="top">
              <t-popconfirm content="确认删除该记录？" @confirm="handleDeleteTryTest(row)">
                <t-link
                  theme="danger"
                  :disabled="
                    isWorkOrderLocked(currentWorkOrder) || currentWorkOrder?.is_cancelled === 1
                  "
                  style="margin-left: 8px"
                >
                  删除
                </t-link>
              </t-popconfirm>
            </t-tooltip>
          </template>
        </t-table>
      </div>
    </t-dialog>

    <!-- 试模表单对话框 -->
    <t-dialog
      v-model:visible="tryTestFormDialogVisible"
      :header="tryTestForm.id ? '编辑试模记录' : '新增试模记录'"
      width="600px"
      @confirm="handleSaveTryTest"
      @cancel="tryTestFormDialogVisible = false"
     :close-on-overlay-click="false">
      <t-form
        ref="tryTestFormRef"
        :model="tryTestForm"
        :rules="tryTestFormRules"
        :label-width="100"
        style="margin-top: 16px"
      >
        <t-form-item label="试模时间" name="testTime">
          <t-date-picker v-model="tryTestForm.testTime" format="YYYY-MM-DD" />
        </t-form-item>
        <t-form-item label="合格数量" name="okQty">
          <t-input-number v-model="tryTestForm.okQty" placeholder="输入数量" style="width: 100%" />
        </t-form-item>
        <t-form-item label="不良数量" name="badQty">
          <t-input-number v-model="tryTestForm.badQty" placeholder="输入数量" style="width: 100%" />
        </t-form-item>
        <t-form-item label="试模结果" name="result">
          <t-select v-model="tryTestForm.result" placeholder="请选择试模结果" clearable>
            <t-option label="合格" value="合格" />
            <t-option label="不合格" value="不合格" />
            <t-option label="需返修" value="需返修" />
          </t-select>
        </t-form-item>
        <t-form-item label="试模方式" name="testMode">
          <t-select v-model="tryTestForm.testMode" placeholder="请选择试模方式" clearable>
            <t-option
              v-for="item in testModeList"
              :key="item.dict_value"
              :label="item.dict_label"
              :value="item.dict_value"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="备注">
          <t-textarea v-model="tryTestForm.remark" placeholder="请输入备注（可选）" :rows="3" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 交付管理对话框 -->
    <t-dialog
      v-model:visible="stockInDialogVisible"
      header="模具入库"
      width="600px"
      @confirm="handleConfirmStockIn"
      @cancel="stockInDialogVisible = false"
     :close-on-overlay-click="false">
      <t-form ref="stockInFormRef" :model="stockInForm" :label-width="100" style="margin-top: 16px">
        <t-form-item label="模具编号">
          <t-input :value="stockInForm.moldCode" disabled />
        </t-form-item>
        <t-form-item label="仓库" name="warehouseId">
          <t-select v-model="stockInForm.warehouseId" placeholder="请选择仓库">
            <t-option
              v-for="warehouse in warehouseList"
              :key="warehouse.id"
              :label="warehouse.wh_name"
              :value="warehouse.id"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="备注">
          <t-textarea v-model="stockInForm.remark" placeholder="请输入备注（可选）" :rows="3" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 取消工单对话框 -->
    <t-dialog
      v-model:visible="cancelDialogVisible"
      header="取消工单"
      width="600px"
      @confirm="handleConfirmCancelWorkOrder"
      @cancel="cancelDialogVisible = false"
     :close-on-overlay-click="false">
      <t-form ref="cancelFormRef" :model="cancelForm" :label-width="100" style="margin-top: 16px">
        <t-form-item label="工单编号">
          <t-input :value="currentWorkOrderForCancel?.work_code" disabled />
        </t-form-item>
        <t-form-item label="模具编号">
          <t-input :value="currentWorkOrderForCancel?.mold?.mold_code" disabled />
        </t-form-item>
        <t-form-item label="取消原因" name="reason">
          <t-textarea
            v-model="cancelForm.reason"
            placeholder="请输入取消原因（必填）"
            :rows="4"
            maxlength="500"
          />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 列设置对话框 - 使用通用组件 -->
    <column-setting-dialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="mold_work_order_table_columns"
      @change="handleColumnsChange"
    />

    <!-- 工序详情侧滑面板 -->
    <t-drawer
      v-model:visible="processDrawerVisible"
      :header="
        currentWorkOrder ? `工序详情 - ${currentWorkOrder.mold?.mold_code || '-'}` : '工序详情'
      "
      size="600px"
      :footer="false"
      @close="closeProcessDrawer"
    >
      <div v-if="currentWorkOrder" class="process-drawer-content">
        <!-- 按分组展示工序 -->
        <div class="process-groups-drawer">
          <template v-for="(group, groupIndex) in groupedProcesses" :key="group.groupId">
            <div class="process-group-card">
              <div class="group-card-header">
                <div class="group-card-title">
                  <span class="group-badge">{{ group.groupId }}</span>
                  <span>{{ group.groupName }}</span>
                  <span class="group-hint">（可并行）</span>
                </div>
                <t-tag v-if="groupIndex > 0" theme="default" size="small">
                  <t-icon name="arrow-down" /> 上一组完成后开始
                </t-tag>
              </div>

              <div class="group-card-processes">
                <div
                  v-for="(template, templateIndex) in group.templates"
                  :key="template.id"
                  class="process-item"
                  :class="
                    getProcessItemClass(
                      currentWorkOrder,
                      template.id,
                      templateIndex,
                      group.templates,
                      groupIndex,
                    )
                  "
                >
                  <div class="process-item-header">
                    <span class="process-item-name">{{ template.process_name }}</span>
                    <t-tag
                      v-if="
                        getProcessStatusValue(currentWorkOrder, template.id) ===
                        PROCESS_STATUS.COMPLETED
                      "
                      theme="success"
                      size="small"
                    >
                      已完成
                    </t-tag>
                    <t-tag
                      v-else-if="
                        getProcessStatusValue(currentWorkOrder, template.id) ===
                        PROCESS_STATUS.SKIPPED
                      "
                      theme="default"
                      size="small"
                      variant="light-outline"
                    >
                      已跳过
                    </t-tag>
                    <t-tag
                      v-else-if="isGroupCanBeStarted(currentWorkOrder, groupIndex)"
                      theme="primary"
                      size="small"
                    >
                      进行中
                    </t-tag>
                    <t-tag v-else theme="default" size="small"> 待开工 </t-tag>
                  </div>
                  <div class="process-item-action">
                    <!-- 未完成状态：显示操作按钮 -->
                    <div
                      v-if="
                        getProcessStatusValue(currentWorkOrder, template.id) ===
                        PROCESS_STATUS.PENDING
                      "
                      class="action-buttons"
                    >
                      <t-tooltip
                        :content="
                          getMutexTip(currentWorkOrder) ||
                          (!isCurrentProcess(currentWorkOrder, template.id)
                            ? '需等待上一组完成后开始'
                            : '')
                        "
                        placement="top"
                      >
                        <t-button
                          theme="success"
                          variant="outline"
                          size="small"
                          shape="round"
                          :disabled="
                            isWorkOrderLocked(currentWorkOrder) ||
                            currentWorkOrder?.is_cancelled === 1 ||
                            !isCurrentProcess(currentWorkOrder, template.id)
                          "
                          class="action-btn-complete"
                          @click="
                            toggleProcessStatus(
                              currentWorkOrder,
                              template.id,
                              PROCESS_STATUS.COMPLETED,
                            )
                          "
                        >
                          <template #icon><check-icon /></template>
                          完成
                        </t-button>
                      </t-tooltip>
                      <t-tooltip
                        :content="
                          getMutexTip(currentWorkOrder) ||
                          (!isCurrentProcess(currentWorkOrder, template.id)
                            ? '需等待上一组完成后开始'
                            : '')
                        "
                        placement="top"
                      >
                        <t-button
                          theme="default"
                          variant="outline"
                          size="small"
                          shape="round"
                          :disabled="
                            isWorkOrderLocked(currentWorkOrder) ||
                            currentWorkOrder?.is_cancelled === 1 ||
                            !isCurrentProcess(currentWorkOrder, template.id)
                          "
                          class="action-btn-skip"
                          @click="
                            toggleProcessStatus(
                              currentWorkOrder,
                              template.id,
                              PROCESS_STATUS.SKIPPED,
                            )
                          "
                        >
                          <template #icon><close-circle-icon /></template>
                          跳过
                        </t-button>
                      </t-tooltip>
                    </div>

                    <!-- 已完成状态：显示取消按钮 -->
                    <div
                      v-else-if="
                        getProcessStatusValue(currentWorkOrder, template.id) ===
                        PROCESS_STATUS.COMPLETED
                      "
                      class="action-buttons"
                    >
                      <t-tooltip
                        :content="
                          getMutexTip(currentWorkOrder) ||
                          (getStatusLevel(currentWorkOrder?.status) >= WORK_ORDER_STATUS.COMPLETED
                            ? '工单已进入后期阶段，禁止回退工序操作'
                            : '')
                        "
                        placement="top"
                      >
                        <t-button
                          theme="default"
                          variant="text"
                          size="small"
                          :disabled="
                            isWorkOrderLocked(currentWorkOrder) ||
                            currentWorkOrder?.is_cancelled === 1 ||
                            getStatusLevel(currentWorkOrder?.status) >= WORK_ORDER_STATUS.COMPLETED
                          "
                          class="action-btn-cancel"
                          @click="
                            toggleProcessStatus(
                              currentWorkOrder,
                              template.id,
                              PROCESS_STATUS.PENDING,
                            )
                          "
                        >
                          撤销
                        </t-button>
                      </t-tooltip>
                    </div>

                    <!-- 已跳过状态：显示取消按钮 -->
                    <div
                      v-else-if="
                        getProcessStatusValue(currentWorkOrder, template.id) ===
                        PROCESS_STATUS.SKIPPED
                      "
                      class="action-buttons"
                    >
                      <t-tooltip
                        :content="
                          getMutexTip(currentWorkOrder) ||
                          (getStatusLevel(currentWorkOrder?.status) >= WORK_ORDER_STATUS.COMPLETED
                            ? '工单已进入后期阶段，禁止回退工序操作'
                            : '')
                        "
                        placement="top"
                      >
                        <t-button
                          theme="default"
                          variant="text"
                          size="small"
                          :disabled="
                            isWorkOrderLocked(currentWorkOrder) ||
                            currentWorkOrder?.is_cancelled === 1 ||
                            getStatusLevel(currentWorkOrder?.status) >= WORK_ORDER_STATUS.COMPLETED
                          "
                          class="action-btn-cancel"
                          @click="
                            toggleProcessStatus(
                              currentWorkOrder,
                              template.id,
                              PROCESS_STATUS.PENDING,
                            )
                          "
                        >
                          撤销
                        </t-button>
                      </t-tooltip>
                    </div>

                    <span
                      v-if="getProcessCompletedAt(currentWorkOrder, template.id)"
                      class="process-item-time"
                    >
                      {{
                        formatCompletedTime(getProcessCompletedAt(currentWorkOrder, template.id))
                      }}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <!-- 组间分隔箭头 -->
            <div v-if="groupIndex < groupedProcesses.length - 1" class="group-arrow-divider">
              <t-icon name="arrow-down" size="24px" />
            </div>
          </template>
        </div>

        <!-- 进度统计 -->
        <t-card class="process-stats" :bordered="false">
          <div class="stats-container">
            <div class="stat-item">
              <div class="stat-value">{{ getCompletedCount(currentWorkOrder) }}</div>
              <div class="stat-label">已完成</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">{{ getInProgressCount(currentWorkOrder) }}</div>
              <div class="stat-label">进行中</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">{{ getPendingCount(currentWorkOrder) }}</div>
              <div class="stat-label">待开工</div>
            </div>
          </div>
        </t-card>
      </div>
    </t-drawer>
  </div>
</template>
<script setup lang="ts">
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next';
import { computed, h, nextTick, onMounted, reactive, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

defineOptions({
  name: 'MoldWorkOrder',
});
/** HTML 转义，防止 XSS */
const escapeHtml = (str: string): string => {
  const map: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  };
  return str.replace(/[&<>"']/g, (c) => map[c]);
};
import 'dayjs/locale/zh-cn';

import dayjs from 'dayjs';
import timezone from 'dayjs/plugin/timezone';
import utc from 'dayjs/plugin/utc';
import { AddIcon, CheckIcon, CloseCircleIcon, SettingIcon } from 'tdesign-icons-vue-next';
import type { TableProps } from 'tdesign-vue-next';

import {
  cancelMoldWorkOrder,
  createMoldTryTest,
  createMoldWorkOrder,
  deleteMoldTryTest,
  deleteMoldWorkOrder,
  getMoldList,
  getMoldOrderList,
  getMoldTryTests,
  getMoldWorkOrderList,
  updateMoldTryTest,
  updateMoldWorkOrder,
} from '@/api/mold';
import { toggleProcess } from '@/api/process-template';
import { getActiveProcessTemplates, getCustomerList, getWarehouseList } from '@/api/reference';
import type { ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import ColumnSettingDialog from '@/components/ColumnSettingDialog/index.vue';
import { OrderTag } from '@/components/OrderTag';
import type { FilterItem } from '@/components/SearchFilter/index.vue';
import SearchFilter from '@/components/SearchFilter/index.vue';
import { ORDER_TYPE, PROCESS_STATUS, WORK_ORDER_STATUS } from '@/constants';
import { useDictRaw } from '@/hooks/useDict';
import { formatDate } from '@/utils/date';
import { getDictColorRenderConfig } from '@/utils/dict-color';
import { OPERATION_COLUMN_WIDTH } from '@/utils/operation-column-width';

// 配置 dayjs
dayjs.extend(utc);
dayjs.extend(timezone);
dayjs.locale('zh-cn');

const router = useRouter();

// 表格相关
const loading = ref(false);
const tableData = ref<any[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

// 搜索表单
const formData = reactive({
  moldCode: '',
  workCode: '',
  customerId: undefined,
  status: undefined,
  orderType: undefined,
  processStage: undefined,
});

// 订单列表
const orderList = ref<any[]>([]);

// 模具列表
const moldList = ref<any[]>([]);

// 客户列表
const customerList = ref<any[]>([]);

// 仓库列表
const warehouseList = ref<any[]>([]);

// 工序模板列表
const processTemplates = ref<any[]>([]);

// 试模方式字典列表（useDictRaw 自动加载 + 全局缓存）
const { items: testModeList } = useDictRaw('test-mode');

// 工单状态字典列表（useDictRaw 自动加载 + 全局缓存，过滤+映射）
const { items: workOrderStatusList } = useDictRaw('work-order-status', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1 && item.dict_value != null && item.dict_value !== '')
    .map((item: any) => ({
      value: item.dict_value,
      label: item.dict_label,
      color: item.color || 'default',
    })),
);

// 订单/工单类型字典列表（useDictRaw 自动加载 + 全局缓存）
const { items: orderTypeList } = useDictRaw('mold_order_type', (items) =>
  (items as any[]).map((item: any) => ({
    value: item.dict_value,
    label: item.dict_label,
    color: item.color || 'default',
  })),
);

// ✨ 优先级字典列表（useDictRaw 自动加载 + 全局缓存）
const { items: priorityList } = useDictRaw('order-priority', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1 && item.dict_value != null && item.dict_value !== '')
    .map((item: any) => ({
      value: item.dict_value,
      label: item.dict_label,
      color: item.color || 'default',
    })),
);

// 工序阶段列表
const processStageList = ref([
  { value: 'design', label: '设计阶段' },
  { value: 'cnc', label: 'CNC加工' },
  { value: 'edm', label: '放电加工' },
  { value: 'welding', label: '线切割' },
  { value: 'polishing', label: '抛光' },
  { value: 'assembly', label: '装配' },
  { value: 'testing', label: '试模' },
  { value: 'inspection', label: '检验' },
]);

// 字典加载（已替换为 useDictRaw，自动加载 + 全局缓存）

// 入库对话框相关

// 入库对话框相关
const stockInDialogVisible = ref(false);
const stockInFormRef = ref();
const currentWorkOrderForStockIn = ref<any>(null);
const stockInForm = reactive({
  moldCode: '',
  warehouseId: undefined,
  remark: '',
});

// 取消工单对话框相关
const cancelDialogVisible = ref(false);
const cancelFormRef = ref();
const currentWorkOrderForCancel = ref<any>(null);
const cancelForm = reactive({
  reason: '',
});

// 工序进度列的实际宽度
const processColumnActualWidth = ref(600);

// 表格列配置（动态生成）
const columns = computed<TableProps['columns']>(() => {
  const baseColumns: TableProps['columns'] = [
    { colKey: 'work_code', title: '工单号', width: 180, fixed: 'left' },
    { colKey: 'order_type', title: '订单类型', width: 100 },
    { colKey: 'mold_order.order_code', title: '模具订单', width: 180 },
    { colKey: 'mold.mold_code', title: '模具', width: 120 },
    { colKey: 'mold.customer.cus_name', title: '客户', width: 120 },
  ];

  baseColumns.push({
    colKey: 'processes',
    title: '工序进度',
    width: processColumnActualWidth.value, // 动态计算的宽度
    ellipsis: false, // 禁用省略号，使用自定义布局
  });

  baseColumns.push(
    { colKey: 'plan_start', title: '计划开始', width: 160 },
    { colKey: 'plan_end', title: '计划结束', width: 160 },
    { colKey: 'progress', title: '进度', width: 150 },
    { colKey: 'status', title: '状态', width: 100 },
    {
      colKey: 'operation',
      title: '操作',
      width: OPERATION_COLUMN_WIDTH.MEDIUM,
      fixed: 'right' as const,
    },
  );

  return baseColumns;
});

// 对话框相关
const dialogVisible = ref(false);
const isEdit = ref(false);
const dialogTitle = ref('新增工单');
const formRef = ref();

// 保存原始表单数据（用于对比是否修改）
const originalFormData = ref<any>(null);

// 对比两个值是否相等（处理各种边界情况）
const isValueEqual = (val1: any, val2: any): boolean => {
  // 都未定义或都为null，视为相等
  if (val1 === undefined && val2 === undefined) return true;
  if (val1 === null && val2 === null) return true;

  // 将undefined和null视为相等
  if ((val1 === undefined || val1 === null) && (val2 === undefined || val2 === null)) return true;

  // 都未定义或为空字符串，视为相等
  if (val1 === '' && val2 === '') return true;

  // 空字符串和undefined/null视为相等
  if (
    (val1 === '' && (val2 === undefined || val2 === null)) ||
    (val2 === '' && (val1 === undefined || val1 === null))
  ) {
    return true;
  }

  return val1 === val2;
};

const hasFormDataChanged = (): boolean => {
  if (!originalFormData.value) {
    return true;
  }

  // 对比表单字段
  const fieldsToCompare = [
    'workCode',
    'orderType',
    'moldOrderId',

    'moldId',
    'planStart',
    'planEnd',

    'remark',
    'newMoldCode',
    'newMoldName',

    'customerId',
    'newMoldType',
    'newMaterialType',
    'newMoldRemark',
  ];

  // 详细对比每个字段
  for (const field of fieldsToCompare) {
    const val1 = formModel[field as keyof typeof formModel];
    const val2 = originalFormData.value[field];

    if (!isValueEqual(val1, val2)) {
      return true;
    }
  }

  return false;
};

const formModel = reactive({
  id: undefined,
  workCode: '',
  orderType: 1,
  moldOrderId: undefined,
  moldId: undefined,
  planStart: '',
  planEnd: '',
  remark: '',
  // 新模工单字段
  newMoldCode: '',
  newMoldName: '',
  customerId: undefined,
  newMoldType: undefined,
  newMaterialType: undefined,
  newMoldRemark: '',
});

// 已创建的模具信息（编辑时显示）
const existingMoldInfo = reactive({
  moldCode: '',
  moldName: '',
  customerName: '',
  moldTypeText: '',
  materialTypeText: '',
});

// 模具订单字段锁定状态（编辑模式下默认锁定）
const moldOrderLocked = ref(true);

// 原始模具订单ID（用于检测是否修改）
const originalMoldOrderId = ref<number | undefined>(undefined);

const formRules = computed(() => {
  const rules: Record<string, any> = {
    orderType: [{ required: true, message: '请选择订单类型', trigger: 'change' as const }],
    status: [{ required: true, message: '请选择状态', trigger: 'change' as const }],
  };

  // 新模工单：模具信息必填
  if (formModel.orderType === ORDER_TYPE.NEW_MOLD) {
    rules.newMoldCode = [
      { required: true, message: '请输入模具编号', trigger: 'blur' as const },
      { max: 50, message: '模具编号长度不能超过50字符', trigger: 'blur' as const },
    ];
    rules.newMoldName = [
      { required: true, message: '请输入产品名称', trigger: 'blur' as const },
      { max: 100, message: '产品名称长度不能超过100字符', trigger: 'blur' as const },
    ];
    rules.customerId = [{ required: true, message: '请选择客户', trigger: 'change' as const }];
  }

  // 修模/改模工单：模具必选
  if (formModel.orderType === ORDER_TYPE.REPAIR || formModel.orderType === ORDER_TYPE.MODIFY) {
    rules.moldId = [{ required: true, message: '请选择模具', trigger: 'change' as const }];
  }

  return rules;
});

// 试模管理相关
const tryTestDialogVisible = ref(false);
const tryTestFormDialogVisible = ref(false);
const tryTestList = ref<any[]>([]);
const currentWorkOrderId = ref<number>();
const tryTestFormRef = ref();

// 试模表单数据
const tryTestForm = reactive({
  id: undefined as number | undefined,
  mold_work_id: undefined as number | undefined,
  testTime: '',
  okQty: undefined as number | undefined,
  badQty: undefined as number | undefined,
  result: '',
  testMode: '',
  remark: '',
});

// 试模表单验证规则
const tryTestFormRules = {
  result: [{ required: true, message: '请选择试模结果' }],
  okQty: [{ validator: (val: any) => val == null || val === '' || (Number.isInteger(val) && val >= 0), message: '请输入纯数字' }],
  badQty: [{ validator: (val: any) => val == null || val === '' || (Number.isInteger(val) && val >= 0), message: '请输入纯数字' }],
};

const tryTestColumns: TableProps['columns'] = [
  {
    colKey: 'test_time',
    title: '试模时间',
    width: 120,
    cell: (h, { row }) => {
      if (!row.test_time) return '-';
      return formatDate(row.test_time);
    },
  },
  { colKey: 'test_mode', title: '试模方式', width: 120 },
  { colKey: 'ok_qty', title: '合格数', width: 80 },
  { colKey: 'bad_qty', title: '不良数', width: 80 },
  { colKey: 'result', title: '结果', width: 100 },
  { colKey: 'operation', title: '操作', width: 150, fixed: 'right' },
];

// 动态列配置（根据用户选择及拖拽排序）
const displayColumns = computed(() => {
  // 1. 获取参与排序的列（即用户在列设置中勾选的列）
  const sortableCols = columns.value.filter((col) => selectedColumns.value.includes(col.colKey));

  // 2. 严格按照 selectedColumns 的顺序（即用户拖拽后的顺序）排序
  sortableCols.sort((a, b) => {
    const indexA = selectedColumns.value.indexOf(a.colKey);
    const indexB = selectedColumns.value.indexOf(b.colKey);
    return indexA - indexB;
  });

  // 3. 获取操作列（始终显示在最后，不参与排序）
  const operationCol = columns.value.find((col) => col.colKey === 'operation');

  // 4. 组合返回：排序后的列 + 操作列
  const result: TableProps['columns'] = [...sortableCols];
  if (operationCol) result.push(operationCol);

  return result;
});

// 所有可选列配置（不包括操作列）
const allColumns: ColumnOption[] = [
  { colKey: 'work_code', title: '工单号', pinned: true }, // ✨ 固定列：不可隐藏/移动
  { colKey: 'order_type', title: '订单类型' },
  { colKey: 'mold_order.order_code', title: '模具订单' },
  { colKey: 'mold.mold_code', title: '模具' },
  { colKey: 'mold.customer.cus_name', title: '客户名称' },
  { colKey: 'plan_start', title: '计划开始' },
  { colKey: 'plan_end', title: '计划结束' },
  { colKey: 'progress', title: '进度' },
  { colKey: 'status', title: '状态' },
  { colKey: 'processes', title: '工序进度', lockHide: true },
];

// 默认显示的列
const DEFAULT_COLUMNS = [
  'work_code',
  'order_type',
  'mold_order.order_code',
  'mold.mold_code',
  'mold.customer.cus_name',
  'processes',
  'plan_start',
  'plan_end',
  'progress',
  'status',
];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([]);

const getWorkOrderList = async () => {
  loading.value = true;
  try {
    const res = await getMoldWorkOrderList({
      page: pagination.current,
      page_size: pagination.pageSize,
      workCode: formData.workCode,
      moldCode: formData.moldCode,
      customerId: formData.customerId,
      status: formData.status,
      orderType: formData.orderType,
      processStage: formData.processStage,
    });
    const workOrders = res.list || [];

    tableData.value = workOrders;
    pagination.total = res.total || 0;
    // 后端已直接返回 is_locked 字段，无需额外查询
  } catch (error) {
    console.error('获取工单列表失败:', error);
    MessagePlugin.error('获取工单列表失败');
    tableData.value = [];
    pagination.total = 0;
  } finally {
    loading.value = false;
  }
};

// 状态标签颜色（使用统一的颜色配置系统）
const getStatusStyle = (status: number) => {
  if (!status) return {};
  const statusItem = workOrderStatusList.value.find((s) => s.value === status);
  if (!statusItem) return {};

  const color = statusItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);

  // 如果是 theme 模式，返回空样式，由 t-tag 的 theme 属性处理
  if (renderConfig.mode === 'theme') {
    return {};
  }

  // style 模式，使用自定义样式
  if (renderConfig.style) {
    return {
      background: `${renderConfig.style.background} !important`,
      color: `${renderConfig.style.color} !important`,
      borderColor: `${renderConfig.style.borderColor} !important`,
    };
  }

  return {};
};

// 状态主题（用于 t-tag 的 theme 属性）
const getStatusTheme = (status: number) => {
  if (!status) return 'default';
  const statusItem = workOrderStatusList.value.find((s) => s.value === status);
  if (!statusItem) return 'default';

  const color = statusItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);

  return renderConfig.mode === 'theme' ? renderConfig.theme : 'default';
};

// 状态文本转换（使用字典数据）
const getStatusText = (status: number) => {
  if (!status) return '未知';
  const statusItem = workOrderStatusList.value.find((s) => s.value === status);
  return statusItem ? statusItem.label : '未知';
};

// 订单类型标签渲染配置
const getOrderTypeTagProps = (orderType: number) => {
  if (!orderType) return {};
  const typeItem = orderTypeList.value.find((t) => t.value === orderType);
  if (!typeItem || !typeItem.color) return {};

  return getDictColorRenderConfig(typeItem.color);
};

// 订单类型文本转换（从字典数据获取）
const getOrderTypeText = (orderType: number) => {
  const typeItem = orderTypeList.value.find((t) => t.value === orderType);
  return typeItem ? typeItem.label : '未知';
};

// ✨ 查看模具订单详情
const handleViewMoldOrder = (orderId: number) => {
  if (orderId) {
    router.push(`/mold/order/detail/${orderId}`);
  }
};

// ✨ 优先级标签样式（使用统一的颜色配置系统）
const getPriorityStyle = (priority: string) => {
  if (!priority) return getDictColorRenderConfig('default');
  const priorityItem = priorityList.value.find((p) => p.value === priority);
  if (!priorityItem) return getDictColorRenderConfig('default');

  const color = priorityItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);

  // 如果是 theme 模式，直接返回空样式，由 t-tag 的 theme 属性处理
  if (renderConfig.mode === 'theme') {
    return {};
  }

  // style 模式，使用自定义样式
  if (renderConfig.style) {
    return {
      background: `${renderConfig.style.background} !important`,
      color: `${renderConfig.style.color} !important`,
      borderColor: `${renderConfig.style.borderColor} !important`,
    };
  }

  return getDictColorRenderConfig('default');
};

// ✨ 获取优先级主题（用于 t-tag 的 theme 属性）
const getPriorityTheme = (priority: string) => {
  if (!priority) return 'default';
  const priorityItem = priorityList.value.find((p) => p.value === priority);
  if (!priorityItem) return 'default';

  const color = priorityItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);

  return renderConfig.mode === 'theme' ? renderConfig.theme : 'default';
};

// ✨ 优先级文本转换
const getPriorityText = (priority: string) => {
  if (!priority) return '未知';
  const priorityItem = priorityList.value.find((p) => p.value === priority);
  return priorityItem ? priorityItem.label : '未知';
};

// 搜索
const handleSearch = () => {
  pagination.current = 1;
  getWorkOrderList();
};

const handleReset = () => {
  formData.moldCode = '';
  formData.workCode = '';
  formData.customerId = undefined;
  formData.status = undefined;
  formData.orderType = undefined;
  formData.processStage = undefined;
  getWorkOrderList();
};

// 新增 - 跳转到表单页面
const handleAdd = () => {
  router.push('/mold/work-order/form');
};

// 编辑
const handleEdit = async (row: any) => {
  isEdit.value = true;
  dialogTitle.value = '编辑工单';
  Object.assign(formModel, {
    id: row.id,
    workCode: row.work_code,
    orderType: row.order_type || 1,
    moldOrderId: row.mold_order_id,
    moldId: row.mold_id,
    planStart: row.plan_start,
    planEnd: row.plan_end,
    status: row.status,
    remark: row.remark,
    // 新模工单字段（编辑时从关联模具中获取）
    newMoldCode: '',
    newMoldName: '',
    customerId: row.mold?.customer_id || undefined,
    newMoldType: row.mold?.mold_type || undefined,
    newMaterialType: row.mold?.material_type || undefined,
    newMoldRemark: '',
  });

  // 如果是新模工单且已关联模具，加载已创建的模具信息
  if (row.order_type === ORDER_TYPE.NEW_MOLD && row.mold_id) {
    await loadExistingMoldInfo(row.mold_id);
  } else {
    Object.assign(existingMoldInfo, {
      moldCode: '',
      moldName: '',
      customerName: '',
      moldTypeText: '',
      materialTypeText: '',
    });
  }

  await Promise.all([fetchCustomerList(), fetchOrderList()]);
  if (row.order_type === ORDER_TYPE.REPAIR || row.order_type === ORDER_TYPE.MODIFY) {
    await fetchMoldList();
  }

  originalFormData.value = JSON.parse(
    JSON.stringify({
      id: formModel.id,
      workCode: formModel.workCode,
      orderType: formModel.orderType,
      moldOrderId: formModel.moldOrderId,
      moldId: formModel.moldId,
      planStart: formModel.planStart,
      planEnd: formModel.planEnd,
      remark: formModel.remark,
      newMoldCode: formModel.newMoldCode,
      newMoldName: formModel.newMoldName,
      customerId: formModel.customerId,
      newMoldType: formModel.newMoldType,
      newMaterialType: formModel.newMaterialType,
      newMoldRemark: formModel.newMoldRemark,
    }),
  );

  originalMoldOrderId.value = formModel.moldOrderId;

  // 编辑模式下，模具订单字段默认锁定
  if (isEdit.value) {
    moldOrderLocked.value = true;
  }

  dialogVisible.value = true;

  // 等待对话框打开并DOM更新后，清除验证状态
  await nextTick();
  setTimeout(() => {
    formRef.value?.clearValidate();
  }, 100);
};

const handleDelete = async (row: any) => {
  // 检查工单状态，如果已进入加工阶段（status >= 2），禁止删除
  if (getStatusLevel(row.status) >= WORK_ORDER_STATUS.PROCESSING) {
    MessagePlugin.warning('工单已进入加工阶段，无法删除，请使用“取消”功能');
    return;
  }

  const dialog = DialogPlugin.confirm({
    header: '确认删除',
    body: `确定要删除工单 "${row.work_code}" 吗？此操作不可恢复！`,
    theme: 'default',
    confirmBtn: '确定删除',
    cancelBtn: '取消',
    onConfirm: async () => {
      try {
        await deleteMoldWorkOrder(row.id);
        MessagePlugin.success('删除成功');
        getWorkOrderList();
        dialog.destroy();
      } catch (error: any) {
        MessagePlugin.error(error.message || '删除失败');
        dialog.destroy();
      }
    },
  });
};

// 打开取消工单对话框
const handleCancelWorkOrder = (row: any) => {
  currentWorkOrderForCancel.value = row;
  cancelForm.reason = '';
  cancelDialogVisible.value = true;
};

// 确认取消工单
const handleConfirmCancelWorkOrder = async () => {
  if (!currentWorkOrderForCancel.value) return;

  try {
    await cancelMoldWorkOrder({
      id: currentWorkOrderForCancel.value.id,
      reason: cancelForm.reason || '无',
    });
    MessagePlugin.success('工单已取消');
    cancelDialogVisible.value = false;
    getWorkOrderList();
  } catch (error: any) {
    MessagePlugin.error(error.message || '取消失败');
  }
};

// 查看模具档案
const handleViewMold = (moldCode: string) => {
  if (!moldCode) {
    MessagePlugin.warning('未找到模具信息');
    return;
  }
  // 跳转到模具档案页面并携带模具编号参数
  window.location.href = `/mold?moldCode=${encodeURIComponent(moldCode)}`;
};

const loadExistingMoldInfo = async (moldId: number) => {
  try {
    const { getMoldById } = await import('@/api/mold/index');
    const mold = await getMoldById(moldId);
    if (mold) {
      Object.assign(existingMoldInfo, {
        moldCode: mold.mold_code || '',
        moldName: mold.mold_name || '',
        customerName: mold.customer?.cus_name || mold.customer?.name || '',
        moldTypeText: getMoldTypeText(mold.mold_type),
        materialTypeText: getMaterialTypeText(mold.material_type),
      });
    } else {
      MessagePlugin.warning('未找到模具信息');
    }
  } catch (error) {
    console.error('加载模具信息失败:', error);
    MessagePlugin.error('加载模具信息失败');
  }
};

// 模具类型文本转换
const getMoldTypeText = (type: number) => {
  const textMap: Record<number, string> = {
    1: '注塑模',
    2: '冲压模',
    3: '压铸模',
  };
  return textMap[type] || '-';
};

// 材质类型文本转换
const getMaterialTypeText = (type: number) => {
  const textMap: Record<number, string> = {
    1: '钢材',
    2: '铝材',
    3: '铜材',
  };
  return textMap[type] || '-';
};

// 订单类型变更
const handleOrderTypeChange = (value: any) => {
  // 切换订单类型时清空相关字段
  if (value === 1) {
    // 新模：清空模具选择
    formModel.moldId = undefined;
  } else {
    // 修模/改模：清空新模具信息
    formModel.newMoldCode = '';
    formModel.newMoldName = '';
    formModel.customerId = undefined;
    formModel.newMoldType = undefined;
    formModel.newMaterialType = undefined;
    fetchMoldList();
  }
};

const isMoldInMaintenancePhase = (moldId?: number): boolean => {
  if (!moldId) return false;
  const mold = moldList.value.find((m) => m.id === moldId);
  if (!mold) return false;
  // 状态为 'REPAIRING' (维修中), 'IN_STOCK' (仓库库存), 'DELIVERED' (已交付), 'SCRAPPED' (报废)
  const maintenanceStatuses = ['REPAIRING', 'IN_STOCK', 'DELIVERED', 'SCRAPPED'];
  return maintenanceStatuses.includes(String(mold.life_cycle_status));
};

const getLifeCycleStatusText = (status: string | undefined): string => {
  if (!status) return '未知';
  const statusMap: Record<string, string> = {
    DESIGNING: '设计中',
    PROCESSING: '加工中',
    TESTING: '待试模',
    TESTED: '已试模',
    COMPLETED: '已完成',
    REPAIRING: '维修中',
    IN_STOCK: '仓库库存',
    DELIVERED: '已交付客户',
    SCRAPPED: '报废',
  };
  return statusMap[String(status)] || '未知';
};

const fetchCustomerList = async () => {
  try {
    const res = await getCustomerList({ page: 1, page_size: 100 });
    customerList.value = (res.list || []).map((item: any) => ({
      id: item.id,
      name: item.cus_name,
    }));
  } catch (error) {
    console.error('获取客户列表失败:', error);
    customerList.value = [];
  }
};

const fetchMoldList = async () => {
  try {
    const res = await getMoldList({ page: 1, page_size: 100 });
    moldList.value = res.list || [];
  } catch (error) {
    console.error('获取模具列表失败:', error);
    moldList.value = [];
  }
};

const filteredOrderList = computed(() => {
  if (!formModel.orderType) return orderList.value;

  // 新模工单：只显示新模订单（order_type=1）
  if (formModel.orderType === ORDER_TYPE.NEW_MOLD) {
    return orderList.value.filter((order) => order.order_type === 1);
  }
  // 修模/改模工单：只显示修模订单（order_type=2或3）
  if (formModel.orderType === ORDER_TYPE.REPAIR || formModel.orderType === ORDER_TYPE.MODIFY) {
    return orderList.value.filter((order) => order.order_type === 2 || order.order_type === 3);
  }

  return orderList.value;
});

const fetchOrderList = async () => {
  try {
    const res = await getMoldOrderList({ page: 1, page_size: 100 });
    orderList.value = res.list || [];
  } catch (error) {
    console.error('获取订单列表失败:', error);
    orderList.value = [];
  }
};

// 模具订单字段锁定/解锁切换
const toggleMoldOrderLock = () => {
  if (moldOrderLocked.value) {
    // 解锁时需要二次确认
    const dialog = DialogPlugin.confirm({
      header: '确认解锁',
      body: () =>
        h('div', [
          '解锁后可以修改模具订单，但请注意：',
          h('br'),
          h(
            'span',
            { style: { color: '#e34d59', fontWeight: 'bold' } },
            '只能选择与当前工单类型匹配的订单',
          ),
          h('br'),
          '• 新模工单只能关联新模订单',
          h('br'),
          '• 修模/改模工单只能关联修模/改模订单',
        ]),
      theme: 'warning',
      confirmBtn: '确认解锁',
      cancelBtn: '取消',
      onConfirm: () => {
        moldOrderLocked.value = false;
        MessagePlugin.info('已解锁，可以修改模具订单');
        dialog.destroy();
      },
      onCancel: () => {
        dialog.destroy();
      },
      onClose: () => {
        dialog.destroy();
      },
    });
  } else {
    // 锁定时检查是否有修改
    if (formModel.moldOrderId !== originalMoldOrderId.value) {
      const dialog = DialogPlugin.confirm({
        header: '确认锁定',
        body: '锁定后将无法修改模具订单，是否确认锁定？',
        theme: 'warning',
        confirmBtn: '确认锁定',
        cancelBtn: '再想想',
        onConfirm: () => {
          moldOrderLocked.value = true;
          MessagePlugin.success('已锁定');
          dialog.destroy();
        },
        onCancel: () => {
          dialog.destroy();
        },
        onClose: () => {
          dialog.destroy();
        },
      });
    } else {
      moldOrderLocked.value = true;
      MessagePlugin.success('已锁定');
    }
  }
};

// 模具订单变更处理
const handleMoldOrderChange = (value: any) => {
  // 如果选择的订单类型与当前工单类型不匹配，给出警告
  if (value) {
    const selectedOrder = orderList.value.find((order) => order.id === value);
    if (selectedOrder) {
      // 新模工单只能关联新模订单
      if (formModel.orderType === ORDER_TYPE.NEW_MOLD && selectedOrder.order_type !== 1) {
        MessagePlugin.warning('新模工单只能关联新模订单');
        formModel.moldOrderId = originalMoldOrderId.value;
        return;
      }
      // 修模/改模工单只能关联修模/改模订单
      if (
        (formModel.orderType === ORDER_TYPE.REPAIR || formModel.orderType === ORDER_TYPE.MODIFY) &&
        selectedOrder.order_type !== 2 &&
        selectedOrder.order_type !== 3
      ) {
        MessagePlugin.warning('修模/改模工单只能关联修模/改模订单');
        formModel.moldOrderId = originalMoldOrderId.value;
      }
    }
  }
};

const fetchProcessTemplates = async () => {
  try {
    const res = await getActiveProcessTemplates();
    processTemplates.value = res || [];

    // 等待 DOM 更新后重新计算列宽
    await nextTick();
    await calculateProcessColumnWidth();
  } catch (error) {
    console.error('获取工序模板失败:', error);
    processTemplates.value = [];
  }
};

const calculateProcessColumnWidth = async () => {
  const measureEl = document.createElement('div');
  measureEl.style.cssText = `
    position: absolute;
    visibility: hidden;
    white-space: nowrap;
    font-size: 14px;
    display: inline-flex;
    align-items: flex-start;
    gap: 12px;
    padding: 4px 8px;
  `;

  let html = '';
  groupedProcesses.value.forEach((group, index) => {
    // 分组块
    html += '<div style="display: flex; flex-direction: column; align-items: center; gap: 6px;">';
    html += `<div style="font-size: 12px; color: #999; line-height: 1;">${escapeHtml(group.groupName)}</div>`;
    html += '<div style="display: flex; gap: 6px;">';
    group.templates.forEach(() => {
      html +=
        '<span style="display: inline-block; padding: 2px 12px; background: #e8f5e9; border-radius: 16px; font-size: 12px;">工序</span>';
    });
    html += '</div></div>';

    // 箭头
    if (index < groupedProcesses.value.length - 1) {
      html += '<span style="color: #ccc; font-size: 14px; margin: 0 4px;">→</span>';
    }
  });

  measureEl.innerHTML = html;
  document.body.appendChild(measureEl);

  // 测量实际宽度
  const actualWidth = measureEl.scrollWidth;
  document.body.removeChild(measureEl);

  // 更新列宽（加上额外宽度用于与后面一列的边距）
  processColumnActualWidth.value = Math.max(200, actualWidth + 60);
};

const getProcessStatusValue = (row: any, templateId: number): number => {
  if (!row.processes) return 0;
  const process = row.processes.find((p: any) => p.process_template_id === templateId);
  return process ? process.is_completed : 0;
};

const getProcessCompletedAt = (row: any, templateId: number): string | null => {
  if (!row.processes) return null;
  const process = row.processes.find((p: any) => p.process_template_id === templateId);
  return process && process.is_completed === 1 ? process.completed_at : null;
};

// 格式化完成时间
const formatCompletedTime = (timeStr: string | null): string => {
  if (!timeStr) return '';
  try {
    return dayjs(timeStr).format('YYYY-MM-DD HH:mm');
  } catch (error) {
    console.error('日期格式化错误:', error);
    return timeStr;
  }
};

const isGroupCompleted = (row: any, groupTemplates: any[]): boolean => {
  if (!row.processes) return false;
  // 组内所有工序都完成才算完成
  return groupTemplates.every((template) => {
    const process = row.processes.find((p: any) => p.process_template_id === template.id);
    return process && process.is_completed === 1;
  });
};

const hasIncompleteProcess = (row: any, groupTemplates: any[]): boolean => {
  if (!row.processes) return true;
  return groupTemplates.some((template) => {
    const process = row.processes.find((p: any) => p.process_template_id === template.id);
    return !process || process.is_completed !== 1;
  });
};

// 判断是否为当前工序（基于分组逻辑）
const isCurrentProcess = (row: any, templateId: number): boolean => {
  if (!row.processes) return false;

  // 找到当前模板所在的分组
  let currentTemplateGroup = null;
  let currentGroupIndex = -1;
  for (let i = 0; i < groupedProcesses.value.length; i++) {
    const group = groupedProcesses.value[i];
    if (group.templates.find((t) => t.id === templateId)) {
      currentTemplateGroup = group;
      currentGroupIndex = i;
      break;
    }
  }

  if (!currentTemplateGroup) return false;

  // 检查当前模板是否已完成
  const process = row.processes.find((p: any) => p.process_template_id === templateId);
  if (process && process.is_completed === 1) return false; // 已完成

  // 检查当前模板所在分组的所有前置分组是否全部完成
  for (let i = 0; i < currentGroupIndex; i++) {
    const prevGroup = groupedProcesses.value[i];
    // 如果前置分组有未完成的工序，当前模板不能开始
    if (hasIncompleteProcess(row, prevGroup.templates)) {
      return false;
    }
  }

  // 所有前置分组都完成了，当前分组内的未完成工序都是当前工序（可并行）
  return true;
};

const calculateProgress = (row: any): number => {
  if (!row.processes || processTemplates.value.length === 0) return 0;

  let validCount = 0;
  let completedCount = 0;

  processTemplates.value.forEach((template) => {
    const process = row.processes.find((p: any) => p.process_template_id === template.id);
    // 排除跳过的工序（is_completed === 2）
    if (!process || process.is_completed !== 2) {
      validCount++;
      if (process && process.is_completed === 1) {
        completedCount++;
      }
    }
  });

  // 基于有效工序总数计算进度
  if (validCount === 0) return 0;
  return Math.round((completedCount * 100) / validCount);
};

// 按分组整理工序模板
const groupedProcesses = computed(() => {
  const groups: Map<number, { groupId: number; groupName: string; templates: any[] }> = new Map();

  processTemplates.value.forEach((template) => {
    const groupId = template.group_id || 0;
    const groupName = template.group_name || '未分组';

    if (!groups.has(groupId)) {
      groups.set(groupId, { groupId, groupName, templates: [] });
    }

    groups.get(groupId)!.templates.push(template);
  });

  // 按 group_id 排序
  return Array.from(groups.values()).sort((a, b) => a.groupId - b.groupId);
});

// 工序详情面板
const processDrawerVisible = ref(false);
const currentWorkOrder = ref<any>(null);

// 显示工序详情
const showProcessDetail = (row: any) => {
  currentWorkOrder.value = row;
  processDrawerVisible.value = true;
};

// 关闭工序详情
const closeProcessDrawer = () => {
  processDrawerVisible.value = false;
  currentWorkOrder.value = null;
};

const getProcessShortName = (name: string): string => {
  const nameMap: Record<string, string> = {
    CNC加工: 'CNC',
    磨床: '磨床',
    放电加工: '放电',
    线切割: '线切',
    钳工: '钳工',
    抛光: '抛光',
    装配: '装配',
    检验: '检验',
  };
  return nameMap[name] || name;
};

// 判断当前分组是否可以开始（前置分组都已全部完成）
const isGroupCanBeStarted = (row: any, currentGroupIndex: number): boolean => {
  for (let i = 0; i < currentGroupIndex; i++) {
    const prevGroup = groupedProcesses.value[i];
    // 如果前置分组有未完成的工序，当前分组不能开始
    if (!isGroupCompleted(row, prevGroup.templates)) {
      return false;
    }
  }
  return true;
};

const getProcessItemClass = (
  row: any,
  templateId: number,
  index: number,
  groupTemplates: any[],
  groupIndex: number,
): string => {
  const status = getProcessStatusValue(row, templateId);
  if (status === 1) return 'process-item-completed';
  if (status === 2) return 'process-item-skipped';
  // 当前分组可以开始（前置分组都已完成），且当前分组未完成
  if (isGroupCanBeStarted(row, groupIndex) && !isGroupCompleted(row, groupTemplates)) {
    return 'process-item-active';
  }
  return 'process-item-pending';
};

const getProcessNameClass = (row: any, templateId: number, _index: number): string => {
  const status = getProcessStatusValue(row, templateId);
  if (status === 1) return 'process-completed';
  if (status === 2) return 'process-skipped';
  if (isCurrentProcess(row, templateId)) return 'process-active';
  return 'process-pending';
};

// 统计已完成工序数
const getCompletedCount = (row: any): number => {
  if (!row.processes) return 0;
  return row.processes.filter((p: any) => p.is_completed === 1).length;
};

// 统计跳过工序数
const getSkippedCount = (row: any): number => {
  if (!row.processes) return 0;
  return row.processes.filter((p: any) => p.is_completed === 2).length;
};

// 统计进行中工序数（第一个未完成的）
const getInProgressCount = (row: any): number => {
  if (!row.processes) return 0;
  let found = false;
  for (let i = 0; i < processTemplates.value.length; i++) {
    const template = processTemplates.value[i];
    const process = row.processes.find((p: any) => p.process_template_id === template.id);
    // 跳过状态的不计入
    if (process && process.is_completed === 2) continue;
    if (!process || process.is_completed !== 1) {
      found = true;
      break;
    }
  }
  return found ? 1 : 0;
};

// 统计待开工工序数
const getPendingCount = (row: any): number => {
  if (!row.processes) return 0;
  const completed = getCompletedCount(row);
  const skipped = getSkippedCount(row);
  const inProgress = getInProgressCount(row);
  return processTemplates.value.length - completed - skipped - inProgress;
};

// 状态优先级映射（后端返回语义化字符串如 "COMPLETED"，前端常量是数字）
// 将状态值统一转换为可比较的数字级别
const STATUS_LEVEL_MAP: Record<string, number> = {
  NOT_STARTED: 1,
  PROCESSING: 2,
  COMPLETED: 3,
  TESTED: 4,
  STOCKED_IN: 5,
  DELIVERED: 6,
  CANCELLED: 7,
};

// 获取状态级别（兼容后端字符串和前端数字）
const getStatusLevel = (status: any): number => {
  if (typeof status === 'number') return status;
  return STATUS_LEVEL_MAP[String(status)] ?? -1;
};

// 判断是否可以试模（工单必须有工序已经开始，不能是未开始状态）
const canTryTest = (row: any): boolean => {
  if (!row) return false;
  // 工单状态必须是加工中、已完成、已试模及以上
  return getStatusLevel(row.status) >= WORK_ORDER_STATUS.PROCESSING;
};

const getTryTestDisabledTip = (row: any): string => {
  if (!row) return '';
  if (getStatusLevel(row.status) === WORK_ORDER_STATUS.NOT_STARTED) {
    return '工序尚未开始，无法添加试模记录';
  }
  return '';
};

// 工单是否被锁定（直接使用后端返回的 is_locked 字段）
const isWorkOrderLocked = (row: any): boolean => {
  if (!row) return false;
  return row.is_locked === true;
};

const getMutexTip = (row: any): string => {
  if (!row) return '';
  if (isWorkOrderLocked(row)) {
    return '该模具当前有其他工单正在进行中，已自动锁定';
  }
  if (row.is_cancelled === 1) {
    return '工单已取消，无法操作';
  }
  return '';
};

// 切换工序状态
const toggleProcessStatus = async (row: any, templateId: number, newValue: number) => {
  const currentStatus = getProcessStatusValue(row, templateId);

  // 如果是从“已完成”或“跳过”切换到其他状态，需要二次确认
  if (currentStatus === 1 || currentStatus === 2) {
    const template = processTemplates.value.find((t) => t.id === templateId);
    const processName = template?.process_name || '该工序';
    const statusText = currentStatus === 1 ? '完成' : '跳过';

    try {
      await new Promise((resolve, reject) => {
        const confirmDialog = DialogPlugin.confirm({
          header: '确认撤销',
          body: () =>
            h('div', [
              '确定要撤销 ',
              h('span', { style: { fontSize: '16px', fontWeight: 'bold' } }, processName),
              ' 的 ',
              h('span', { style: { fontSize: '16px', fontWeight: 'bold' } }, statusText),
              ' 状态吗？',
            ]),
          confirmBtn: '确定',
          cancelBtn: '取消',
          onConfirm: () => {
            confirmDialog.destroy();
            resolve(true);
          },
          onCancel: () => {
            confirmDialog.destroy();
            reject(new Error('User cancelled'));
          },
          onClose: () => {
            confirmDialog.destroy();
            reject(new Error('User cancelled'));
          },
        });
      });
    } catch {
      // 用户取消了操作，不执行任何逻辑
      return;
    }
  }

  try {
    const _result = await toggleProcess({
      work_order_id: row.id,
      process_template_id: templateId,
      is_completed: newValue,
    });
    MessagePlugin.success('操作成功');

    // ✅ 性能优化：只更新当前行的工序数据，不刷新整个列表
    const currentRow = tableData.value.find((w) => w.id === row.id);
    if (currentRow) {
      if (!currentRow.processes) {
        currentRow.processes = [];
      }

      let process = currentRow.processes.find((p: any) => p.process_template_id === templateId);
      if (!process) {
        process = { process_template_id: templateId, is_completed: 0, completed_at: null };
        currentRow.processes.push(process);
      }
      process.is_completed = newValue;

      // ✅ 更新完成时间：完成时设置当前时间（东八区），撤销或跳过时清空
      if (newValue === 1) {
        process.completed_at = dayjs().tz('Asia/Shanghai').format('YYYY-MM-DD HH:mm:ss');
      } else {
        process.completed_at = null;
      }

      // 重新计算进度
      const validCount = currentRow.processes.filter((p: any) => p.is_completed !== 2).length;
      const completedCount = currentRow.processes.filter((p: any) => p.is_completed === 1).length;
      currentRow.progress = validCount > 0 ? Math.floor((completedCount * 100) / validCount) : 0;

      // 根据进度更新状态（使用语义化字符串，与后端一致）
      if (currentRow.progress === 0 && getStatusLevel(currentRow.status) < 4) {
        currentRow.status = 'NOT_STARTED'; // 未开始
      } else if (currentRow.progress > 0 && currentRow.progress < 100 && getStatusLevel(currentRow.status) < 4) {
        currentRow.status = 'PROCESSING'; // 加工中
      } else if (currentRow.progress === 100 && getStatusLevel(currentRow.status) < 4) {
        currentRow.status = 'COMPLETED'; // 已完成（待试模）
      }

      // 更新当前工序详情的数据
      if (currentWorkOrder.value?.id === row.id) {
        currentWorkOrder.value = { ...currentRow };
      }
    }
  } catch (error: any) {
    console.error('切换工序状态失败:', error);
    MessagePlugin.error(error?.message || '操作失败');
  }
};

// 试模管理
const handleTryTest = async (row: any) => {
  currentWorkOrderId.value = row.id;
  currentWorkOrder.value = row; // 同步设置当前工单对象

  tryTestDialogVisible.value = true;
  getTryTestList();
};

const getTryTestList = async () => {
  if (!currentWorkOrderId.value) return;
  try {
    const res = await getMoldTryTests({ work_order_id: currentWorkOrderId.value });
    tryTestList.value = res.list || [];
  } catch {
    MessagePlugin.error('获取试模记录失败');
  }
};

const handleAddTryTest = () => {
  tryTestForm.id = undefined;
  tryTestForm.mold_work_id = currentWorkOrderId.value;
  tryTestForm.testTime = dayjs().tz('Asia/Shanghai').format('YYYY-MM-DD');
  tryTestForm.okQty = undefined;
  tryTestForm.badQty = undefined;
  tryTestForm.result = '';
  tryTestForm.testMode = '';
  tryTestForm.remark = '';

  setTimeout(() => {
    tryTestFormRef.value?.reset();
  }, 0);

  tryTestFormDialogVisible.value = true;
};

const handleEditTryTest = (row: any) => {
  // 填充表单
  tryTestForm.id = row.id;
  tryTestForm.mold_work_id = row.mold_work_id;
  // 格式化时间：只显示日期部分
  if (row.test_time) {
    tryTestForm.testTime = formatDate(row.test_time);
  } else {
    tryTestForm.testTime = '';
  }
  tryTestForm.okQty = row.ok_qty ?? undefined;
  tryTestForm.badQty = row.bad_qty ?? undefined;
  tryTestForm.result = row.result || '';
  tryTestForm.testMode = row.test_mode || '';
  tryTestForm.remark = row.remark || '';

  setTimeout(() => {
    tryTestFormRef.value?.reset();
  }, 0);

  tryTestFormDialogVisible.value = true;
};

const handleSaveTryTest = async () => {
  try {
    // 手动验证必填字段
    if (!tryTestForm.result || tryTestForm.result.trim() === '') {
      MessagePlugin.error('请选择试模结果');
      return;
    }

    // 将前端驼峰命名转换为后端下划线命名
    const submitData = {
      mold_work_id: tryTestForm.mold_work_id,
      test_time: tryTestForm.testTime
        ? dayjs(tryTestForm.testTime).tz('Asia/Shanghai').format('YYYY-MM-DD HH:mm:ss')
        : null,
      ok_qty: tryTestForm.okQty,
      bad_qty: tryTestForm.badQty,
      result: tryTestForm.result,
      test_mode: tryTestForm.testMode,
      remark: tryTestForm.remark,
    };

    if (tryTestForm.id) {
      // 编辑
      await updateMoldTryTest(tryTestForm.id, submitData);
      MessagePlugin.success('更新成功');
    } else {
      // 新增
      await createMoldTryTest(submitData);
      MessagePlugin.success('创建成功');

      // 如果是新模工单且试模结果为"合格"，则更新工单状态为已试模
      const currentWorkOrder = tableData.value.find((w) => w.id === currentWorkOrderId.value);
      if (
        currentWorkOrder &&
        getStatusLevel(currentWorkOrder.status) < WORK_ORDER_STATUS.TESTED &&
        tryTestForm.result === '合格'
      ) {
        await updateMoldWorkOrder({
          id: currentWorkOrderId.value,
          status: 'TESTED',
        });
        MessagePlugin.success('工单状态已更新为已试模');
      }
    }

    tryTestFormDialogVisible.value = false;
    getTryTestList();
    getWorkOrderList(); // 刷新工单列表
  } catch (error: any) {
    // 如果是表单验证失败，error 会包含验证错误信息
    console.error('保存失败:', error);
    MessagePlugin.error(`保存失败: ${error?.message || '未知错误'}`);
  }
};

const handleDeleteTryTest = async (row: any) => {
  try {
    await deleteMoldTryTest(row.id);
    MessagePlugin.success('删除成功');
    getTryTestList();
    getWorkOrderList(); // 刷新工单列表以更新状态
  } catch {
    MessagePlugin.error('删除失败');
  }
};

const canStockIn = (row: any): boolean => {
  if (!row.mold_id) return false;

  // 使用统一的状态级别比较，兼容字符串和数字
  return getStatusLevel(row.status) === WORK_ORDER_STATUS.TESTED;
};

// 入库操作
const handleStockIn = async (row: any) => {
  currentWorkOrderForStockIn.value = row;
  stockInForm.moldCode = row.mold?.mold_code || '';
  stockInForm.warehouseId = undefined;
  stockInForm.remark = '';
  await fetchWarehouseList();
  stockInDialogVisible.value = true;
};

// 确认入库
const handleConfirmStockIn = async () => {
  if (!stockInForm.warehouseId) {
    MessagePlugin.warning('请选择仓库');
    return;
  }

  try {
    const { createMoldStock } = await import('@/api/mold');
    await createMoldStock({
      mold_id: currentWorkOrderForStockIn.value.mold_id,
      warehouse_id: stockInForm.warehouseId,
      stock_type: 1, // 1=入库
      operate_time: dayjs().tz('Asia/Shanghai').format('YYYY-MM-DD HH:mm:ss'),
      quantity: 1,
      remark: stockInForm.remark || `工单 ${currentWorkOrderForStockIn.value.work_code} 入库`,
    });

    const { updateMoldWorkOrder } = await import('@/api/mold');
    await updateMoldWorkOrder({
      id: currentWorkOrderForStockIn.value.id,
      status: 'STOCKED_IN', // ✨ 使用语义标签
      warehouse_id: stockInForm.warehouseId,
    });

    MessagePlugin.success('入库成功');
    stockInDialogVisible.value = false;
    getWorkOrderList();
  } catch {
    MessagePlugin.error('入库失败');
  }
};

const fetchWarehouseList = async () => {
  try {
    const res = await getWarehouseList({ page: 1, page_size: 100 });
    warehouseList.value = res.list || [];
  } catch (error) {
    console.error('获取仓库列表失败:', error);
    warehouseList.value = [];
  }
};
// 提交表单
const handleSubmit = async () => {
  try {
    // 表单验证（TDesign: 成功返回undefined，失败返回错误对象）
    const validateResult = await formRef.value?.validate();
    if (validateResult) {
      MessagePlugin.warning('请检查表单填写是否完整');
      return;
    }

    // 编辑模式下，检查数据是否有修改
    if (isEdit.value && !hasFormDataChanged()) {
      MessagePlugin.info('数据未做修改，无需提交');
      dialogVisible.value = false;
      return;
    }

    // 根据订单类型验证必填字段
    if (formModel.orderType === ORDER_TYPE.NEW_MOLD) {
      // 新模工单：验证新模具信息
      // 编辑时如果已关联模具，则不需要验证新模具信息（模具已创建）
      if (!isEdit.value || !formModel.moldId) {
        if (!formModel.newMoldCode) {
          MessagePlugin.warning('请输入模具编号');
          return;
        }
        if (!formModel.newMoldName) {
          MessagePlugin.warning('请输入产品名称');
          return;
        }
        if (!formModel.customerId) {
          MessagePlugin.warning('请选择客户');
          return;
        }
      }
    } else if (
      formModel.orderType === ORDER_TYPE.REPAIR ||
      formModel.orderType === ORDER_TYPE.MODIFY
    ) {
      // 修模/改模工单：验证模具选择
      if (!formModel.moldId) {
        MessagePlugin.warning('请选择模具');
        return;
      }
      // 🔒 强约束：检查模具是否处于运维阶段（禁止对已进入运维阶段的模具创建修模工单）
      if (isMoldInMaintenancePhase(formModel.moldId)) {
        const mold = moldList.value.find((m) => m.id === formModel.moldId);
        MessagePlugin.error(
          `该模具当前状态为“${getLifeCycleStatusText(mold?.life_cycle_status)}”，无法创建修模工单`,
        );
        return;
      }
    }

    const data: any = {
      work_code: formModel.workCode,
      order_type: formModel.orderType,
    };

    if (formModel.moldOrderId) data.mold_order_id = formModel.moldOrderId;
    if (formModel.moldId) data.mold_id = formModel.moldId;
    if (formModel.planStart) data.plan_start = formModel.planStart;
    if (formModel.planEnd) data.plan_end = formModel.planEnd;
    if (formModel.remark) data.remark = formModel.remark;

    // 新模工单：添加新模具信息
    if (formModel.orderType === ORDER_TYPE.NEW_MOLD) {
      data.new_mold_code = formModel.newMoldCode;
      data.new_mold_name = formModel.newMoldName;
      data.customer_id = formModel.customerId;
      if (formModel.newMoldType) data.new_mold_type = formModel.newMoldType;
      if (formModel.newMaterialType) data.new_material_type = formModel.newMaterialType;
      if (formModel.newMoldRemark) data.new_mold_remark = formModel.newMoldRemark;
    }

    if (isEdit.value) {
      data.id = formModel.id;
      await updateMoldWorkOrder(data);
      MessagePlugin.success('更新成功');
    } else {
      await createMoldWorkOrder(data);
      MessagePlugin.success('创建成功');
    }

    dialogVisible.value = false;
    getWorkOrderList();
  } catch (error: any) {
    console.error('提交工单失败:', error);
    if (error?.message) {
      MessagePlugin.error(error.message);
    } else {
      MessagePlugin.error('操作失败');
    }
  }
};

// 取消按钮
const handleCancel = () => {
  dialogVisible.value = false;
  moldOrderLocked.value = true;
};

// 对话框完全打开后清除验证（编辑模式下避免误报验证错误）
const handleDialogOpened = async () => {
  await nextTick();
  // 清除表单验证状态，避免编辑模式下打开对话框时误报验证错误
  formRef.value?.clearValidate();
};

// 分页变化
const rehandlePageChange = (curr: any) => {
  pagination.current = curr.current;
  pagination.pageSize = curr.pageSize;
  getWorkOrderList();
};

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('mold_work_order_table_columns');
    // 获取强制可见的列（lockHide 列）
    const forcedVisibleKeys = allColumns.filter((col) => col.lockHide).map((col) => col.colKey);

    if (saved) {
      const savedData = JSON.parse(saved);
      // 兼容新格式 { order: [], visible: [] } 和旧格式数组
      if (Array.isArray(savedData)) {
        // 旧格式：直接是数组
        selectedColumns.value = savedData;
      } else if (savedData.visible) {
        // 新格式：提取 visible 字段
        selectedColumns.value = savedData.visible;
      } else {
        selectedColumns.value = [...DEFAULT_COLUMNS];
      }
    } else {
      selectedColumns.value = [...DEFAULT_COLUMNS];
    }

    // 确保强制可见列始终在 selectedColumns 中
    if (forcedVisibleKeys.length > 0) {
      selectedColumns.value = [...new Set([...selectedColumns.value, ...forcedVisibleKeys])];
    }
  } catch (error) {
    console.error('加载列配置失败:', error);
    selectedColumns.value = [...DEFAULT_COLUMNS];
  }
};

// 打开列设置对话框
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// 列配置变化时的回调
const handleColumnsChange = (columns: string[]) => {
  selectedColumns.value = columns;
};

// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'moldCode', placeholder: '请输入模具编号' },
  { type: 'input', field: 'workCode', placeholder: '请输入工单号' },
  {
    type: 'select',
    field: 'customerId',
    placeholder: '请选择客户',
    options: customerList,
    labelKey: 'name',
    valueKey: 'id',
  },
  { type: 'select', field: 'status', placeholder: '请选择状态', options: workOrderStatusList },
  { type: 'select', field: 'orderType', placeholder: '请选择订单类型', options: orderTypeList },
  {
    type: 'select',
    field: 'processStage',
    placeholder: '请选择工序阶段',
    options: processStageList,
  },
];

onMounted(async () => {
  await Promise.all([
    fetchProcessTemplates(),
    // 字典已通过 useDictRaw 自动加载
    fetchCustomerList(), // 加载客户列表
  ]);

  // 等待 DOM 更新后测量工序进度列的实际宽度
  await nextTick();
  await calculateProcessColumnWidth();

  // 检查 URL 参数，如果有模具编号则自动填充并搜索
  const route = window.location.search;
  const urlParams = new URLSearchParams(route);
  const moldCodeParam = urlParams.get('moldCode');

  if (moldCodeParam) {
    formData.moldCode = moldCodeParam;
  }

  getWorkOrderList();
  loadColumnConfig();
});

// 监听工序模板数据变化，自动重新计算列宽
watch(
  () => processTemplates.value,
  async () => {
    await nextTick();
    await calculateProcessColumnWidth();
  },
  { deep: true },
);
</script>
<style scoped>
/* 强制工序列内容完整显示 */
:deep(.t-table td[data-colkey='processes']) {
  overflow: visible !important;
  white-space: nowrap !important;
  padding: 16px 8px !important; /* 增加表格行高度，给胶囊更多垂直空间 */
  vertical-align: middle !important; /* 确保单元格内容垂直居中 */
}

:deep(.t-table td[data-colkey='processes'] .t-table__cell-content) {
  overflow: visible !important;
  white-space: nowrap !important;
}

.mold-work-order-container {
  padding: 16px;
}

.search-card,
.toolbar-card {
  margin-bottom: 16px;
}

.toolbar-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.toolbar-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.process-content {
  padding: 16px 0;
}

/* 列设置样式 */
.column-setting-content {
  padding: 16px 0;
}

.column-item {
  padding: 8px 0;
  border-bottom: 1px solid var(--td-border-level-1-color, #e7e7e7);
  display: flex;
  align-items: center;
  justify-content: space-between;

  &:last-child {
    border-bottom: none;
  }
}

.draggable-item {
  cursor: move;
  transition: background-color 0.2s;

  &:hover {
    background-color: var(--td-bg-color-container-hover, #f0f0f0);
  }
}

.drag-handle {
  color: var(--td-text-color-placeholder, #999);
  cursor: grab;
  user-select: none;
  font-size: 16px;
  letter-spacing: -2px;

  &:active {
    cursor: grabbing;
  }
}

.column-setting-footer {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.process-names {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
  gap: 4px;
  font-size: 13px;
  overflow-x: auto;
}

.process-name-item {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  padding: 4px 10px;
  border-radius: 16px;
  font-weight: 500;
  white-space: nowrap;
  transition: all 0.3s ease;
}

/* 已完成工序 - 绿色 */
.process-completed {
  background: var(--td-success-color-1, #e3f9e9);
  color: var(--td-success-color, #00a870);
  border: 1px solid var(--td-success-color-2, #b3ebd2);
}

/* 进行中工序 - 蓝色 */
.process-active {
  background: var(--td-brand-color-1, #e6f0ff);
  color: var(--td-brand-color, #0052d9);
  border: 1px solid var(--td-brand-color-2, #b3d1ff);
  animation: process-blink 2s ease-in-out infinite;
}

@keyframes process-blink {
  0%,
  100% {
    box-shadow: 0 0 0 0 rgb(0 82 217 / 30%);
  }

  50% {
    box-shadow: 0 0 0 4px rgb(0 82 217 / 0%);
  }
}

/* 待开工工序 - 灰色 */
.process-pending {
  background: var(--td-gray-color-1, #f5f5f5);
  color: var(--td-text-color-placeholder, #bbb);
  border: 1px solid var(--td-gray-color-3, #e7e7e7);
}

/* 跳过工序 - 灰色虚线+透明度 */
.process-skipped {
  background: var(--td-gray-color-1, #f5f5f5);
  color: var(--td-text-color-placeholder, #bbb);
  border: 1px dashed var(--td-gray-color-4, #c5c5c5);
  opacity: 0.7;
}

.process-arrow {
  color: var(--td-text-color-placeholder, #ccc);
  margin: 0 4px;
  font-size: 14px;
  font-weight: normal;
}

/* 工序详情抽屉样式 */
.process-drawer-content {
  padding: 16px 0;
}

.process-timeline {
  margin-bottom: 24px;
}

.timeline-item-active .t-timeline-item__dot {
  animation: timeline-pulse 2s ease-in-out infinite;
}

.timeline-item-completed .t-timeline-item__dot {
  background: var(--td-success-color, #00a870) !important;
}

@keyframes timeline-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 0 rgb(0 82 217 / 40%);
  }

  50% {
    box-shadow: 0 0 0 6px rgb(0 82 217 / 0%);
  }
}

.timeline-content {
  padding-left: 12px;
}

.timeline-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.timeline-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.timeline-actions {
  margin-top: 8px;
}

/* 进度统计卡片 */
.process-stats {
  margin-top: 16px;
}

.stats-container {
  display: flex;
  justify-content: space-around;
  align-items: center;
}

.stat-item {
  text-align: center;
  padding: 16px 24px;
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: var(--td-brand-color, #0052d9);
  margin-bottom: 4px;
}

.stat-label {
  font-size: 13px;
  color: var(--td-text-color-secondary, #888);
}

/* 分组式工序展示样式 */
.process-groups-drawer {
  margin-bottom: 24px;
}

.process-group-card {
  background: var(--td-bg-color-container, #fff);
  border: 1px solid var(--td-border-level-1-color, #e7e7e7);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
}

.group-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--td-border-level-1-color, #e7e7e7);
}

.group-card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.group-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: var(--td-brand-color, #0052d9);
  color: #fff;
  border-radius: 50%;
  font-size: 12px;
  font-weight: 600;
}

.group-hint {
  font-size: 12px;
  color: var(--td-text-color-placeholder, #bbb);
  font-weight: normal;
}

.group-card-processes {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.process-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background: var(--td-bg-color-container-hover, #f5f5f5);
  border-radius: 6px;
  transition: all 0.3s ease;
  border: 1px solid var(--td-border-level-1-color, #e7e7e7);
}

/* 已完成工序 - 绿色边框 */
.process-item-completed {
  border-color: var(--td-success-color, #00a870) !important;
}

/* 跳过工序 - 灰色虚线边框 */
.process-item-skipped {
  border-color: var(--td-gray-color-5, #c5c5c5) !important;
  border-style: dashed !important;
  opacity: 0.7;
}

/* 进行中工序 - 蓝色边框 */
.process-item-active {
  border-color: var(--td-brand-color, #0052d9) !important;
}

/* 待开工工序 - 灰色边框 */
.process-item-pending {
  border-color: var(--td-gray-color-3, #e7e7e7) !important;
}

.process-item:hover {
  background: var(--td-bg-color-container-active, #ebebeb);
}

.process-item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.process-item-name {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
  flex: 1;
}

.process-item-action {
  margin-top: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

/* 操作按钮组 */
.action-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
}

/* 完成按钮 - 绿色胶囊 */
.action-btn-complete {
  padding: 4px 16px !important;
  font-size: 13px !important;
  border-radius: 20px !important;
  transition: all 0.2s ease;
  min-width: 80px !important;
}

.action-btn-complete:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgb(27 200 138 / 30%);
}

/* 跳过按钮 - 灰色胶囊 */
.action-btn-skip {
  padding: 4px 16px !important;
  font-size: 13px !important;
  border-radius: 20px !important;
  transition: all 0.2s ease;
  min-width: 80px !important;
}

.action-btn-skip:hover {
  background-color: var(--td-bg-color-container-active, #ebebeb) !important;
  transform: translateY(-1px);
}

/* 取消按钮 - 文字链接样式 */
.action-btn-cancel {
  font-size: 12px !important;
  color: var(--td-text-color-secondary, #888) !important;
  padding: 4px 8px !important;
  transition: all 0.2s ease;
}

.action-btn-cancel:hover {
  color: var(--td-brand-color, #0052d9) !important;
}

.process-item-time {
  font-size: 14px;
  font-weight: 500;
  color: var(--td-text-color-primary);
  white-space: nowrap;
}

.group-arrow-divider {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 8px 0;
  color: var(--td-text-color-placeholder, #ccc);
}

/* TDesign 开关样式覆盖 - 改为绿色 */
:deep(.t-switch.t-is-checked) {
  background-color: var(--td-success-color, #00a870) !important;
}

:deep(.t-switch.t-is-checked .t-switch__handle) {
  border-color: var(--td-success-color, #00a870) !important;
}

/* 表格中的分组展示 */
.process-summary {
  cursor: pointer;
  padding: 0; /* 移除padding，让胶囊自然居中 */
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center; /* 垂直居中 */
  height: 100%; /* 占满整个单元格高度 */
}

/* 第一行：工序胶囊行 */
.process-row-capsules {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 4px 8px;
  border-radius: 6px;
  white-space: nowrap;
  transition: background 0.3s ease;
}

/* 鼠标悬停时显示灰色背景 */
.process-summary:hover .process-row-capsules {
  background: var(--td-bg-color-secondary, #f5f5f5);
}

/* 第二行：阶段名称行 */
.process-row-labels {
  display: inline-flex;
  align-items: center;
  gap: 12px; /* 与第一行保持一致 */
  padding: 0 8px;
  white-space: nowrap;
}

/* 工序胶囊组 */
.group-processes-list {
  display: flex;
  gap: 6px;
  flex-wrap: nowrap;
  align-items: center;
  flex-shrink: 0;
}

/* 分组名称（第二行） */
.group-name-label {
  font-size: 12px;
  color: var(--td-text-color-placeholder, #999);
  white-space: nowrap;
  line-height: 1;
  text-align: center;
  flex-shrink: 0;
  min-width: fit-content;
}

/* 分组间的箭头（第一行） */
.group-arrow {
  color: var(--td-text-color-placeholder, #ccc);
  font-size: 14px;
  font-weight: bold;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
}

/* 箭头占位符（第二行，保持与第一行对齐） */
.group-arrow-placeholder {
  flex-shrink: 0;
  width: 20px;
  height: 1px;
}

/* 模具编号样式 */
.mold-code-link {
  font-weight: 600;
  font-size: 15px;
}

/* 锁定输入框样式 */
.locked-input :deep(.t-input) {
  background-color: var(--td-bg-color-container, #f5f5f5) !important;
  cursor: not-allowed !important;
  opacity: 0.6;
}

.locked-input :deep(.t-input__inner) {
  cursor: not-allowed !important;
}
</style>
