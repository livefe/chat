package com.nearby.chat.feature.chat.ui

import android.content.Context
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.ui.Modifier
import com.nearby.chat.R
import com.nearby.chat.core.chat.messageListKey
import com.nearby.chat.core.model.MessageItem
import java.text.SimpleDateFormat
import java.util.Calendar
import java.util.Date
import java.util.Locale

private const val MESSAGE_TIME_DIVIDER_GAP_MS = 5 * 60 * 1000L

private fun shouldShowTimeDivider(prevTs: Long?, currTs: Long, isFirst: Boolean): Boolean {
    val curr = normalizeTs(currTs)
    if (curr <= 0) return false
    if (isFirst) return true
    val prev = normalizeTs(prevTs ?: 0)
    if (prev <= 0) return false
    if (curr - prev >= MESSAGE_TIME_DIVIDER_GAP_MS) return true
    return !isSameDay(prev, curr)
}

fun LazyListScope.chatMessageItems(
    messages: List<MessageItem>,
    selfId: Long,
    hideName: Boolean = false,
    newestFirst: Boolean = false,
    onAvatarClick: ((Long) -> Unit)? = null,
    onReply: ((MessageItem) -> Unit)? = null,
    onForward: ((MessageItem) -> Unit)? = null,
    onReport: ((MessageItem) -> Unit)? = null,
    onRecall: ((MessageItem) -> Unit)? = null,
    onPostClick: ((Long) -> Unit)? = null,
) {
    itemsIndexed(messages, key = { _, m -> messageListKey(m) }) { index, msg ->
        val older = if (newestFirst) {
            messages.getOrNull(index + 1)
        } else {
            messages.getOrNull(index - 1)
        }
        val isOldest = if (newestFirst) index == messages.lastIndex else index == 0
        val showDivider = shouldShowTimeDivider(older?.createdAt, msg.createdAt, isOldest)

        Column(modifier = Modifier.fillMaxWidth()) {
            if (showDivider) {
                ChatTimeDivider(msg.createdAt)
            }
            MessageBubble(
                message = msg,
                isMine = msg.fromUserId == selfId,
                hideName = hideName,
                onAvatarClick = onAvatarClick,
                onReply = onReply,
                onForward = onForward,
                onReport = onReport,
                onRecall = onRecall,
                onPostClick = onPostClick,
            )
        }
    }
}

internal fun formatChatBubbleTime(context: Context, raw: Long): String {
    val ms = normalizeTs(raw)
    if (ms <= 0) return ""
    val d = Calendar.getInstance().apply { timeInMillis = ms }
    val now = Calendar.getInstance()
    val hm = SimpleDateFormat("HH:mm", Locale.getDefault()).format(Date(ms))
    if (isSameCalendarDay(d, now)) return hm
    val yesterday = Calendar.getInstance().apply { add(Calendar.DAY_OF_YEAR, -1) }
    if (isSameCalendarDay(d, yesterday)) return context.getString(R.string.time_yesterday, hm)
    val pattern = if (d.get(Calendar.YEAR) == now.get(Calendar.YEAR)) {
        context.getString(R.string.time_pattern_month_day_hm)
    } else {
        context.getString(R.string.time_pattern_ymd_hm)
    }
    return SimpleDateFormat(pattern, Locale.getDefault()).format(Date(ms))
}

private fun normalizeTs(ts: Long): Long {
    if (ts <= 0) return 0
    return ts * 1000
}

private fun isSameDay(aMs: Long, bMs: Long): Boolean {
    val a = Calendar.getInstance().apply { timeInMillis = aMs }
    val b = Calendar.getInstance().apply { timeInMillis = bMs }
    return isSameCalendarDay(a, b)
}

private fun isSameCalendarDay(a: Calendar, b: Calendar): Boolean = a.get(Calendar.YEAR) == b.get(Calendar.YEAR) &&
    a.get(Calendar.DAY_OF_YEAR) == b.get(Calendar.DAY_OF_YEAR)
