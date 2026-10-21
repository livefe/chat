package com.nearby.chat.core.ui

data class ImageUploadItem(
    val id: String,
    val url: String,
    val localUri: String,
    val uploading: Boolean = false,
    val failed: Boolean = false,
)
