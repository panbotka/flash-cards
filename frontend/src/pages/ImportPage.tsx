import { useState, useRef } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { importPreview, importCommit } from '../api/client.ts'
import type { ImportPreviewResponse, ImportPreviewCard } from '../api/client.ts'

/* ------------------------------------------------------------------ */
/*  Preview Card Row                                                   */
/* ------------------------------------------------------------------ */

function PreviewRow({ card }: { card: ImportPreviewCard }) {
  return (
    <div className="flex items-center gap-3 rounded-xl bg-[#1a1a1a] border border-[#2a2a2a] p-3">
      {/* Status icon */}
      <div className="shrink-0">
        {card.isDuplicate ? (
          <span className="flex h-5 w-5 items-center justify-center rounded-full bg-orange-500/20">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#ff9f0a" strokeWidth="3" strokeLinecap="round">
              <line x1="12" y1="5" x2="12" y2="14" />
              <circle cx="12" cy="18" r="0.5" fill="#ff9f0a" />
            </svg>
          </span>
        ) : (
          <span className="flex h-5 w-5 items-center justify-center rounded-full bg-green-500/20">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#30d158" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="20 6 9 17 4 12" />
            </svg>
          </span>
        )}
      </div>

      {/* Text */}
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-white truncate">{card.czech}</p>
        <p className="text-xs text-text-secondary truncate">{card.english}</p>
      </div>

      {/* Badge */}
      {card.isDuplicate && (
        <span className="shrink-0 rounded-full bg-orange-500/20 px-2 py-0.5 text-[10px] font-medium text-orange-400">
          Duplicate
        </span>
      )}
    </div>
  )
}

/* ------------------------------------------------------------------ */
/*  Import Page                                                        */
/* ------------------------------------------------------------------ */

