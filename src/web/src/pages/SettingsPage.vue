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

        <div class="setting-item">
          <div class="setting-info">
            <span class="setting-label">Default View</span>
            <span class="setting-desc">Preferred collection view on mobile / PWA</span>
          </div>
          <div class="theme-toggle">
            <button
              class="theme-btn"
              :class="{ active: defaultView === 'swipe' }"
              @click="setDefaultView('swipe')"
            >Swipe</button>
            <button
              class="theme-btn"
              :class="{ active: defaultView === 'grid' }"
              @click="setDefaultView('grid')"
            >Grid</button>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <span class="setting-label">Default Sort</span>
            <span class="setting-desc">How coins are sorted by default</span>
          </div>
          <select v-model="defaultSort" class="form-select sort-select" @change="saveDefaultSort">
            <option value="updated_at_desc">Last Updated</option>
            <option value="created_at_desc">Newest First</option>
            <option value="created_at_asc">Oldest First</option>
            <option value="current_value_desc">Price: High → Low</option>
            <option value="current_value_asc">Price: Low → High</option>
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

      <!-- API Keys -->
      <section class="settings-section card">
        <h2>API Keys</h2>
        <p class="setting-desc" style="margin-bottom: 1rem">
          Generate API keys to access your collection from external tools and scripts. Use the <code>X-API-Key</code> header to authenticate.
        </p>

        <!-- Generate form -->
        <div class="apikey-generate">
          <input
            v-model="apiKeyName"
            type="text"
            class="form-input"
            placeholder="Key name (e.g. My Script)"
            :disabled="generatingKey"
          />
          <button
            class="btn btn-primary btn-sm"
            :disabled="!apiKeyName.trim() || generatingKey"
            @click="handleGenerateKey"
          >
            {{ generatingKey ? 'Generating...' : '🔑 Generate Key' }}
          </button>
        </div>

        <!-- Newly generated key (shown once) -->
        <div v-if="newlyGeneratedKey" class="apikey-reveal">
          <p class="apikey-reveal-warning">
            ⚠️ Copy this key now — it will not be shown again.
          </p>
          <div class="apikey-reveal-box">
            <code class="apikey-reveal-value">{{ newlyGeneratedKey }}</code>
            <button class="btn btn-secondary btn-sm" @click="copyKey">
              {{ keyCopied ? '✓ Copied' : '📋 Copy' }}
            </button>
          </div>
        </div>

        <p v-if="apiKeyMsg" class="msg" :class="{ error: apiKeyError }">{{ apiKeyMsg }}</p>

        <!-- Existing keys list -->
        <div v-if="apiKeys.length" class="apikey-list">
          <div
            v-for="key in apiKeys"
            :key="key.id"
            class="apikey-item"
            :class="{ revoked: key.revokedAt }"
          >
            <div class="apikey-item-info">
              <span class="apikey-item-name">{{ key.name }}</span>
              <span class="apikey-item-meta">
                ...{{ key.keyPrefix }}
                · Created {{ formatDate(key.createdAt) }}
                <template v-if="key.lastUsedAt"> · Last used {{ formatDate(key.lastUsedAt) }}</template>
              </span>
            </div>
            <span v-if="key.revokedAt" class="apikey-item-badge revoked-badge">Revoked</span>
            <button
              v-else
              class="btn btn-danger btn-sm"
              @click="handleRevokeKey(key.id)"
            >
              Revoke
            </button>
          </div>
        </div>
        <p v-else-if="!generatingKey" class="setting-desc" style="margin-top: 0.5rem">No API keys yet.</p>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { changePassword, exportCollection, importCollection, generateApiKey, listApiKeys, revokeApiKey } from '@/api/client'
import type { Coin, Theme, ApiKey } from '@/types'


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

// Default view
const defaultView = ref<'swipe' | 'grid'>((localStorage.getItem('defaultView') as 'swipe' | 'grid') || 'swipe')

