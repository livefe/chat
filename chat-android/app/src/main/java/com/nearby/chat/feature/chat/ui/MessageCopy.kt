package com.nearby.chat.feature.chat.ui

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import com.nearby.chat.R
import com.nearby.chat.core.chat.MessageContentTypes
import com.nearby.chat.core.chat.chatMediaDisplayUrl
import com.nearby.chat.core.chat.messageClipboardText
import com.nearby.chat.core.chat.messagePreview
import com.nearby.chat.core.model.MessageItem
import com.nearby.chat.core.ui.Toaster

internal fun copyMessage(context: Context, toaster: Toaster, message: MessageItem) {
    val text = when (message.contentType) {
        MessageContentTypes.IMAGE, MessageContentTypes.VIDEO ->
            chatMediaDisplayUrl(message.content).ifBlank {
                messagePreview(message.contentType, message.content).resolve(context)
            }

        else -> messageClipboardText(message.contentType, message.content).resolve(context)
    }
    if (text.isBlank()) {
        toaster.show(R.string.copy_nothing)
        return
    }
    context.getSystemService(ClipboardManager::class.java)
        ?.setPrimaryClip(ClipData.newPlainText("message", text))
    toaster.show(R.string.common_copied)
}
