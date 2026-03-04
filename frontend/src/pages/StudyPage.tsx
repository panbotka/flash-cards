import { useMemo } from 'react'
import { FlashCard } from '../components/FlashCard'
import { RatingButtons } from '../components/RatingButtons'
import { useStudySession } from '../hooks/useStudySession'
import { useKeyboardShortcuts } from '../hooks/useKeyboardShortcuts'
import { useSwipeRating } from '../hooks/useSwipeRating'

export function StudyPage() {
  const { card, flipped, revealed, flip, rate, isDone, newAvailable, showNewCards, isLoading, isRating } =
    useStudySession()

  const handlers = useMemo(
    () => ({
      space: flip,
      '1': () => rate(2),
      '2': () => rate(3),
      '3': () => rate(4),
    }),
    [flip, rate],
  )

  useKeyboardShortcuts(handlers)

  const { swipeRef, swipeStyle, swipeIndicator } = useSwipeRating({
    onRate: rate,
    enabled: revealed && !isRating,
  })

  return (
    <div className="flex flex-col items-center min-h-screen bg-[#0a0a0a] px-4 pt-12 pb-20">
      {/* Loading skeleton */}
      {isLoading && (
        <div className="w-full max-w-md min-h-[280px] rounded-2xl bg-[#1a1a1a] border border-[#2a2a2a] animate-pulse" />
      )}

      {/* Study card */}
      {card && !isLoading && (
        <>
          {/* Tag chips */}
          {card.card.tags && card.card.tags.length > 0 && (
            <div className="flex flex-wrap gap-2 mb-6">
              {card.card.tags.map((t) => (
                <span
                  key={t}
                  className="text-xs px-3 py-1 rounded-full bg-[#1a1a1a] border border-[#2a2a2a] text-[#a1a1a6]"
                >
                  {t}
                </span>
              ))}
            </div>
          )}

          <div ref={swipeRef} style={swipeStyle} className="relative">
            {swipeIndicator && swipeIndicator.opacity > 0 && (
              <div
                className="absolute inset-0 z-10 flex items-center justify-center pointer-events-none rounded-2xl"
                style={{
                  backgroundColor: `${swipeIndicator.color}20`,
                  border: `2px solid ${swipeIndicator.color}`,
                  opacity: swipeIndicator.opacity,
                }}
              >
                <span
                  className="text-2xl font-bold"
                  style={{ color: swipeIndicator.color }}
                >
                  {swipeIndicator.label}
                </span>
              </div>
            )}
            <FlashCard
              front={card.card.czech}
              back={card.card.english}
              flipped={flipped}
              onFlip={flip}
            />
          </div>

          {/* Rating buttons with slide-up animation */}
          <div
            className={`mt-6 w-full flex justify-center transition-all duration-300 ease-out ${
              revealed
                ? 'opacity-100 translate-y-0'
                : 'opacity-0 translate-y-4 pointer-events-none'
            }`}
          >
            <RatingButtons
              intervalHints={card.intervalHints}
              onRate={rate}
              disabled={isRating || !revealed}
            />
          </div>

          {/* Keyboard shortcut hint */}
          <div className="mt-8 text-xs text-[#6e6e73] hidden sm:block">
            <span className="mr-4">Space = flip</span>
            <span>1-3 = rate</span>
          </div>

          {/* Mobile swipe hint */}
          <div className="mt-6 text-xs text-[#6e6e73] sm:hidden text-center">
            Swipe to rate: ← Hard · ↑ Good · → Easy
          </div>
        </>
      )}

      {/* Done state */}
      {isDone && !isLoading && (
        <div className="flex flex-col items-center justify-center flex-1 mt-20 text-center">
          <div className="text-5xl mb-6">&#10003;</div>
          <h2 className="text-2xl font-semibold text-white mb-2">All caught up!</h2>
          <p className="text-[#a1a1a6] mb-8">
            {newAvailable > 0
              ? `${newAvailable} new card${newAvailable === 1 ? '' : 's'} available to learn.`
              : 'No more cards to review today.'}
          </p>
          <div className="flex flex-col gap-3 w-full max-w-xs">
            {newAvailable > 0 && (
              <button
                onClick={showNewCards}
                className="w-full py-3 px-6 rounded-xl bg-[#5e9eff] text-white font-medium transition-all duration-150 active:scale-95 hover:bg-[#4a8af0]"
              >
                Continue with new cards
              </button>
            )}
            <button
              onClick={() => window.history.back()}
              className="w-full py-3 px-6 rounded-xl bg-[#1a1a1a] border border-[#2a2a2a] text-[#a1a1a6] font-medium transition-all duration-150 active:scale-95 hover:bg-[#222222]"
            >
              Done for today
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
