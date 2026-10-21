package com.nearby.chat.feature.chat.ui

import androidx.compose.foundation.lazy.LazyListState
import com.nearby.chat.core.model.MessageItem

private fun LazyListState.isShowingLatest(reverseLayout: Boolean): Boolean {
    val visible = layoutInfo.visibleItemsInfo
    if (visible.isEmpty()) return true
    return if (reverseLayout) {
        visible.any { it.index == 0 }
    } else {
        val last = layoutInfo.totalItemsCount - 1
        last < 0 || visible.any { it.index == last }
    }
}

internal suspend fun LazyListState.scrollToContentEnd(index: Int) {
    if (index < 0) return
    scrollToItem(index)
    val item = layoutInfo.visibleItemsInfo.lastOrNull { it.index == index } ?: return
    val delta = item.offset + item.size - layoutInfo.viewportEndOffset
    if (delta != 0) scroll { scrollBy(delta.toFloat()) }
}

suspend fun LazyListState.scrollToLatestIfFollowing(
    last: MessageItem?,
    selfId: Long,
    reverseLayout: Boolean,
    latestIndex: Int,
    animate: Boolean = false,
) {
    if (last == null || latestIndex < 0) return
    if (last.fromUserId != selfId && !isShowingLatest(reverseLayout)) return
    if (animate) animateScrollToItem(latestIndex) else scrollToItem(latestIndex)
}
