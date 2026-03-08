import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getNextCard, getNewCard, submitReview, undoReview, isStudyCard } from '../api/client'
import type { StudyCardResponse, StudyDoneResponse } from '../api/client'

export function useStudySession(tag?: string, direction: string = 'cz_en') {
  const queryClient = useQueryClient()
  const [flipped, setFlipped] = useState(false)
  const [showingNew, setShowingNew] = useState(false)
  const [canUndo, setCanUndo] = useState(false)
  const [undoneCard, setUndoneCard] = useState<StudyCardResponse | null>(null)

  const queryKey = ['study', 'next', { tag, direction, showingNew }]

  const { data, isLoading } = useQuery<StudyCardResponse | StudyDoneResponse>({
    queryKey,
    queryFn: () =>
      showingNew
        ? getNewCard({ tag, direction })
        : getNextCard({ tag, direction }),
    enabled: !undoneCard,
  })

  const reviewMutation = useMutation({
    mutationFn: ({ srsStateId, rating }: { srsStateId: number; rating: number }) =>
      submitReview(srsStateId, rating),
    onSuccess: () => {
      setFlipped(false)
      setUndoneCard(null)
      setCanUndo(true)
      queryClient.invalidateQueries({ queryKey: ['study', 'next'] })
    },
  })

  const undoMutation = useMutation({
    mutationFn: () => undoReview({ tag, direction }),
    onSuccess: (restoredCard) => {
      setCanUndo(false)
      setFlipped(true)
      setUndoneCard(restoredCard)
      queryClient.invalidateQueries({ queryKey: ['study', 'next'] })
    },
  })

  const activeData = undoneCard ?? data
  const card = activeData && isStudyCard(activeData) ? activeData : null
  const doneData = !undoneCard && data && !isStudyCard(data) ? data : null

  const flip = useCallback(() => {
    if (!card) return
    setFlipped((prev) => !prev)
  }, [card])

  const rate = useCallback(
    (rating: number) => {
      if (!card || reviewMutation.isPending) return
      reviewMutation.mutate({ srsStateId: card.srsState.id, rating })
    },
    [card, reviewMutation],
  )

  const undo = useCallback(() => {
    if (!canUndo || undoMutation.isPending) return
    undoMutation.mutate()
  }, [canUndo, undoMutation])

  const showNewCards = useCallback(() => {
    setShowingNew(true)
    setFlipped(false)
    setCanUndo(false)
    setUndoneCard(null)
  }, [])

  return {
    card,
    flipped,
    flip,
    rate,
    isDone: !!doneData,
    newAvailable: doneData?.newAvailable ?? 0,
    showNewCards,
    isLoading: isLoading && !undoneCard,
    isRating: reviewMutation.isPending,
    canUndo,
    undo,
    isUndoing: undoMutation.isPending,
  }
}
