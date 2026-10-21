export type RichSegment = string | { strong: string }
export type RichText = RichSegment | readonly RichSegment[]

export type LegalBlock =
  | { type: 'p'; text: RichText }
  | { type: 'list'; items: readonly RichText[] }
  | { type: 'table'; columns: readonly string[]; rows: readonly (readonly RichText[])[] }

export interface LegalSubsection {
  heading: string
  blocks: readonly LegalBlock[]
}

export interface LegalSection {
  id: string
  heading: string
  blocks?: readonly LegalBlock[]
  subsections?: readonly LegalSubsection[]
}

export interface LegalDoc {
  title: string
  updatedAt: string
  effectiveAt: string
  preamble: readonly LegalBlock[]
  sections: readonly LegalSection[]
}

export const b = (text: string): RichSegment => ({ strong: text })

export const p = (...segments: RichSegment[]): LegalBlock => ({ type: 'p', text: segments })

export const list = (...items: RichText[]): LegalBlock => ({ type: 'list', items })

export const table = (columns: string[], rows: RichText[][]): LegalBlock => ({
  type: 'table',
  columns,
  rows,
})

export function segmentsOf(text: RichText): readonly RichSegment[] {
  return Array.isArray(text) ? text : [text as RichSegment]
}
