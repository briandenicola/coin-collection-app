<template>
  <div class="image-processor">
    <!-- Step 1: Image Input -->
    <div v-if="!sourceImage" class="input-section card">
      <h3>Load Image</h3>
      <div class="input-methods">
        <label class="drop-zone" :class="{ dragging }" @dragover.prevent="dragging = true"
          @dragleave="dragging = false" @drop.prevent="handleDrop">
          <Upload :size="32" />
          <span>Drop an image here or click to browse</span>
          <input type="file" accept="image/*" hidden @change="handleFileSelect" />
        </label>
        <div class="url-input-row">
          <input v-model="imageUrl" type="url" class="form-input" placeholder="Or paste an image URL..."
            @keydown.enter="handleUrlLoad" />
          <button class="btn btn-primary btn-sm" :disabled="!imageUrl || urlLoading" @click="handleUrlLoad">
            {{ urlLoading ? 'Loading...' : 'Fetch' }}
          </button>
        </div>
        <p v-if="inputError" class="msg error">{{ inputError }}</p>
      </div>
    </div>

    <!-- Step 2: Processing & Preview -->
    <div v-if="sourceImage" class="processing-section">
      <div class="step-header">
        <button class="btn btn-secondary btn-sm" @click="reset">← Start Over</button>
        <div class="step-indicators">
          <span class="step" :class="{ active: step === 'preview', done: step !== 'preview' }">1. Original</span>
          <span class="step" :class="{ active: step === 'removing', done: step === 'crop' || step === 'done' }">2.
            Remove BG</span>
          <span class="step" :class="{ active: step === 'crop', done: step === 'done' }">3. Crop</span>
        </div>
      </div>

      <!-- Original preview + remove BG button -->
      <div v-if="step === 'preview'" class="preview-panel">
        <div class="image-preview">
          <img :src="sourceImage" alt="Original" />
        </div>
        <button class="btn btn-primary" @click="removeBackground">Remove Background</button>
      </div>

      <!-- Processing spinner -->
      <div v-if="step === 'removing'" class="preview-panel">
        <div class="image-preview processing">
          <img :src="sourceImage" alt="Processing" class="processing-img" />
          <div class="processing-overlay">
            <div class="spinner"></div>
            <p>Removing background...</p>
            <p class="processing-hint">First run downloads the ML model (~40MB)</p>
          </div>
        </div>
      </div>

      <!-- Crop step -->
      <div v-if="step === 'crop' || step === 'done'" class="crop-panel">
        <div class="crop-workspace">
          <canvas ref="cropCanvas" class="crop-canvas" @pointerdown="startCropDrag" @pointermove="onCropDrag"
            @pointerup="endCropDrag" />
        </div>
        <div class="crop-controls">
          <button class="btn btn-secondary btn-sm" @click="autoCrop">Auto Crop</button>
          <button class="btn btn-secondary btn-sm" @click="resetCrop">Reset Crop</button>
          <label class="padding-control">
            <span>Padding</span>
            <input v-model.number="cropPadding" type="range" min="0" max="50" />
            <span class="padding-value">{{ cropPadding }}px</span>
          </label>
        </div>

        <!-- Result preview -->
        <div class="result-row">
          <div class="result-preview">
            <h4>Result</h4>
            <canvas ref="resultCanvas" class="result-canvas" />
          </div>
          <div class="save-controls">
            <div v-if="coinId" class="save-to-coin">
              <h4>Save to Coin</h4>
              <div class="save-options">
                <label class="type-select-label">Image Type</label>
                <select v-model="saveImageType" class="form-select">
                  <option value="obverse">Obverse</option>
                  <option value="reverse">Reverse</option>
                  <option value="detail">Detail</option>
                  <option value="other">Other</option>
                </select>
                <label class="checkbox-label">
                  <input v-model="savePrimary" type="checkbox" />
                  Set as primary image
                </label>
                <button class="btn btn-primary" :disabled="saving" @click="saveToCoIn">
                  {{ saving ? 'Saving...' : 'Upload to Coin' }}
                </button>
              </div>
            </div>
            <div class="download-section">
              <h4>Download</h4>
              <button class="btn btn-secondary" @click="downloadResult">💾 Download PNG</button>
            </div>
            <p v-if="saveMsg" class="msg" :class="{ error: saveError }">{{ saveMsg }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onUnmounted } from 'vue'
