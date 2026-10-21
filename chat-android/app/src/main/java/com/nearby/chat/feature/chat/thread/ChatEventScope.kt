package com.nearby.chat.feature.chat.thread

import com.nearby.chat.core.model.ChatEvent

val ChatEvent.messageScope: Int?
    get() = when (this) {
        is ChatEvent.Message -> scope
        is ChatEvent.Recall -> scope
        else -> null
    }
