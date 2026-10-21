package com.nearby.chat.core.model

data class SendFailedEvent(
    val clientMsgId: String,
    val reason: String = "",
)
