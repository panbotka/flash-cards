import { useState, useMemo } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { FlashCard } from '../components/FlashCard'
import { RatingButtons } from '../components/RatingButtons'
import { useStudySession } from '../hooks/useStudySession'
import { useKeyboardShortcuts } from '../hooks/useKeyboardShortcuts'
import { useSwipeRating } from '../hooks/useSwipeRating'
import { getTags, getStatsSummary, getDailyGoal, setDailyGoal } from '../api/client'

function ProgressRing({ reviewed, goal, onClick }: { reviewed: number; goal: number; onClick: () => void }) {
  const size = 44
  const stroke = 4
  const radius = (size - stroke) / 2
  const circumference = 2 * Math.PI * radius
  const progress = Math.min(reviewed / goal, 1)
  const offset = circumference * (1 - progress)
  const completed = reviewed >= goal

  return (
    <button
      onClick={onClick}
      className="relative flex items-center gap-2 shrink-0"
      title="Edit daily goal"
    >
      <svg width={size} height={size} className="-rotate-90">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="#2a2a2a"
          strokeWidth={stroke}
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke={completed ? '#30d158' : '#5e9eff'}
          strokeWidth={stroke}
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          strokeLinecap="round"
          className="transition-all duration-300"
        />
      </svg>
      <span className="absolute inset-0 flex items-center justify-center" style={{ width: size, height: size }}>
        {completed
          ? <span className="text-[#30d158] text-sm">&#10003;</span>
          : <span className="text-[#a1a1a6] text-[10px] font-medium">{reviewed}</span>
        }
      </span>
      <span className="text-[10px] text-[#6e6e73]">{reviewed}/{goal}</span>
    </button>
  )
}

function GoalEditor({ current, onSave, onClose }: { current: number; onSave: (g: number) => void; onClose: () => void }) {
  const [value, setValue] = useState(String(current || ''))

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
      <div className="bg-[#1a1a1a] border border-[#2a2a2a] rounded-2xl p-6 w-72" onClick={(e) => e.stopPropagation()}>
        <h3 className="text-white text-sm font-semibold mb-4">Daily review goal</h3>
        <input
          type="number"
          min="0"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder="0 = disabled"
          autoFocus
          className="w-full bg-[#0a0a0a] border border-[#2a2a2a] text-white text-sm rounded-lg px-3 py-2 outline-none focus:border-[#5e9eff] mb-4"
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              onSave(Math.max(0, parseInt(value) || 0))
            }
          }}
        />
        <div className="flex gap-2">
          <button
            onClick={onClose}
            className="flex-1 py-2 text-xs rounded-lg bg-[#2a2a2a] text-[#a1a1a6] hover:bg-[#333]"
          >
            Cancel
          </button>
          <button
            onClick={() => onSave(Math.max(0, parseInt(value) || 0))}
            className="flex-1 py-2 text-xs rounded-lg bg-[#5e9eff] text-white hover:bg-[#4a8af0]"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  )
}

