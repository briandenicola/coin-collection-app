<template>
  <div class="app">
    <nav v-if="auth.isAuthenticated" class="nav-bar">
      <div class="nav-content">
        <router-link to="/" class="nav-brand">
          <img src="/coin-logo.jpg" alt="Ancient Coins" class="nav-logo" />
          <span class="nav-title">Ancient Coins</span>
        </router-link>
        <div class="nav-links">
          <router-link to="/" class="nav-link" active-class="active">
            <span class="nav-icon">🏛️</span>
            <span class="nav-label">Collection</span>
          </router-link>
          <router-link to="/wishlist" class="nav-link" active-class="active">
            <span class="nav-icon">⭐</span>
            <span class="nav-label">Wishlist</span>
          </router-link>
          <router-link to="/stats" class="nav-link" active-class="active">
            <span class="nav-icon">📊</span>
            <span class="nav-label">Stats</span>
          </router-link>
          <router-link to="/add" class="nav-link add-link" active-class="active">
            <span class="nav-icon">➕</span>
            <span class="nav-label">Add</span>
          </router-link>
        </div>
        <div class="nav-right">
          <router-link to="/settings" class="nav-link" active-class="active">
            <span class="nav-icon">⚙️</span>
            <span class="nav-label">Settings</span>
          </router-link>
          <router-link v-if="auth.isAdmin" to="/admin" class="nav-link" active-class="active">
            <span class="nav-icon">🛡️</span>
            <span class="nav-label">Admin</span>
          </router-link>
          <button class="btn-logout" @click="handleLogout">Logout</button>
        </div>
      </div>
    </nav>
    <main class="main-content" :class="{ 'with-nav': auth.isAuthenticated }">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

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
</style>
