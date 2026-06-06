<template>
  <div class="sets-page">
    <div class="sets-header">
      <h1>My Sets</h1>
      <button class="btn btn-primary" @click="showCreateModal = true">
        Create Set
      </button>
    </div>

    <div v-if="loading" class="loading-state">
      Loading sets...
    </div>

    <div v-else-if="sets.length === 0" class="empty-state">
      <FolderOpen :size="48" />
      <h2>No sets yet</h2>
      <p>Create your first set to organize your collection</p>
      <button class="btn btn-primary" @click="showCreateModal = true">
        Create Set
      </button>
    </div>

    <div v-else class="sets-grid">
      <SetDashboardCard
        v-for="set in sets"
        :key="set.id"
        :set="set"
        @click="goToSet(set.id)"
      />
    </div>

    <!-- Create Set Modal - placeholder for T026 wizard -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal-content">
        <h2>Create New Set</h2>
        <SetCreationWizard
          :initial-value="newSet"
          submit-label="Create"
          @submit="createSet"
          @cancel="showCreateModal = false"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { FolderOpen } from 'lucide-vue-next'
import { getSets, createSet as createSetApi, createSetFromCsv } from '@/api/client'
import type { CoinSetSummary, CreateCoinSetRequest } from '@/types'
import SetDashboardCard from '@/components/sets/SetDashboardCard.vue'
import SetCreationWizard from '@/components/sets/SetCreationWizard.vue'

const router = useRouter()
const loading = ref(true)
const sets = ref<CoinSetSummary[]>([])
const showCreateModal = ref(false)
const newSet = ref({
  name: '',
  description: '',
  color: '#6b7280',
  setType: 'open' as const,
})

onMounted(async () => {
  await loadSets()
})

async function loadSets() {
  loading.value = true
  try {
    const res = await getSets()
    sets.value = res.data.sets
  } catch (error) {
    console.error('Failed to load sets:', error)
  } finally {
    loading.value = false
  }
}

async function createSet(value: CreateCoinSetRequest, csv?: string) {
  try {
    if (csv) {
      await createSetFromCsv({ ...value, csv })
    } else {
      await createSetApi(value)
    }
    showCreateModal.value = false
    newSet.value = {
      name: '',
      description: '',
      color: '#6b7280',
      setType: 'open',
    }
    await loadSets()
  } catch (error) {
    console.error('Failed to create set:', error)
    alert('Failed to create set')
  }
}

function goToSet(id: number) {
  router.push({ name: 'set-detail', params: { id } })
}
</script>

<style scoped>
.sets-page {
  padding: 1.5rem;
  max-width: 1200px;
  margin: 0 auto;
}

.sets-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.sets-header h1 {
  margin: 0;
}

.loading-state,
.empty-state {
  text-align: center;
  padding: 3rem 1rem;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.empty-state h2 {
  margin: 0;
}

.empty-state p {
  color: var(--text-secondary);
  margin: 0;
}

.sets-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--modal-backdrop, rgba(0, 0, 0, 0.6));
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  padding: 2rem;
  border-radius:var(--radius-md);
  max-width: 500px;
  width: 90%;
  box-shadow: var(--shadow-card);
}

.modal-content h2 {
  margin-top: 0;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid var(--border-subtle);
  border-radius:var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-primary);
  font-family: inherit;
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}
</style>