import { removeBackground as removeBg } from '@imgly/background-removal'
import { Upload } from 'lucide-vue-next'
import { proxyImage, uploadImage } from '@/api/client'

const props = defineProps<{
  coinId?: number
}>()

const emit = defineEmits<{
  saved: []
}>()

// Input state
const sourceImage = ref<string | null>(null)
const imageUrl = ref('')
const urlLoading = ref(false)
const inputError = ref('')
const dragging = ref(false)

// Processing state
type Step = 'preview' | 'removing' | 'crop' | 'done'
const step = ref<Step>('preview')
const processedBlob = ref<Blob | null>(null)
const processedImage = ref<HTMLImageElement | null>(null)

// Crop state
const cropCanvas = ref<HTMLCanvasElement | null>(null)
const resultCanvas = ref<HTMLCanvasElement | null>(null)
const cropPadding = ref(10)
const cropRect = ref({ x: 0, y: 0, w: 0, h: 0 })
const cropDragging = ref(false)
const cropDragType = ref<'move' | 'nw' | 'ne' | 'sw' | 'se' | null>(null)
const cropDragStart = ref({ x: 0, y: 0, rx: 0, ry: 0, rw: 0, rh: 0 })
let canvasScale = 1

// Save state
const saveImageType = ref('obverse')
const savePrimary = ref(false)
const saving = ref(false)
const saveMsg = ref('')
const saveError = ref(false)

// --- Input Methods ---

function loadImageFromFile(file: File) {
  inputError.value = ''
  if (!file.type.startsWith('image/')) {
    inputError.value = 'Please select an image file'
    return
  }
  const reader = new FileReader()
  reader.onload = (e) => {
    sourceImage.value = e.target?.result as string
    step.value = 'preview'
  }
  reader.readAsDataURL(file)
}

function handleFileSelect(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (file) loadImageFromFile(file)
}

function handleDrop(e: DragEvent) {
  dragging.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file) loadImageFromFile(file)
}

async function handleUrlLoad() {
  if (!imageUrl.value) return
  inputError.value = ''
  urlLoading.value = true
  try {
    const res = await proxyImage(imageUrl.value)
    const blob = res.data as Blob
    const reader = new FileReader()
    reader.onload = (e) => {
      sourceImage.value = e.target?.result as string
      step.value = 'preview'
    }
    reader.readAsDataURL(blob)
  } catch {
    inputError.value = 'Failed to fetch image from URL'
  } finally {
    urlLoading.value = false
  }
}

// --- Background Removal ---

async function removeBackground() {
  if (!sourceImage.value) return
  step.value = 'removing'

  try {
    const response = await fetch(sourceImage.value)
    const srcBlob = await response.blob()
    const result = await removeBg(srcBlob, {
      output: { format: 'image/png', quality: 1 },
    })
    processedBlob.value = result
    const img = new Image()
    img.onload = () => {
      processedImage.value = img
      step.value = 'crop'
      nextTick(() => {
        autoCrop()
      })
    }
    img.src = URL.createObjectURL(result)
  } catch (err) {
    console.error('Background removal failed:', err)
    inputError.value = 'Background removal failed. Please try again.'
    step.value = 'preview'
  }
}

// --- Crop Logic ---