export function ImportPage() {
  const queryClient = useQueryClient()
  const fileRef = useRef<HTMLInputElement>(null)

  const [content, setContent] = useState('')
  const [batchTags, setBatchTags] = useState('')
  const [preview, setPreview] = useState<ImportPreviewResponse | null>(null)
  const [importResult, setImportResult] = useState<{ imported: number; skipped: number } | null>(null)

  /* ---- Mutations ---- */

  const previewMut = useMutation({
    mutationFn: (text: string) => importPreview(text),
    onSuccess: (data) => {
      setPreview(data)
      setImportResult(null)
    },
  })

  const commitMut = useMutation({
    mutationFn: ({ cards, tags }: { cards: { czech: string; english: string }[]; tags?: string[] }) =>
      importCommit(cards, tags),
    onSuccess: (data) => {
      setImportResult(data)
      setPreview(null)
      setContent('')
      setBatchTags('')
      queryClient.invalidateQueries({ queryKey: ['cards'] })
    },
  })

  /* ---- Handlers ---- */

  function handleFileUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (ev) => {
      const text = ev.target?.result
      if (typeof text === 'string') {
        setContent(text)
        setPreview(null)
        setImportResult(null)
      }
    }
    reader.readAsText(file)

    // Reset input so the same file can be selected again
    e.target.value = ''
  }

  function handlePreview() {
    if (!content.trim()) return
    previewMut.mutate(content)
  }

  function handleImport() {
    if (!preview) return

    const newCards = preview.cards
      .filter((c) => !c.isDuplicate)
      .map((c) => ({ czech: c.czech, english: c.english }))

    if (newCards.length === 0) return

    const tags = batchTags
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean)

    commitMut.mutate({
      cards: newCards,
      tags: tags.length > 0 ? tags : undefined,
    })
  }

  const newCardCount = preview ? preview.cards.filter((c) => !c.isDuplicate).length : 0

  return (
    <div className="min-h-screen bg-[#0a0a0a] pb-20">
      <div className="mx-auto max-w-lg px-4 pt-6">
        {/* Header */}
        <h1 className="text-2xl font-bold text-white mb-2">Import</h1>
        <p className="text-sm text-text-secondary mb-5">
          Paste tab-separated or CSV content with Czech and English columns.
        </p>

        {/* Success banner */}
        {importResult && (
          <div className="mb-5 rounded-xl bg-green-500/10 border border-green-500/20 p-4">
            <div className="flex items-center gap-2">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#30d158" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                <polyline points="22 4 12 14.01 9 11.01" />
              </svg>
              <p className="text-sm font-medium text-green-400">
                Imported {importResult.imported} card{importResult.imported !== 1 ? 's' : ''}
                {importResult.skipped > 0 && (
                  <span className="text-text-secondary font-normal">
                    {' '}({importResult.skipped} skipped)
                  </span>
                )}
              </p>
            </div>
          </div>
        )}

        {/* Textarea */}
        <textarea
          value={content}
          onChange={(e) => {
            setContent(e.target.value)
            setPreview(null)
            setImportResult(null)
          }}
          placeholder={"pes\tdog\nkocka\tcat\njablko\tapple"}
          className="w-full min-h-[200px] rounded-xl bg-[#1a1a1a] border border-[#2a2a2a] px-4 py-3 text-sm text-white placeholder-text-tertiary font-mono resize-y focus:outline-none focus:border-accent transition-colors"
        />

        {/* File upload */}
        <input
          ref={fileRef}
          type="file"
          accept=".csv,.tsv,.txt"
          onChange={handleFileUpload}
          className="hidden"
        />
        <button
          onClick={() => fileRef.current?.click()}
          className="mt-3 flex w-full items-center justify-center gap-2 rounded-xl bg-[#1a1a1a] border border-[#2a2a2a] py-3 text-sm text-text-secondary hover:bg-[#222] transition-colors"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="17 8 12 3 7 8" />
            <line x1="12" y1="3" x2="12" y2="15" />
          </svg>
          Upload file (.csv, .tsv, .txt)
        </button>

        {/* Batch tags */}
        <div className="mt-4">
          <label className="block text-xs text-text-tertiary mb-1.5">
            Tags for all imported cards (optional, comma-separated)
          </label>
          <input
            type="text"
            value={batchTags}
            onChange={(e) => setBatchTags(e.target.value)}
            placeholder="lesson-1, verbs"
            className="w-full rounded-xl bg-[#1a1a1a] border border-[#2a2a2a] px-4 py-3 text-sm text-white placeholder-text-tertiary focus:outline-none focus:border-accent transition-colors"
          />
        </div>

        {/* Preview button */}
        {!preview && (
          <button
            onClick={handlePreview}
            disabled={!content.trim() || previewMut.isPending}
            className="mt-4 w-full rounded-xl bg-accent py-3.5 text-sm font-semibold text-white hover:bg-accent-hover transition-colors disabled:opacity-40"
          >
            {previewMut.isPending ? 'Parsing...' : 'Preview'}
          </button>
        )}

        {/* Error */}
        {previewMut.isError && (
          <p className="mt-3 text-sm text-red-400">
            Failed to parse: {previewMut.error instanceof Error ? previewMut.error.message : 'Unknown error'}
          </p>
        )}

        {commitMut.isError && (
          <p className="mt-3 text-sm text-red-400">
            Import failed: {commitMut.error instanceof Error ? commitMut.error.message : 'Unknown error'}
          </p>
        )}

        {/* Preview results */}
        {preview && (
          <div className="mt-5">
            {/* Summary */}
            <div className="mb-4 rounded-xl bg-[#1a1a1a] border border-[#2a2a2a] p-4">
              <p className="text-sm text-white">
                <span className="font-semibold">{preview.total}</span> card{preview.total !== 1 ? 's' : ''} found
                {preview.duplicates > 0 && (
                  <span className="text-orange-400">
                    , <span className="font-semibold">{preview.duplicates}</span> duplicate{preview.duplicates !== 1 ? 's' : ''} will be skipped
                  </span>
                )}
              </p>
            </div>

            {/* Card list */}
            <div className="space-y-2 mb-4">
              {preview.cards.map((card, i) => (
                <PreviewRow key={`${card.czech}-${card.english}-${i}`} card={card} />
              ))}
            </div>

            {/* Import button */}
            <div className="flex gap-3">
              <button
                onClick={() => {
                  setPreview(null)
                }}
                className="flex-1 rounded-xl bg-[#2a2a2a] py-3.5 text-sm font-medium text-text-secondary hover:bg-[#333] transition-colors"
              >
                Back
              </button>
              <button
                onClick={handleImport}
                disabled={newCardCount === 0 || commitMut.isPending}
                className="flex-1 rounded-xl bg-green-600 py-3.5 text-sm font-semibold text-white hover:bg-green-500 transition-colors disabled:opacity-40"
              >
                {commitMut.isPending
                  ? 'Importing...'
                  : `Import ${newCardCount} card${newCardCount !== 1 ? 's' : ''}`}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
