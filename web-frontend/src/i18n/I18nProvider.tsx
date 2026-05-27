import { createContext, useContext, useState } from 'react'
import type { ReactNode } from 'react'
import zh from './zh.json'
import en from './en.json'

type Locale = 'zh' | 'en'
const messages: Record<Locale, Record<string, string>> = { zh, en }

interface I18nContextType {
  locale: Locale
  t: (key: string) => string
  setLocale: (l: Locale) => void
}

const I18nContext = createContext<I18nContextType>({ locale: 'zh', t: (k) => k, setLocale: () => {} })

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocale] = useState<Locale>('zh')
  const t = (key: string) => messages[locale][key] || key
  return <I18nContext.Provider value={{ locale, t, setLocale }}>{children}</I18nContext.Provider>
}

export function useI18n() {
  return useContext(I18nContext)
}