function autoCrop() {
  if (!processedImage.value || !cropCanvas.value) return

  // Draw to offscreen canvas to read pixels
  const img = processedImage.value
  const offscreen = document.createElement('canvas')
  offscreen.width = img.naturalWidth
  offscreen.height = img.naturalHeight
  const ctx = offscreen.getContext('2d')!
  ctx.drawImage(img, 0, 0)

  const data = ctx.getImageData(0, 0, offscreen.width, offscreen.height).data
  let minX = offscreen.width, minY = offscreen.height, maxX = 0, maxY = 0

  for (let y = 0; y < offscreen.height; y++) {
    for (let x = 0; x < offscreen.width; x++) {
      const alpha = data[(y * offscreen.width + x) * 4 + 3]
      if (alpha > 10) {
        if (x < minX) minX = x
        if (x > maxX) maxX = x
        if (y < minY) minY = y
        if (y > maxY) maxY = y
      }
    }
  }

  if (maxX <= minX || maxY <= minY) {
    // No visible content found, use full image
    cropRect.value = { x: 0, y: 0, w: img.naturalWidth, h: img.naturalHeight }
  } else {
    const pad = cropPadding.value
    cropRect.value = {
      x: Math.max(0, minX - pad),
      y: Math.max(0, minY - pad),
      w: Math.min(img.naturalWidth - Math.max(0, minX - pad), maxX - minX + 1 + pad * 2),
      h: Math.min(img.naturalHeight - Math.max(0, minY - pad), maxY - minY + 1 + pad * 2),
    }
  }

  drawCropCanvas()
  drawResultCanvas()
}

function resetCrop() {
  if (!processedImage.value) return
  cropRect.value = {
    x: 0, y: 0,
    w: processedImage.value.naturalWidth,
    h: processedImage.value.naturalHeight,
  }
  drawCropCanvas()
  drawResultCanvas()
}

function drawCropCanvas() {
  const canvas = cropCanvas.value
  const img = processedImage.value
  if (!canvas || !img) return

  // Scale to fit container (max 500px wide)
  const maxW = Math.min(500, canvas.parentElement?.clientWidth || 500)
  canvasScale = maxW / img.naturalWidth
  const dispH = img.naturalHeight * canvasScale

  canvas.width = maxW
  canvas.height = dispH
  canvas.style.width = maxW + 'px'
  canvas.style.height = dispH + 'px'

  const ctx = canvas.getContext('2d')!

  // Checkerboard background for transparency
  drawCheckerboard(ctx, maxW, dispH)

  // Draw image
  ctx.drawImage(img, 0, 0, maxW, dispH)

  // Dim outside crop area
  const r = cropRect.value
  const sx = r.x * canvasScale
  const sy = r.y * canvasScale
  const sw = r.w * canvasScale
  const sh = r.h * canvasScale

  ctx.fillStyle = 'rgba(0, 0, 0, 0.5)'
  ctx.fillRect(0, 0, maxW, sy) // top
  ctx.fillRect(0, sy + sh, maxW, dispH - sy - sh) // bottom
  ctx.fillRect(0, sy, sx, sh) // left
  ctx.fillRect(sx + sw, sy, maxW - sx - sw, sh) // right

  // Crop border
  ctx.strokeStyle = '#c9a84c'
  ctx.lineWidth = 2
  ctx.strokeRect(sx, sy, sw, sh)

  // Corner handles
  const handleSize = 8
  ctx.fillStyle = '#c9a84c'
  for (const [hx, hy] of [[sx, sy], [sx + sw, sy], [sx, sy + sh], [sx + sw, sy + sh]]) {
    ctx.fillRect(hx - handleSize / 2, hy - handleSize / 2, handleSize, handleSize)
  }
}

function drawResultCanvas() {
  const canvas = resultCanvas.value
  const img = processedImage.value
  if (!canvas || !img) return

  const r = cropRect.value
  const w = Math.max(1, Math.round(r.w))
  const h = Math.max(1, Math.round(r.h))

  // Scale result preview to max 200px
  const maxDim = 200
  const scale = Math.min(maxDim / w, maxDim / h, 1)
  canvas.width = Math.round(w * scale)
  canvas.height = Math.round(h * scale)

  const ctx = canvas.getContext('2d')!
  drawCheckerboard(ctx, canvas.width, canvas.height)
  ctx.drawImage(img, r.x, r.y, r.w, r.h, 0, 0, canvas.width, canvas.height)
}

function drawCheckerboard(ctx: CanvasRenderingContext2D, w: number, h: number) {
  const size = 8
  for (let y = 0; y < h; y += size) {
    for (let x = 0; x < w; x += size) {
      ctx.fillStyle = (Math.floor(x / size) + Math.floor(y / size)) % 2 === 0 ? '#2a2a3e' : '#1e1e30'
      ctx.fillRect(x, y, size, size)
    }
  }
}

