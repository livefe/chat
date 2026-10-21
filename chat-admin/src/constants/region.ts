export const REGION_LEVEL_OPTIONS = [
  { label: '省级', value: 1 },
  { label: '市级', value: 2 },
  { label: '区县级', value: 3 },
]

export function regionLevelText(level: number) {
  return REGION_LEVEL_OPTIONS.find((item) => item.value === level)?.label ?? `L${level}`
}
