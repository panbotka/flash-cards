import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getNextCard, getNewCard, submitReview, isStudyCard } from '../api/client'
import type { StudyCardResponse, StudyDoneResponse } from '../api/client'

export function useStudySession(tag?: string, direction?: string) {
  const queryClient = useQueryClient()
  const [flipped, setFlipped] = useState(false)
  const [showingNew, setShowingNew] = useState(false)

  const queryKey = ['study', 'next', { tag, direction, showingNew }]

  const { data, isLoading } = useQuery<StudyCardResponse | StudyDoneResponse>({
    queryKey,
    queryFn: () =>
      showingNew
        ? getNewCard({ tag, direction })
        : getNextCard({ tag, direction }),
  })

  const reviewMutation = useMutation({
    mutationFn: ({ srsStateId, rating }: { srsStateId: number; rating: number }) =>
      submitReview(srsStateId, rating),
    onSuccess: () => {
      setFlipped(false)
      queryClient.invalidateQueries({ queryKey: ['study', 'next'] })
    },
  })

  const card = data && isStudyCard(data) ? data : null
  const doneData = data && !isStudyCard(data) ? data : null

  const flip = useCallback(() => {
    if (card && !flipped) {
      setFlipped(true)
    }
  }, [card, flipped])

  const rate = useCallback(
    (rating: number) => {
      if (!card || !flipped || reviewMutation.isPending) return
      reviewMutation.mutate({ srsStateId: card.srsState.id, rating })
    },
    [card, flipped, reviewMutation],
  )

  const showNewCards = useCallback(() => {
    setShowingNew(true)
    setFlipped(false)
  }, [])

  return {
    card,
    flipped,
    flip,
    rate,
    isDone: !!doneData,
    newAvailable: doneData?.newAvailable ?? 0,
    showNewCards,
    isLoading,
    isRating: reviewMutation.isPending,
  }
}