// --- Crop Drag ---

function getCanvasPos(e: PointerEvent) {
  const canvas = cropCanvas.value!
  const rect = canvas.getBoundingClientRect()
  return {
    x: (e.clientX - rect.left) / canvasScale,
    y: (e.clientY - rect.top) / canvasScale,
  }
}

function startCropDrag(e: PointerEvent) {
  const pos = getCanvasPos(e)
  const r = cropRect.value
  const handleThreshold = 12 / canvasScale

  // Check corners
  if (Math.abs(pos.x - r.x) < handleThreshold && Math.abs(pos.y - r.y) < handleThreshold) {
    cropDragType.value = 'nw'
  } else if (Math.abs(pos.x - (r.x + r.w)) < handleThreshold && Math.abs(pos.y - r.y) < handleThreshold) {
    cropDragType.value = 'ne'
  } else if (Math.abs(pos.x - r.x) < handleThreshold && Math.abs(pos.y - (r.y + r.h)) < handleThreshold) {
    cropDragType.value = 'sw'
  } else if (Math.abs(pos.x - (r.x + r.w)) < handleThreshold && Math.abs(pos.y - (r.y + r.h)) < handleThreshold) {
    cropDragType.value = 'se'
  } else if (pos.x >= r.x && pos.x <= r.x + r.w && pos.y >= r.y && pos.y <= r.y + r.h) {
    cropDragType.value = 'move'
  } else {
    return
  }

  cropDragging.value = true
  cropDragStart.value = { x: pos.x, y: pos.y, rx: r.x, ry: r.y, rw: r.w, rh: r.h }
  cropCanvas.value?.setPointerCapture(e.pointerId)
}

function onCropDrag(e: PointerEvent) {
  if (!cropDragging.value || !processedImage.value) return
  const pos = getCanvasPos(e)
  const s = cropDragStart.value
  const dx = pos.x - s.x
  const dy = pos.y - s.y
  const imgW = processedImage.value.naturalWidth
  const imgH = processedImage.value.naturalHeight

  const r = { ...cropRect.value }

  switch (cropDragType.value) {
    case 'move':
      r.x = Math.max(0, Math.min(imgW - s.rw, s.rx + dx))
      r.y = Math.max(0, Math.min(imgH - s.rh, s.ry + dy))
      break
    case 'nw':
      r.x = Math.max(0, Math.min(s.rx + s.rw - 20, s.rx + dx))
      r.y = Math.max(0, Math.min(s.ry + s.rh - 20, s.ry + dy))
      r.w = s.rw - (r.x - s.rx)
      r.h = s.rh - (r.y - s.ry)
      break
    case 'ne':
      r.w = Math.max(20, Math.min(imgW - s.rx, s.rw + dx))
      r.y = Math.max(0, Math.min(s.ry + s.rh - 20, s.ry + dy))
      r.h = s.rh - (r.y - s.ry)
      break
    case 'sw':
      r.x = Math.max(0, Math.min(s.rx + s.rw - 20, s.rx + dx))
      r.w = s.rw - (r.x - s.rx)
      r.h = Math.max(20, Math.min(imgH - s.ry, s.rh + dy))
      break
    case 'se':
      r.w = Math.max(20, Math.min(imgW - s.rx, s.rw + dx))
      r.h = Math.max(20, Math.min(imgH - s.ry, s.rh + dy))
      break
  }

  cropRect.value = r
  drawCropCanvas()
  drawResultCanvas()
}

function endCropDrag() {
  cropDragging.value = false
  cropDragType.value = null
}

// Redraw when padding changes
watch(cropPadding, () => {
  if (step.value === 'crop' || step.value === 'done') {
    autoCrop()
  }
})

// --- Save / Download ---

function getResultBlob(): Promise<Blob> {
  return new Promise((resolve) => {
    const img = processedImage.value!
    const r = cropRect.value
    const canvas = document.createElement('canvas')
    canvas.width = Math.round(r.w)
    canvas.height = Math.round(r.h)
    const ctx = canvas.getContext('2d')!
    ctx.drawImage(img, r.x, r.y, r.w, r.h, 0, 0, canvas.width, canvas.height)
    canvas.toBlob((blob) => resolve(blob!), 'image/png')
  })
}

