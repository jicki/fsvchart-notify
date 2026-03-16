<template>
  <div>
    <div class="page-header">
      <div>
        <h3>用户管理</h3>
        <p>查看用户列表、修改角色与重置密码</p>
      </div>
    </div>

    <div class="card">
      <table class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>显示名</th>
            <th>邮箱</th>
            <th>角色</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td>{{ u.id }}</td>
            <td>{{ u.username }}</td>
            <td>{{ u.display_name || '-' }}</td>
            <td>{{ u.email || '-' }}</td>
            <td>
              <select
                class="form-input role-select"
                :value="u.role"
                :disabled="isSelf(u.username)"
                @change="handleRoleChange(u, ($event.target as HTMLSelectElement).value)"
              >
                <option value="admin">admin</option>
                <option value="user">user</option>
              </select>
            </td>
            <td>{{ formatTime(u.created_at) }}</td>
            <td>
              <button class="btn btn-sm btn-secondary" @click="openResetModal(u)">
                重置密码
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="users.length === 0" class="empty">暂无用户</div>
    </div>

    <!-- 重置密码弹窗 -->
    <ModalDialog
      :visible="showResetModal"
      :title="`重置密码 - ${resetUser?.username}`"
      max-width="400px"
      @close="closeResetModal"
    >
      <div class="form-group">
        <label>新密码</label>
        <input
          class="form-input"
          type="password"
          v-model="newPassword"
          placeholder="至少 6 位"
          minlength="6"
        />
      </div>
      <div class="modal-actions">
        <button class="btn btn-primary" @click="handleResetPassword" :disabled="!newPassword || newPassword.length < 6">
          确认重置
        </button>
        <button class="btn btn-secondary" @click="closeResetModal">取消</button>
      </div>
    </ModalDialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { get, put } from '../utils/api'
import { useNotification } from '../composables/useNotification'
import { usePolling } from '../composables/usePolling'
import { useAuthStore } from '../stores/auth'
import ModalDialog from '../components/ModalDialog.vue'

interface UserItem {
  id: number
  username: string
  display_name: string
  email: string
  role: string
  created_at: string
  updated_at: string
}

const authStore = useAuthStore()
const { showSuccess, showError } = useNotification()

const users = ref<UserItem[]>([])

async function fetchUsers() {
  try {
    const data = await get<UserItem[]>('/api/users')
    if (Array.isArray(data)) {
      users.value = data
    }
  } catch (err) {
    console.error('获取用户列表失败:', err)
  }
}

function isSelf(username: string): boolean {
  return authStore.user?.username === username
}

function formatTime(t: string): string {
  if (!t) return '-'
  const d = new Date(t)
  if (isNaN(d.getTime())) return t
  return d.toLocaleString('zh-CN', { hour12: false })
}

// 角色修改
async function handleRoleChange(user: UserItem, newRole: string) {
  if (newRole === user.role) return
  try {
    await put(`/api/users/${user.id}/role`, { role: newRole })
    showSuccess(`已将 ${user.username} 角色修改为 ${newRole}`)
    await fetchUsers()
  } catch (err) {
    showError('修改角色失败')
    console.error(err)
  }
}

// 重置密码
const showResetModal = ref(false)
const resetUser = ref<UserItem | null>(null)
const newPassword = ref('')

function openResetModal(user: UserItem) {
  resetUser.value = user
  newPassword.value = ''
  showResetModal.value = true
}

function closeResetModal() {
  showResetModal.value = false
  resetUser.value = null
  newPassword.value = ''
}

async function handleResetPassword() {
  if (!resetUser.value || newPassword.value.length < 6) return
  try {
    await put(`/api/users/${resetUser.value.id}/password`, { new_password: newPassword.value })
    showSuccess(`已重置 ${resetUser.value.username} 的密码`)
    closeResetModal()
  } catch (err) {
    showError('重置密码失败')
    console.error(err)
  }
}

usePolling(fetchUsers, 30000)
</script>

<style scoped>
.role-select {
  width: auto;
  padding: 4px 8px;
  font-size: 13px;
}
</style>
