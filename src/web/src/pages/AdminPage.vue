<template>
  <div class="container">
    <div class="page-header">
      <h1>Admin</h1>
    </div>

    <div v-if="!auth.isAdmin" class="empty-state">
      <h3>Access Denied</h3>
      <p>Admin privileges required</p>
    </div>

    <div v-else class="admin-layout">
      <!-- Tab Nav -->
      <div class="tab-nav">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          class="tab-btn"
          :class="{ active: activeTab === tab.id }"
          @click="activeTab = tab.id"
        ><component :is="tabIcons[tab.id]" :size="16" /> {{ tab.label }}</button>
      </div>

      <!-- Users Tab -->
      <section v-if="activeTab === 'users'" class="admin-section card">
        <h2>User Management</h2>
        <div v-if="usersLoading" class="loading-overlay"><div class="spinner"></div></div>
        <table v-else class="users-table">
          <thead>
            <tr>
              <th>Username</th>
              <th>Role</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.id">
              <td>
                <span class="username">{{ user.username }}</span>
                <span v-if="user.id === auth.user?.id" class="you-badge">(you)</span>
              </td>
              <td>
                <span class="badge" :class="`badge-${user.role === 'admin' ? 'roman' : 'modern'}`">
                  {{ user.role }}
                </span>
              </td>
              <td class="date-cell">{{ formatDate(user.createdAt) }}</td>
              <td>
                <div v-if="user.id !== auth.user?.id" class="action-btns">
                  <button class="btn btn-secondary btn-sm" @click="openResetModal(user)">
                    Reset
                  </button>
                  <button class="btn btn-danger btn-sm" @click="handleDeleteUser(user)">
                    Delete
                  </button>
                </div>
                <span v-else class="text-muted">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </section>

      <!-- Ollama / AI Tab -->
      <section v-if="activeTab === 'ai'" class="admin-section card">
        <h2>AI Configuration</h2>
        <form @submit.prevent="saveSettings">
          <div class="form-group">
            <label class="form-label">Ollama URL</label>
            <input v-model="settings.OllamaURL" class="form-input" placeholder="http://localhost:11434" />
          </div>
          <div class="form-group">
            <label class="form-label">Vision Model</label>
            <input v-model="settings.OllamaModel" class="form-input" placeholder="llava" />
            <span class="form-hint">e.g. llava, llama3.2-vision, bakllava</span>
          </div>
          <div class="form-group">
            <label class="form-label">Custom Analysis Prompt</label>
            <textarea
              v-model="settings.AiAnalysisPrompt"
              class="form-textarea"
              rows="6"
              placeholder="Leave empty for default numismatic analysis prompt..."
            />
            <span class="form-hint">Override the default coin analysis prompt. Coin context is appended automatically.</span>
          </div>
          <p v-if="settingsMsg" class="msg" :class="{ error: settingsError }">{{ settingsMsg }}</p>
          <button type="submit" class="btn btn-primary btn-sm" :disabled="settingsSaving">
            {{ settingsSaving ? 'Saving...' : 'Save AI Settings' }}
          </button>
        </form>
      </section>

      <!-- System Tab -->
      <section v-if="activeTab === 'system'" class="admin-section card">
        <h2>System Settings</h2>
        <form @submit.prevent="saveSettings">
          <div class="form-group">
            <label class="form-label">Log Level</label>
            <select v-model="settings.LogLevel" class="form-select">
              <option v-for="level in LOG_LEVELS" :key="level" :value="level">{{ level }}</option>
            </select>
          </div>
          <p v-if="settingsMsg" class="msg" :class="{ error: settingsError }">{{ settingsMsg }}</p>
          <button type="submit" class="btn btn-primary btn-sm" :disabled="settingsSaving">
            {{ settingsSaving ? 'Saving...' : 'Save System Settings' }}
          </button>
        </form>
      </section>

      <!-- Logs Tab -->
      <section v-if="activeTab === 'logs'" class="admin-section card">
        <h2>Application Logs</h2>
        <div class="logs-toolbar">
          <select v-model="logsFilter" class="form-select logs-filter" @change="loadLogs">
            <option value="">All Levels</option>
            <option v-for="level in ['TRACE','DEBUG','INFO','WARN','ERROR']" :key="level" :value="level">{{ level }}</option>
          </select>
          <button class="btn btn-secondary btn-sm" @click="loadLogs" :disabled="logsLoading">
            {{ logsLoading ? 'Loading...' : 'Refresh' }}
          </button>
          <button
            class="btn btn-sm"
            :class="logsAutoRefresh ? 'btn-primary' : 'btn-secondary'"
            @click="toggleAutoRefresh"
          >
            {{ logsAutoRefresh ? 'Auto ●' : 'Auto ○' }}
          </button>
        </div>
        <div class="logs-container">
          <div v-if="logs.length === 0 && !logsLoading" class="logs-empty">
            No log entries. Click Refresh to load.
          </div>
          <div
            v-for="(entry, i) in logs"
            :key="i"
            class="log-entry"
            :class="logLevelClass(entry.level)"
          >
            <span class="log-time">{{ entry.timestamp.substring(11, 19) }}</span>
            <span class="log-level-badge">{{ entry.level }}</span>
            <span class="log-msg">{{ entry.message }}</span>
          </div>
        </div>
      </section>

      <!-- Reset Password Modal -->
      <div v-if="resetTarget" class="modal-overlay" @click.self="resetTarget = null">
        <div class="modal card">
          <h3>Reset Password for {{ resetTarget.username }}</h3>
          <form @submit.prevent="handleResetPassword">
            <div class="form-group">
              <label class="form-label">New Password</label>
              <input v-model="resetNewPassword" type="password" class="form-input" required minlength="6" />
            </div>
            <p v-if="resetMsg" class="msg" :class="{ error: resetError }">{{ resetMsg }}</p>
            <div class="modal-actions">
              <button type="button" class="btn btn-secondary btn-sm" @click="resetTarget = null">Cancel</button>
              <button type="submit" class="btn btn-primary btn-sm" :disabled="resetLoading">
                {{ resetLoading ? 'Resetting...' : 'Reset Password' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, type Component } from 'vue'
import { useAuthStore } from '@/stores/auth'
import {
  getUsers, deleteUser, resetUserPassword,
  getAppSettings, updateAppSettings, getAdminLogs,
} from '@/api/client'
import { LOG_LEVELS } from '@/types'
import type { UserInfo, AppSettings, LogEntry } from '@/types'
import { Users, Cpu, Wrench, ScrollText } from 'lucide-vue-next'

const tabIcons: Record<string, Component> = { users: Users, ai: Cpu, system: Wrench, logs: ScrollText }

const auth = useAuthStore()

const tabs = [
  { id: 'users', icon: 'users', label: 'Users' },
  { id: 'ai', icon: 'cpu', label: 'AI Config' },
  { id: 'system', icon: 'wrench', label: 'System' },
  { id: 'logs', icon: 'scroll-text', label: 'Logs' },
]
const activeTab = ref('users')

// Users
const users = ref<UserInfo[]>([])
const usersLoading = ref(true)

async function loadUsers() {
  usersLoading.value = true
  try {
    const res = await getUsers()
    users.value = res.data
  } finally {
    usersLoading.value = false
  }
}

async function handleDeleteUser(user: UserInfo) {
  if (!confirm(`Delete user "${user.username}" and all their data? This cannot be undone.`)) return
  try {
    await deleteUser(user.id)
    users.value = users.value.filter((u) => u.id !== user.id)
  } catch {
    alert('Failed to delete user')
  }
}

// Reset password modal
const resetTarget = ref<UserInfo | null>(null)
const resetNewPassword = ref('')
const resetMsg = ref('')
const resetError = ref(false)
const resetLoading = ref(false)

function openResetModal(user: UserInfo) {
  resetTarget.value = user
  resetNewPassword.value = ''
  resetMsg.value = ''
  resetError.value = false
}

async function handleResetPassword() {
  if (!resetTarget.value) return
  resetLoading.value = true
  resetMsg.value = ''
  try {
    await resetUserPassword(resetTarget.value.id, resetNewPassword.value)
    resetMsg.value = 'Password reset successfully'
    setTimeout(() => { resetTarget.value = null }, 1200)
  } catch {
    resetMsg.value = 'Failed to reset password'
    resetError.value = true
  } finally {
    resetLoading.value = false
  }
}

// Settings
const settings = ref<AppSettings>({
  OllamaURL: 'http://localhost:11434',
  OllamaModel: 'llava',
  AiAnalysisPrompt: '',
  LogLevel: 'info',
})
const settingsMsg = ref('')
const settingsError = ref(false)
const settingsSaving = ref(false)

async function loadSettings() {
  try {
    const res = await getAppSettings()
    settings.value = { ...settings.value, ...res.data }
  } catch { /* use defaults */ }
}

async function saveSettings() {
  settingsSaving.value = true
  settingsMsg.value = ''
  settingsError.value = false
  try {
    const entries = Object.entries(settings.value).map(([key, value]) => ({ key, value }))
    await updateAppSettings(entries)
    settingsMsg.value = 'Settings saved'
  } catch {
    settingsMsg.value = 'Failed to save settings'
    settingsError.value = true
  } finally {
    settingsSaving.value = false
  }
}

// Logs
const logs = ref<LogEntry[]>([])
const logsLoading = ref(false)
const logsFilter = ref('')
const logsAutoRefresh = ref(false)
let logsInterval: ReturnType<typeof setInterval> | null = null

async function loadLogs() {
  logsLoading.value = true
  try {
    const res = await getAdminLogs(500, logsFilter.value || undefined)
    logs.value = res.data.logs || []
  } catch { /* ignore */ } finally {
    logsLoading.value = false
  }
}

function toggleAutoRefresh() {
  logsAutoRefresh.value = !logsAutoRefresh.value
  if (logsAutoRefresh.value) {
    logsInterval = setInterval(loadLogs, 3000)
  } else if (logsInterval) {
    clearInterval(logsInterval)
    logsInterval = null
  }
}

function logLevelClass(level: string) {
  switch (level) {
    case 'ERROR': return 'log-error'
    case 'WARN': return 'log-warn'
    case 'DEBUG': return 'log-debug'
    case 'TRACE': return 'log-trace'
    default: return 'log-info'
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString()
}

onMounted(() => {
  loadUsers()
  loadSettings()
})

onUnmounted(() => {
  if (logsInterval) clearInterval(logsInterval)
})
</script>

<style scoped>
.admin-layout {
  max-width: 800px;
  margin-left: auto;
  margin-right: auto;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.tab-nav {
  display: flex;
  gap: 0.25rem;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 0.3rem;
}

.tab-btn {
  flex: 1;
  padding: 0.6rem 1rem;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.tab-btn.active {
  background: var(--accent-gold-dim);
  color: var(--accent-gold);
}

.tab-btn:hover:not(.active) {
  color: var(--text-primary);
}

.admin-section h2 {
  font-size: 1.1rem;
  margin-bottom: 1.25rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border-subtle);
}

.users-table {
  width: 100%;
  border-collapse: collapse;
}

.users-table th,
.users-table td {
  text-align: left;
  padding: 0.75rem 0.5rem;
  border-bottom: 1px solid var(--border-subtle);
}

.users-table th {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  font-weight: 600;
}

.username {
  font-weight: 500;
}

.you-badge {
  font-size: 0.7rem;
  color: var(--text-muted);
  margin-left: 0.3rem;
}

.date-cell {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.action-btns {
  display: flex;
  gap: 0.4rem;
}

.text-muted {
  color: var(--text-muted);
}

.form-hint {
  display: block;
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: 0.25rem;
}

.msg {
  font-size: 0.85rem;
  color: var(--accent-gold);
  margin: 0.5rem 0;
}

.msg.error {
  color: #e74c3c;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  padding: 1rem;
}

.modal {
  width: 100%;
  max-width: 400px;
}

.modal h3 {
  margin-bottom: 1rem;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 1rem;
}

@media (max-width: 640px) {
  .tab-nav {
    flex-wrap: wrap;
  }
  .users-table {
    font-size: 0.85rem;
  }
  .action-btns {
    flex-direction: column;
  }
}

/* Logs */
.logs-toolbar {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 1rem;
}

.logs-filter {
  width: auto;
  min-width: 120px;
}

.logs-container {
  max-height: 500px;
  overflow-y: auto;
  background: var(--bg-body);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 0.5rem;
  font-family: 'Courier New', Courier, monospace;
  font-size: 0.78rem;
  line-height: 1.5;
}

.logs-empty {
  text-align: center;
  padding: 2rem;
  color: var(--text-muted);
  font-family: 'Inter', sans-serif;
}

.log-entry {
  display: flex;
  gap: 0.5rem;
  padding: 0.15rem 0.25rem;
  border-radius: 2px;
}

.log-entry:hover {
  background: var(--bg-card);
}

.log-time {
  color: var(--text-muted);
  flex-shrink: 0;
}

.log-level-badge {
  flex-shrink: 0;
  min-width: 48px;
  text-align: center;
  font-weight: 600;
  border-radius: 2px;
  padding: 0 4px;
}

.log-msg {
  word-break: break-word;
}

.log-error .log-level-badge { color: #e74c3c; }
.log-error .log-msg { color: #e74c3c; }
.log-warn .log-level-badge { color: #f39c12; }
.log-debug .log-level-badge { color: #3498db; }
.log-trace .log-level-badge { color: #7f8c8d; }
.log-info .log-level-badge { color: #2ecc71; }
</style>
