import { useState, useRef, useEffect, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getCards,
  createCard,
  updateCard,
  deleteCard,
  suspendCard,
  restoreCard,
} from '../api/client.ts'
import type { Card } from '../api/client.ts'
import { TagFilter } from '../components/TagFilter.tsx'

/* ------------------------------------------------------------------ */
/*  Helpers                                                            */
/* ------------------------------------------------------------------ */

function getWorstStatus(card: Card): 'new' | 'learning' | 'review' {
  const states = card.srsStates
  if (!states || states.length === 0) return 'new'

  const priority: Record<string, number> = { new: 0, learning: 1, review: 2 }
  let worst = 'review'
  for (const s of states) {
    const st = s.status.toLowerCase()
    if ((priority[st] ?? 3) < (priority[worst] ?? 3)) {
      worst = st
    }
  }
  return worst as 'new' | 'learning' | 'review'
}

const statusConfig = {
  new: { label: 'New', color: 'bg-blue-500/20 text-blue-400' },
  learning: { label: 'Learning', color: 'bg-orange-500/20 text-orange-400' },
  review: { label: 'Review', color: 'bg-green-500/20 text-green-400' },
} as const

/* ------------------------------------------------------------------ */
/*  Card Form Modal                                                    */
/* ------------------------------------------------------------------ */

interface CardFormData {
  czech: string
  english: string
  tags: string
}

