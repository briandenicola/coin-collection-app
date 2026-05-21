# Skill: Adding Admin Schedule Panel

## Context

The Admin Schedules page (`AdminSchedulesSection.vue`) is where system-wide scheduled jobs are configured. Each scheduler (wishlist availability, collection valuation, auction-ending alerts) has its own panel with enable toggle, start time, interval, and save button.

## When to Use

When Cassius or another backend dev adds a new daily/periodic scheduler that needs user configuration (enabled/disabled, start time, interval), add a panel to the Schedules UI following this recipe.

## Recipe

### 1. Identify Setting Keys

The backend scheduler reads three keys from the `AppSettings` table:
- `{Feature}CheckEnabled` — boolean stored as string `'true'` or `'false'`, default `'false'`
- `{Feature}CheckStartTime` — string `"HH:MM"`, default `"08:00"`
- `{Feature}CheckInterval` — integer minutes stored as string, default `"1440"` (24 hours)

**Example:** For auction-ending alerts, the keys are `AuctionEndingCheckEnabled`, `AuctionEndingCheckStartTime`, `AuctionEndingCheckInterval`.

Check `.squad/decisions/inbox/` for backend docs if Cassius chose different names.

### 2. Add UI Panel in AdminSchedulesSection.vue

Open `src/web/src/components/admin/AdminSchedulesSection.vue`.

**Add the panel** after an existing section (e.g., after wishlist, before valuation):

```vue
<hr class="section-divider" />

<!-- {Feature Name} -->
<h3 class="subsection-title">{Panel Title}</h3>
<p class="subsection-desc">{One-line description of what the scheduler does.}</p>
<div class="avail-settings">
  <div class="form-group avail-toggle-row">
    <label class="form-label">Enable Automatic {Action}</label>
    <label class="toggle-switch">
      <input
        type="checkbox"
        :checked="settings.{Feature}CheckEnabled === 'true'"
        @change="settings.{Feature}CheckEnabled = ($event.target as HTMLInputElement).checked ? 'true' : 'false'"
      />
      <span class="toggle-slider"></span>
    </label>
  </div>
  <div class="form-group">
    <label class="form-label">Start Time (daily anchor)</label>
    <input
      v-model="settings.{Feature}CheckStartTime"
      class="form-input avail-interval-input"
      type="time"
    />
    <span class="form-hint">The first check runs at this time each day.</span>
  </div>
  <div class="form-group">
    <label class="form-label">Repeat Interval (minutes)</label>
    <input
      v-model="settings.{Feature}CheckInterval"
      class="form-input avail-interval-input"
      type="number"
      min="60"
      step="60"
    />
    <span class="form-hint">How often to repeat after the start time. Default {default interval description}.</span>
  </div>
  <div class="avail-save-row">
    <button class="btn btn-primary btn-sm" :disabled="settingsSaving" @click="$emit('save')">
      {{ settingsSaving ? 'Saving...' : 'Save {Feature} Settings' }}
    </button>
    <span v-if="{feature}SettingsMsg" class="avail-save-msg" :class="{ 'avail-save-error': {feature}SettingsError }">{{ {feature}SettingsMsg }}</span>
  </div>
</div>
```

**Add props** to the `defineProps` block:

```ts
const props = defineProps<{
  settings: AppSettings
  settingsSaving: boolean
  availSettingsMsg: string
  availSettingsError: boolean
  {feature}SettingsMsg: string  // ADD THIS
  {feature}SettingsError: boolean  // ADD THIS
  valSettingsMsg: string
  valSettingsError: boolean
}>()
```

### 3. Update useAdminConfig.ts Composable

Open `src/web/src/composables/useAdminConfig.ts`.

**Add state refs:**

```ts
// Schedule-tab save messages (cleared alongside main settingsMsg)
const availSettingsMsg = ref('')
const availSettingsError = ref(false)
const {feature}SettingsMsg = ref('')  // ADD THIS
const {feature}SettingsError = ref(false)  // ADD THIS
const valSettingsMsg = ref('')
const valSettingsError = ref(false)
```

**Update loadSettings()** to apply defaults:

```ts
async function loadSettings() {
  try {
    const [settingsRes, defaultsRes] = await Promise.all([
      getAppSettings(),
      getAppSettingDefaults(),
    ])
    settingDefaults.value = { ...settingDefaults.value, ...defaultsRes.data }
    settings.value = { ...settings.value, ...settingsRes.data }

    // Apply defaults for {feature} settings if not set
    if (!settings.value.{Feature}CheckEnabled) {
      settings.value.{Feature}CheckEnabled = 'false'
    }
    if (!settings.value.{Feature}CheckStartTime) {
      settings.value.{Feature}CheckStartTime = '08:00'
    }
    if (!settings.value.{Feature}CheckInterval) {
      settings.value.{Feature}CheckInterval = '1440'
    }

    // ... rest of loadSettings
  } catch { /* use defaults */ }
}
```

**Update saveSettings()** to clear/set new messages:

