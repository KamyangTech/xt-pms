<template>
  <div class="payment-stats">
    <t-row :gutter="16">
      <t-col :span="6">
        <t-card class="stat-card contract-total">
          <div class="stat-content">
            <div class="stat-label">合同总额</div>
            <div class="stat-value">¥{{ formatAmount(stats?.contract_total || 0) }}</div>
          </div>
        </t-card>
      </t-col>
      <t-col :span="6">
        <t-card class="stat-card received-total">
          <div class="stat-content">
            <div class="stat-label">已收总额</div>
            <div class="stat-value success">¥{{ formatAmount(stats?.actual_total || 0) }}</div>
          </div>
        </t-card>
      </t-col>
      <t-col :span="6">
        <t-card class="stat-card pending-total">
          <div class="stat-content">
            <div class="stat-label">未收总额</div>
            <div class="stat-value warning">¥{{ formatAmount(stats?.pending_total || 0) }}</div>
          </div>
        </t-card>
      </t-col>
      <t-col :span="6">
        <t-card class="stat-card progress">
          <div class="stat-content">
            <div class="stat-label">收款进度</div>
            <div class="stat-value">{{ calculateProgress().toFixed(1) }}%</div>
            <t-progress
              :percentage="calculateProgress()"
              :color="getProgressColor()"
              :label="false"
              class="progress-bar"
            />
          </div>
        </t-card>
      </t-col>
    </t-row>
  </div>
</template>

<script setup lang="ts">
import type { PaymentStats } from '@/api/mold/payment';

interface Props {
  stats?: PaymentStats;
}

const props = withDefaults(defineProps<Props>(), {
  stats: undefined,
});

// 格式化金额
function formatAmount(amount: number) {
  return amount.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

function calculateProgress(): number {
  if (!props.stats || !props.stats.contract_total) {
    return 0;
  }
  const progress = (props.stats.actual_total / props.stats.contract_total) * 100;
  return Math.min(progress, 100);
}

function getProgressColor() {
  const progress = calculateProgress();
  if (progress < 50) {
    return '#e34d59';
  } else if (progress < 80) {
    return '#ed7b2f';
  } else {
    return '#00a870';
  }
}
</script>

<style scoped>
.payment-stats {
  margin-bottom: 16px;
}

.stat-card {
  border-radius: 4px;
  transition: all 0.3s;
}

.stat-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stat-content {
  text-align: center;
}

.stat-label {
  font-size: 14px;
  color: var(--td-text-color-secondary);
  margin-bottom: 8px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.stat-value.success {
  color: #00a870;
}

.stat-value.warning {
  color: #ed7b2f;
}

.progress-bar {
  margin-top: 12px;
}
</style>
