package com.nearby.chat.core.format

import android.content.Context
import androidx.annotation.StringRes
import com.nearby.chat.R
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

fun formatUnixTime(context: Context, seconds: Long, @StringRes pattern: Int = R.string.time_pattern_full): String {
    if (seconds <= 0) return ""
    return SimpleDateFormat(context.getString(pattern), Locale.getDefault()).format(Date(seconds * 1000))
}
