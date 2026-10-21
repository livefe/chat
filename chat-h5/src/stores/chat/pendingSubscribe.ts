export function createPendingSubscribeWait(opts: { timeoutMs: number; timeoutMessage: string }) {
  interface Waiter {
    promise: Promise<void>
    resolveFn: () => void
    rejectFn: (err: Error) => void
    timer: ReturnType<typeof setTimeout>
  }

  const waiters = new Map<number, Waiter>()

  function clearWaiter(id: number) {
    const w = waiters.get(id)
    if (!w) return
    clearTimeout(w.timer)
    waiters.delete(id)
  }

  function resolve(id: number) {
    const w = waiters.get(id)
    if (!w) return
    clearWaiter(id)
    w.resolveFn()
  }

  function reject(id: number, err: Error) {
    const w = waiters.get(id)
    if (!w) return
    clearWaiter(id)
    w.rejectFn(err)
  }

  function rejectAll(err: Error) {
    for (const id of [...waiters.keys()]) reject(id, err)
  }

  function wait(id: number): Promise<void> {
    if (id <= 0) return Promise.resolve()
    const existing = waiters.get(id)
    if (existing) return existing.promise

    let resolveFn!: () => void
    let rejectFn!: (err: Error) => void
    const promise = new Promise<void>((res, rej) => {
      resolveFn = res
      rejectFn = rej
    })
    const timer = setTimeout(() => {
      reject(id, new Error(opts.timeoutMessage))
    }, opts.timeoutMs)
    waiters.set(id, { promise, resolveFn, rejectFn, timer })
    return promise
  }

  function matches(id: number): boolean {
    return waiters.has(id)
  }

  return { wait, resolve, reject, rejectAll, matches }
}
