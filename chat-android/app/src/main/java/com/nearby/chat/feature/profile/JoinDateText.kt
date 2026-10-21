package com.nearby.chat.feature.profile

import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import com.nearby.chat.R

private const val DATE_PART_LENGTH = 10

@Composable
internal fun joinDateText(raw: String): String {
    if (raw.isBlank()) return ""
    val datePart = raw.take(DATE_PART_LENGTH)
    val parts = datePart.split("-")
    if (parts.size < 3) return datePart
    return stringResource(
        R.string.profile_join_date,
        parts[0],
        parts[1].padStart(2, '0'),
        parts[2].padStart(2, '0'),
    )
}
