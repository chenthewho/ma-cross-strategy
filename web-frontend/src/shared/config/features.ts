export type AppFeature = 'dashboard' | 'strategies' | 'agents' | 'risk' | 'backtesting' | 'settings' | 'evolution'

const defaultFeatures: AppFeature[] = ['dashboard', 'strategies', 'agents', 'backtesting', 'settings', 'evolution']

export function hasFeature(feature: AppFeature): boolean {
  return defaultFeatures.includes(feature)
}
