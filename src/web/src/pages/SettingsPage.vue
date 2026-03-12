<template>
  <div class="container">
    <div class="page-header">
      <h1>Settings</h1>
    </div>

    <div class="settings-layout">
      <!-- Account -->
      <section class="settings-section card">
        <h2>Account</h2>
        <div class="setting-item">
          <div class="setting-info">
            <span class="setting-label">Username</span>
            <span class="setting-value">{{ auth.user?.username }}</span>
          </div>
        </div>
        <div class="setting-item">
          <div class="setting-info">
            <span class="setting-label">Role</span>
            <span class="setting-value badge" :class="`badge-${auth.user?.role === 'admin' ? 'roman' : 'modern'}`">
              {{ auth.user?.role }}
            </span>
          </div>
        </div>

        <h3>Change Password</h3>
        <form class="password-form" @submit.prevent="handleChangePassword">
          <div class="form-group">
            <label class="form-label">Current Password</label>
            <input v-model="currentPassword" type="password" class="form-input" required />
          </div>
          <div class="form-group">
            <label class="form-label">New Password</label>
            <input v-model="newPassword" type="password" class="form-input" required minlength="6" />
          </div>
          <div class="form-group">
            <label class="form-label">Confirm New Password</label>
            <input v-model="confirmPassword" type="password" class="form-input" required />
          </div>
          <p v-if="passwordMsg" class="msg" :class="{ error: passwordError }">{{ passwordMsg }}</p>
          <button type="submit" class="btn btn-primary btn-sm" :disabled="passwordLoading">
            {{ passwordLoading ? 'Changing...' : 'Change Password' }}
          </button>
        </form>
      </section>

      <!-- Appearance -->
      <section class="settings-section card">
        <h2>Appearance</h2>
        <div class="setting-item">
          <div class="setting-info">
            <span class="setting-label">Theme</span>
            <span class="setting-desc">Choose your preferred color scheme</span>
          </div>
          <div class="theme-toggle">
            <button
              class="theme-btn"
              :class="{ active: theme === 'dark' }"
              @click="setTheme('dark')"
            >Dark</button>
            <button
              class="theme-btn"
              :class="{ active: theme === 'light' }"
              @click="setTheme('light')"
            >Light</button>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <span class="setting-label">Timezone</span>
            <span class="setting-desc">Used for date display</span>
          </div>
          <select v-model="timezone" class="form-select tz-select" @change="saveTimezone">
            <option v-for="tz in timezones" :key="tz" :value="tz">{{ tz }}</option>
          </select>
        </div>
      </section>

      <!-- Data Management -->
      <section class="settings-section card">
        <h2>Data Management</h2>
        <div class="setting-item">
          <div class="setting-info">
            <span class="setting-label">Export Collection</span>
            <span class="setting-desc">Download your collection data and photos as a zip archive</span>
          </div>
          <button class="btn btn-secondary btn-sm" :disabled="exporting" @click="handleExport">
            {{ exporting ? 'Exporting...' : '📥 Export' }}
          </button>
        </div>
        <div class="setting-item">
          <div class="setting-info">
            <span class="setting-label">Import Collection</span>
            <span class="setting-desc">Import coins from a JSON file</span>
          </div>
          <label class="btn btn-secondary btn-sm import-btn">
            📤 Import
            <input type="file" accept=".json" hidden @change="handleImport" />
          </label>
        </div>
        <p v-if="dataMsg" class="msg" :class="{ error: dataError }">{{ dataMsg }}</p>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { changePassword, exportCollection, importCollection } from '@/api/client'
import type { Coin, Theme } from '@/types'


const auth = useAuthStore()

// Password
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordMsg = ref('')
const passwordError = ref(false)
const passwordLoading = ref(false)

async function handleChangePassword() {
  passwordMsg.value = ''
  passwordError.value = false

  if (newPassword.value !== confirmPassword.value) {
    passwordMsg.value = 'New passwords do not match'
    passwordError.value = true
    return
  }

  passwordLoading.value = true
  try {
    await changePassword(currentPassword.value, newPassword.value)
    passwordMsg.value = 'Password changed successfully'
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch {
    passwordMsg.value = 'Failed — check your current password'
    passwordError.value = true
  } finally {
    passwordLoading.value = false
  }
}

// Theme
const theme = ref<Theme>((localStorage.getItem('theme') as Theme) || 'dark')

function setTheme(t: Theme) {
  theme.value = t
  localStorage.setItem('theme', t)
  document.documentElement.setAttribute('data-theme', t)
}

// Timezone
const timezones = (Intl as any).supportedValuesOf('timeZone') as string[]
const timezone = ref(localStorage.getItem('timezone') || Intl.DateTimeFormat().resolvedOptions().timeZone)

function saveTimezone() {
  localStorage.setItem('timezone', timezone.value)
}

// Data
const exporting = ref(false)
const dataMsg = ref('')
const dataError = ref(false)

async function handleExport() {
  exporting.value = true
  dataMsg.value = ''
  try {
    const res = await exportCollection()
    const blob = new Blob([res.data], { type: 'application/zip' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `ancient-coins-export-${new Date().toISOString().slice(0, 10)}.zip`
    a.click()
    URL.revokeObjectURL(url)
    dataMsg.value = 'Export downloaded'
  } catch {
    dataMsg.value = 'Export failed'
    dataError.value = true
  } finally {
    exporting.value = false
  }
}

async function handleImport(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return

  dataMsg.value = ''
  dataError.value = false

  try {
    const text = await file.text()
    const coins: Coin[] = JSON.parse(text)
    const res = await importCollection(coins)
    dataMsg.value = `Imported ${res.data.imported} coins`
  } catch {
    dataMsg.value = 'Import failed — ensure valid JSON format'
    dataError.value = true
  }
}
</script>

<style scoped>
.settings-layout {
  max-width: 700px;
  margin-left: auto;
  margin-right: auto;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.settings-section h2 {
  font-size: 1.1rem;
  margin-bottom: 1.25rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border-subtle);
}

.settings-section h3 {
  font-size: 0.95rem;
  margin-top: 1.25rem;
  margin-bottom: 0.75rem;
  color: var(--text-secondary);
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--border-subtle);
  gap: 1rem;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.setting-label {
  font-size: 0.9rem;
  font-weight: 500;
}

.setting-desc {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.setting-value {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.password-form {
  max-width: 350px;
}

.theme-toggle {
  display: flex;
  gap: 0.25rem;
  background: var(--bg-primary);
  border-radius: var(--radius-full);
  padding: 0.2rem;
}

.theme-btn {
  padding: 0.35rem 0.75rem;
  border: none;
  border-radius: var(--radius-full);
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.theme-btn.active {
  background: var(--accent-gold-dim);
  color: var(--accent-gold);
}

.tz-select {
  max-width: 250px;
}

.import-btn {
  cursor: pointer;
}

.msg {
  font-size: 0.85rem;
  color: var(--accent-gold);
  margin: 0.5rem 0;
}

.msg.error {
  color: #e74c3c;
}

@media (max-width: 640px) {
  .setting-item {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
