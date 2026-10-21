package com.nearby.chat.core.chat

import com.nearby.chat.R
import com.nearby.chat.core.ui.UiText
import org.json.JSONObject

private const val REPLY_QUOTE_MAX_LENGTH = 80

fun messagePreview(contentType: Int, content: String): UiText = when (contentType) {
    MessageContentTypes.IMAGE -> UiText.of(R.string.msg_preview_image)

    MessageContentTypes.VOICE -> voiceLabel(content)

    MessageContentTypes.VIDEO -> UiText.of(R.string.msg_preview_video)

    MessageContentTypes.FILE -> {
        val name = runCatching { JSONObject(content).optString("name") }.getOrDefault("")
        if (name.isNotBlank()) {
            UiText.of(R.string.msg_preview_file_named, name)
        } else {
            UiText.of(R.string.msg_preview_file)
        }
    }

    MessageContentTypes.LOCATION -> UiText.of(R.string.msg_preview_location)

    MessageContentTypes.POST -> UiText.of(R.string.msg_preview_post)

    else -> UiText.of(content.trim())
}

fun replyQuoteText(contentType: Int, content: String): UiText {
    if (contentType == MessageContentTypes.IMAGE || contentType == MessageContentTypes.VIDEO) {
        return messagePreview(contentType, content)
    }
    if (contentType == MessageContentTypes.LOCATION) {
        return UiText.of(content.trim())
    }
    if (contentType == MessageContentTypes.VOICE) {
        return voiceLabel(content)
    }
    if (contentType == MessageContentTypes.POST) {
        return chatPostQuoteLabel(parseChatPost(content))
    }
    return UiText.Ellipsized(messagePreview(contentType, content), REPLY_QUOTE_MAX_LENGTH)
}

fun replyQuoteMediaUrl(contentType: Int, content: String): String {
    if (contentType != MessageContentTypes.IMAGE && contentType != MessageContentTypes.VIDEO) {
        return ""
    }
    return chatMediaDisplayUrl(content)
}

fun messageClipboardText(contentType: Int, content: String): UiText {
    if (contentType == MessageContentTypes.LOCATION) {
        return parseChatLocation(content)?.displayText() ?: UiText.of(R.string.msg_preview_location)
    }
    if (contentType == MessageContentTypes.VOICE) {
        return voiceLabel(content)
    }
    if (contentType == MessageContentTypes.POST) {
        return chatPostQuoteLabel(parseChatPost(content))
    }
    if (contentType == MessageContentTypes.IMAGE || contentType == MessageContentTypes.VIDEO || contentType == MessageContentTypes.FILE) {
        return messagePreview(contentType, content)
    }
    return UiText.of(content.trim())
}

private fun voiceLabel(content: String): UiText {
    val voice = parseChatVoice(content)
    return if (voice != null) voiceQuoteLabel(voice.durationMs) else UiText.of(R.string.msg_preview_voice)
}
