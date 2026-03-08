const BASE = '/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (res.status === 401) {
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  if (!res.ok) {
    const body = await res.text()
    throw new Error(body || res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

// Auth
export function authCheck() {
  return request<{ authenticated: boolean; authRequired: boolean }>('/auth/check')
}
export function authLogin(password: string) {
  return request<{ ok: boolean }>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ password }),
  })
}
export function authLogout() {
  return request<{ ok: boolean }>('/auth/logout', { method: 'POST' })
}

// Cards
export interface Card {
  id: number
  czech: string
  english: string
  deletedAt?: string | null
  suspended: boolean
  createdAt: string
  updatedAt: string
  tags?: string[]
  srsStates?: SRSState[]
}

export interface SRSState {
  id: number
  cardId: number
  direction: string
  easeFactor: number
  intervalDays: number
  repetitions: number
  nextReview: string
  status: string
  learningStep: number
}

export interface StudyCardResponse {
  card: Card
  srsState: SRSState
  intervalHints: Record<string, string>
}

export interface StudyDoneResponse {
  done: boolean
  newAvailable: number
}

export interface ReviewResponse {
  srsState: SRSState
  nextInterval: string
}

export interface ImportPreviewCard {
  czech: string
  english: string
  isDuplicate: boolean
}

export interface ImportPreviewResponse {
  cards: ImportPreviewCard[]
  duplicates: number
  total: number
}

export interface StatsResponse {
  reviewsToday: number
  totalCards: number
  streak: number
  accuracyToday: number
}

export interface HeatmapEntry {
  date: string
  count: number
}

export interface AccuracyEntry {
  date: string
  accuracy: number
  total: number
}

export interface MaturityData {
  new: number
  learning: number
  young: number
  mature: number
}

export interface ForecastEntry {
  date: string
  count: number
}

export interface HardestCard {
  cardId: number
  czech: string
  english: string
  totalReviews: number
  againCount: number
  accuracy: number
}

export interface ReviewEvent {
  id: number
  srsStateId: number
  cardId: number
  direction: string
  rating: number
  reviewedAt: string
  intervalBefore?: number | null
  intervalAfter?: number | null
  easeBefore?: number | null
  easeAfter?: number | null
}

export interface CardHistoryResponse {
  card: Card
  reviews: ReviewEvent[]
}

export function getCardHistory(id: number) {
  return request<CardHistoryResponse>(`/cards/${id}/history`)
}

export function getTags() {
  return request<string[]>('/tags')
}

export function getCards(params?: { tag?: string; search?: string }) {
  const sp = new URLSearchParams()
  if (params?.tag) sp.set('tag', params.tag)
  if (params?.search) sp.set('search', params.search)
  const qs = sp.toString()
  return request<Card[]>(`/cards${qs ? `?${qs}` : ''}`)
}

export function getCard(id: number) {
  return request<Card>(`/cards/${id}`)
}

export function createCard(data: { czech: string; english: string; tags?: string[] }) {
  return request<Card>('/cards', { method: 'POST', body: JSON.stringify(data) })
}

export function updateCard(id: number, data: { czech?: string; english?: string; tags?: string[] }) {
  return request<Card>(`/cards/${id}`, { method: 'PUT', body: JSON.stringify(data) })
}

export function deleteCard(id: number) {
  return request<void>(`/cards/${id}`, { method: 'DELETE' })
}

export function suspendCard(id: number) {
  return request<Card>(`/cards/${id}/suspend`, { method: 'POST' })
}

export function restoreCard(id: number) {
  return request<Card>(`/cards/${id}/restore`, { method: 'POST' })
}

export function renameTag(oldName: string, newName: string) {
  return request<{ ok: boolean }>('/tags', {
    method: 'PUT',
    body: JSON.stringify({ oldName, newName }),
  })
}

export function exportCards() {
  window.location.href = `${BASE}/cards/export`
}

export function deleteTagWithCards(tag: string) {
  return request<{ deleted: number }>(`/tags/${encodeURIComponent(tag)}`, {
    method: 'DELETE',
  })
}

// Study
export function getNextCard(params?: { tag?: string; direction?: string; mode?: string; exclude?: number[] }) {
  const sp = new URLSearchParams()
  if (params?.tag) sp.set('tag', params.tag)
  if (params?.direction) sp.set('direction', params.direction)
  if (params?.mode) sp.set('mode', params.mode)
  if (params?.exclude && params.exclude.length > 0) sp.set('exclude', params.exclude.join(','))
  const qs = sp.toString()
  return request<StudyCardResponse | StudyDoneResponse>(`/study/next${qs ? `?${qs}` : ''}`)
}

export function getNewCard(params?: { tag?: string; direction?: string }) {
  const sp = new URLSearchParams()
  if (params?.tag) sp.set('tag', params.tag)
  if (params?.direction) sp.set('direction', params.direction)
  const qs = sp.toString()
  return request<StudyCardResponse | StudyDoneResponse>(`/study/new${qs ? `?${qs}` : ''}`)
}

export function submitReview(srsStateId: number, rating: number, cram?: boolean) {
  return request<ReviewResponse>('/study/review', {
    method: 'POST',
    body: JSON.stringify({ srsStateId, rating, ...(cram ? { cram: true } : {}) }),
  })
}

export function undoReview(params?: { tag?: string; direction?: string }) {
  const sp = new URLSearchParams()
  if (params?.tag) sp.set('tag', params.tag)
  if (params?.direction) sp.set('direction', params.direction)
  const qs = sp.toString()
  return request<StudyCardResponse>(`/study/undo${qs ? `?${qs}` : ''}`, {
    method: 'POST',
  })
}

// Import
export function importPreview(content: string) {
  return request<ImportPreviewResponse>('/cards/import/preview', {
    method: 'POST',
    body: JSON.stringify({ content }),
  })
}

export function importCommit(cards: { czech: string; english: string }[], tags?: string[]) {
  return request<{ imported: number; skipped: number }>('/cards/import/commit', {
    method: 'POST',
    body: JSON.stringify({ cards, tags }),
  })
}

// Stats
export function getStatsSummary() {
  return request<StatsResponse>('/stats/summary')
}
export function getStatsHeatmap() {
  return request<HeatmapEntry[]>('/stats/heatmap')
}
export function getStatsAccuracy() {
  return request<AccuracyEntry[]>('/stats/accuracy')
}
export function getStatsMaturity() {
  return request<MaturityData>('/stats/maturity')
}
export function getStatsForecast() {
  return request<ForecastEntry[]>('/stats/forecast')
}
export function getStatsHardest() {
  return request<HardestCard[]>('/stats/hardest')
}

export function isStudyCard(r: StudyCardResponse | StudyDoneResponse): r is StudyCardResponse {
  return 'card' in r
}
