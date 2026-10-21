package com.nearby.chat.feature.post

data class PostImageItem(
    val id: String,
    val localUri: String,
    val url: String = "",
    val mediaId: Long = 0,
    val uploading: Boolean = true,
    val failed: Boolean = false,
)
