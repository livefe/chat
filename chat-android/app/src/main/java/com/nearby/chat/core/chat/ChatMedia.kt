package com.nearby.chat.core.chat

import com.nearby.chat.core.media.MediaUrl
import org.json.JSONObject

data class ChatMediaContent(
    val mediaId: Long,
    val url: String,
)

fun parseChatMedia(raw: String?): ChatMediaContent? {
    val t = raw?.trim().orEmpty()
    if (t.isEmpty() || !t.startsWith("{")) return null
    return runCatching {
        val o = JSONObject(t)
        val mediaId = (o.opt("mediaId") as? Number)?.toLong() ?: 0L
        val url = o.optString("url").trim()
        if (mediaId < 1L || url.isEmpty()) return null
        ChatMediaContent(mediaId = mediaId, url = url)
    }.getOrNull()
}

fun chatMediaDisplayUrl(raw: String?): String = parseChatMedia(raw)?.url?.let { MediaUrl.resolve(it) }.orEmpty()

fun encodeChatFile(mediaId: Long, name: String): String = JSONObject()
    .put("mediaId", mediaId)
    .put("name", name.trim().ifBlank { "file" })
    .toString()

fun chatFileDisplayName(raw: String?): String {
    val t = raw?.trim().orEmpty()
    if (t.isEmpty()) return ""
    return runCatching {
        val o = JSONObject(t)
        val name = o.optString("name").trim()
        if (name.isNotBlank()) return@runCatching name
        val url = o.optString("url").trim()
        if (url.isNotBlank()) {
            val last = url.substringAfterLast('/').substringBefore('?')
            if (last.isNotBlank()) return@runCatching last
        }
        ""
    }.getOrDefault("")
}

fun truncateFileDisplayName(name: String, maxLen: Int = 28): String {
    val n = name.trim()
    if (n.length <= maxLen) return n
    val dot = n.lastIndexOf('.')
    val ext = if (dot > 0 && n.length - dot <= 9) n.substring(dot) else ""
    val budget = (maxLen - ext.length - 1).coerceAtLeast(8)
    return n.take(budget) + "…" + ext
}
