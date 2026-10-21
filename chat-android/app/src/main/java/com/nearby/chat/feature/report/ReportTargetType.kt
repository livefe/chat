package com.nearby.chat.feature.report

import androidx.annotation.StringRes
import com.nearby.chat.R

object ReportTargetType {
    const val USER = 1
    const val MESSAGE = 2
    const val CONTENT = 3

    @StringRes
    fun label(type: Int): Int = when (type) {
        USER -> R.string.report_target_user
        MESSAGE -> R.string.report_target_message
        else -> R.string.report_target_content
    }
}
