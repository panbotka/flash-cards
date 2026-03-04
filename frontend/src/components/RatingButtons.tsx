interface RatingButtonsProps {
  intervalHints: Record<string, string>
  onRate: (rating: number) => void
  disabled: boolean
}

const ratings = [
  { rating: 4, label: 'Easy', color: '#30d158' },
  { rating: 3, label: 'Good', color: '#ffd60a' },
  { rating: 2, label: 'Hard', color: '#ff3b30' },
] as const

export function RatingButtons({ intervalHints, onRate, disabled }: RatingButtonsProps) {
  return (
    <div className="grid grid-cols-3 gap-2 w-full max-w-md">
      {ratings.map(({ rating, label, color }) => (
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
            {intervalHints[String(rating)] ?? ''}
          </span>
        </button>
      ))}
    </div>
  )
}
