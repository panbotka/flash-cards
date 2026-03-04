import { useRef, useCallback, useState, type CSSProperties, type RefCallback } from 'react'

type Direction = 'left' | 'right' | 'up' | null

interface UseSwipeRatingOptions {
  onRate: (rating: number) => void
  enabled: boolean
}

interface SwipeIndicator {
  label: string
  color: string
  opacity: number
}

const DIRECTION_THRESHOLD = 50
const COMMIT_THRESHOLD = 100

const directionConfig: Record<NonNullable<Direction>, { rating: number; label: string; color: string }> = {
  left: { rating: 2, label: 'Hard', color: '#ff9f0a' },
  up: { rating: 3, label: 'Good', color: '#30d158' },
  right: { rating: 4, label: 'Easy', color: '#5e9eff' },
}

export function useSwipeRating({ onRate, enabled }: UseSwipeRatingOptions) {
  const [swipeStyle, setSwipeStyle] = useState<CSSProperties>({})
  const [swipeIndicator, setSwipeIndicator] = useState<SwipeIndicator | null>(null)

  const startX = useRef(0)
  const startY = useRef(0)
  const lockedDir = useRef<Direction>(null)
  const animating = useRef(false)
  const elRef = useRef<HTMLDivElement | null>(null)

  const handleTouchStart = useCallback(
    (e: TouchEvent) => {
      if (!enabled || animating.current) return
      const touch = e.touches[0]
      startX.current = touch.clientX
      startY.current = touch.clientY
      lockedDir.current = null
      setSwipeStyle({ transition: 'none' })
      setSwipeIndicator(null)
    },
    [enabled],
  )

  const handleTouchMove = useCallback(
    (e: TouchEvent) => {
      if (!enabled || animating.current) return
      const touch = e.touches[0]
      const dx = touch.clientX - startX.current
      const dy = touch.clientY - startY.current

      // Detect direction once past threshold
      if (!lockedDir.current) {
        const absDx = Math.abs(dx)
        const absDy = Math.abs(dy)
        if (absDx < DIRECTION_THRESHOLD && absDy < DIRECTION_THRESHOLD) {
          return
        }
        if (absDy > absDx && dy < 0) {
          lockedDir.current = 'up'
        } else if (absDx > absDy) {
          lockedDir.current = dx < 0 ? 'left' : 'right'
        } else {
          return // down swipe — ignore
        }
      }

      e.preventDefault()

      let translateX = 0
      let translateY = 0
      let distance = 0

      if (lockedDir.current === 'left' || lockedDir.current === 'right') {
        translateX = dx
        distance = Math.abs(dx)
      } else if (lockedDir.current === 'up') {
        translateY = Math.min(0, dy)
        distance = Math.abs(dy)
      }

      const rotation = lockedDir.current !== 'up' ? (translateX / window.innerWidth) * 15 : 0

      requestAnimationFrame(() => {
        setSwipeStyle({
          transition: 'none',
          transform: `translate(${translateX}px, ${translateY}px) rotate(${rotation}deg)`,
        })

        const config = directionConfig[lockedDir.current!]
        const opacity = Math.min(1, (distance - DIRECTION_THRESHOLD) / (COMMIT_THRESHOLD - DIRECTION_THRESHOLD))
        setSwipeIndicator({
          label: config.label,
          color: config.color,
          opacity: Math.max(0, opacity),
        })
      })
    },
    [enabled],
  )

  const handleTouchEnd = useCallback(() => {
    if (!enabled || !lockedDir.current) {
      setSwipeStyle({})
      setSwipeIndicator(null)
      return
    }

    const dir = lockedDir.current
    const el = elRef.current
    if (!el) return

    // Get current translation from style
    const transform = el.style.transform
    let distance = 0
    if (dir === 'up') {
      const match = transform.match(/translate\([^,]+,\s*(-?[\d.]+)px/)
      distance = match ? Math.abs(parseFloat(match[1])) : 0
    } else {
      const match = transform.match(/translate\((-?[\d.]+)px/)
      distance = match ? Math.abs(parseFloat(match[1])) : 0
    }

    if (distance >= COMMIT_THRESHOLD) {
      // Animate off-screen
      animating.current = true
      const config = directionConfig[dir]
      let offX = 0
      let offY = 0
      if (dir === 'left') offX = -window.innerWidth
      else if (dir === 'right') offX = window.innerWidth
      else if (dir === 'up') offY = -window.innerHeight

      const rotation = dir !== 'up' ? (offX / window.innerWidth) * 15 : 0

      setSwipeStyle({
        transition: 'transform 0.3s ease-out',
        transform: `translate(${offX}px, ${offY}px) rotate(${rotation}deg)`,
      })
      setSwipeIndicator((prev) => prev ? { ...prev, opacity: 1 } : null)

      setTimeout(() => {
        onRate(config.rating)
        animating.current = false
        setSwipeStyle({})
        setSwipeIndicator(null)
      }, 300)
    } else {
      // Snap back
      setSwipeStyle({
        transition: 'transform 0.25s ease-out',
        transform: 'translate(0px, 0px) rotate(0deg)',
      })
      setSwipeIndicator(null)
      lockedDir.current = null
    }
  }, [enabled, onRate])

  const swipeRef: RefCallback<HTMLDivElement> = useCallback(
    (node) => {
      // Cleanup previous
      if (elRef.current) {
        elRef.current.removeEventListener('touchstart', handleTouchStart)
        elRef.current.removeEventListener('touchmove', handleTouchMove)
        elRef.current.removeEventListener('touchend', handleTouchEnd)
      }
      elRef.current = node
      if (node) {
        node.addEventListener('touchstart', handleTouchStart, { passive: true })
        node.addEventListener('touchmove', handleTouchMove, { passive: false })
        node.addEventListener('touchend', handleTouchEnd)
      }
    },
    [handleTouchStart, handleTouchMove, handleTouchEnd],
  )

  return { swipeRef, swipeStyle, swipeIndicator }
}
