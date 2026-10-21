package com.nearby.chat.data.remote.dto

import kotlinx.serialization.Serializable

@Serializable
data class UploadMediaDto(
    val id: Long = 0,
    val url: String = "",
    val key: String = "",
    val kind: String = "",
    val mimeType: String = "",
    val sizeBytes: Long = 0,
)
