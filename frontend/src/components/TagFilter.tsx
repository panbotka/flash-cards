import { useState, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getTags, getCards, renameTag, deleteTagWithCards } from '../api/client.ts'

interface TagFilterProps {
  selectedTag: string | null
  onSelectTag: (tag: string | null) => void
}

/* ------------------------------------------------------------------ */
/*  Bottom Sheet                                                       */
/* ------------------------------------------------------------------ */

function BottomSheet({
  open,
  onClose,
  children,
}: {
  open: boolean
  onClose: () => void
  children: React.ReactNode
}) {
  if (!open) return null

  return (
    <div className="fixed inset-0 z-50">
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm animate-[fadeIn_200ms_ease-out]"
        onClick={onClose}
      />
      <div className="absolute bottom-0 left-0 right-0 rounded-t-2xl bg-[#1a1a1a] border-t border-[#2a2a2a] animate-[slideUp_300ms_ease-out]">
        <div className="mx-auto w-10 h-1 rounded-full bg-[#333] mt-2 mb-1" />
        {children}
      </div>
    </div>
  )
}

/* ------------------------------------------------------------------ */
/*  Tag Filter                                                         */
/* ------------------------------------------------------------------ */

export function TagFilter({ selectedTag, onSelectTag }: TagFilterProps) {
  const queryClient = useQueryClient()

  const { data: tags } = useQuery<string[]>({
    queryKey: ['tags'],
    queryFn: () => getTags(),
  })

  const [editMode, setEditMode] = useState(false)
  const [sheetTag, setSheetTag] = useState<string | null>(null)
  const [sheetView, setSheetView] = useState<'actions' | 'rename' | 'delete'>('actions')
  const [renameValue, setRenameValue] = useState('')
  const renameRef = useRef<HTMLInputElement>(null)

  // Count cards for delete confirmation
  const { data: tagCards } = useQuery({
    queryKey: ['cards', { tag: sheetTag }],
    queryFn: () => getCards({ tag: sheetTag! }),
    enabled: sheetView === 'delete' && sheetTag !== null,
  })

  const renameMut = useMutation({
    mutationFn: ({ oldName, newName }: { oldName: string; newName: string }) =>
      renameTag(oldName, newName),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] })
      queryClient.invalidateQueries({ queryKey: ['cards'] })
      if (selectedTag === sheetTag) {
        onSelectTag(renameValue.trim())
      }
      closeSheet()
    },
  })

  const deleteMut = useMutation({
    mutationFn: (tag: string) => deleteTagWithCards(tag),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] })
      queryClient.invalidateQueries({ queryKey: ['cards'] })
      if (selectedTag === sheetTag) {
        onSelectTag(null)
      }
      closeSheet()
      setEditMode(false)
    },
  })

  function openSheet(tag: string) {
    setSheetTag(tag)
    setSheetView('actions')
  }

  function closeSheet() {
    setSheetTag(null)
    setSheetView('actions')
    setRenameValue('')
  }

  function startRename() {
    setRenameValue(sheetTag ?? '')
    setSheetView('rename')
    setTimeout(() => renameRef.current?.focus(), 100)
  }

  function submitRename() {
    const trimmed = renameValue.trim()
    if (!trimmed || !sheetTag || trimmed === sheetTag) return
    renameMut.mutate({ oldName: sheetTag, newName: trimmed })
  }

  const allTags = tags ?? []
  if (allTags.length === 0 && !editMode) return null

  return (
    <>
      <div className="flex gap-2 overflow-x-auto py-1 scrollbar-none items-center">
        {!editMode && (
          <button
            onClick={() => onSelectTag(null)}
            className={`shrink-0 rounded-full px-3 py-1 text-xs font-medium transition-colors ${
              selectedTag === null
                ? 'bg-accent text-white'
                : 'bg-[#1a1a1a] text-text-secondary border border-[#2a2a2a] hover:bg-[#222]'
            }`}
          >
            All
          </button>
        )}

        {allTags.map((tag) => (
          <button
            key={tag}
            onClick={() => {
              if (editMode) {
                openSheet(tag)
              } else {
                onSelectTag(tag === selectedTag ? null : tag)
              }
            }}
            className={`shrink-0 rounded-full px-3 py-1 text-xs font-medium transition-all duration-200 active:scale-95 ${
              editMode
                ? 'bg-[#1a1a1a] text-text-secondary border border-accent/40 hover:bg-[#222]'
                : selectedTag === tag
                  ? 'bg-accent text-white'
                  : 'bg-[#1a1a1a] text-text-secondary border border-[#2a2a2a] hover:bg-[#222]'
            }`}
          >
            {tag}
          </button>
        ))}

        {/* Edit/Done toggle */}
        {allTags.length > 0 && (
          <button
            onClick={() => setEditMode(!editMode)}
            className="shrink-0 rounded-full px-2.5 py-1 text-xs font-medium text-text-tertiary hover:text-white transition-colors"
          >
            {editMode ? (
              'Done'
            ) : (
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 20h9" />
                <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
              </svg>
            )}
          </button>
        )}
      </div>

      {/* Action Sheet */}
      <BottomSheet open={sheetTag !== null && sheetView === 'actions'} onClose={closeSheet}>
        <div className="px-6 pt-4 pb-8">
          <p className="text-sm font-semibold text-white mb-4">{sheetTag}</p>
          <button
            onClick={startRename}
            className="w-full text-left px-4 py-4 text-sm text-white hover:bg-[#222] rounded-xl transition-colors active:scale-[0.98]"
          >
            Rename
          </button>
          <button
            onClick={() => setSheetView('delete')}
            className="w-full text-left px-4 py-4 text-sm text-red-400 hover:bg-[#222] rounded-xl transition-colors active:scale-[0.98]"
          >
            Delete tag & all cards
          </button>
          <button
            onClick={closeSheet}
            className="w-full mt-2 px-4 py-3 text-sm text-text-tertiary hover:bg-[#222] rounded-xl transition-colors text-center"
          >
            Cancel
          </button>
        </div>
      </BottomSheet>

      {/* Rename Sheet */}
      <BottomSheet open={sheetTag !== null && sheetView === 'rename'} onClose={closeSheet}>
        <div className="px-6 pt-4 pb-8">
          <p className="text-sm font-semibold text-white mb-4">Rename "{sheetTag}"</p>
          <input
            ref={renameRef}
            type="text"
            value={renameValue}
            onChange={(e) => setRenameValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') submitRename()
            }}
            className="w-full rounded-xl bg-[#0a0a0a] border border-[#2a2a2a] px-4 py-3 text-white placeholder-text-tertiary focus:outline-none focus:border-accent transition-colors mb-4"
            placeholder="New tag name"
          />
          <div className="flex gap-3">
            <button
              onClick={closeSheet}
              className="flex-1 rounded-xl bg-[#2a2a2a] py-3 text-sm font-medium text-text-secondary hover:bg-[#333] transition-colors active:scale-95"
            >
              Cancel
            </button>
            <button
              onClick={submitRename}
              disabled={renameMut.isPending || !renameValue.trim() || renameValue.trim() === sheetTag}
              className="flex-1 rounded-xl bg-accent py-3 text-sm font-medium text-white hover:bg-accent-hover transition-colors disabled:opacity-40 active:scale-95"
            >
              {renameMut.isPending ? 'Saving...' : 'Save'}
            </button>
          </div>
        </div>
      </BottomSheet>

      {/* Delete Confirmation Sheet */}
      <BottomSheet open={sheetTag !== null && sheetView === 'delete'} onClose={closeSheet}>
        <div className="px-6 pt-4 pb-8">
          <p className="text-sm font-semibold text-white mb-2">Delete "{sheetTag}"?</p>
          <p className="text-sm text-text-tertiary mb-6">
            All cards with this tag will be deleted.
          </p>
          <div className="flex gap-3">
            <button
              onClick={closeSheet}
              className="flex-1 rounded-xl bg-[#2a2a2a] py-3 text-sm font-medium text-text-secondary hover:bg-[#333] transition-colors active:scale-95"
            >
              Cancel
            </button>
            <button
              onClick={() => sheetTag && deleteMut.mutate(sheetTag)}
              disabled={deleteMut.isPending}
              className="flex-1 rounded-xl bg-red-500 py-3 text-sm font-medium text-white hover:bg-red-600 transition-colors active:scale-95 disabled:opacity-40"
            >
              {deleteMut.isPending
                ? 'Deleting...'
                : `Delete${tagCards ? ` · ${tagCards.length} card${tagCards.length !== 1 ? 's' : ''}` : ''}`}
            </button>
          </div>
        </div>
      </BottomSheet>
    </>
  )
}