function StudyContent({ direction, tag, cram }: { direction: 'cz_en' | 'en_cz'; tag?: string; cram: boolean }) {
  const { card, flipped, flip, rate, isDone, newAvailable, showNewCards, isLoading, isRating, canUndo, undo } =
    useStudySession(tag, direction, cram)

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

          {/* Undo button */}
          {canUndo && (
            <button
              onClick={undo}
              className="mt-4 text-xs text-[#6e6e73] hover:text-[#a1a1a6] transition-colors"
            >
              Undo last rating
            </button>
          )}

          {/* Keyboard shortcut hint */}
          <div className="mt-4 text-xs text-[#6e6e73] hidden sm:block">
            <span className="mr-4">Space = flip</span>
            <span>1-3 = rate</span>
          </div>

          {/* Mobile swipe hint */}
          <div className="mt-4 text-xs text-[#6e6e73] sm:hidden text-center">
            Swipe to rate: &#8592; Hard · &#8593; Good · &#8594; Easy
          </div>
        </>
      )}

      {/* Done state */}
      {isDone && !isLoading && (
        <div className="flex flex-col items-center justify-center flex-1 mt-20 text-center">
          <div className="text-5xl mb-6">&#10003;</div>
          <h2 className="text-2xl font-semibold text-white mb-2">
            {cram ? 'All cards reviewed!' : 'All caught up!'}
          </h2>
          <p className="text-[#a1a1a6] mb-8">
            {cram
              ? 'You\'ve gone through all cards in this cram session.'
              : newAvailable > 0
                ? `${newAvailable} new card${newAvailable === 1 ? '' : 's'} available to learn.`
                : 'No more cards to review today.'}
          </p>
          <div className="flex flex-col gap-3 w-full max-w-xs">
            {!cram && newAvailable > 0 && (
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
  const [cramMode, setCramMode] = useState(false)
  const [showGoalEditor, setShowGoalEditor] = useState(false)
  const queryClient = useQueryClient()

  const { data: tags } = useQuery({
    queryKey: ['tags'],
    queryFn: getTags,
  })

  const { data: summary } = useQuery({
    queryKey: ['stats', 'summary'],
    queryFn: getStatsSummary,
    refetchInterval: 30000,
  })

  const { data: goalData } = useQuery({
    queryKey: ['settings', 'daily-goal'],
    queryFn: getDailyGoal,
  })

  const goal = goalData?.goal ?? 0
  const reviewsToday = summary?.reviewsToday ?? 0

  const handleSaveGoal = async (newGoal: number) => {
    await setDailyGoal(newGoal)
    queryClient.invalidateQueries({ queryKey: ['settings', 'daily-goal'] })
    setShowGoalEditor(false)
  }

  return (
    <div className="fixed inset-0 flex flex-col bg-[#0a0a0a] px-4 pb-20 overflow-hidden" style={{ paddingTop: 'max(env(safe-area-inset-top, 0px), 0.75rem)' }}>
      {/* Cram mode banner */}
      {cramMode && (
        <div className="w-full max-w-md mx-auto mb-2 shrink-0 text-center">
          <span className="inline-block text-xs font-medium px-3 py-1 rounded-full bg-[#ff9f0a]/15 text-[#ff9f0a] border border-[#ff9f0a]/30">
            Cram Mode — SRS not affected
          </span>
        </div>
      )}

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
            CZ&#8594;EN
          </button>
          <button
            onClick={() => setDirection('en_cz')}
            className={`px-3 py-1.5 text-xs font-medium transition-colors ${
              direction === 'en_cz'
                ? 'bg-[#5e9eff] text-white'
                : 'bg-[#1a1a1a] text-[#a1a1a6] hover:bg-[#222]'
            }`}
          >
            EN&#8594;CZ
          </button>
        </div>

        <div className="flex items-center gap-2">
          {/* Daily goal progress */}
          {goal > 0 && (
            <ProgressRing reviewed={reviewsToday} goal={goal} onClick={() => setShowGoalEditor(true)} />
          )}

          {/* Goal setter (when no goal) */}
          {goal === 0 && (
            <button
              onClick={() => setShowGoalEditor(true)}
              className="p-1.5 text-[#6e6e73] hover:text-[#a1a1a6] transition-colors"
              title="Set daily goal"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="10" />
                <circle cx="12" cy="12" r="6" />
                <circle cx="12" cy="12" r="2" />
              </svg>
            </button>
          )}

          {/* Cram toggle */}
          <button
            onClick={() => setCramMode((v) => !v)}
            className={`px-3 py-1.5 text-xs font-medium rounded-lg border transition-colors ${
              cramMode
                ? 'bg-[#ff9f0a]/15 text-[#ff9f0a] border-[#ff9f0a]/30'
                : 'bg-[#1a1a1a] text-[#a1a1a6] border-[#2a2a2a] hover:bg-[#222]'
            }`}
          >
            Cram
          </button>

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
      </div>

      {/* Key resets flipped/showingNew state when direction or tag changes */}
      <div className="flex-1 min-h-0 flex flex-col items-center overflow-hidden">
        <StudyContent
          key={`${direction}-${selectedTag}-${cramMode}`}
          direction={direction}
          tag={selectedTag || undefined}
          cram={cramMode}
        />
      </div>

      {/* Goal editor modal */}
      {showGoalEditor && (
        <GoalEditor
          current={goal}
          onSave={handleSaveGoal}
          onClose={() => setShowGoalEditor(false)}
        />
      )}
    </div>
  )
}
