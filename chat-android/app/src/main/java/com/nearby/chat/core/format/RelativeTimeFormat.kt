package com.nearby.chat.core.format

import android.content.Context
import com.nearby.chat.R
import java.text.SimpleDateFormat
import java.util.Calendar
import java.util.Date
import java.util.Locale

private const val MINUTE_MS = 60_000L
private const val HOUR_MS = 3_600_000L

fun formatConversationTime(context: Context, seconds: Long): String {
    if (seconds <= 0) return ""
    val ts = seconds * 1000
    val diff = System.currentTimeMillis() - ts
    if (diff < MINUTE_MS) return context.getString(R.string.time_just_now)
    if (diff < HOUR_MS) return context.getString(R.string.time_minutes_ago, (diff / MINUTE_MS).toInt())
    val date = Date(ts)
    val cal = Calendar.getInstance().apply { time = date }
    val now = Calendar.getInstance()
    val timeFmt = SimpleDateFormat(context.getString(R.string.time_pattern_hm), Locale.getDefault())
    val sameDay = cal.get(Calendar.YEAR) == now.get(Calendar.YEAR) &&
        cal.get(Calendar.DAY_OF_YEAR) == now.get(Calendar.DAY_OF_YEAR)
    if (sameDay) return timeFmt.format(date)
    val yesterday = Calendar.getInstance().apply { add(Calendar.DAY_OF_YEAR, -1) }
    val isYesterday = cal.get(Calendar.YEAR) == yesterday.get(Calendar.YEAR) &&
        cal.get(Calendar.DAY_OF_YEAR) == yesterday.get(Calendar.DAY_OF_YEAR)
    if (isYesterday) return context.getString(R.string.time_yesterday, timeFmt.format(date))
    val pattern = if (cal.get(Calendar.YEAR) == now.get(Calendar.YEAR)) {
        R.string.time_pattern_month_day
    } else {
        R.string.time_pattern_ymd
    }
    return SimpleDateFormat(context.getString(pattern), Locale.getDefault()).format(date)
}
