package com.nearby.chat.feature.post

import androidx.annotation.StringRes
import com.nearby.chat.R

data class PostCheckInScanUiState(
    val busy: Boolean = false,
    @StringRes val tip: Int = R.string.scan_tip,
)
