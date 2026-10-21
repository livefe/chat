package com.nearby.chat.feature.update

import com.nearby.chat.core.model.AppUpdateInfo

data class AppUpdateUiState(
    val checking: Boolean = false,
    val prompt: AppUpdateInfo? = null,
)
