import { useEffect } from 'react'

export function useKeyboardShortcuts(handlers: Record<string, () => void>) {
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      const target = e.target as HTMLElement
      if (
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.tagName === 'SELECT' ||
        target.isContentEditable
      ) {
        return
      }

      if (e.key === ' ' || e.code === 'Space') {
        e.preventDefault()
        handlers['space']?.()
      } else if (e.key >= '1' && e.key <= '4') {
        e.preventDefault()
        handlers[e.key]?.()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handlers])
}
