package com.nearby.chat.core.model

data class UploadedChatMedia(
    val mediaId: Long,
    val url: String,
    val contentType: Int,
    val fileName: String = "",
)
