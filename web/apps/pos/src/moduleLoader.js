import { shallowRef, ref } from 'vue'

// Load the business modules currently supported by the POS shell.
const moduleLoaders = {
  retail: () => import('../../../modules/retail/src/index.js'),
  pharmacy: () => import('../../../modules/pharmacy/src/index.js'),
  hardware: () => import('../../../modules/hardware/src/index.js'),
  restaurant: () => import('../../../modules/restaurant/src/index.js'),
}

const slotRegistry = ref({}) // { [slotName]: [{ moduleId, component }] }
const loadedModules = shallowRef([])

export async function loadEnabledModules(router, i18n) {
  const loaded = []
  const newSlots = {}

  for (const [moduleId, importFn] of Object.entries(moduleLoaders)) {
    try {
      const moduleExport = await importFn()
      const mod = moduleExport.default || moduleExport
      loaded.push(mod)

      // Register routes if router is provided
      if (router && Array.isArray(mod.routes)) {
        for (const route of mod.routes) {
          if (!router.hasRoute(route.name)) {
            router.addRoute('workspace', route)
          }
        }
      }

      // Register slots
      if (mod.slots && typeof mod.slots === 'object') {
        for (const [slotName, slotComponent] of Object.entries(mod.slots)) {
          if (!newSlots[slotName]) newSlots[slotName] = []
          newSlots[slotName].push({
            moduleId: mod.id || moduleId,
            component: slotComponent,
          })
        }
      }

      // Merge i18n messages
      if (i18n && mod.i18n) {
        if (mod.i18n.en) {
          i18n.global.mergeLocaleMessage('en', { [mod.id || moduleId]: mod.i18n.en, ...mod.i18n.en })
        }
        if (mod.i18n.sw) {
          i18n.global.mergeLocaleMessage('sw', { [mod.id || moduleId]: mod.i18n.sw, ...mod.i18n.sw })
        }
      }
    } catch (err) {
      console.warn(`[ModuleLoader] Failed to load module ${moduleId}:`, err)
    }
  }

  slotRegistry.value = newSlots
  loadedModules.value = loaded
  return loaded
}

export function getSlotContributions(slotName, activeModuleId = null) {
  const contributions = slotRegistry.value[slotName] || []
  if (!activeModuleId) return contributions
  return contributions.filter((c) => c.moduleId === activeModuleId)
}

export function useLoadedModules() {
  return {
    modules: loadedModules,
    slots: slotRegistry,
    getSlotContributions,
  }
}
