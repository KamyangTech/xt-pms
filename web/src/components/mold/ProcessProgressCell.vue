<template>
  <div class="process-summary" :class="{ clickable: clickable }" @click="handleClick">
    <div class="process-row-capsules">
      <template v-for="(group, groupIndex) in groupedProcesses" :key="`group-${group.groupId}`">
        <div class="group-processes-list">
          <template v-for="template in group.templates" :key="template.id">
            <span
              class="process-name-item"
              :class="getProcessNameClass(template.id)"
              :title="template.process_name"
            >
              {{ getProcessShortName(template.process_name) }}
            </span>
          </template>
        </div>
        <span v-if="groupIndex < groupedProcesses.length - 1" class="group-arrow">→</span>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = withDefaults(defineProps<{
  processes?: any[]; // work_order_processes 列表，含 process_template_id + is_completed
  templates?: any[]; // 工序模板列表（含 process_group、group_name 等）
  clickable?: boolean;
}>(), {
  processes: () => [],
  templates: () => [],
  clickable: false,
});

const emit = defineEmits<{
  (e: 'click'): void;
}>();

const handleClick = () => {
  if (props.clickable) {
    emit('click');
  }
};

// 工序简称映射
const processShortNameMap: Record<string, string> = {
  'CNC加工': 'CNC',
  '磨床加工': '磨床',
  '放电加工': '放电',
  '线切割': '线切',
  '钳工加工': '钳工',
  '抛光': '抛光',
  '装配': '装配',
  '合模': '合模',
  '咬花': '咬花',
  '省模': '省模',
  '试模': '试模',
  '热处理': '热处',
  '车床加工': '车床',
  '铣床加工': '铣床',
  '电火花': '电火花',
  '深孔钻': '深孔',
  '磨床': '磨床',
  '车床': '车床',
  '铣床': '铣床',
  '冲压': '冲压',
  '喷涂': '喷涂',
  '电镀': '电镀',
  '阳极氧化': '阳极',
  '丝印': '丝印',
  '镭雕': '镭雕',
  '热转印': '热转',
};

const getProcessShortName = (name: string): string => {
  if (!name) return '';
  if (processShortNameMap[name]) return processShortNameMap[name];
  return name.length > 2 ? name.substring(0, 2) : name;
};

const getProcessStatusValue = (templateId: number): number => {
  const p = props.processes.find((wp: any) => wp.process_template_id === templateId);
  return p ? (p.is_completed ?? 0) : 0;
};

const getProcessNameClass = (templateId: number): string => {
  const status = getProcessStatusValue(templateId);
  if (status === 1) return 'process-completed';
  if (status === 2) return 'process-skipped';
  return 'process-pending';
};

const groupedProcesses = computed(() => {
  const groups: { groupId: number; groupName: string; templates: any[] }[] = [];
  const map = new Map<number, { groupId: number; groupName: string; templates: any[] }>();

  props.templates.forEach((template) => {
    const groupKey = template.group_id ?? template.process_group ?? 0;
    const groupName = template.group_name || template.process_group_name || '工序';
    if (!map.has(groupKey)) {
      const g = { groupId: groupKey, groupName, templates: [] as any[] };
      map.set(groupKey, g);
      groups.push(g);
    }
    map.get(groupKey)!.templates.push(template);
  });

  return groups;
});
</script>

<style scoped>
.process-summary {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 0;
}

.process-summary.clickable {
  cursor: pointer;
}

.process-summary.clickable:hover .process-row-capsules {
  background: var(--td-bg-color-secondary, #f5f5f5);
}

.process-row-capsules {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 4px 8px;
  border-radius: 6px;
  white-space: nowrap;
  transition: background 0.3s ease;
}

.group-processes-list {
  display: flex;
  gap: 6px;
  flex-wrap: nowrap;
}

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

.process-completed {
  background: var(--td-success-color-1, #e3f9e9);
  color: var(--td-success-color, #00a870);
  border: 1px solid var(--td-success-color-2, #b3ebd2);
}

.process-pending {
  background: var(--td-bg-color-component, #f0f0f0);
  color: var(--td-text-color-placeholder, #999);
  border: 1px solid var(--td-border-level-2-color, #e0e0e0);
}

.process-skipped {
  background: #fafafa;
  color: #bbb;
  border: 1px dashed #ddd;
}
</style>
