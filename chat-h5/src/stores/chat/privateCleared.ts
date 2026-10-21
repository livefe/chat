import { ref } from 'vue'

export function createPrivateClearedHelpers() {
  const privateClearedBeforeId = ref<Record<number, number>>({})

  function setPrivateClearedBefore(peerId: number, clearedBeforeId: number) {
    if (clearedBeforeId <= 0) return
    privateClearedBeforeId.value = {
      ...privateClearedBeforeId.value,
      [peerId]: clearedBeforeId,
    }
  }

  function isClearedMessage(peerId: number, messageId: number): boolean {
    if (messageId <= 0) return false
    return messageId <= (privateClearedBeforeId.value[peerId] ?? 0)
  }

  function clearAllPrivateCleared() {
    privateClearedBeforeId.value = {}
  }

  return {
    setPrivateClearedBefore,
    isClearedMessage,
    clearAllPrivateCleared,
  }
}
