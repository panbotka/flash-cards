import { useQuery } from '@tanstack/react-query'
import { getCards } from '../api/client.ts'
import type { Card } from '../api/client.ts'

interface TagFilterProps {
  selectedTag: string | null
  onSelectTag: (tag: string | null) => void
}

export function TagFilter({ selectedTag, onSelectTag }: TagFilterProps) {
  const { data: cards } = useQuery<Card[]>({
    queryKey: ['cards'],
    queryFn: () => getCards(),
  })

  const allTags = cards
    ? [...new Set(cards.flatMap((c) => c.tags ?? []))].sort()
    : []

  if (allTags.length === 0) return null

  return (
    <div className="flex gap-2 overflow-x-auto py-1 scrollbar-none">
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
      {allTags.map((tag) => (
        <button
          key={tag}
          onClick={() => onSelectTag(tag === selectedTag ? null : tag)}
          className={`shrink-0 rounded-full px-3 py-1 text-xs font-medium transition-colors ${
            selectedTag === tag
              ? 'bg-accent text-white'
              : 'bg-[#1a1a1a] text-text-secondary border border-[#2a2a2a] hover:bg-[#222]'
          }`}
        >
          {tag}
        </button>
      ))}
    </div>
  )
}
