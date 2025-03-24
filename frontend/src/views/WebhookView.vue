<template>
  <div class="tab-content">
    <h3>飞书 WebHooks</h3>
    <div>
      <label>名称: <input v-model="newWebhookName"/></label>
      <label>URL: <input v-model="newWebhookURL"/></label>
      <button @click="addFeishuWebhook">添加 WebHook</button>
    </div>

    <table>
      <thead>
        <tr>
          <th>ID</th><th>名称</th><th>URL</th><th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="wb in feishuWebhooks" :key="wb.id">
          <td>{{ wb.id }}</td>
          <td v-if="editWebhookID===wb.id">
            <input v-model="editWebhookName"/>
          </td>
          <td v-else>
            {{ wb.name }}
          </td>

          <td v-if="editWebhookID===wb.id">
            <input v-model="editWebhookURL"/>
          </td>
          <td v-else>
            {{ wb.url }}
          </td>

          <td>
            <div v-if="editWebhookID===wb.id">
              <button @click="saveEditWebhook(wb.id)">保存</button>
              <button @click="cancelEditWebhook">取消</button>
            </div>
            <div v-else>
              <button @click="startEditWebhook(wb)">编辑</button>
              <button @click="deleteFeishuWebhook(wb.id)">删除</button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { get, post, put, del } from '../utils/api'

const emit = defineEmits(['webhook-updated'])

const feishuWebhooks = ref<any[]>([])
const newWebhookName = ref('')
const newWebhookURL = ref('')
const editWebhookID = ref<number|null>(null)
const editWebhookName = ref('')
const editWebhookURL = ref('')

async function fetchFeishuWebhooks() {
  try {
    const data = await get('/api/feishu_webhook')
    if (Array.isArray(data)) {
      feishuWebhooks.value = data
      console.log('成功获取webhook列表:', data.length)
    } else {
      console.error('获取webhook返回格式错误:', data)
    }
  } catch (err) {
    console.error('获取webhook失败:', err)
    feishuWebhooks.value = [] // 确保失败时设置为空数组
  }
}

async function addFeishuWebhook() {
  if (!newWebhookName.value || !newWebhookURL.value) {
    alert('名称或URL不能为空')
    return
  }
  
  try {
    const body = { name: newWebhookName.value, url: newWebhookURL.value }
    const result = await post('/api/feishu_webhook', body)
    console.log('添加webhook成功:', result)
    
    // 重置表单
    newWebhookName.value = ''
    newWebhookURL.value = ''
    
    // 刷新列表并通知更新
    await fetchFeishuWebhooks()
    emit('webhook-updated')
  } catch (err) {
    console.error('添加webhook失败:', err)
    alert('添加webhook失败，请重试')
  }
}

function startEditWebhook(wb: any) {
  editWebhookID.value = wb.id
  editWebhookName.value = wb.name
  editWebhookURL.value = wb.url
}

function cancelEditWebhook() {
  editWebhookID.value = null
  editWebhookName.value = ''
  editWebhookURL.value = ''
}

async function saveEditWebhook(id: number) {
  if (!editWebhookName.value || !editWebhookURL.value) {
    alert('名称或URL不能为空')
    return
  }

  try {
    const body = { name: editWebhookName.value, url: editWebhookURL.value }
    const result = await put(`/api/feishu_webhook/${id}`, body)
    console.log('更新webhook成功:', result)
    
    // 重置编辑状态
    editWebhookID.value = null
    editWebhookName.value = ''
    editWebhookURL.value = ''
    
    // 刷新列表并通知更新
    await fetchFeishuWebhooks()
    emit('webhook-updated')
  } catch (err) {
    console.error('更新webhook失败:', err)
    alert('更新webhook失败，请重试')
  }
}

async function deleteFeishuWebhook(id: number) {
  if (!confirm(`确认删除WebHook ID=${id}?`)) return
  
  try {
    await del(`/api/feishu_webhook/${id}`)
    console.log('删除webhook成功:', id)
    
    // 刷新列表并通知更新
    await fetchFeishuWebhooks()
    emit('webhook-updated')
  } catch (err) {
    console.error('删除webhook失败:', err)
    alert('删除webhook失败，请重试')
  }
}

// 定期刷新数据
let refreshInterval: number

onMounted(() => {
  // 初始加载
  fetchFeishuWebhooks()
  
  // 设置定期刷新
  refreshInterval = setInterval(() => {
    fetchFeishuWebhooks()
  }, 30000) // 每30秒刷新一次
})

onUnmounted(() => {
  // 清理定时器
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<style scoped>
</style>
