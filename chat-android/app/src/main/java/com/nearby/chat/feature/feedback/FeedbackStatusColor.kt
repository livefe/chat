package com.nearby.chat.feature.feedback

import androidx.compose.ui.graphics.Color
import com.nearby.chat.core.theme.AppColors

private const val STATUS_HANDLED = 1
private const val STATUS_CLOSED = 2

internal fun feedbackStatusColor(status: Int): Color = when (status) {
    STATUS_HANDLED -> AppColors.Success
    STATUS_CLOSED -> AppColors.TextTertiary
    else -> AppColors.Danger
}
