<template>
  <div class="app">
    <nav v-if="auth.isAuthenticated" class="nav-bar" :class="{ 'pwa-mode': isPwa }">
      <div class="nav-content">
        <component :is="isPwa ? 'router-link' : 'router-link'" to="/" class="nav-brand">
          <img src="/coin-logo.jpg" alt="Ancient Coins" class="nav-logo" />
          <span class="nav-title">Coin Collection</span>
        </component>
        <div class="nav-links">
          <router-link v-if="!isPwa" to="/" class="nav-link" active-class="active">
            <Landmark :size="18" />
            <span class="nav-label">Collection</span>
          </router-link>
          <router-link to="/wishlist" class="nav-link" active-class="active">
            <Bookmark :size="18" />
            <span class="nav-label">Wishlist</span>
          </router-link>
          <router-link to="/sold" class="nav-link" active-class="active">
            <BadgeDollarSign :size="18" />
            <span class="nav-label">Sold</span>
          </router-link>
          <router-link v-if="isPwa" to="/add" class="nav-link add-link" active-class="active">
            <CirclePlus :size="18" />
            <span class="nav-label">Add</span>
          </router-link>
          <router-link to="/followers" class="nav-link" active-class="active">
            <UsersIcon :size="18" />
            <span class="nav-label">Followers</span>
          </router-link>
          <button v-if="!isPwa" class="nav-link" @click="showChat = true">
            <Bot :size="18" />
            <span class="nav-label">Agent</span>
          </button>
          <router-link to="/stats" class="nav-link" active-class="active">
            <BarChart3 :size="18" />
            <span class="nav-label">Stats</span>
          </router-link>
          <router-link to="/timeline" class="nav-link" active-class="active">
            <Clock :size="18" />
            <span class="nav-label">Timeline</span>
          </router-link>
        </div>
        <div class="nav-right">
          <router-link to="/settings" class="nav-link" active-class="active">
            <Settings :size="18" />
            <span class="nav-label">Settings</span>
          </router-link>
          <router-link v-if="auth.isAdmin" to="/admin" class="nav-link" active-class="active">
            <ShieldCheck :size="18" />
            <span class="nav-label">Admin</span>
          </router-link>
          <button v-if="!isPwa" class="btn-logout" @click="handleLogout">
            <LogOut :size="16" />
          </button>
        </div>
      </div>
    </nav>
    <main class="main-content" :class="{ 'with-nav': auth.isAuthenticated }">
      <router-view />
    </main>

    <!-- PWA floating agent button -->
    <button
      v-if="isPwa && auth.isAuthenticated && !showChat"
      class="agent-fab"
      @click="showChat = true"
      aria-label="Open AI Agent"
    >
      <Bot :size="22" />
    </button>

    <!-- AI Agent Chat -->
    <CoinSearchChat v-if="showChat" @close="showChat = false" />

    <!-- Email prompt modal for legacy users -->
    <div v-if="showEmailPrompt" class="modal-overlay" @click.self="dismissEmailPrompt">
      <div class="modal card">
        <h3>📧 Add Your Email</h3>
        <p>An email address is now required. Please add yours to continue using all features.</p>
        <div class="form-group" style="margin: 1rem 0">
          <input v-model="promptEmail" type="email" class="form-input" placeholder="you@example.com" />
        </div>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="dismissEmailPrompt">Later</button>
          <button class="btn btn-primary" @click="savePromptEmail" :disabled="!promptEmail">Save</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { Landmark, Bookmark, BadgeDollarSign, BarChart3, CirclePlus, Settings, ShieldCheck, LogOut, Users as UsersIcon, Clock, Bot } from 'lucide-vue-next'
import { updateProfile, getMe } from '@/api/client'
import CoinSearchChat from '@/components/CoinSearchChat.vue'

const auth = useAuthStore()
const router = useRouter()
const isPwa = window.matchMedia('(display-mode: standalone)').matches
  || (window.navigator as any).standalone === true

const showChat = ref(false)
const showEmailPrompt = ref(false)
const promptEmail = ref('')

onMounted(async () => {
  if (auth.isAuthenticated) {
    try {
      const res = await getMe()
      if (res.data.emailMissing) {
        const dismissed = localStorage.getItem('emailPromptDismissed')
        if (!dismissed || Date.now() - parseInt(dismissed) > 7 * 24 * 60 * 60 * 1000) {
          showEmailPrompt.value = true
        }
      }
    } catch { /* ignore */ }
  }
})

function dismissEmailPrompt() {
  showEmailPrompt.value = false
  localStorage.setItem('emailPromptDismissed', Date.now().toString())
}

async function savePromptEmail() {
  if (!promptEmail.value) return
  try {
    await updateProfile({ email: promptEmail.value })
    showEmailPrompt.value = false
    localStorage.removeItem('emailPromptDismissed')
  } catch { /* ignore */ }
}

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.app {
  min-height: 100vh;
}

.nav-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  background: rgba(15, 15, 26, 0.95);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border-subtle);
}

.nav-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  text-decoration: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.nav-logo {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--accent-gold-dim);
}

.nav-title {
  font-family: 'Cinzel', serif;
  font-size: 1.1rem;
  color: var(--accent-gold);
  font-weight: 600;
  white-space: nowrap;
}

.nav-links {
  display: flex;
  gap: 0.25rem;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 0.8rem;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 0.85rem;
  transition: all var(--transition-fast);
  text-decoration: none;
}

.nav-link:hover,
.nav-link.active {
  color: var(--accent-gold);
  background: var(--accent-gold-glow);
}

.add-link {
  color: var(--accent-gold);
}

.nav-right {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.btn-logout {
  padding: 0.4rem 0.8rem;
  background: transparent;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-logout:hover {
  border-color: var(--border-accent);
  color: var(--text-primary);
}

.main-content {
  min-height: 100vh;
}

.main-content.with-nav {
  padding-top: 76px;
}

@media (max-width: 640px) {
  .nav-title { display: none; }
  .nav-label { display: none; }
  .nav-link { padding: 0.5rem; }
  .nav-icon { font-size: 1.2rem; }
}

.pwa-mode .nav-links {
  flex: 1;
  justify-content: space-evenly;
  gap: 0;
}

.pwa-mode .nav-right {
  gap: 0;
}

.pwa-mode .nav-content {
  gap: 0.5rem;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  max-width: 360px;
  width: 90%;
  padding: 2rem;
}

.modal h3 {
  margin-bottom: 0.5rem;
}

.modal p {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin-bottom: 0;
}

.modal-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

/* Agent nav button — match <a> nav-link appearance */
button.nav-link {
  background: none;
  border: none;
  cursor: pointer;
  font: inherit;
  font-size: 0.85rem;
  color: var(--text-secondary);
  padding: 0.5rem 0.8rem;
  line-height: normal;
  letter-spacing: inherit;
}

/* PWA floating agent button */
.agent-fab {
  position: fixed;
  bottom: 24px;
  right: 24px;
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: var(--accent-gold);
  color: #1a1a2e;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
  z-index: 1000;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.agent-fab:active {
  transform: scale(0.92);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}
</style>
