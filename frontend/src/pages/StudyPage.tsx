import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { FlashCard } from '../components/FlashCard'
import { RatingButtons } from '../components/RatingButtons'
import { useStudySession } from '../hooks/useStudySession'
import { useKeyboardShortcuts } from '../hooks/useKeyboardShortcuts'
import { useSwipeRating } from '../hooks/useSwipeRating'
import { getTags } from '../api/client'

function StudyContent({ direction, tag }: { direction: 'cz_en' | 'en_cz'; tag?: string }) {
  const { card, flipped, flip, rate, isDone, newAvailable, showNewCards, isLoading, isRating } =
    useStudySession(tag, direction)

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
    enabled: !!card && !isRating,
  })

  const front = card
    ? direction === 'cz_en' ? card.card.czech : card.card.english
    : ''
  const back = card
    ? direction === 'cz_en' ? card.card.english : card.card.czech
    : ''

  return (
    <>
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

          <div ref={swipeRef} style={swipeStyle} className="relative w-full max-w-md touch-none">
            {swipeIndicator && swipeIndicator.opacity > 0 && (
              <div
                className="absolute inset-0 z-10 pointer-events-none rounded-2xl"
                style={{
                  backgroundColor: `${swipeIndicator.color}20`,
                  border: `2px solid ${swipeIndicator.color}`,
                  opacity: swipeIndicator.opacity,
                }}
              />
            )}
            <FlashCard
              front={front}
              back={back}
              flipped={flipped}
              onFlip={flip}
            />
          </div>

          {/* Rating buttons */}
          <div className="mt-6 w-full flex justify-center">
            <RatingButtons
              intervalHints={card.intervalHints}
              onRate={rate}
              disabled={isRating}
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
    </>
  )
}

export function StudyPage() {
  const [direction, setDirection] = useState<'cz_en' | 'en_cz'>('cz_en')
  const [selectedTag, setSelectedTag] = useState<string>('')

  const { data: tags } = useQuery({
    queryKey: ['tags'],
    queryFn: getTags,
  })

  return (
    <div className="fixed inset-0 flex flex-col bg-[#0a0a0a] px-4 pt-3 pb-20 overflow-hidden">
      {/* Control bar */}
      <div className="w-full max-w-md mx-auto flex items-center justify-between mb-4 shrink-0">
        {/* Direction toggle */}
        <div className="flex rounded-lg overflow-hidden border border-[#2a2a2a]">
          <button
            onClick={() => setDirection('cz_en')}
            className={`px-3 py-1.5 text-xs font-medium transition-colors ${
              direction === 'cz_en'
                ? 'bg-[#5e9eff] text-white'
                : 'bg-[#1a1a1a] text-[#a1a1a6] hover:bg-[#222]'
            }`}
          >
            CZ→EN
          </button>
          <button
            onClick={() => setDirection('en_cz')}
            className={`px-3 py-1.5 text-xs font-medium transition-colors ${
              direction === 'en_cz'
                ? 'bg-[#5e9eff] text-white'
                : 'bg-[#1a1a1a] text-[#a1a1a6] hover:bg-[#222]'
            }`}
          >
            EN→CZ
          </button>
        </div>

        {/* Tag filter */}
        <select
          value={selectedTag}
          onChange={(e) => setSelectedTag(e.target.value)}
          className="bg-[#1a1a1a] border border-[#2a2a2a] text-[#a1a1a6] text-xs rounded-lg px-3 py-1.5 outline-none focus:border-[#5e9eff] max-w-[140px]"
        >
          <option value="">All tags</option>
          {tags?.map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>
      </div>

      {/* Key resets flipped/showingNew state when direction or tag changes */}
      <div className="flex-1 min-h-0 flex flex-col items-center overflow-hidden">
        <StudyContent
          key={`${direction}-${selectedTag}`}
          direction={direction}
          tag={selectedTag || undefined}
        />
      </div>
    </div>
  )
}
