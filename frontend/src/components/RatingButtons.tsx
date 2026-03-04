interface RatingButtonsProps {
  intervalHints: Record<string, string>
  onRate: (rating: number) => void
  disabled: boolean
}

const ratings = [
  { rating: 1, key: 'again', label: 'Again', color: '#ff453a' },
  { rating: 2, key: 'hard', label: 'Hard', color: '#ff9f0a' },
  { rating: 3, key: 'good', label: 'Good', color: '#30d158' },
  { rating: 4, key: 'easy', label: 'Easy', color: '#5e9eff' },
] as const

export function RatingButtons({ intervalHints, onRate, disabled }: RatingButtonsProps) {
  return (
    <div className="grid grid-cols-4 gap-2 w-full max-w-md">
      {ratings.map(({ rating, key, label, color }) => (
        <button
          key={rating}
          onClick={() => onRate(rating)}
          disabled={disabled}
          className="flex flex-col items-center gap-0.5 rounded-xl py-3 px-2 transition-all duration-150 active:scale-95 disabled:opacity-40 disabled:pointer-events-none"
          style={{
            backgroundColor: `${color}10`,
            borderWidth: '1px',
            borderColor: `${color}40`,
          }}
        >
          <span className="text-sm font-medium" style={{ color }}>
            {label}
          </span>
          <span className="text-xs text-[#a1a1a6]">
            {intervalHints[key] ?? ''}
          </span>
        </button>
      ))}
    </div>
  )
}
