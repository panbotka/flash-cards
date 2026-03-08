import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getCardHistory } from '../api/client'
import type { ReviewEvent, SRSState } from '../api/client'

const ratingConfig: Record<number, { label: string; color: string }> = {
  2: { label: 'Hard', color: 'text-[#ff3b30]' },
  3: { label: 'Good', color: 'text-[#ffd60a]' },
  4: { label: 'Easy', color: 'text-[#30d158]' },
}

const directionLabel: Record<string, string> = {
  cz_en: 'CZ \u2192 EN',
  en_cz: 'EN \u2192 CZ',
}

function formatDate(iso: string) {
  const d = new Date(iso)
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

function formatTime(iso: string) {
  const d = new Date(iso)
  return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}

function formatInterval(days: number): string {
  if (days < 1) {
    const mins = Math.round(days * 24 * 60)
    return mins <= 0 ? '<1m' : `${mins}m`
  }
  if (days < 30) return `${Math.round(days)}d`
  if (days < 365) return `${Math.round(days / 30)}mo`
  return `${(days / 365).toFixed(1)}y`
}

function SRSStateSummary({ state }: { state: SRSState }) {
  const statusColors: Record<string, string> = {
    new: 'bg-blue-500/20 text-blue-400',
    learning: 'bg-orange-500/20 text-orange-400',
    review: 'bg-green-500/20 text-green-400',
  }
  const nextReview = new Date(state.nextReview)
  const now = new Date()
  const isDue = nextReview <= now

  return (
    <div className="rounded-xl bg-[#1a1a1a] border border-[#2a2a2a] p-4">
      <div className="flex items-center gap-2 mb-3">
        <span className="text-sm font-medium text-white">{directionLabel[state.direction] ?? state.direction}</span>
        <span className={`rounded-full px-2 py-0.5 text-[10px] font-medium ${statusColors[state.status] ?? 'text-text-secondary'}`}>
          {state.status.charAt(0).toUpperCase() + state.status.slice(1)}
        </span>
      </div>
      <div className="grid grid-cols-3 gap-3 text-xs">
        <div>
          <p className="text-text-tertiary mb-0.5">Interval</p>
          <p className="text-white font-medium">{formatInterval(state.intervalDays)}</p>
        </div>
        <div>
          <p className="text-text-tertiary mb-0.5">Ease</p>
          <p className="text-white font-medium">{(state.easeFactor * 100).toFixed(0)}%</p>
        </div>
        <div>
          <p className="text-text-tertiary mb-0.5">Next review</p>
          <p className={`font-medium ${isDue ? 'text-[#ffd60a]' : 'text-white'}`}>
            {state.status === 'new' ? '—' : isDue ? 'Due' : formatDate(state.nextReview)}
          </p>
        </div>
      </div>
    </div>
  )
}

function ReviewRow({ review }: { review: ReviewEvent }) {
  const cfg = ratingConfig[review.rating] ?? { label: `${review.rating}`, color: 'text-text-secondary' }

  return (
    <div className="flex items-center gap-3 py-2.5 border-b border-[#1a1a1a] last:border-0">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className={`text-sm font-medium ${cfg.color}`}>{cfg.label}</span>
          <span className="text-[10px] text-text-tertiary rounded-full bg-[#2a2a2a] px-2 py-0.5">
            {directionLabel[review.direction] ?? review.direction}
          </span>
        </div>
      </div>
      <div className="text-right shrink-0">
        <p className="text-xs text-text-secondary">{formatDate(review.reviewedAt)}</p>
        <p className="text-[10px] text-text-tertiary">{formatTime(review.reviewedAt)}</p>
      </div>
    </div>
  )
}

export function CardHistoryPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const cardId = Number(id)

  const { data, isLoading } = useQuery({
    queryKey: ['cardHistory', cardId],
    queryFn: () => getCardHistory(cardId),
    enabled: !isNaN(cardId),
  })

  return (
    <div className="fixed inset-0 flex flex-col bg-[#0a0a0a]">
      {/* Header */}
      <div className="shrink-0 px-4" style={{ paddingTop: 'max(env(safe-area-inset-top, 0px), 1rem)' }}>
        <div className="max-w-lg mx-auto w-full">
          <button
            onClick={() => navigate('/cards')}
            className="flex items-center gap-1 text-accent text-sm py-2 -ml-1"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="15 18 9 12 15 6" />
            </svg>
            Cards
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto px-4 pb-24">
        <div className="max-w-lg mx-auto w-full">
          {isLoading ? (
            <div className="flex items-center justify-center py-20">
              <div className="h-6 w-6 animate-spin rounded-full border-2 border-accent border-t-transparent" />
            </div>
          ) : !data ? (
            <p className="text-text-tertiary text-center py-20">Card not found</p>
          ) : (
            <>
              {/* Card info */}
              <div className="mb-6">
                <h1 className="text-xl font-bold text-white">{data.card.czech}</h1>
                <p className="text-base text-text-secondary mt-1">{data.card.english}</p>
                {data.card.tags && data.card.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1.5 mt-3">
                    {data.card.tags.map((tag) => (
                      <span
                        key={tag}
                        className="rounded-full bg-[#2a2a2a] px-2.5 py-0.5 text-[11px] font-medium text-text-secondary"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>
                )}
              </div>

              {/* SRS State Summary */}
              <h2 className="text-xs uppercase tracking-wider text-[#6e6e73] mb-2">SRS Status</h2>
              <div className="space-y-2 mb-6">
                {(data.card.srsStates ?? []).map((state) => (
                  <SRSStateSummary key={state.id} state={state} />
                ))}
              </div>

              {/* Review History */}
              <h2 className="text-xs uppercase tracking-wider text-[#6e6e73] mb-2">Review History</h2>
              {data.reviews.length === 0 ? (
                <div className="rounded-xl bg-[#1a1a1a] border border-[#2a2a2a] p-8 text-center">
                  <p className="text-sm text-text-tertiary">No reviews yet</p>
                </div>
              ) : (
                <div className="rounded-xl bg-[#1a1a1a] border border-[#2a2a2a] px-4">
                  {data.reviews.map((review) => (
                    <ReviewRow key={review.id} review={review} />
                  ))}
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  )
}
