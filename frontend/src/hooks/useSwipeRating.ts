import { useRef, useCallback, useState, useEffect, type CSSProperties } from 'react'

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

const DIRECTION_THRESHOLD = 30
const COMMIT_THRESHOLD = 80

const directionConfig: Record<NonNullable<Direction>, { rating: number; label: string; color: string }> = {
  left: { rating: 2, label: 'Hard', color: '#ff3b30' },
  up: { rating: 3, label: 'Good', color: '#ffd60a' },
  right: { rating: 4, label: 'Easy', color: '#30d158' },
}

export function useSwipeRating({ onRate, enabled }: UseSwipeRatingOptions) {
  const [swipeStyle, setSwipeStyle] = useState<CSSProperties>({})
  const [swipeIndicator, setSwipeIndicator] = useState<SwipeIndicator | null>(null)
  const [el, setEl] = useState<HTMLDivElement | null>(null)

  const startX = useRef(0)
  const startY = useRef(0)
  const lockedDir = useRef<Direction>(null)
  const animating = useRef(false)
  const enabledRef = useRef(enabled)
  const onRateRef = useRef(onRate)

  // Keep refs in sync via effect so event handlers always see latest values
  useEffect(() => {
    enabledRef.current = enabled
    onRateRef.current = onRate
  })

  // Register native touch listeners (passive: false needed for iOS preventDefault)
  useEffect(() => {
    if (!el) return

    function handleTouchStart(e: TouchEvent) {
      if (!enabledRef.current || animating.current) return
      const touch = e.touches[0]
      startX.current = touch.clientX
      startY.current = touch.clientY
      lockedDir.current = null
      setSwipeStyle({ transition: 'none' })
      setSwipeIndicator(null)
    }

    function handleTouchMove(e: TouchEvent) {
      if (!enabledRef.current || animating.current) return

      // Prevent default immediately so iOS Safari doesn't claim the gesture
      // for scrolling or edge-swipe navigation
      e.preventDefault()

      const touch = e.touches[0]
      const dx = touch.clientX - startX.current
      const dy = touch.clientY - startY.current

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
    }

    function handleTouchEnd() {
      if (!enabledRef.current || !lockedDir.current) {
        setSwipeStyle({})
        setSwipeIndicator(null)
        return
      }

      const dir = lockedDir.current

      const transform = el!.style.transform
      let distance = 0
      if (dir === 'up') {
        const match = transform.match(/translate\([^,]+,\s*(-?[\d.]+)px/)
        distance = match ? Math.abs(parseFloat(match[1])) : 0
      } else {
        const match = transform.match(/translate\((-?[\d.]+)px/)
        distance = match ? Math.abs(parseFloat(match[1])) : 0
      }

      if (distance >= COMMIT_THRESHOLD) {
        animating.current = true
        const config = directionConfig[dir]
        let offX = 0
        let offY = 0
        if (dir === 'left') offX = -window.innerWidth
        else if (dir === 'right') offX = window.innerWidth
        else if (dir === 'up') offY = -window.innerHeight

        const rotation = dir !== 'up' ? (offX / window.innerWidth) * 15 : 0

        // Fly off and fade out
        setSwipeStyle({
          transition: 'transform 0.25s ease-out, opacity 0.25s ease-out',
          transform: `translate(${offX}px, ${offY}px) rotate(${rotation}deg)`,
          opacity: 0,
        })
        setSwipeIndicator((prev) => prev ? { ...prev, opacity: 1 } : null)

        setTimeout(() => {
          onRateRef.current(config.rating)
          // Start new card invisible, then fade in
          setSwipeStyle({ opacity: 0, transition: 'none' })
          setSwipeIndicator(null)
          requestAnimationFrame(() => {
            requestAnimationFrame(() => {
              setSwipeStyle({
                opacity: 1,
                transition: 'opacity 0.2s ease-in',
              })
              animating.current = false
            })
          })
        }, 250)
      } else {
        setSwipeStyle({
          transition: 'transform 0.25s ease-out',
          transform: 'translate(0px, 0px) rotate(0deg)',
        })
        setSwipeIndicator(null)
        lockedDir.current = null
      }
    }

    el.addEventListener('touchstart', handleTouchStart, { passive: true })
    el.addEventListener('touchmove', handleTouchMove, { passive: false })
    el.addEventListener('touchend', handleTouchEnd)

    return () => {
      el.removeEventListener('touchstart', handleTouchStart)
      el.removeEventListener('touchmove', handleTouchMove)
      el.removeEventListener('touchend', handleTouchEnd)
    }
  }, [el])

  const swipeRef = useCallback((node: HTMLDivElement | null) => {
    setEl(node)
  }, [])

  return { swipeRef, swipeStyle, swipeIndicator }
}
