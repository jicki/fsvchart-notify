<template>
  <div class="task-form" :class="{ card: !hideActions, 'edit-form': isEditing }">
    <h4 v-if="!hideActions">{{ isEditing ? '编辑任务' : '创建新的任务' }}</h4>

    <div class="form-section">
      <div class="form-group">
        <label>任务名称</label>
        <input class="form-input" type="text" v-model="form.name.value" placeholder="任务名称" />
      </div>

      <div class="form-group">
        <label>数据源</label>
        <select class="form-input" v-model="form.sourceId.value">
          <option value="">请选择数据源</option>
          <option v-for="source in sources" :key="source.id" :value="source.id">
            {{ source.name }} ({{ source.url }})
          </option>
        </select>
      </div>
    </div>

    <PromqlSelector
      :promqls="promqls"
      :id-prefix="isEditing ? 'edit' : 'new'"
      v-model:selected-ids="form.selectedPromQLs.value"
      v-model:configs="form.promqlConfigs.value"
    />

    <div class="form-section">
      <div class="form-group">
        <label>时间范围</label>
        <select class="form-input time-range-select" v-model="form.timeRange.value">
          <option value="1d">1 天</option>
          <option value="5d">5 天</option>
          <option value="7d">7 天</option>
          <option value="15d">15 天</option>
          <option value="30d">30 天</option>
        </select>
        <div class="form-hint">数据点已针对每个时间范围优化，确保图表清晰</div>
      </div>

      <div class="form-group">
        <label>图表类型</label>
        <select class="form-input" v-model="form.chartType.value">
          <option value="area">面积图</option>
          <option value="line">折线图</option>
          <option value="bar">柱状图</option>
        </select>
      </div>

      <div class="form-group">
        <label>消息标题</label>
        <input class="form-input" v-model="form.cardTitle.value" />
      </div>

      <div class="form-group">
        <label>卡片模板</label>
        <select class="form-input" v-model="form.cardTemplate.value">
          <option value="red">红色</option>
          <option value="carmine">粉色</option>
          <option value="orange">橙色</option>
          <option value="blue">蓝色</option>
          <option value="green">绿色</option>
          <option value="turquoise">青色</option>
          <option value="purple">紫色</option>
          <option value="violet">紫红</option>
          <option value="grey">灰色</option>
        </select>
      </div>

      <div class="form-group">
        <label>按钮文本</label>
        <input class="form-input" v-model="form.buttonText.value" placeholder="例如: 节点池资源总览" />
        <div class="form-hint">自定义卡片底部按钮的文本，留空则使用默认值</div>
      </div>

      <div class="form-group">
        <label>按钮链接</label>
        <input class="form-input" v-model="form.buttonURL.value" placeholder="例如: https://grafana.example.com/d/xxx" />
        <div class="form-hint">自定义卡片底部按钮的链接URL，留空则使用默认值</div>
      </div>
    </div>

    <WebhookSelector
      :webhooks="webhooks"
      :id-prefix="isEditing ? 'edit' : 'new'"
      v-model:selected-ids="form.webhookIds.value"
    />

    <SendTimeEditor
      :send-times="form.sendTimes.value"
      @add="form.addSendTime"
      @remove="form.removeSendTime"
    />

    <div class="form-group checkbox-group">
      <label class="toggle">
        <input type="checkbox" v-model="form.showDataLabel.value">
        <span class="toggle-track"></span>
        <span>显示曲线数值</span>
      </label>
      <div class="form-hint">在图表中显示数据点的具体数值</div>
    </div>

    <div v-if="!hideActions" class="form-actions">
      <button v-if="isEditing" class="btn btn-primary" @click.prevent="handleSubmit">保存修改</button>
      <button v-if="isEditing" class="btn btn-secondary" @click.prevent="$emit('cancel')">取消</button>
      <button v-if="!isEditing" class="btn btn-primary" @click.prevent="handleSubmit">创建并发送</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import PromqlSelector from './PromqlSelector.vue'
import WebhookSelector from './WebhookSelector.vue'
import SendTimeEditor from './SendTimeEditor.vue'
import type { MetricsSource, FeishuWebhook, ChartTemplate, PromQL } from '../../types'
import type { usePushTaskForm } from '../../composables/usePushTaskForm'

const props = defineProps<{
  form: ReturnType<typeof usePushTaskForm>
  sources: MetricsSource[]
  webhooks: FeishuWebhook[]
  chartTemplates: ChartTemplate[]
  promqls: PromQL[]
  isEditing: boolean
  hideActions?: boolean
}>()

const emit = defineEmits<{
  submit: [payload: Record<string, unknown>]
  cancel: []
}>()

function handleSubmit() {
  if (!props.form.validate()) return
  const payload = props.form.buildPayload(props.promqls, props.chartTemplates)
  emit('submit', payload)
}
</script>

<style scoped>
.task-form {
  margin-bottom: var(--spacing-lg);
}

.task-form h4 {
  margin: 0 0 var(--spacing-lg);
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--color-border);
  font-weight: 600;
}

.edit-form h4 {
  color: var(--color-accent);
  border-bottom-color: var(--color-accent);
}

.form-section {
  margin-bottom: var(--spacing-md);
}

.form-input {
  width: 100%;
}

.time-range-select {
  max-width: 200px;
}

.form-hint {
  font-size: 13px;
  color: var(--color-text-muted);
  margin-top: 4px;
}

.checkbox-group {
  margin-top: var(--spacing-md);
}

.form-actions {
  display: flex;
  gap: 8px;
  margin-top: var(--spacing-lg);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border);
}
</style>
