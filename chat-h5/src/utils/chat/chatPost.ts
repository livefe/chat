export type ChatPostContent = {
  postId: number
  title: string
  cover: string
}

export function encodeChatPost(post: { postId: number; title?: string; cover?: string }): string {
  return JSON.stringify({
    postId: post.postId,
    title: (post.title || '').trim(),
    cover: (post.cover || '').trim(),
  })
}

export function parseChatPost(content: string | undefined): ChatPostContent | null {
  if (!content?.trim()) return null
  try {
    const raw = JSON.parse(content) as { postId?: unknown; title?: unknown; cover?: unknown }
    if (typeof raw.postId !== 'number' || !Number.isInteger(raw.postId) || raw.postId <= 0)
      return null
    return {
      postId: raw.postId,
      title: (typeof raw.title === 'string' ? raw.title.trim() : '') || '动态',
      cover: typeof raw.cover === 'string' ? raw.cover.trim() : '',
    }
  } catch {
    return null
  }
}

export function chatPostQuoteLabel(post: ChatPostContent | null): string {
  if (!post) return '[动态]'
  const title = post.title.trim()
  return title ? `[动态] ${title}` : '[动态]'
}
