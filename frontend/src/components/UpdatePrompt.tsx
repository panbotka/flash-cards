import { useState } from 'react'
import { useRegisterSW } from 'virtual:pwa-register/react'

export function UpdatePrompt() {
  const [dismissed, setDismissed] = useState(false)

  const {
    needRefresh: [needRefresh],
    updateServiceWorker,
  } = useRegisterSW()

  if (!needRefresh || dismissed) return null

  return (
    <div
      style={{ paddingTop: 'max(env(safe-area-inset-top, 0px), 8px)' }}
      className="fixed top-0 left-0 right-0 z-[100] flex items-center justify-between gap-3 px-4 pb-2 bg-[#1a1a1a] border-b border-[#2a2a2a] shadow-lg"
    >
      <span className="text-sm text-white/80">New version available</span>
      <div className="flex items-center gap-2">
        <button
          onClick={() => updateServiceWorker(true)}
          className="px-3 py-1 text-sm font-medium text-white bg-[#5e9eff] rounded-md active:opacity-80"
        >
          Reload
        </button>
        <button
          onClick={() => setDismissed(true)}
          className="p-1 text-white/50 active:text-white/80"
          aria-label="Dismiss"
        >
          ✕
        </button>
      </div>
    </div>
  )
}
