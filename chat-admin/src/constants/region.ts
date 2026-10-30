export const REGION_LEVEL_OPTIONS = [
  { label: '省级', value: 1 },
  { label: '市级', value: 2 },
  { label: '区县级', value: 3 },
]

export const REGION_LEAF_LEVEL = 3

const REGION_LEVEL_TEXT: Record<number, string> = {
  1: '省级',
  2: '市级',
  3: '区县级',
  4: '乡镇级',
}

export function regionLevelText(level: number) {
  return REGION_LEVEL_TEXT[level] ?? `L${level}`
}
