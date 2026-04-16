<template>
  <div class="app">
    <!-- Nav bar — brand + hamburger for both desktop and PWA -->
    <nav v-if="auth.isAuthenticated" class="nav-bar" :class="{ 'pwa-mode': isPwa }">
      <div class="nav-content">
        <button class="nav-brand" @click="sidebarOpen = !sidebarOpen">
          <img src="/coin-logo.jpg" alt="Ancient Coins" class="nav-logo" />
          <span class="nav-title">Coin Collection</span>
        </button>
      </div>
    </nav>

    <!-- Sidebar overlay -->
    <Transition name="sidebar-fade">
      <div v-if="sidebarOpen" class="sidebar-overlay" @click="sidebarOpen = false"></div>
    </Transition>

    <!-- Slide-in sidebar -->
    <Transition name="sidebar-slide">
      <aside v-if="sidebarOpen" class="sidebar">
        <div class="sidebar-header">
          <img src="/coin-logo.jpg" alt="Ancient Coins" class="sidebar-logo" />
          <span class="sidebar-title">Coin Collection</span>
          <button class="sidebar-close" @click="sidebarOpen = false">
            <X :size="20" />
          </button>
        </div>
        <nav class="sidebar-nav">
          <router-link to="/" class="sidebar-link" active-class="active" @click="sidebarOpen = false">
            <Landmark :size="20" />
            <span>Collection</span>
          </router-link>
          <router-link v-if="isPwa" to="/add" class="sidebar-link" active-class="active" @click="sidebarOpen = false">
            <CirclePlus :size="20" />
            <span>Add Coin</span>
          </router-link>
          <router-link to="/wishlist" class="sidebar-link" active-class="active" @click="sidebarOpen = false">
            <Bookmark :size="20" />
            <span>Wishlist</span>
          </router-link>
          <router-link to="/sold" class="sidebar-link" active-class="active" @click="sidebarOpen = false">
            <BadgeDollarSign :size="20" />
            <span>Sold</span>
          </router-link>
          <router-link to="/auctions" class="sidebar-link" active-class="active" @click="sidebarOpen = false">
            <Gavel :size="20" />
            <span>Auctions</span>
          </router-link>
          <router-link to="/followers" class="sidebar-link" active-class="active" @click="sidebarOpen = false">
            <UsersIcon :size="20" />
            <span>Followers</span>
          </router-link>
          <a href="#" class="sidebar-link" @click.prevent="showChat = true; sidebarOpen = false">
            <Bot :size="20" />
            <span>Agent</span>
          </a>
          <router-link to="/stats" class="sidebar-link" active-class="active" @click="sidebarOpen = false">
            <BarChart3 :size="20" />
            <span>Stats</span>
          </router-link>
          <router-link to="/timeline" class="sidebar-link" active-class="active" @click="sidebarOpen = false">
            <Clock :size="20" />
            <span>Timeline</span>
          </router-link>
        </nav>
        <div class="sidebar-footer">
          <router-link to="/settings" class="sidebar-link" active-class="active" @click="sidebarOpen = false">
            <Settings :size="20" />
            <span>Settings</span>
          </router-link>
          <router-link v-if="auth.isAdmin" to="/admin" class="sidebar-link" active-class="active" @click="sidebarOpen = false">
            <ShieldCheck :size="20" />
            <span>Admin</span>
          </router-link>
          <button class="sidebar-link sidebar-logout" @click="handleLogout">
            <LogOut :size="20" />
            <span>Logout</span>
          </button>
        </div>
      </aside>
    </Transition>

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

    <AppDialog />
    <PwaInstallPrompt />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { Landmark, Bookmark, BadgeDollarSign, BarChart3, CirclePlus, Settings, ShieldCheck, LogOut, Users as UsersIcon, Clock, Bot, Gavel, X } from 'lucide-vue-next'
import { updateProfile, getMe } from '@/api/client'
import CoinSearchChat from '@/components/CoinSearchChat.vue'
import AppDialog from '@/components/AppDialog.vue'
import PwaInstallPrompt from '@/components/PwaInstallPrompt.vue'

const auth = useAuthStore()
const router = useRouter()
const isPwa = window.matchMedia('(display-mode: standalone)').matches
  || (window.navigator as any).standalone === true

const showChat = ref(false)
const sidebarOpen = ref(false)
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

/* ── Top nav bar (shared structure) ── */
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
  padding: 0.4rem 0.6rem;
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
  flex-shrink: 0;
}

.nav-brand:hover {
  background: var(--accent-gold-glow);
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

/* ── Sidebar ── */
.sidebar-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 199;
}

.sidebar {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 280px;
  background: var(--bg-card);
  border-right: 1px solid var(--border-subtle);
  z-index: 200;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1.25rem 1.25rem 1rem;
  border-bottom: 1px solid var(--border-subtle);
}

.sidebar-logo {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--accent-gold-dim);
}

.sidebar-title {
  font-family: 'Cinzel', serif;
  font-size: 1.05rem;
  color: var(--accent-gold);
  font-weight: 600;
  flex: 1;
}

.sidebar-close {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0.3rem;
  border-radius: var(--radius-sm);
  transition: color var(--transition-fast), background var(--transition-fast);
}

.sidebar-close:hover {
  color: var(--text-primary);
  background: var(--accent-gold-glow);
}

.sidebar-nav {
  flex: 1;
  padding: 0.75rem 0;
}

.sidebar-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.7rem 1.25rem;
  color: var(--text-secondary);
  font-size: 0.92rem;
  text-decoration: none;
  transition: all var(--transition-fast);
  border: none;
  background: none;
  width: 100%;
  cursor: pointer;
  font-family: inherit;
}

.sidebar-link:hover {
  color: var(--accent-gold);
  background: var(--accent-gold-glow);
}

.sidebar-link.active {
  color: var(--accent-gold);
  background: var(--accent-gold-glow);
  border-right: 3px solid var(--accent-gold);
}

.sidebar-footer {
  border-top: 1px solid var(--border-subtle);
  padding: 0.5rem 0;
}

.sidebar-logout {
  color: var(--text-secondary);
}

.sidebar-logout:hover {
  color: #f87171;
  background: rgba(248, 113, 113, 0.1);
}

/* Sidebar transitions */
.sidebar-slide-enter-active,
.sidebar-slide-leave-active {
  transition: transform 0.25s ease;
}

.sidebar-slide-enter-from,
.sidebar-slide-leave-to {
  transform: translateX(-100%);
}

.sidebar-fade-enter-active,
.sidebar-fade-leave-active {
  transition: opacity 0.25s ease;
}

.sidebar-fade-enter-from,
.sidebar-fade-leave-to {
  opacity: 0;
}

.main-content {
  min-height: 100vh;
}

.main-content.with-nav {
  padding-top: 76px;
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
