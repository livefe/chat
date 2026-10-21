export function isInAppShell(query: Record<string, unknown> | undefined | null): boolean {
  return String(query?.from || '') === 'app'
}
