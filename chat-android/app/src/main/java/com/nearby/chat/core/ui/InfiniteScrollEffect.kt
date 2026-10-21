package com.nearby.chat.core.ui

import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.staggeredgrid.LazyStaggeredGridState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.snapshotFlow
import kotlinx.coroutines.flow.distinctUntilChanged

@Composable
fun InfiniteScrollEffect(
    listState: LazyListState,
    enabled: Boolean = true,
    buffer: Int = 3,
    nearStart: Boolean = false,
    onNearEnd: () -> Unit,
) {
    LaunchedEffect(listState, enabled, buffer, nearStart) {
        if (!enabled) return@LaunchedEffect
        snapshotFlow {
            val info = listState.layoutInfo
            val total = info.totalItemsCount
            if (total <= 0) return@snapshotFlow -1
            val near = if (nearStart) {
                (info.visibleItemsInfo.firstOrNull()?.index ?: Int.MAX_VALUE) <= buffer
            } else {
                (info.visibleItemsInfo.lastOrNull()?.index ?: 0) >= total - buffer
            }

            if (near) total else -1
        }.distinctUntilChanged().collect { nearTotal ->
            if (nearTotal >= 0) onNearEnd()
        }
    }
}

@Composable
fun InfiniteScrollEffect(
    gridState: LazyStaggeredGridState,
    enabled: Boolean = true,
    buffer: Int = 3,
    onNearEnd: () -> Unit,
) {
    LaunchedEffect(gridState, enabled, buffer) {
        if (!enabled) return@LaunchedEffect
        snapshotFlow {
            val info = gridState.layoutInfo
            val total = info.totalItemsCount
            if (total <= 0) return@snapshotFlow -1
            val last = info.visibleItemsInfo.lastOrNull()?.index ?: 0
            val near = last >= total - buffer
            if (near) total else -1
        }.distinctUntilChanged().collect { nearTotal ->
            if (nearTotal >= 0) onNearEnd()
        }
    }
}