async function saveToCoIn() {
  if (!props.coinId) return
  saving.value = true
  saveMsg.value = ''
  saveError.value = false
  try {
    const blob = await getResultBlob()
    const file = new File([blob], `${saveImageType.value}.png`, { type: 'image/png' })
    await uploadImage(props.coinId, file, saveImageType.value, savePrimary.value)
    saveMsg.value = `Saved as ${saveImageType.value} image!`
    emit('saved')
  } catch {
    saveMsg.value = 'Failed to save image'
    saveError.value = true
  } finally {
    saving.value = false
  }
}

async function downloadResult() {
  const blob = await getResultBlob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `coin-${saveImageType.value}-processed.png`
  a.click()
  URL.revokeObjectURL(url)
}

function reset() {
  if (processedImage.value) URL.revokeObjectURL(processedImage.value.src)
  sourceImage.value = null
  processedBlob.value = null
  processedImage.value = null
  step.value = 'preview'
  inputError.value = ''
  saveMsg.value = ''
}

onUnmounted(() => {
  if (processedImage.value) URL.revokeObjectURL(processedImage.value.src)
})
</script>

<style scoped>
.image-processor {
  max-width: 700px;
  margin: 0 auto;
}

.input-section h3 {
  margin-bottom: 1rem;
}

.input-methods {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.drop-zone {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 2.5rem 1.5rem;
  border: 2px dashed var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
  text-align: center;
}

.drop-zone:hover,
.drop-zone.dragging {
  border-color: var(--accent-gold);
  color: var(--accent-gold);
  background: var(--accent-gold-glow);
}

.url-input-row {
  display: flex;
  gap: 0.5rem;
}

.url-input-row .form-input {
  flex: 1;
}

/* Steps */
.step-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
}

.step-indicators {
  display: flex;
  gap: 0.5rem;
}

.step {
  font-size: 0.75rem;
  padding: 0.3rem 0.6rem;
  border-radius: var(--radius-full);
  color: var(--text-muted);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
}

.step.active {
  color: var(--accent-gold);
  border-color: var(--accent-gold);
  background: var(--accent-gold-glow);
}

.step.done {
  color: var(--text-secondary);
  border-color: var(--accent-gold-dim);
}

/* Preview */
.preview-panel {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.image-preview {
  position: relative;
  max-width: 500px;
  width: 100%;
  border-radius: var(--radius-md);
  overflow: hidden;
  border: 1px solid var(--border-subtle);
}

.image-preview img {
  width: 100%;
  display: block;
}

.image-preview.processing {
  position: relative;
}

.processing-img {
  opacity: 0.3;
  filter: blur(2px);
}

.processing-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  color: var(--accent-gold);
}

.processing-overlay p {
  font-size: 0.9rem;
}

.processing-hint {
  font-size: 0.75rem !important;
  color: var(--text-muted) !important;
}

/* Crop */
.crop-panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.crop-workspace {
  display: flex;
  justify-content: center;
}

.crop-canvas {
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  cursor: crosshair;
  max-width: 100%;
  touch-action: none;
}

.crop-controls {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.padding-control {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  color: var(--text-secondary);
  margin-left: auto;
}

.padding-control input[type="range"] {
  width: 100px;
  accent-color: var(--accent-gold);
}

.padding-value {
  min-width: 32px;
  text-align: right;
}

/* Result */
.result-row {
  display: flex;
  gap: 1.5rem;
  padding: 1rem;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.result-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.result-preview h4,
.save-controls h4 {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-bottom: 0.25rem;
}

.result-canvas {
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle);
}

.save-controls {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.save-options {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.type-select-label {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.85rem;
  color: var(--text-secondary);
  cursor: pointer;
}

.download-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.msg {
  font-size: 0.85rem;
  color: var(--accent-gold);
}

.msg.error {
  color: #e74c3c;
}

@media (max-width: 640px) {
  .result-row {
    flex-direction: column;
  }

  .step-header {
    flex-direction: column;
    gap: 0.75rem;
    align-items: stretch;
  }

  .step-indicators {
    justify-content: center;
  }
}
</style>
