import { computed, type Ref } from 'vue'

type DateRange = [string, string] | null

export function useDateRangeModel(start: Ref<string | undefined>, end: Ref<string | undefined>) {
  return computed<DateRange>({
    get: (): DateRange => (start.value && end.value ? [start.value, end.value] : null),
    set: (v: DateRange) => {
      start.value = v?.[0] ?? ''
      end.value = v?.[1] ?? ''
    },
  })
}
