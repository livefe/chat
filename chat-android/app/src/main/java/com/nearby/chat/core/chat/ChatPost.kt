package com.nearby.chat.core.chat

import com.nearby.chat.R
import com.nearby.chat.core.ui.UiText
import org.json.JSONObject

data class ChatPostContent(
    val postId: Long,
    val title: String,
    val cover: String,
)

fun encodeChatPost(postId: Long, title: String = "", cover: String = ""): String = JSONObject()
    .put("postId", postId)
    .put("title", title.trim())
    .put("cover", cover.trim())
    .toString()

fun parseChatPost(content: String?): ChatPostContent? {
    if (content.isNullOrBlank()) return null
    return runCatching {
        val o = JSONObject(content)
        val postId = (o.opt("postId") as? Number)?.toLong() ?: 0L
        if (postId <= 0L) return null
        ChatPostContent(
            postId = postId,
            title = o.optString("title").trim(),
            cover = o.optString("cover").trim(),
        )
    }.getOrNull()
}

fun chatPostTitle(post: ChatPostContent): UiText = if (post.title.isBlank()) UiText.of(R.string.chat_post_default_title) else UiText.of(post.title)

fun chatPostQuoteLabel(post: ChatPostContent?): UiText {
    if (post == null) return UiText.of(R.string.msg_preview_post)
    return UiText.of(R.string.msg_preview_post_titled, chatPostTitle(post))
}
