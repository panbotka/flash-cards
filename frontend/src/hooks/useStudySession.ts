import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getNextCard, getNewCard, submitReview, isStudyCard } from '../api/client'
import type { StudyCardResponse, StudyDoneResponse } from '../api/client'

export function useStudySession(tag?: string) {
  const queryClient = useQueryClient()
  const [flipped, setFlipped] = useState(false)
  const [revealed, setRevealed] = useState(false)
  const [showingNew, setShowingNew] = useState(false)

  const queryKey = ['study', 'next', { tag, showingNew }]

  const { data, isLoading } = useQuery<StudyCardResponse | StudyDoneResponse>({
    queryKey,
    queryFn: () =>
      showingNew
        ? getNewCard({ tag })
        : getNextCard({ tag }),
  })

  const reviewMutation = useMutation({
    mutationFn: ({ srsStateId, rating }: { srsStateId: number; rating: number }) =>
      submitReview(srsStateId, rating),
    onSuccess: () => {
      setFlipped(false)
      setRevealed(false)
      queryClient.invalidateQueries({ queryKey: ['study', 'next'] })
    },
  })

  const card = data && isStudyCard(data) ? data : null
  const doneData = data && !isStudyCard(data) ? data : null

  const flip = useCallback(() => {
    if (!card) return
    setFlipped((prev) => !prev)
    setRevealed(true)
  }, [card])

  const rate = useCallback(
    (rating: number) => {
      if (!card || !revealed || reviewMutation.isPending) return
      reviewMutation.mutate({ srsStateId: card.srsState.id, rating })
    },
    [card, revealed, reviewMutation],
  )

  const showNewCards = useCallback(() => {
    setShowingNew(true)
    setFlipped(false)
    setRevealed(false)
  }, [])

  return {
    card,
    flipped,
    revealed,
    flip,
    rate,
    isDone: !!doneData,
    newAvailable: doneData?.newAvailable ?? 0,
    showNewCards,
    isLoading,
    isRating: reviewMutation.isPending,
  }
}
