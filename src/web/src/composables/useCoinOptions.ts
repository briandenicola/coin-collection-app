import { ref } from 'vue'
import { getAppSettings } from '@/api/client'
import { CATEGORIES, MATERIALS, COIN_ERAS } from '@/types'
import { parseOptionList } from '@/utils/options'

export function useCoinOptions() {
  const settingsLoaded = ref(false)
  const categoryOptions = ref<string[]>([...CATEGORIES])
  const eraOptions = ref<string[]>([...COIN_ERAS])
  const materialOptions = ref<string[]>([...MATERIALS])

  async function loadOptions() {
    try {
      const res = await getAppSettings()
      const settings = res.data

      categoryOptions.value = parseOptionList(settings.CoinCategories, CATEGORIES)
      eraOptions.value = parseOptionList(settings.CoinEras, COIN_ERAS)
      materialOptions.value = [...MATERIALS]

      settingsLoaded.value = true
    } catch {
      // Use defaults on error
      categoryOptions.value = [...CATEGORIES]
      eraOptions.value = [...COIN_ERAS]
      materialOptions.value = [...MATERIALS]
      settingsLoaded.value = true
    }
  }

  return {
    settingsLoaded,
    categoryOptions,
    eraOptions,
    materialOptions,
    loadOptions,
  }
}
