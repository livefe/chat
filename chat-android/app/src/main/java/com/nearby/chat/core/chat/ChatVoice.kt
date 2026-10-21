package com.nearby.chat.core.chat

import com.nearby.chat.R
import com.nearby.chat.core.ui.UiText
import org.json.JSONObject
import kotlin.math.round

const val VOICE_MAX_DURATION_MS = 60_000L
const val VOICE_MIN_DURATION_MS = 500L

data class ChatVoiceContent(
    val mediaId: Long = 0,
    val url: String,
    val durationMs: Long,
)

fun parseChatVoice(raw: String?): ChatVoiceContent? {
    val t = raw?.trim().orEmpty()
    if (t.isEmpty()) return null
    return runCatching {
        val o = JSONObject(t)
        val mediaId = (o.opt("mediaId") as? Number)?.toLong() ?: 0L
        val url = o.optString("url").trim()
        val durationMs = (o.opt("durationMs") as? Number)?.let { round(it.toDouble()).toLong() } ?: 0L
        if (mediaId < 1L || durationMs < 1L) return null
        ChatVoiceContent(mediaId = mediaId, url = url, durationMs = durationMs)
    }.getOrNull()
}

fun encodeChatVoice(mediaId: Long, durationMs: Long): String {
    val clamped = durationMs.coerceIn(1L, VOICE_MAX_DURATION_MS)
    return JSONObject()
        .put("mediaId", mediaId)
        .put("durationMs", clamped)
        .toString()
}

fun formatVoiceDuration(durationMs: Long): String {
    val sec = (durationMs / 1000.0).roundToIntAtLeastOne()
    return "${sec}\""
}

fun voiceQuoteLabel(durationMs: Long): UiText = UiText.of(R.string.msg_preview_voice_duration, formatVoiceDuration(durationMs))

private fun Double.roundToIntAtLeastOne(): Int {
    val n = round(this).toInt()
    return n.coerceAtLeast(1)
}