```ts
async function saveSettings() {
  settingsSaving.value = true
  settingsMsg.value = ''
  settingsError.value = false
  availSettingsMsg.value = ''
  availSettingsError.value = false
  {feature}SettingsMsg.value = ''  // ADD THIS
  {feature}SettingsError.value = false  // ADD THIS
  valSettingsMsg.value = ''
  valSettingsError.value = false
  try {
    const entries = Object.entries(settings.value).map(([key, value]) => ({ key, value: String(value) }))
    await updateAppSettings(entries)
    settingsMsg.value = 'Settings saved'
    availSettingsMsg.value = 'Settings saved'
    {feature}SettingsMsg.value = 'Settings saved'  // ADD THIS
    valSettingsMsg.value = 'Settings saved'
    if (saveTimerId) clearTimeout(saveTimerId)
    saveTimerId = setTimeout(() => { availSettingsMsg.value = ''; {feature}SettingsMsg.value = ''; valSettingsMsg.value = '' }, 3000)
  } catch {
    settingsMsg.value = 'Failed to save settings'
    settingsError.value = true
    availSettingsMsg.value = 'Failed to save settings'
    availSettingsError.value = true
    {feature}SettingsMsg.value = 'Failed to save settings'  // ADD THIS
    {feature}SettingsError.value = true  // ADD THIS
    valSettingsMsg.value = 'Failed to save settings'
    valSettingsError.value = true
  } finally {
    settingsSaving.value = false
  }
}
```

**Update return object:**

```ts
return {
  // ... existing exports
  // Schedule messages
  availSettingsMsg,
  availSettingsError,
  {feature}SettingsMsg,  // ADD THIS
  {feature}SettingsError,  // ADD THIS
  valSettingsMsg,
  valSettingsError,
  // ... rest
}
```

### 4. Update AdminPage.vue Parent

Open `src/web/src/pages/AdminPage.vue`.

**Destructure new refs** from `useAdminConfig()`:

```ts
const {
  settings, settingDefaults, settingsMsg, settingsError, settingsSaving,
  ollamaTesting, ollamaTestResult, ollamaTestOk,
  anthropicTesting, anthropicTestResult, anthropicTestOk, anthropicModels,
  searxngTesting, searxngTestResult, searxngTestOk,
  coinSearchPromptDefault, coinShowsPromptDefault, valuationPromptDefault,
  availSettingsMsg, availSettingsError, {feature}SettingsMsg, {feature}SettingsError, valSettingsMsg, valSettingsError,  // ADD {feature} HERE
  loadSettings, saveSettings,
  testOllamaConnection, testAnthropicConn, testSearxngConn,
  cleanup: cleanupAdminConfig,
} = useAdminConfig()
```

**Update AdminSchedulesSection binding** in template:

```vue
<AdminSchedulesSection
  v-if="activeTab === 'schedules'"
  :settings="settings"
  :settings-saving="settingsSaving"
  :avail-settings-msg="availSettingsMsg"
  :avail-settings-error="availSettingsError"
  :{feature}-settings-msg="{feature}SettingsMsg"
  :{feature}-settings-error="{feature}SettingsError"
  :val-settings-msg="valSettingsMsg"
  :val-settings-error="valSettingsError"
  @save="saveSettings"
  @update:val-settings-msg="valSettingsMsg = $event"
  @update:val-settings-error="valSettingsError = $event"
/>
```

### 5. Run Type-Check

```bash
cd src/web
npx vue-tsc --noEmit
```

Must pass clean before committing.

## Design Guidelines

- **No emojis** in UI text
- **Use global classes** from `main.css` — `.btn`, `.btn-primary`, `.form-label`, `.form-input`, `.form-hint`, `.toggle-switch`, `.avail-settings`, `.subsection-title`
- **Subsection description** should be one sentence, plain English, no marketing language
- **Interval hint** should suggest realistic defaults (e.g., "Default 1440 (daily)" or "e.g. 120 = every 2 hours")
- **Start Time hint** should clarify it's a daily anchor (first run of the day)

## Example — Auction Ending Alerts

See `.squad/decisions/inbox/aurelia-journal-and-auction-schedule.md` for the complete implementation of `AuctionEndingCheckEnabled`, `AuctionEndingCheckStartTime`, `AuctionEndingCheckInterval`.

**Panel title:** "Auction Ending Alerts"  
**Description:** "Sends a Pushover alert each day for auction lots you are bidding on that end today."  
**Defaults:** `false`, `"08:00"`, `"1440"`

## Files Touched

1. `src/web/src/components/admin/AdminSchedulesSection.vue` — add panel and props
2. `src/web/src/composables/useAdminConfig.ts` — add state, defaults, clear/set in save, expose in return
3. `src/web/src/pages/AdminPage.vue` — destructure and pass props

## Notes

- The `AppSettings` interface in `src/web/src/types/index.ts` has an index signature `[key: string]: string`, so new keys don't require type changes — they're dynamically accessible.
- If the backend hasn't implemented the scheduler yet, the settings will still save/load — they just won't have any effect until the backend reads them.
- Always check for backend decision docs in `.squad/decisions/inbox/cassius-*.md` to confirm the exact setting key names before implementing.