function CardFormModal({
  card,
  onClose,
  onSave,
  saving,
}: {
  card: Card | null
  onClose: () => void
  onSave: (data: CardFormData) => void
  saving: boolean
}) {
  const [czech, setCzech] = useState(card?.czech ?? '')
  const [english, setEnglish] = useState(card?.english ?? '')
  const [tags, setTags] = useState(card?.tags?.join(', ') ?? '')
  const czechRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    czechRef.current?.focus()
  }, [])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!czech.trim() || !english.trim()) return
    onSave({ czech: czech.trim(), english: english.trim(), tags })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center sm:items-center">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
      <div className="relative w-full max-w-lg rounded-t-2xl sm:rounded-2xl bg-[#1a1a1a] border border-[#2a2a2a] p-6">
        <h2 className="text-lg font-semibold text-white mb-5">
          {card ? 'Edit Card' : 'Add Card'}
        </h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs text-text-tertiary mb-1.5">Czech</label>
            <input
              ref={czechRef}
              type="text"
              value={czech}
              onChange={(e) => setCzech(e.target.value)}
              className="w-full rounded-xl bg-[#0a0a0a] border border-[#2a2a2a] px-4 py-3 text-white placeholder-text-tertiary focus:outline-none focus:border-accent transition-colors"
              placeholder="Czech word or phrase"
            />
          </div>

          <div>
            <label className="block text-xs text-text-tertiary mb-1.5">English</label>
            <input
              type="text"
              value={english}
              onChange={(e) => setEnglish(e.target.value)}
              className="w-full rounded-xl bg-[#0a0a0a] border border-[#2a2a2a] px-4 py-3 text-white placeholder-text-tertiary focus:outline-none focus:border-accent transition-colors"
              placeholder="English translation"
            />
          </div>

          <div>
            <label className="block text-xs text-text-tertiary mb-1.5">Tags (comma-separated)</label>
            <input
              type="text"
              value={tags}
              onChange={(e) => setTags(e.target.value)}
              className="w-full rounded-xl bg-[#0a0a0a] border border-[#2a2a2a] px-4 py-3 text-white placeholder-text-tertiary focus:outline-none focus:border-accent transition-colors"
              placeholder="food, verbs, travel"
            />
          </div>

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 rounded-xl bg-[#2a2a2a] py-3 text-sm font-medium text-text-secondary hover:bg-[#333] transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving || !czech.trim() || !english.trim()}
              className="flex-1 rounded-xl bg-accent py-3 text-sm font-medium text-white hover:bg-accent-hover transition-colors disabled:opacity-40"
            >
              {saving ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

/* ------------------------------------------------------------------ */
/*  Action Menu                                                        */
/* ------------------------------------------------------------------ */

function ActionMenu({
  card,
  onEdit,
  onDelete,
  onSuspendToggle,
  onClose,
}: {
  card: Card
  onEdit: () => void
  onDelete: () => void
  onSuspendToggle: () => void
  onClose: () => void
}) {
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onClose()
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [onClose])

  return (
    <div
      ref={menuRef}
      className="absolute right-4 top-12 z-40 min-w-[160px] rounded-xl bg-[#2a2a2a] border border-[#333] shadow-xl overflow-hidden"
    >
      <button
        onClick={onEdit}
        className="w-full text-left px-4 py-2.5 text-sm text-white hover:bg-[#333] transition-colors"
      >
        Edit
      </button>
      <button
        onClick={onSuspendToggle}
        className="w-full text-left px-4 py-2.5 text-sm text-white hover:bg-[#333] transition-colors"
      >
        {card.suspended ? 'Unsuspend' : 'Suspend'}
      </button>
      <button
        onClick={onDelete}
        className="w-full text-left px-4 py-2.5 text-sm text-red-400 hover:bg-[#333] transition-colors"
      >
        Delete
      </button>
    </div>
  )
}

/* ------------------------------------------------------------------ */
/*  Card Row                                                           */
/* ------------------------------------------------------------------ */

function CardRow({
  card,
  onEdit,
  onDelete,
  onSuspendToggle,
}: {
  card: Card
  onEdit: () => void
  onDelete: () => void
  onSuspendToggle: () => void
}) {
  const [menuOpen, setMenuOpen] = useState(false)
  const status = getWorstStatus(card)
  const cfg = statusConfig[status]

  return (
    <div
      className={`relative rounded-xl bg-[#1a1a1a] border border-[#2a2a2a] mb-2 p-4 transition-colors ${
        card.suspended ? 'opacity-50' : ''
      }`}
    >
      <div className="flex items-start justify-between gap-3">
        {/* Left: text + tags */}
        <button
          type="button"
          className="flex-1 min-w-0 text-left"
          onClick={onEdit}
        >
          <p className="text-base font-medium text-white truncate">{card.czech}</p>
          <p className="text-sm text-text-secondary truncate mt-0.5">{card.english}</p>

          <div className="flex flex-wrap items-center gap-1.5 mt-2">
            {(card.tags ?? []).map((tag) => (
              <span
                key={tag}
                className="rounded-full bg-[#2a2a2a] px-2 py-0.5 text-[10px] font-medium text-text-secondary"
              >
                {tag}
              </span>
            ))}

            <span className={`rounded-full px-2 py-0.5 text-[10px] font-medium ${cfg.color}`}>
              {cfg.label}
            </span>

            {card.suspended && (
              <span className="rounded-full bg-red-500/15 px-2 py-0.5 text-[10px] font-medium text-red-400">
                Suspended
              </span>
            )}
          </div>
        </button>

        {/* Right: action trigger */}
        <button
          onClick={(e) => {
            e.stopPropagation()
            setMenuOpen(!menuOpen)
          }}
          className="shrink-0 rounded-lg p-1.5 hover:bg-[#2a2a2a] transition-colors text-text-tertiary"
          aria-label="Card actions"
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
            <circle cx="12" cy="5" r="2" />
            <circle cx="12" cy="12" r="2" />
            <circle cx="12" cy="19" r="2" />
          </svg>
        </button>
      </div>

      {menuOpen && (
        <ActionMenu
          card={card}
          onEdit={() => {
            setMenuOpen(false)
            onEdit()
          }}
          onDelete={() => {
            setMenuOpen(false)
            onDelete()
          }}
          onSuspendToggle={() => {
            setMenuOpen(false)
            onSuspendToggle()
          }}
          onClose={() => setMenuOpen(false)}
        />
      )}
    </div>
  )
}

/* ------------------------------------------------------------------ */
/*  Cards Page                                                         */
/* ------------------------------------------------------------------ */

export function CardsPage() {
  const queryClient = useQueryClient()

  const [search, setSearch] = useState('')
  const [selectedTag, setSelectedTag] = useState<string | null>(null)
  // undefined = modal closed, null = creating new card, Card = editing existing
  const [modalCard, setModalCard] = useState<Card | null | undefined>(undefined)

  const { data: cards, isLoading } = useQuery<Card[]>({
    queryKey: ['cards', { tag: selectedTag, search: search || undefined }],
    queryFn: () =>
      getCards({
        tag: selectedTag ?? undefined,
        search: search || undefined,
      }),
  })

  /* ---- Mutations ---- */

  const invalidateCards = useCallback(
    () => queryClient.invalidateQueries({ queryKey: ['cards'] }),
    [queryClient],
  )

  const createMut = useMutation({
    mutationFn: (data: { czech: string; english: string; tags?: string[] }) => createCard(data),
    onSuccess: invalidateCards,
  })

  const updateMut = useMutation({
    mutationFn: ({ id, data }: { id: number; data: { czech?: string; english?: string; tags?: string[] } }) =>
      updateCard(id, data),
    onSuccess: invalidateCards,
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => deleteCard(id),
    onSuccess: invalidateCards,
  })

  const suspendMut = useMutation({
    mutationFn: (id: number) => suspendCard(id),
    onSuccess: invalidateCards,
  })

  const restoreMut = useMutation({
    mutationFn: (id: number) => restoreCard(id),
    onSuccess: invalidateCards,
  })

  const saving = createMut.isPending || updateMut.isPending

  /* ---- Handlers ---- */

  function handleSave(data: CardFormData) {
    const parsedTags = data.tags
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean)

    if (modalCard) {
      updateMut.mutate(
        { id: modalCard.id, data: { czech: data.czech, english: data.english, tags: parsedTags } },
        { onSuccess: () => setModalCard(undefined) },
      )
    } else {
      createMut.mutate(
        { czech: data.czech, english: data.english, tags: parsedTags.length > 0 ? parsedTags : undefined },
        { onSuccess: () => setModalCard(undefined) },
      )
    }
  }

  function handleDelete(card: Card) {
    deleteMut.mutate(card.id)
  }

  function handleSuspendToggle(card: Card) {
    if (card.suspended) {
      restoreMut.mutate(card.id)
    } else {
      suspendMut.mutate(card.id)
    }
  }

  return (
    <div className="min-h-screen bg-[#0a0a0a] pb-20">
      <div className="mx-auto max-w-lg px-4 pt-6">
        {/* Header */}
        <h1 className="text-2xl font-bold text-white mb-5">Cards</h1>

        {/* Search */}
        <div className="relative mb-3">
          <svg
            className="absolute left-3.5 top-1/2 -translate-y-1/2 text-text-tertiary"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search cards..."
            className="w-full rounded-xl bg-[#1a1a1a] border border-[#2a2a2a] pl-10 pr-4 py-3 text-sm text-white placeholder-text-tertiary focus:outline-none focus:border-accent transition-colors"
          />
        </div>

        {/* Tag filter */}
        <div className="mb-4">
          <TagFilter selectedTag={selectedTag} onSelectTag={setSelectedTag} />
        </div>

        {/* Card count */}
        {cards && (
          <p className="text-xs text-text-tertiary mb-3">
            {cards.length} card{cards.length !== 1 ? 's' : ''}
          </p>
        )}

        {/* Card list */}
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-accent border-t-transparent" />
          </div>
        ) : cards && cards.length > 0 ? (
          <div>
            {cards.map((card) => (
              <CardRow
                key={card.id}
                card={card}
                onEdit={() => setModalCard(card)}
                onDelete={() => handleDelete(card)}
                onSuspendToggle={() => handleSuspendToggle(card)}
              />
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-text-tertiary">
            <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="mb-3 opacity-40">
              <rect x="3" y="3" width="18" height="18" rx="3" />
              <line x1="9" y1="9" x2="15" y2="15" />
              <line x1="15" y1="9" x2="9" y2="15" />
            </svg>
            <p className="text-sm">No cards found</p>
          </div>
        )}
      </div>

      {/* FAB - Add Card */}
      <button
        onClick={() => setModalCard(null)}
        className="fixed bottom-24 right-6 z-30 flex h-14 w-14 items-center justify-center rounded-full bg-accent shadow-lg shadow-accent/25 hover:bg-accent-hover transition-colors active:scale-95"
        aria-label="Add card"
      >
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round">
          <line x1="12" y1="5" x2="12" y2="19" />
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
      </button>

      {/* Modal */}
      {modalCard !== undefined && (
        <CardFormModal
          card={modalCard}
          onClose={() => setModalCard(undefined)}
          onSave={handleSave}
          saving={saving}
        />
      )}
    </div>
  )
}
