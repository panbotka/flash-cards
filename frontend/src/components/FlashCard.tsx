interface FlashCardProps {
  front: string
  back: string
  flipped: boolean
  onFlip: () => void
}

export function FlashCard({ front, back, flipped, onFlip }: FlashCardProps) {
  return (
    <div
      className="perspective-1000 w-full max-w-md min-h-[280px] cursor-pointer select-none"
      onClick={onFlip}
    >
      <div
        className={`preserve-3d relative w-full min-h-[280px] transition-transform duration-[600ms] ease-in-out ${
          flipped ? 'rotate-y-180' : ''
        }`}
      >
        {/* Front face */}
        <div className="backface-hidden absolute inset-0 flex flex-col items-center justify-center rounded-2xl bg-[#1a1a1a] border border-[#2a2a2a] px-8 py-10">
          <span className="text-3xl font-semibold text-white text-center leading-relaxed">
            {front}
          </span>
          <span className="absolute bottom-6 text-sm text-[#6e6e73]">
            Tap to flip
          </span>
        </div>

        {/* Back face */}
        <div className="backface-hidden rotate-y-180 absolute inset-0 flex flex-col items-center justify-center rounded-2xl bg-[#1a1a1a] border border-[#2a2a2a] px-8 py-10">
          <div className={`flex flex-col items-center transition-opacity duration-0 ${flipped ? 'opacity-100 delay-300' : 'opacity-0'}`}>
            <span className="text-3xl font-semibold text-white text-center leading-relaxed">
              {back}
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}
