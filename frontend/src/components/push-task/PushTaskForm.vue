<template>
  <div class="form-container" :class="{ 'edit-form': isEditing }">
    <h4>{{ isEditing ? '编辑任务' : '创建新的任务' }}</h4>

    <div class="form-group">
      <label>任务名称:</label>
      <input type="text" v-model="form.name.value" placeholder="任务名称" />
    </div>

    <div class="form-group">
      <label>数据源:</label>
      <select v-model="form.sourceId.value">
        <option value="">请选择数据源</option>
        <option v-for="source in sources" :key="source.id" :value="source.id">
          {{ source.name }} ({{ source.url }})
        </option>
      </select>
    </div>

    <PromqlSelector
      :promqls="promqls"
      :id-prefix="isEditing ? 'edit' : 'new'"
      v-model:selected-ids="form.selectedPromQLs.value"
      v-model:configs="form.promqlConfigs.value"
    />

    <div class="form-group">
      <label>时间范围:</label>
      <select v-model="form.timeRange.value" class="time-range-select">
        <option value="1d">1 天</option>
        <option value="5d">5 天</option>
        <option value="7d">7 天</option>
        <option value="15d">15 天</option>
        <option value="30d">30 天</option>
      </select>
      <div class="form-hint">数据点已针对每个时间范围优化，确保图表清晰</div>
    </div>

    <div class="form-group">
      <label>选择图表模板:
        <select v-model="form.chartTemplateId.value">
          <option value="">-- 请选择图表模板 --</option>
          <option v-for="tmpl in chartTemplates" :key="tmpl.id" :value="tmpl.id">
            {{ tmpl.name }} ({{ tmpl.chart_type }})
          </option>
        </select>
      </label>
    </div>

    <div class="form-group">
      <label>消息标题: <input v-model="form.cardTitle.value" /></label>
    </div>

    <div class="form-group">
      <label>卡片模板:
        <select v-model="form.cardTemplate.value">
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
      </label>
    </div>

    <div class="form-group">
      <label>按钮文本:
        <input v-model="form.buttonText.value" placeholder="例如: 节点池资源总览" />
      </label>
      <small>自定义卡片底部按钮的文本，留空则使用默认值</small>
    </div>

    <div class="form-group">
      <label>按钮链接:
        <input v-model="form.buttonURL.value" placeholder="例如: https://grafana.example.com/d/xxx" />
      </label>
      <small>自定义卡片底部按钮的链接URL，留空则使用默认值</small>
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

    <div class="form-group">
      <label>
        <input type="checkbox" v-model="form.showDataLabel.value" />
        显示曲线数值
      </label>
      <small>在图表中显示数据点的具体数值</small>
    </div>

    <div v-if="isEditing" class="edit-actions">
      <button @click.prevent="handleSubmit" class="update-btn">保存修改</button>
      <button @click.prevent="$emit('cancel')" class="cancel-btn">取消</button>
    </div>
    <button v-else @click.prevent="handleSubmit" class="create-btn">创建并发送</button>
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
}>()

const emit = defineEmits<{
  submit: [payload: Record<string, unknown>]
  cancel: []
}>()

function handleSubmit() {
  if (!props.form.validate()) return
  const payload = props.form.buildPayload(props.promqls)
  emit('submit', payload)
}
</script>

<style scoped>
.form-container { padding: 15px; border: 1px solid var(--color-border, #ddd); border-radius: 4px; background-color: var(--color-bg-light, #f8f9fa); margin-bottom: 20px; }
.edit-form { margin-top: 20px; border: 1px solid #dee2e6; }
.edit-form h4 { margin-top: 0; color: var(--color-success, #28a745); border-bottom: 2px solid var(--color-success, #28a745); padding-bottom: 10px; margin-bottom: 20px; }
.form-group { margin-bottom: 15px; }
.form-hint { font-size: 0.85em; color: var(--color-text-secondary, #666); margin-top: 5px; }
.time-range-select { width: 200px; padding: 8px; border: 1px solid #ccc; border-radius: 4px; font-size: 14px; background-color: white; cursor: pointer; }
.create-btn { margin-top: 15px; padding: 10px 20px; background-color: var(--color-success, #4CAF50); color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 16px; font-weight: bold; }
.create-btn:hover { background-color: var(--color-success-hover, #45a049); }
.edit-actions { display: flex; justify-content: space-between; margin-top: 15px; }
.update-btn, .cancel-btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
.update-btn { background-color: var(--color-success, #28a745); color: white; }
.update-btn:hover { background-color: var(--color-success-hover, #218838); }
.cancel-btn { background-color: var(--color-danger, #dc3545); color: white; }
.cancel-btn:hover { background-color: var(--color-danger-hover, #c82333); }
</style>