function setDefaultView(v: 'swipe' | 'grid') {
  defaultView.value = v
  localStorage.setItem('defaultView', v)
}

// Default sort
const defaultSort = ref(localStorage.getItem('defaultSort') || 'updated_at_desc')

function saveDefaultSort() {
  localStorage.setItem('defaultSort', defaultSort.value)
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

// API Keys
const apiKeys = ref<ApiKey[]>([])
const apiKeyName = ref('')
const newlyGeneratedKey = ref('')
const keyCopied = ref(false)
const generatingKey = ref(false)
const apiKeyMsg = ref('')
const apiKeyError = ref(false)

async function loadApiKeys() {
  try {
    const res = await listApiKeys()
    apiKeys.value = res.data
  } catch {
    // silently fail on load
  }
}

async function handleGenerateKey() {
  if (!apiKeyName.value.trim()) return

  generatingKey.value = true
  apiKeyMsg.value = ''
  apiKeyError.value = false
  newlyGeneratedKey.value = ''
  keyCopied.value = false

  try {
    const res = await generateApiKey(apiKeyName.value.trim())
    newlyGeneratedKey.value = res.data.key
    apiKeyName.value = ''
    await loadApiKeys()
  } catch {
    apiKeyMsg.value = 'Failed to generate API key'
    apiKeyError.value = true
  } finally {
    generatingKey.value = false
  }
}

async function copyKey() {
  try {
    await navigator.clipboard.writeText(newlyGeneratedKey.value)
    keyCopied.value = true
    setTimeout(() => { keyCopied.value = false }, 3000)
  } catch {
    // Fallback for non-HTTPS contexts
    const textarea = document.createElement('textarea')
    textarea.value = newlyGeneratedKey.value
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    keyCopied.value = true
    setTimeout(() => { keyCopied.value = false }, 3000)
  }
}

async function handleRevokeKey(id: number) {
  apiKeyMsg.value = ''
  apiKeyError.value = false
  try {
    await revokeApiKey(id)
    await loadApiKeys()
    newlyGeneratedKey.value = ''
  } catch {
    apiKeyMsg.value = 'Failed to revoke key'
    apiKeyError.value = true
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString(undefined, {
    year: 'numeric', month: 'short', day: 'numeric',
  })
}

onMounted(loadApiKeys)
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

.sort-select {
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

.apikey-generate {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  margin-bottom: 0.75rem;
}

.apikey-generate .form-input {
  flex: 1;
  max-width: 280px;
}

.apikey-reveal {
  background: var(--bg-primary);
  border: 1px solid var(--accent-gold-dim);
  border-radius: var(--radius-sm);
  padding: 0.75rem 1rem;
  margin-bottom: 0.75rem;
}

.apikey-reveal-warning {
  font-size: 0.8rem;
  color: var(--accent-gold);
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.apikey-reveal-box {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.apikey-reveal-value {
  flex: 1;
  font-size: 0.78rem;
  background: var(--bg-card);
  padding: 0.4rem 0.6rem;
  border-radius: var(--radius-sm);
  word-break: break-all;
  user-select: all;
}

.apikey-list {
  margin-top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.apikey-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.6rem 0;
  border-bottom: 1px solid var(--border-subtle);
  gap: 0.75rem;
}

.apikey-item:last-child {
  border-bottom: none;
}

.apikey-item.revoked {
  opacity: 0.5;
}

.apikey-item-info {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 0;
}

.apikey-item-name {
  font-size: 0.9rem;
  font-weight: 500;
}

.apikey-item-meta {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.revoked-badge {
  font-size: 0.7rem;
  padding: 0.15rem 0.5rem;
  background: var(--bg-primary);
  border-radius: var(--radius-full);
  color: var(--text-muted);
}

.btn-danger {
  background: #e74c3c;
  color: #fff;
  border: none;
  cursor: pointer;
}

.btn-danger:hover {
  background: #c0392b;
}

@media (max-width: 640px) {
  .setting-item {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
