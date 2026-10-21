package com.nearby.chat.feature.post

import android.graphics.Bitmap

data class PostCheckInCodeUiState(
    val loading: Boolean = true,
    val postTitle: String = "",
    val checkedIn: Boolean = false,
    val checkedInAtText: String = "",
    val userNickname: String = "",
    val signupAt: String = "",
    val qrBitmap: Bitmap? = null,
)
